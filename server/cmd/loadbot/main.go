package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/app"
	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/loadbot"
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
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: loadbot serve|run|local [flags]")
	os.Exit(2)
}

func serverConfig() (app.Config, error) {
	config, err := app.ConfigFromEnv(os.LookupEnv)
	if err != nil {
		return config, err
	}
	return config, nil
}

func startServer(ctx context.Context, config app.Config) (string, <-chan error, error) {
	handler := loadbot.NewHandler(rules.DefaultConfig(), bot.DefaultConfig())
	a := app.New(config, app.NewLogger(os.Stderr, config.LogLevel), handler)
	handler.Bind(a.Rooms())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	select {
	case addr := <-a.Addr():
		return addr, done, nil
	case err := <-done:
		return "", done, err
	}
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
	out := loadFlags(flags, &config)
	flags.Parse(args)
	level, _ := os.LookupEnv("HD_LOG_LEVEL")
	return execute(ctx, config, *out, level)
}

func heapBytes() uint64 {
	runtime.GC()
	samples := []metrics.Sample{{Name: "/gc/heap/live:bytes"}}
	metrics.Read(samples)
	if samples[0].Value.Kind() != metrics.KindUint64 {
		return 0
	}
	return samples[0].Value.Uint64()
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
	config.Heap = heapBytes
	runErr := execute(ctx, config, *out, serverCfg.LogLevel)
	stopServer()
	select {
	case <-done:
	case <-time.After(serverCfg.ShutdownTimeout + serverCfg.DrainTimeout):
	}
	return runErr
}
