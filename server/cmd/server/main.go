package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/attestkey"
	"github.com/Alievs-corp/hypedomino/server/internal/auth"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/dailylogin"
	"github.com/Alievs-corp/hypedomino/server/internal/db"
	"github.com/Alievs-corp/hypedomino/server/internal/economy"
	"github.com/Alievs-corp/hypedomino/server/internal/friends"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
	"github.com/Alievs-corp/hypedomino/server/internal/integrity"
	"github.com/Alievs-corp/hypedomino/server/internal/inventory"
	"github.com/Alievs-corp/hypedomino/server/internal/leaderboard"
	"github.com/Alievs-corp/hypedomino/server/internal/livecontrol"
	"github.com/Alievs-corp/hypedomino/server/internal/matchreward"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/medal"
	"github.com/Alievs-corp/hypedomino/server/internal/mission"
	"github.com/Alievs-corp/hypedomino/server/internal/panics"
	"github.com/Alievs-corp/hypedomino/server/internal/pass"
	"github.com/Alievs-corp/hypedomino/server/internal/penalties"
	"github.com/Alievs-corp/hypedomino/server/internal/profile"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/purchase"
	"github.com/Alievs-corp/hypedomino/server/internal/push"
	"github.com/Alievs-corp/hypedomino/server/internal/quickmessage"
	"github.com/Alievs-corp/hypedomino/server/internal/ratelimit"
	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/report"
	"github.com/Alievs-corp/hypedomino/server/internal/reward"
	"github.com/Alievs-corp/hypedomino/server/internal/risk"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/secrand"
	"github.com/Alievs-corp/hypedomino/server/internal/session"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
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
	slog.SetDefault(logger)
	sessions := session.New(config.Session, systemClock{}, secrand.Source{})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	a, release, err := build(ctx, config, logger, sessions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer release()
	if err := a.Run(ctx); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func build(ctx context.Context, config app.Config, logger *slog.Logger, sessions *session.Manager) (*app.App, func(), error) {
	var client *redis.Client
	if config.Redis != "" {
		options, err := redis.ParseURL(config.Redis)
		if err != nil {
			return nil, nil, fmt.Errorf("HD_REDIS_URL: %w", err)
		}
		client = redis.NewClient(options)
	}
	a, members, err := node(config, logger, sessions, client)
	if err != nil {
		return nil, nil, err
	}
	limiter := ratelimit.New(config.RateLimit, buckets(client, config.Cluster.KeyPrefix), logger, time.Now)
	a.Limit(limiter)
	if config.Database == "" {
		return a, func() {}, nil
	}
	pool, err := db.Open(ctx, config.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("HD_DATABASE_URL: %w", err)
	}
	a.Expose(db.Gauges(pool))
	var nonces idtoken.NonceStore
	if client != nil {
		nonces = auth.NewRedisNonces(client, config.Cluster.KeyPrefix)
	}
	verifier := idtoken.New(idtoken.Options{Config: config.IDToken, Nonces: nonces})
	tokens, err := token.NewService(config.Token, token.NewStore(pool), time.Now)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("HD_TOKEN_KEYS: %w", err)
	}
	sessions.Authenticate(tokens)
	signals := risk.NewStore(pool)
	sessions.ObserveClocks(risk.NewClockSkew(signals, config.Risk, systemClock{}, logger))
	sessions.ObserveReplays(risk.NewReplays(signals, config.Risk, systemClock{}, logger))
	sessions.ObserveDevices(integrity.NewDeviceReports(config.IntegrityPolicy, risk.NewIntegrity(signals, systemClock{}, logger)))
	macFailures := risk.NewMacFailures(signals, config.Risk, systemClock{}, logger)
	critical := func(next http.Handler) http.Handler { return token.RequireMac(tokens, macFailures, next) }
	limiter.Authenticate(tokens)
	accounts := account.New(pool)
	signIn := auth.New(verifier, accounts, tokens, sessions, logger)
	signIn.SetCountryHeader(config.CountryHeader)
	a.Handle("POST /auth/", signIn)
	profileStore := profile.NewStore(pool, config.Profile)
	moderation := snapshot.New(pool, profile.ModerationDomain, profile.ParseModeration, logger)
	go panics.Supervise(ctx, "config moderation", config.ConfigPoll, moderation.Run)
	profileStore.Moderate(func() profile.Moderation {
		current, _, _ := moderation.Current()
		return current
	})
	profiles := profile.NewHandler(tokens, profileStore, logger, time.Now)
	profiles.SetCountries(accounts, config.CountryCooldown)
	a.Handle("/profile/", profiles)
	catalog, err := loadCatalog(config.Cosmetics)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	inventoryStore := inventory.NewStore(pool, catalog, config.Inventory)
	manifests := snapshot.New(pool, cosmetics.Domain, cosmetics.Parse, logger)
	go panics.Supervise(ctx, "config cosmetics", config.ConfigPoll, manifests.Run)
	inventoryStore.Follow(manifests.Current)
	unowned := risk.NewUnownedEquips(signals, config.Risk, systemClock{}, logger)
	inventoryHandler := inventory.NewHandler(tokens, inventoryStore, logger, time.Now)
	inventoryHandler.FlagUnowned(unowned)
	items := critical(inventoryHandler)
	a.Handle("/inventory", items)
	a.Handle("/inventory/", items)
	friendStore := friends.NewStore(pool, config.Friends)
	if client != nil {
		presence := friends.NewPresence(client, config.Cluster.KeyPrefix, 3*config.Gateway.PingInterval)
		go panics.Supervise(ctx, "friend presence", config.Gateway.PingInterval, func(ctx context.Context, interval time.Duration) {
			presence.Run(ctx, interval, sessions.Presence, logger)
		})
		friendStore.ShowPresence(presence)
	}
	friendsHandler := friends.NewHandler(tokens, friendStore, logger, time.Now)
	a.Handle("/friends", friendsHandler)
	a.Handle("/friends/", friendsHandler)
	a.Handle("/push/", push.NewHandler(tokens, push.NewStore(pool), logger, time.Now))
	a.Handle("/reports", report.NewHandler(tokens, report.NewStore(pool, config.Reports), logger, time.Now))
	walletStore := wallet.NewStore(pool)
	a.Handle("/wallet", wallet.NewHandler(tokens, walletStore, logger))
	passConfigs := pass.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config pass", config.ConfigPoll, passConfigs.Run)
	passes := pass.NewStore(pool, passConfigs)
	profiles.SetStanding(passes.Standing)
	medalCatalog := medal.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config medals", config.ConfigPoll, medalCatalog.Run)
	medals := medal.NewStore(pool, medalCatalog)
	profiles.SetMedals(func(ctx context.Context, account string) (any, error) { return medals.Public(ctx, account) })
	engine := reward.NewEngine(pool, walletStore, inventoryStore, medals, passes, config.Reward)
	passes.SetRewards(engine)
	passes.SetWallet(walletStore)
	missionCatalog := mission.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config missions", config.ConfigPoll, missionCatalog.Run)
	missions := mission.NewStore(pool, missionCatalog, engine)
	calendar := dailylogin.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config daily login", config.ConfigPoll, calendar.Run)
	matchRewardConfigs := matchreward.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config match rewards", config.ConfigPoll, matchRewardConfigs.Run)
	matchRewards := matchstore.NewRelay(matchstore.New(pool), matchreward.NewStore(pool, matchRewardConfigs, walletStore), matchstore.TopicEconomy, 0, time.Now)
	go panics.Supervise(ctx, "outbox economy", config.OutboxPoll, func(ctx context.Context, every time.Duration) {
		matchRewards.Run(ctx, every, func(err error) { logger.Error("match reward relay failed", "error", err) })
	})
	passXp := matchstore.NewRelay(matchstore.New(pool), passes, matchstore.TopicXp, 0, time.Now)
	go panics.Supervise(ctx, "outbox xp", config.OutboxPoll, func(ctx context.Context, every time.Duration) {
		passXp.Run(ctx, every, func(err error) { logger.Error("pass xp relay failed", "error", err) })
	})
	missionProgress := matchstore.NewRelay(matchstore.New(pool), missions, matchstore.TopicMissions, 0, time.Now)
	go panics.Supervise(ctx, "outbox missions", config.OutboxPoll, func(ctx context.Context, every time.Duration) {
		missionProgress.Run(ctx, every, func(err error) { logger.Error("mission progress relay failed", "error", err) })
	})
	logins := critical(dailylogin.NewHandler(tokens, dailylogin.NewStore(pool, calendar, engine), logger, time.Now))
	a.Handle("/daily-login", logins)
	a.Handle("/daily-login/", logins)
	offers := snapshot.New(pool, shop.Domain, shop.Parser(config.Shop), logger)
	if config.ShopCatalog != "" {
		if err := offers.SeedFile(config.ShopCatalog); err != nil {
			pool.Close()
			return nil, nil, fmt.Errorf("HD_SHOP_CATALOG: %w", err)
		}
	}
	go panics.Supervise(ctx, "config shop", config.ConfigPoll, offers.Run)
	controls := snapshot.New(pool, livecontrol.Domain, livecontrol.Parse, logger)
	controls.Seed(livecontrol.Controls{MinimumAppVersion: config.Gateway.MinimumAppVersion}, 0)
	if _, err := controls.Load(ctx); err != nil && !errors.Is(err, snapshot.ErrNone) {
		pool.Close()
		return nil, nil, fmt.Errorf("live controls: %w", err)
	}
	controls.OnApplied(a.EnforceAdmission)
	go panics.Supervise(ctx, "config live controls", config.ConfigPoll, controls.Run)
	a.Admit(func() gateway.Admission {
		current, _, _ := controls.Current()
		return current.Admission()
	})
	activeStrings := snapshot.New(pool, gamestrings.Domain, gamestrings.NewStore(pool).ActiveParser(ctx), logger)
	go panics.Supervise(ctx, "config game strings", config.ConfigPoll, activeStrings.Run)
	a.Localize(func() gateway.Strings {
		active, _, _ := activeStrings.Current()
		bundles, chains := active.Hello()
		return gateway.Strings{Bundles: bundles, Chains: chains}
	})
	quickMessages := snapshot.New(pool, quickmessage.Domain, quickmessage.Parse, logger)
	go panics.Supervise(ctx, "config quick messages", config.ConfigPoll, quickMessages.Run)
	a.Rooms().AllowQuickMessages(func(key string) bool {
		catalog, _, _ := quickMessages.Current()
		return catalog.Contains(key)
	})
	tuning := snapshot.New(pool, rulesconfig.Domain, rulesconfig.Parse, logger)
	go panics.Supervise(ctx, "config rules", config.ConfigPoll, tuning.Run)
	a.Rooms().Tune(func(base rules.Config) rules.Config {
		current, _, _ := tuning.Current()
		return current.Apply(base)
	})
	ratingConfigs := rating.NewWatcher(pool, logger)
	go panics.Supervise(ctx, "config rating", config.ConfigPoll, ratingConfigs.Run)
	ratings := rating.NewStore(pool, ratingConfigs)
	profiles.SetRanked(func(ctx context.Context, account string) (any, error) { return ratings.Public(ctx, account) })
	penaltyRecords := penaltyStore(ctx, config, pool, logger)
	ratings.UseOffensePenalties(func() (int, int) {
		current, _ := penaltyRecords.Config()
		return current.OffenseRatingPenalty, current.PartnerLossReductionPercent
	})
	wave := risk.NewWave(pool, config.Risk, logger)
	wave.OnShadowban(ratings.HideFromBoards)
	go panics.Supervise(ctx, "ban wave", config.SeasonPoll, func(ctx context.Context, every time.Duration) { wave.Run(ctx, every, time.Now) })
	monitor := economy.NewMonitor(pool, config.Economy, logger)
	go panics.Supervise(ctx, "economy monitor", config.SeasonPoll, func(ctx context.Context, every time.Duration) { monitor.Run(ctx, every, time.Now) })
	ratings.SetRewards(engine)
	ratings.SetCollusionWeight(config.Risk.CollusionWeight)
	go panics.Supervise(ctx, "apex recalc", config.SeasonPoll, func(ctx context.Context, _ time.Duration) { ratings.RunApexRecalc(ctx, logger, time.Now) })
	go panics.Supervise(ctx, "rating season end", config.SeasonPoll, func(ctx context.Context, every time.Duration) {
		ratings.RunSeasonEnd(ctx, logger, every, time.Now)
	})
	signIn.EnableDeletion(auth.Deletion{Accounts: accounts, Grace: config.DeletionGrace, InMatch: sessions.InMatch, Hide: ratings.HideFromBoards, Now: time.Now})
	go panics.Supervise(ctx, "account deletions", config.SeasonPoll, func(ctx context.Context, every time.Duration) {
		accounts.RunDeletions(ctx, logger, every, config.DeletionGrace, time.Now)
	})
	ratingResults := matchstore.NewRelay(matchstore.New(pool), ratings, matchstore.TopicRating, 0, time.Now)
	go panics.Supervise(ctx, "outbox rating", config.OutboxPoll, func(ctx context.Context, every time.Duration) {
		ratingResults.Run(ctx, every, func(err error) { logger.Error("rating relay failed", "error", err) })
	})
	a.Handle("/penalties", critical(penalties.NewHandler(tokens, penaltyRecords, logger, time.Now)))
	if members != nil {
		startMatchmaking(ctx, config, logger, matchmakingDeps{pool: pool, client: client, members: members, sessions: sessions, rooms: a.Rooms(), ratings: ratings, profiles: profileStore, inventory: inventoryStore, penalties: penaltyRecords, rankedOpen: func() bool {
			current, _, _ := controls.Current()
			return !current.RankedDisabled
		}})
	}
	var reader leaderboard.Reader
	if client != nil {
		boards := leaderboard.New(client, ratings, ratings, leaderboard.DefaultConfig())
		ratings.SetBoards(boards, logger)
		reader = boards
	}
	a.Handle("GET /leaderboard/", leaderboard.NewHandler(tokens, reader, ratings, friendStore, logger, time.Now))
	a.Handle("/ranked/", rating.NewHandler(tokens, ratings, logger, time.Now))
	go panics.Supervise(ctx, "pass season end", config.SeasonPoll, func(ctx context.Context, every time.Duration) {
		passes.RunSeasonEnd(ctx, logger, every, time.Now)
	})
	passHandler := critical(pass.NewHandler(tokens, passes, logger, time.Now))
	a.Handle("/pass", passHandler)
	a.Handle("/pass/", passHandler)
	medalRequests := medal.NewHandler(tokens, medals, logger, time.Now)
	medalRequests.FlagUnowned(unowned)
	medalHandler := critical(medalRequests)
	a.Handle("/medals", medalHandler)
	a.Handle("/medals/", medalHandler)
	missionHandler := critical(mission.NewHandler(tokens, missions, logger, time.Now))
	a.Handle("/missions", missionHandler)
	a.Handle("/missions/", missionHandler)
	shopStore := shop.NewStore(pool, walletStore, inventoryStore, engine, offers, config.Shop)
	shopHandler := shop.NewHandler(tokens, shopStore, logger, time.Now)
	shopHandler.CloseCheckoutWhen(func() bool {
		current, _, _ := controls.Current()
		return current.PurchasesDisabled
	})
	store := critical(shopHandler)
	a.Handle("/shop", store)
	a.Handle("/shop/", store)
	validators, google, apple, err := stores(config)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	purchases := purchase.NewStore(pool, walletStore, engine, offers, validators, config.Purchase, logger)
	shopStore.Guard(purchases)
	shopStore.WithPass(passes)
	purchases.WithPass(passes)
	purchases.Flag(signals, config.Risk)
	var integrityStore integrity.NonceStore = integrity.NewMemoryNonceStore(systemClock{})
	if client != nil {
		integrityStore = integrity.NewRedisNonceStore(client, config.Cluster.KeyPrefix, systemClock{})
	}
	integrityNonces, err := integrity.NewNonces(config.Integrity, integrityStore, systemClock{}, rand.Reader)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("HD_INTEGRITY_NONCE_TTL: %w", err)
	}
	attestation := integrity.NewHandler(tokens, integrityNonces, logger)
	gate, err := integrityGate(config, attestation, integrityNonces, pool, risk.NewIntegrity(signals, systemClock{}, logger), logger)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	a.Handle("POST /integrity/", attestation)
	if gate.Enabled() {
		sessions.CheckIntegrity(gate)
	}
	receipts := purchase.NewHandler(tokens, purchases, google, apple, config.PushSecret, logger, time.Now)
	a.Handle("POST /shop/complete", critical(guarded(gate, tokens, integrity.PurposePurchase, receipts)))
	a.Handle("GET /purchases", receipts)
	a.Handle("POST /store/", receipts)
	promos, err := promos(pool, engine, catalog, config)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	a.Handle("POST /promo/redeem", critical(promo.NewHandler(tokens, promos, logger, time.Now)))
	return a, pool.Close, nil
}

