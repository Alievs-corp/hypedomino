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
	for _, config := range []risk.Config{{ClockSkewTolerance: 0, ClockSkewWeight: 1}, {ClockSkewTolerance: time.Minute, ClockSkewWeight: -1}, {ClockSkewTolerance: time.Minute, FakeReceiptWeight: -1}, {ClockSkewTolerance: time.Minute, RefundWeight: -1}} {
		if err := config.Validate(); !errors.Is(err, risk.ErrInvalidConfig) {
			t.Fatalf("%+v: %v", config, err)
		}
	}
	if config := risk.DefaultConfig(); config.Validate() != nil || config.FakeReceiptWeight != 20 || config.RefundWeight != 3 {
		t.Fatalf("defaults %+v", config)
	}
}
