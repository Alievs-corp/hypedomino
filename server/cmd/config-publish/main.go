package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/db"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

func main() {
	flags := flag.NewFlagSet("config-publish", flag.ExitOnError)
	domain := flags.String("domain", shop.Domain, "config domain to publish")
	file := flags.String("file", "", "JSON file with the new snapshot")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: config-publish -domain shop -file shop.json")
		flags.PrintDefaults()
	}
	flags.Parse(os.Args[1:])
	url, ok := os.LookupEnv("HD_DATABASE_URL")
	if !ok || url == "" || *file == "" {
		fmt.Fprintln(os.Stderr, "HD_DATABASE_URL and -file are required")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := run(ctx, url, *domain, *file, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var errDomain = errors.New("unknown config domain")

func run(ctx context.Context, url, domain, file string, out io.Writer) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err := validate(domain, raw); err != nil {
		return err
	}
	pool, err := db.Open(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Close()
	if domain == cosmetics.Domain {
		if err := followsPublished(ctx, pool, raw); err != nil {
			return err
		}
	}
	version, err := snapshot.Publish(ctx, pool, domain, raw)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s version %d\n", domain, version)
	return nil
}

func validate(domain string, raw []byte) error {
	switch domain {
	case shop.Domain:
		_, err := shop.Parse(raw, shop.DefaultConfig())
		return err
	case rulesconfig.Domain:
		_, err := rulesconfig.Parse(raw)
		return err
	case cosmetics.Domain:
		_, err := cosmetics.ParseManifest(raw)
		return err
	}
	return fmt.Errorf("%w: %s", errDomain, domain)
}

func followsPublished(ctx context.Context, pool *pgxpool.Pool, raw []byte) error {
	_, published, err := snapshot.Latest(ctx, pool, cosmetics.Domain)
	if errors.Is(err, snapshot.ErrNone) {
		return nil
	}
	if err != nil {
		return err
	}
	previous, err := cosmetics.ParseManifest(published)
	if err != nil {
		return err
	}
	next, err := cosmetics.ParseManifest(raw)
	if err != nil {
		return err
	}
	return cosmetics.CheckSuccessor(previous, next)
}
