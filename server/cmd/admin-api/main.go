package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/adminsession"
	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/approval"
	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/auth"
	"github.com/Alievs-corp/hypedomino/server/internal/db"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

const shutdownTimeout = 10 * time.Second

type config struct {
	listenAddr string
	logLevel   string
	database   string
	redis      string
	audiences  []string
	api        adminapi.Config
}

type lookup func(name string) (string, bool)

func configFromEnv(env lookup) (config, error) {
	c := config{listenAddr: ":8090", logLevel: "info", api: adminapi.Config{Session: adminsession.DefaultConfig(), ApprovalTTL: approval.DefaultConfig().TTL}}
	text := func(name string, value *string) {
		if raw, ok := env(name); ok {
			*value = strings.TrimSpace(raw)
		}
	}
	var errs []error
	duration := func(name string, value *time.Duration) {
		raw, ok := env(name)
		if !ok {
			return
		}
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			errs = append(errs, fmt.Errorf("%s: %q is not a positive duration", name, raw))
			return
		}
		*value = parsed
	}
	text("HD_ADMIN_LISTEN_ADDR", &c.listenAddr)
	text("HD_LOG_LEVEL", &c.logLevel)
	text("HD_DATABASE_URL", &c.database)
	text("HD_REDIS_URL", &c.redis)
	text("HD_ADMIN_ENVIRONMENT", &c.api.Environment)
	text("HD_ADMIN_GOOGLE_DOMAIN", &c.api.GoogleDomain)
	var audiences string
	text("HD_ADMIN_GOOGLE_AUDIENCES", &audiences)
	for _, audience := range strings.Split(audiences, ",") {
		if audience = strings.TrimSpace(audience); audience != "" {
			c.audiences = append(c.audiences, audience)
		}
	}
	duration("HD_ADMIN_SESSION_IDLE", &c.api.Session.Idle)
	duration("HD_ADMIN_SESSION_MAX", &c.api.Session.Max)
	duration("HD_ADMIN_STEP_UP", &c.api.Session.StepUp)
	duration("HD_ADMIN_APPROVAL_TTL", &c.api.ApprovalTTL)
	if c.database == "" {
		errs = append(errs, errors.New("HD_DATABASE_URL is required"))
	}
	if c.api.Environment != audit.Staging && c.api.Environment != audit.Production {
		errs = append(errs, fmt.Errorf("HD_ADMIN_ENVIRONMENT: %q is not staging or production", c.api.Environment))
	}
	if c.api.GoogleDomain == "" {
		errs = append(errs, errors.New("HD_ADMIN_GOOGLE_DOMAIN is required"))
	}
	if len(c.audiences) == 0 {
		errs = append(errs, errors.New("HD_ADMIN_GOOGLE_AUDIENCES is required"))
	}
	return c, errors.Join(errs...)
}

func main() {
	flags := flag.NewFlagSet("admin-api", flag.ExitOnError)
	email := flags.String("bootstrap-email", "", "create the first super_admin with this company email and exit")
	name := flags.String("bootstrap-name", "", "display name of the first super_admin")
	flags.Parse(os.Args[1:])
	c, err := configFromEnv(os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	logger := app.NewLogger(os.Stdout, c.logLevel)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if *email != "" {
		if err := bootstrap(ctx, c, *email, *name, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(ctx, c, logger); err != nil {
		logger.Error("admin api failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c config, logger *slog.Logger) error {
	pool, err := db.Open(ctx, c.database)
	if err != nil {
		return fmt.Errorf("HD_DATABASE_URL: %w", err)
	}
	defer pool.Close()
	nonces, err := nonceStore(c.redis)
	if err != nil {
		return err
	}
	tokens := idtoken.DefaultConfig()
	tokens.Google.Audiences = c.audiences
	verifier := idtoken.New(idtoken.Options{Config: tokens, Nonces: nonces})
	server := &http.Server{Addr: c.listenAddr, Handler: adminapi.New(c.api, verifier, pool, time.Now, logger), ReadHeaderTimeout: 10 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()
	logger.Info("admin api listening", "addr", c.listenAddr, "environment", c.api.Environment)
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func bootstrap(ctx context.Context, c config, email, name string, out io.Writer) error {
	pool, err := db.Open(ctx, c.database)
	if err != nil {
		return fmt.Errorf("HD_DATABASE_URL: %w", err)
	}
	defer pool.Close()
	account, err := adminapi.Bootstrap(ctx, pool, c.api, email, name, time.Now())
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "super_admin %s created for %s\n", account.ID, account.Email)
	return nil
}

func nonceStore(url string) (idtoken.NonceStore, error) {
	if url == "" {
		return idtoken.NewMemoryNonces(), nil
	}
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("HD_REDIS_URL: %w", err)
	}
	return auth.NewRedisNonces(redis.NewClient(options), "hd:admin:"), nil
}
