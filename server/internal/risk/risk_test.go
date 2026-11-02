package risk_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/risk"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

func TestASkewedDeviceClockRaisesOneSoftSignalPerDay(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	var account string
	if err := pool.QueryRow(ctx, `insert into accounts (created_at) values (now()) returning id`).Scan(&account); err != nil {
		t.Fatal(err)
	}
	clock := &fixedClock{now: time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)}
	store := risk.NewStore(pool)
	skew := risk.NewClockSkew(store, risk.DefaultConfig(), clock, slog.New(slog.NewTextHandler(io.Discard, nil)))
	at := func(offset time.Duration) uint64 { return uint64(clock.now.Add(offset).UnixMilli()) }

	cases := []struct {
		name   string
		client uint64
		day    time.Duration
		total  int
	}{
		{"no device time", 0, 0, 0},
		{"inside the tolerance ahead", at(5 * time.Minute), 0, 0},
		{"inside the tolerance behind", at(-4 * time.Minute), 0, 0},
		{"two days ahead", at(48 * time.Hour), 0, 1},
		{"again the same day", at(72 * time.Hour), 0, 1},
		{"behind on the next day", at(24*time.Hour - 10*time.Minute), 24 * time.Hour, 2},
	}
	for _, c := range cases {
		clock.now = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC).Add(c.day)
		skew.Observe(ctx, account, c.client)
		signals, err := store.Signals(ctx, account)
		if err != nil || len(signals) != c.total {
			t.Fatalf("%s: %d signals %v", c.name, len(signals), err)
		}
	}
	signals, _ := store.Signals(ctx, account)
	first, second := signals[0], signals[1]
	if first.Kind != risk.KindClockSkew || first.Weight != 1 || first.Dedupe != "2026-11-02" || first.Evidence["skew_seconds"] != float64(48*3600) || first.Evidence["server_time"] != "2026-11-02T12:00:00Z" {
		t.Fatalf("first signal %+v", first)
	}
	if second.Dedupe != "2026-11-03" || second.Evidence["skew_seconds"] != float64(-600) {
		t.Fatalf("second signal %+v", second)
	}
}

func TestTheRiskConfigIsChecked(t *testing.T) {
	for name, change := range map[string]func(*risk.Config){
		"no skew tolerance":   func(c *risk.Config) { c.ClockSkewTolerance = 0 },
		"negative skew":       func(c *risk.Config) { c.ClockSkewWeight = -1 },
		"negative receipt":    func(c *risk.Config) { c.FakeReceiptWeight = -1 },
		"negative refund":     func(c *risk.Config) { c.RefundWeight = -1 },
		"no replay threshold": func(c *risk.Config) { c.ReplayThreshold = 0 },
		"no replay window":    func(c *risk.Config) { c.ReplayWindow = 0 },
		"negative replay":     func(c *risk.Config) { c.ReplayWeight = -1 },
	} {
		config := risk.DefaultConfig()
		change(&config)
		if err := config.Validate(); !errors.Is(err, risk.ErrInvalidConfig) {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if config := risk.DefaultConfig(); config.Validate() != nil || config.FakeReceiptWeight != 20 || config.RefundWeight != 3 || config.ReplayThreshold != 5 || config.ReplayWindow != 10*time.Minute || config.ReplayWeight != 5 {
		t.Fatalf("defaults %+v", config)
	}
}

func TestRepeatedReplaysRaiseOneSoftSignalPerWindow(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	var ayla, bora string
	for _, id := range []*string{&ayla, &bora} {
		if err := pool.QueryRow(ctx, `insert into accounts (created_at) values (now()) returning id`).Scan(id); err != nil {
			t.Fatal(err)
		}
	}
	start := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	clock := &fixedClock{now: start}
	store := risk.NewStore(pool)
	replays := risk.NewReplays(store, risk.DefaultConfig(), clock, slog.New(slog.NewTextHandler(io.Discard, nil)))
	steps := []struct {
		account string
		at      time.Duration
		ayla    int
	}{
		{ayla, 0, 0}, {ayla, time.Minute, 0}, {ayla, 2 * time.Minute, 0}, {bora, 2 * time.Minute, 0}, {ayla, 3 * time.Minute, 0},
		{ayla, 13*time.Minute + time.Second, 0},
		{ayla, 14 * time.Minute, 0}, {ayla, 15 * time.Minute, 0}, {ayla, 16 * time.Minute, 0},
		{ayla, 17 * time.Minute, 1},
		{ayla, 17 * time.Minute, 1}, {ayla, 18 * time.Minute, 1}, {ayla, 18 * time.Minute, 1}, {ayla, 19 * time.Minute, 1},
		{ayla, 19 * time.Minute, 1},
		{ayla, 21 * time.Minute, 1}, {ayla, 21 * time.Minute, 1}, {ayla, 22 * time.Minute, 1}, {ayla, 22 * time.Minute, 1}, {ayla, 23 * time.Minute, 2},
	}
	for i, step := range steps {
		clock.now = start.Add(step.at)
		replays.Observe(ctx, step.account)
		signals, err := store.Signals(ctx, ayla)
		if err != nil || len(signals) != step.ayla {
			t.Fatalf("step %d: %d signals %v", i, len(signals), err)
		}
	}
	signals, _ := store.Signals(ctx, ayla)
	first := signals[0]
	if first.Kind != risk.KindReplay || first.Weight != 5 || first.Evidence["rejections"] != float64(5) || first.Evidence["window_seconds"] != float64(600) || first.Dedupe != "2026-11-02T12:10:00Z" || signals[1].Dedupe != "2026-11-02T12:20:00Z" {
		t.Fatalf("signals %+v", signals)
	}
	if others, _ := store.Signals(ctx, bora); len(others) != 0 {
		t.Fatalf("one rejection flagged bora: %+v", others)
	}
	clock.now = start.Add(time.Hour)
	replays.Observe(ctx, ayla)
	if replays.Tracked() != 1 {
		t.Fatalf("%d accounts still tracked after the window", replays.Tracked())
	}
}
