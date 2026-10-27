package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/secrand"
	"github.com/Alievs-corp/hypedomino/server/internal/session"
)

type systemClock struct{}

func (systemClock) Now() time.Time                         { return time.Now() }
func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

func main() {
	config, err := app.ConfigFromEnv(os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	logger := app.NewLogger(os.Stdout, config.LogLevel)
	sessions := session.New(config.Session, systemClock{}, secrand.Source{})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := app.New(config, logger, sessions).Run(ctx); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
