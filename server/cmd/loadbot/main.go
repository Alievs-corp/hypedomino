package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/db"
	"github.com/Alievs-corp/hypedomino/server/internal/loadbot"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	var err error
	switch os.Args[1] {
	case "serve":
		err = serve(ctx, os.Args[2:])
	case "run":
		err = run(ctx, os.Args[2:])
	case "local":
		err = local(ctx, os.Args[2:])
	case "record":
		err = record(ctx, os.Args[2:])
	case "replay":
		err = replay(ctx, os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: loadbot serve|run|local|record|replay [flags]")
	os.Exit(2)
}

func serverConfig() (app.Config, error) {
	return app.ConfigFromEnv(func(key string) (string, bool) {
		if key == "HD_DATABASE_URL" {
			return "", false
		}
		return os.LookupEnv(key)
	})
}

func startServer(ctx context.Context, config app.Config) (string, <-chan error, error) {
	logger := app.NewLogger(os.Stderr, config.LogLevel)
	handler := loadbot.NewHandler(rules.DefaultConfig(), bot.DefaultConfig())
	a := app.New(config, logger, handler)
	handler.Bind(a.Rooms())
	closeStores, err := stores(ctx, config, logger, a, handler)
	if err != nil {
		return "", nil, err
	}
	done := make(chan error, 1)
	go func() {
		err := a.Run(ctx)
		closeStores()
		done <- err
	}()
	select {
	case addr := <-a.Addr():
		return addr, done, nil
	case err := <-done:
		return "", done, err
	}
}

func stores(ctx context.Context, config app.Config, logger *slog.Logger, a *app.App, handler *loadbot.Handler) (func(), error) {
	database := os.Getenv("HD_DATABASE_URL")
	if database == "" && config.Redis == "" {
		return func() {}, nil
	}
	var archive loadbot.Archive
	closeStores := func() {}
	if database != "" {
		pool, err := db.Open(ctx, database)
		if err != nil {
			return nil, fmt.Errorf("HD_DATABASE_URL: %w", err)
		}
		archive, closeStores = matchstore.New(pool), pool.Close
	}
	var router loadbot.Router
	if config.Redis != "" {
		options, err := redis.ParseURL(config.Redis)
		if err != nil {
			closeStores()
			return nil, fmt.Errorf("HD_REDIS_URL: %w", err)
		}
		members, err := cluster.New(config.Cluster, redis.NewClient(options), logger)
		if err != nil {
			closeStores()
			return nil, err
		}
		a.Join(members)
		router = members
	}
	s := loadbot.NewStores(archive, router)
	handler.Store(s)
	a.Expose(s)
	return closeStores, nil
}

func serve(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := flags.String("addr", "", "listen address; HD_LISTEN_ADDR when empty")
	flags.Parse(args)
	config, err := serverConfig()
	if err != nil {
		return err
	}
	if *addr != "" {
		config.ListenAddr = *addr
	}
	_, done, err := startServer(ctx, config)
	if err != nil {
		return err
	}
	return <-done
}

func loadFlags(flags *flag.FlagSet, config *loadbot.Config) *string {
	flags.IntVar(&config.Rooms, "rooms", config.Rooms, "rooms to keep active")
	flags.IntVar(&config.TwoVersusTwoPct, "2v2", config.TwoVersusTwoPct, "share of 2v2 rooms in percent")
	flags.DurationVar(&config.Ramp, "ramp", config.Ramp, "time to open all rooms")
	flags.DurationVar(&config.Hold, "hold", config.Hold, "time to hold the load after the ramp")
	flags.DurationVar(&config.ThinkMin, "think-min", config.ThinkMin, "shortest pause before a move")
	flags.DurationVar(&config.ThinkMax, "think-max", config.ThinkMax, "longest pause before a move")
	flags.DurationVar(&config.ThinkFollow, "think-follow", config.ThinkFollow, "pause before a further action in the same turn, such as a draw")
	flags.Float64Var(&config.DropsPerMinute, "drops", config.DropsPerMinute, "connection drops per connection per minute")
	flags.DurationVar(&config.StormAt, "storm", config.StormAt, "time from the start when every connection drops at once and reconnects; 0 turns it off")
	flags.DurationVar(&config.Window, "window", config.Window, "latency window length")
	flags.DurationVar(&config.Warmup, "warmup", config.Warmup, "time discarded before the steady state")
	flags.DurationVar(&config.ReportEvery, "report", config.ReportEvery, "progress log interval")
	flags.DurationVar(&config.Budget, "budget", config.Budget, "p99 budget")
	flags.Uint64Var(&config.Seed, "seed", config.Seed, "seed for modes, think times and drops")
	flags.StringVar(&config.AppVersion, "app-version", config.AppVersion, "app version sent in the hello")
	return flags.String("out", "", "file for the JSON report")
}

func execute(ctx context.Context, config loadbot.Config, out string, level string) error {
	logger := app.NewLogger(os.Stderr, level)
	report, err := loadbot.New(config, logger).Run(ctx)
	if err != nil {
		return err
	}
	fmt.Print(report.Text())
	if out != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
	}
	if !report.Pass {
		return fmt.Errorf("load budget not met")
	}
	return nil
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	config := loadbot.DefaultConfig()
	flags.StringVar(&config.Target, "target", "ws://127.0.0.1:8080/ws", "WebSocket URL of a loadbot serve node")
	metricsURL := flags.String("metrics", "", "URL of the node's /metrics, sampled for heap and goroutine drift; empty turns it off")
	out := loadFlags(flags, &config)
	flags.Parse(args)
	if *metricsURL != "" {
		config.Node = loadbot.ScrapeNode(*metricsURL)
	}
	level, _ := os.LookupEnv("HD_LOG_LEVEL")
	return execute(ctx, config, *out, level)
}

