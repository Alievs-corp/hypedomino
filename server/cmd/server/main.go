package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
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
	a, err := build(config, logger, sessions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := a.Run(ctx); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func build(config app.Config, logger *slog.Logger, sessions *session.Manager) (*app.App, error) {
	if config.Redis == "" {
		return app.New(config, logger, sessions), nil
	}
	options, err := redis.ParseURL(config.Redis)
	if err != nil {
		return nil, fmt.Errorf("HD_REDIS_URL: %w", err)
	}
	members, err := cluster.New(config.Cluster, redis.NewClient(options), logger)
	if err != nil {
		return nil, err
	}
	sessions.Route(members)
	var handler gateway.Handler = cluster.NewRelay(members, sessions, config.Relay)
	a := app.New(config, logger, handler)
	a.Join(members)
	return a, nil
}