func integrityGate(config app.Config, handler *integrity.Handler, nonces *integrity.Nonces, pool *pgxpool.Pool, signals integrity.SignalSink, logger *slog.Logger) (*integrity.Gate, error) {
	gate := integrity.NewGate(config.IntegrityPolicy, signals, logger)
	if len(config.PlayIntegrity.CertificateDigests) > 0 {
		decoder, err := integrity.NewPlayDecoder(integrity.PlayDecoderConfig{PackageName: config.Play.PackageName, ServiceAccountFile: config.Play.ServiceAccountFile})
		if err != nil {
			return nil, fmt.Errorf("HD_PLAY_SERVICE_ACCOUNT: %w", err)
		}
		play, err := integrity.NewPlayIntegrity(config.PlayIntegrity, decoder, nonces, systemClock{})
		if err != nil {
			return nil, fmt.Errorf("HD_PLAY_CERT_DIGESTS: %w", err)
		}
		gate.WithPlay(play)
	}
	if config.AppAttest.TeamID == "" {
		return gate, nil
	}
	root, err := os.ReadFile(config.AppAttestRoot)
	if err != nil {
		return nil, fmt.Errorf("HD_APP_ATTEST_ROOT_CA: %w", err)
	}
	attestConfig := config.AppAttest
	attestConfig.RootCertificatePEM = root
	verifier, err := integrity.NewAppAttest(attestConfig, nonces, systemClock{})
	if err != nil {
		return nil, fmt.Errorf("HD_APP_ATTEST_*: %w", err)
	}
	keys := attestkey.NewStore(pool)
	handler.EnableAppAttest(verifier, keys, systemClock{})
	gate.WithAppAttest(verifier, keys)
	return gate, nil
}