func withTarget(ctx context.Context, target string, use func(target string) error) error {
	if target != "" {
		return use(target)
	}
	config, err := serverConfig()
	if err != nil {
		return err
	}
	config.ListenAddr = "127.0.0.1:0"
	config.DrainTimeout = time.Second
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	addr, done, err := startServer(serverCtx, config)
	if err != nil {
		return err
	}
	useErr := use("ws://" + addr + "/ws")
	stopServer()
	select {
	case <-done:
	case <-time.After(config.ShutdownTimeout + config.DrainTimeout):
	}
	return useErr
}

func roomOrNow(room uint64) uint64 {
	if room != 0 {
		return room
	}
	return uint64(time.Now().UnixMilli()) * 1000
}

func record(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("record", flag.ExitOnError)
	config := loadbot.RecordConfig{AppVersion: "1.0.0", Mode: rules.Mode2v2, Seed: 1, ThinkMin: time.Second, ThinkMax: 3 * time.Second, ThinkFollow: 200 * time.Millisecond, Length: 3 * time.Minute}
	target := flags.String("target", "", "WebSocket URL of a loadbot serve node; empty starts one in process")
	mode := flags.String("mode", config.Mode.String(), "1v1 or 2v2")
	flags.Uint64Var(&config.Seed, "seed", config.Seed, "match seed")
	flags.Uint64Var(&config.Room, "room", 0, "room id on the node; 0 picks one from the clock")
	flags.DurationVar(&config.ThinkMin, "think-min", config.ThinkMin, "shortest pause before a move")
	flags.DurationVar(&config.ThinkMax, "think-max", config.ThinkMax, "longest pause before a move")
	flags.DurationVar(&config.ThinkFollow, "think-follow", config.ThinkFollow, "pause before a further action in the same turn")
	flags.DurationVar(&config.Length, "length", config.Length, "stop recording after this time; 0 records the whole match")
	flags.StringVar(&config.AppVersion, "app-version", config.AppVersion, "app version sent in the hello")
	out := flags.String("out", "", "file for the trace (required)")
	flags.Parse(args)
	if *out == "" {
		return fmt.Errorf("record: -out is required")
	}
	parsed, err := rules.ParseMode(*mode)
	if err != nil {
		return err
	}
	config.Mode, config.Room = parsed, roomOrNow(config.Room)
	return withTarget(ctx, *target, func(target string) error {
		config.Target = target
		trace, err := loadbot.Record(ctx, config)
		if err != nil {
			return err
		}
		fmt.Printf("recorded %s seed %d: %d intents\n", trace.Mode, trace.Seed, trace.Intents())
		return trace.Write(*out)
	})
}

