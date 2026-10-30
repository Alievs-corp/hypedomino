package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alievs-corp/hypedomino/server/internal/db"
)

func main() {
	flags := flag.NewFlagSet("migrate", flag.ExitOnError)
	to := flags.Int64("to", 0, "target version for down")
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: migrate up|down|status|version [-to N]")
		flags.PrintDefaults()
	}
	if len(os.Args) < 2 {
		flags.Usage()
		os.Exit(2)
	}
	flags.Parse(os.Args[2:])
	url, ok := os.LookupEnv("HD_DATABASE_URL")
	if !ok || url == "" {
		fmt.Fprintln(os.Stderr, "HD_DATABASE_URL is not set")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := run(ctx, os.Args[1], url, *to); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, command, url string, to int64) error {
	migrator, err := db.NewMigrator(url)
	if err != nil {
		return err
	}
	defer migrator.Close()
	switch command {
	case "up":
		version, err := migrator.Up(ctx)
		if err != nil {
			return err
		}
		fmt.Println("version", version)
	case "down":
		version, err := migrator.DownTo(ctx, to)
		if err != nil {
			return err
		}
		fmt.Println("version", version)
	case "version":
		version, err := migrator.Version(ctx)
		if err != nil {
			return err
		}
		fmt.Println("version", version)
	case "status":
		statuses, err := migrator.Status(ctx)
		if err != nil {
			return err
		}
		for _, s := range statuses {
			state := "pending"
			if s.Applied {
				state = "applied"
			}
			fmt.Printf("%-8s %5d %s\n", state, s.Version, s.Name)
		}
	default:
		return fmt.Errorf("unknown command %q", command)
	}
	return nil
}
