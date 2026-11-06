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
	"strconv"
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
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/db"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
	"github.com/Alievs-corp/hypedomino/server/internal/objectstore"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/ratelimit"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
)

const shutdownTimeout = 10 * time.Second

type config struct {
	listenAddr string
	logLevel   string
	database   string
	redis      string
	audiences  []string
	readLimit  time.Duration
	idleLimit  time.Duration
	api        adminapi.Config
}

const defaultAssetURLTTL = 15 * time.Minute

type lookup func(name string) (string, bool)

func configFromEnv(env lookup) (config, error) {
	c := config{listenAddr: ":8090", logLevel: "info", readLimit: ratelimit.DefaultConfig().ReadTimeout, idleLimit: ratelimit.DefaultConfig().IdleTimeout, api: adminapi.Config{Session: adminsession.DefaultConfig(), ApprovalTTL: approval.DefaultConfig().TTL, Shop: shop.DefaultConfig(), GrantApprovalCoins: adminapi.DefaultGrantApprovalCoins, Limits: adminapi.DefaultLimits()}}
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
	integer := func(name string, value *int) {
		raw, ok := env(name)
		if !ok {
			return
		}
		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %q is not an integer", name, raw))
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
	duration("HD_READ_TIMEOUT", &c.readLimit)
	duration("HD_IDLE_TIMEOUT", &c.idleLimit)
	duration("HD_ADMIN_SESSION_IDLE", &c.api.Session.Idle)
	duration("HD_ADMIN_SESSION_MAX", &c.api.Session.Max)
	duration("HD_ADMIN_STEP_UP", &c.api.Session.StepUp)
	duration("HD_ADMIN_APPROVAL_TTL", &c.api.ApprovalTTL)
	grantApprovalCoins := int(c.api.GrantApprovalCoins)
	integer("HD_ADMIN_GRANT_APPROVAL_COINS", &grantApprovalCoins)
	if grantApprovalCoins < 1 {
		errs = append(errs, fmt.Errorf("HD_ADMIN_GRANT_APPROVAL_COINS: %d is not a positive coin amount", grantApprovalCoins))
	}
	c.api.GrantApprovalCoins = int64(grantApprovalCoins)
	var tempBanHours int
	integer("HD_TEMP_BAN_HOURS", &tempBanHours)
	if tempBanHours < 0 {
		errs = append(errs, fmt.Errorf("HD_TEMP_BAN_HOURS: %d is negative", tempBanHours))
	}
	c.api.TempBan = time.Duration(tempBanHours) * time.Hour
	integer("HD_SHOP_DAILY_SLOTS", &c.api.Shop.DailySlots)
	integer("HD_SHOP_WEEKLY_SLOTS", &c.api.Shop.WeeklySlots)
	integer("HD_SHOP_RESET_HOUR_UTC", &c.api.Shop.ResetHourUTC)
	if err := c.api.Shop.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("HD_SHOP_DAILY_SLOTS, HD_SHOP_WEEKLY_SLOTS and HD_SHOP_RESET_HOUR_UTC: %w", err))
	}
	errs = append(errs, promoFromEnv(&c.api, text, integer)...)
	errs = append(errs, assetStoreFromEnv(&c.api, text, duration)...)
	integer("HD_ADMIN_RATE_ACCOUNT_PER_MINUTE", &c.api.Limits.Account.PerMinute)
	integer("HD_ADMIN_RATE_ACCOUNT_BURST", &c.api.Limits.Account.Burst)
	integer("HD_ADMIN_RATE_SIGNIN_PER_MINUTE", &c.api.Limits.SignIn.PerMinute)
	integer("HD_ADMIN_RATE_SIGNIN_BURST", &c.api.Limits.SignIn.Burst)
	if err := c.api.Limits.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("HD_ADMIN_RATE_*: %w", err))
	}
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

func assetStoreFromEnv(api *adminapi.Config, text func(string, *string), duration func(string, *time.Duration)) []error {
	config := objectstore.Config{URLTTL: defaultAssetURLTTL}
	text("HD_ASSET_STORE_URL", &config.BaseURL)
	if config.BaseURL == "" {
		return nil
	}
	text("HD_ASSET_STORE_REGION", &config.Region)
	text("HD_ASSET_STORE_ACCESS_KEY", &config.AccessKey)
	text("HD_ASSET_STORE_SECRET_KEY", &config.SecretKey)
	duration("HD_ASSET_STORE_URL_TTL", &config.URLTTL)
	store, err := objectstore.New(config)
	if err != nil {
		return []error{fmt.Errorf("HD_ASSET_STORE_*: %w", err)}
	}
	api.Assets = store
	return nil
}

func promoFromEnv(api *adminapi.Config, text func(string, *string), integer func(string, *int)) []error {
	text("HD_PROMO_KEY", &api.PromoKey)
	if api.PromoKey == "" {
		return nil
	}
	api.Promo = promo.DefaultConfig()
	var errs []error
	if _, err := promo.NewHasher(api.PromoKey); err != nil {
		errs = append(errs, fmt.Errorf("HD_PROMO_KEY: %w", err))
	}
	integer("HD_PROMO_CODE_LENGTH", &api.Promo.CodeLength)
	integer("HD_PROMO_MAX_VALIDITY_DAYS", &api.Promo.MaxValidityDays)
	integer("HD_PROMO_COINS_APPROVAL_THRESHOLD", &api.Promo.CoinsApprovalThreshold)
	if err := api.Promo.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("HD_PROMO_*: %w", err))
	}
	api.Cosmetics = cosmetics.Builtin()
	var manifest string
	text("HD_COSMETICS_MANIFEST", &manifest)
	if manifest == "" {
		return errs
	}
	catalog, err := cosmetics.Load(manifest)
	if err != nil {
		return append(errs, fmt.Errorf("HD_COSMETICS_MANIFEST: %w", err))
	}
	api.Cosmetics = catalog
	return errs
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
	client, err := redisClient(c.redis)
	if err != nil {
		return err
	}
	nonces, buckets := stores(client)
	tokens := idtoken.DefaultConfig()
	tokens.Google.Audiences = c.audiences
	verifier := idtoken.New(idtoken.Options{Config: tokens, Nonces: nonces})
	server := &http.Server{Addr: c.listenAddr, Handler: adminapi.New(c.api, verifier, buckets, pool, time.Now, logger), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: c.readLimit, IdleTimeout: c.idleLimit}
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

func redisClient(url string) (*redis.Client, error) {
	if url == "" {
		return nil, nil
	}
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("HD_REDIS_URL: %w", err)
	}
	return redis.NewClient(options), nil
}

func stores(client *redis.Client) (idtoken.NonceStore, ratelimit.Buckets) {
	if client == nil {
		return idtoken.NewMemoryNonces(), ratelimit.NewMemory()
	}
	return auth.NewRedisNonces(client, "hd:admin:"), ratelimit.NewRedis(client, "hd:admin:")
}