func replay(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("replay", flag.ExitOnError)
	config := loadbot.ReplayConfig{AppVersion: "1.0.0", Runs: 3, Parallel: 1, Budget: 100 * time.Millisecond, RunTimeout: 15 * time.Minute}
	target := flags.String("target", "", "WebSocket URL of a loadbot serve node; empty starts one in process")
	path := flags.String("trace", "", "trace written by record (required)")
	flags.IntVar(&config.Runs, "runs", config.Runs, "replays played back to back")
	flags.IntVar(&config.Parallel, "parallel", config.Parallel, "copies of each replay played at once, each in its own room")
	flags.Uint64Var(&config.Room, "room", 0, "first room id on the node; 0 picks one from the clock")
	flags.DurationVar(&config.Budget, "budget", config.Budget, "p99 budget of every run")
	flags.DurationVar(&config.RunTimeout, "run-timeout", config.RunTimeout, "longest time one replay may take")
	flags.StringVar(&config.AppVersion, "app-version", config.AppVersion, "app version sent in the hello")
	out := flags.String("out", "", "file for the JSON report")
	flags.Parse(args)
	if *path == "" {
		return fmt.Errorf("replay: -trace is required")
	}
	trace, err := loadbot.ReadTrace(*path)
	if err != nil {
		return err
	}
	config.Room = roomOrNow(config.Room)
	return withTarget(ctx, *target, func(target string) error {
		config.Target = target
		report, err := loadbot.Replay(ctx, trace, config)
		if err != nil {
			return err
		}
		fmt.Print(report.Text())
		if *out != "" {
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(*out, data, 0o644); err != nil {
				return err
			}
		}
		if !report.Pass {
			return fmt.Errorf("replay budget not met")
		}
		return nil
	})
}

func localNode() (loadbot.NodeSample, error) {
	runtime.GC()
	samples := []metrics.Sample{{Name: "/gc/heap/live:bytes"}}
	metrics.Read(samples)
	if samples[0].Value.Kind() != metrics.KindUint64 {
		return loadbot.NodeSample{}, fmt.Errorf("loadbot: no live heap metric")
	}
	return loadbot.NodeSample{HeapBytes: samples[0].Value.Uint64(), Goroutines: uint64(runtime.NumGoroutine())}, nil
}

func local(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("local", flag.ExitOnError)
	config := loadbot.DefaultConfig()
	out := loadFlags(flags, &config)
	flags.Parse(args)
	serverCfg, err := serverConfig()
	if err != nil {
		return err
	}
	serverCfg.ListenAddr = "127.0.0.1:0"
	serverCfg.DrainTimeout = time.Second
	serverCfg.Gateway.IntentsPerSecond = max(serverCfg.Gateway.IntentsPerSecond, 1000)
	serverCfg.Gateway.IntentBurst = max(serverCfg.Gateway.IntentBurst, 1000)
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	addr, done, err := startServer(serverCtx, serverCfg)
	if err != nil {
		return err
	}
	config.Target = "ws://" + addr + "/ws"
	config.Node = localNode
	runErr := execute(ctx, config, *out, serverCfg.LogLevel)
	stopServer()
	select {
	case <-done:
	case <-time.After(serverCfg.ShutdownTimeout + serverCfg.DrainTimeout):
	}
	return runErr
}