func guarded(gate *integrity.Gate, verifier token.Verifier, purpose integrity.Purpose, next http.Handler) http.Handler {
	if !gate.Enabled() {
		return next
	}
	return gate.Require(verifier, purpose, next)
}

func stores(config app.Config) (map[purchase.Platform]purchase.Validator, purchase.Notifier, purchase.Notifier, error) {
	validators := map[purchase.Platform]purchase.Validator{}
	var google, apple purchase.Notifier
	if config.Play.PackageName != "" {
		play, err := purchase.NewGoogle(config.Play)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("HD_PLAY_SERVICE_ACCOUNT: %w", err)
		}
		validators[purchase.PlatformGoogle], google = play, play
	}
	if config.Apple.BundleID != "" {
		ios, err := purchase.NewApple(config.Apple)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("HD_APPLE_BUNDLE_ID: %w", err)
		}
		validators[purchase.PlatformApple], apple = ios, ios
	}
	return validators, google, apple, nil
}

func promos(pool *pgxpool.Pool, engine *reward.Engine, catalog *cosmetics.Catalog, config app.Config) (promo.Promos, error) {
	if config.PromoKey == "" {
		return nil, nil
	}
	hasher, err := promo.NewHasher(config.PromoKey)
	if err != nil {
		return nil, fmt.Errorf("HD_PROMO_KEY: %w", err)
	}
	return promo.NewStore(pool, engine, catalog, hasher, config.Promo), nil
}

func loadCatalog(path string) (*cosmetics.Catalog, error) {
	if path == "" {
		return cosmetics.Builtin(), nil
	}
	return cosmetics.Load(path)
}

func buckets(client *redis.Client, prefix string) ratelimit.Buckets {
	if client == nil {
		return ratelimit.NewMemory()
	}
	return ratelimit.NewRedis(client, prefix)
}

func node(config app.Config, logger *slog.Logger, sessions *session.Manager, client *redis.Client) (*app.App, *cluster.Cluster, error) {
	if client == nil {
		return app.New(config, logger, sessions), nil, nil
	}
	members, err := cluster.New(config.Cluster, client, logger)
	if err != nil {
		return nil, nil, err
	}
	sessions.Route(members)
	var handler gateway.Handler = cluster.NewRelay(members, sessions, config.Relay)
	a := app.New(config, logger, handler)
	a.Join(members)
	return a, members, nil
}
