package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	"github.com/Alievs-corp/hypedomino/server/internal/bottoken"
)

var errNotTestEnvironment = errors.New("HD_ENVIRONMENT: bot tokens are minted for dev or staging only")

type options struct {
	bot    string
	server string
	device string
}

func main() {
	flags := flag.NewFlagSet("bot-token", flag.ExitOnError)
	var o options
	flags.StringVar(&o.bot, "bot", "", "bot name: 1 to 64 lowercase letters, digits, '-' or '_'; one name is one bot account")
	flags.StringVar(&o.server, "server", "", "base URL of a dev or staging server; when set, signs in and prints the session instead of the bot token")
	flags.StringVar(&o.device, "device", "", "device id of the sign-in; defaults to bot-token-<bot>")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bot-token -bot <name> [-server <url>] [-device <id>]")
		flags.PrintDefaults()
	}
	flags.Parse(os.Args[1:])
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := run(ctx, os.LookupEnv, o, time.Now(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(ctx context.Context, lookup app.Lookup, o options, now time.Time, out io.Writer) error {
	environment, _ := lookup("HD_ENVIRONMENT")
	if environment != app.EnvironmentDev && environment != app.EnvironmentStaging {
		return errNotTestEnvironment
	}
	encoded, _ := lookup("HD_BOT_TOKEN_SECRET")
	secret, err := bottoken.ParseSecret(encoded)
	if err != nil {
		return fmt.Errorf("HD_BOT_TOKEN_SECRET: %w", err)
	}
	minted, err := bottoken.Mint(secret, o.bot, now)
	if err != nil {
		return err
	}
	if o.server == "" {
		_, err := fmt.Fprintln(out, minted)
		return err
	}
	device := o.device
	if device == "" {
		device = "bot-token-" + o.bot
	}
	session, err := botclient.SignIn(ctx, &http.Client{Timeout: 10 * time.Second}, o.server, minted, device)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(session)
}
