package risk

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/anticheat"
	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
)

type stoppedClock struct{ now time.Time }

func (c *stoppedClock) Now() time.Time { return c.now }

func TestIllegalIntentsRaiseOneSoftSignalPerWindowAtTheThreshold(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	store := NewStore(pool)
	start := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		rule    anticheat.IllegalIntents
		at      []time.Duration
		signals int
		count   int
	}{
		{"disabled when the threshold is 0", anticheat.IllegalIntents{Threshold: 0, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 0, 0},
		{"disabled without a window", anticheat.IllegalIntents{Threshold: 1, Weight: 5}, []time.Duration{0, 1, 2}, 0, 0},
		{"below the threshold", anticheat.IllegalIntents{Threshold: 3, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, time.Minute}, 0, 0},
		{"at the threshold", anticheat.IllegalIntents{Threshold: 3, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, time.Minute, 2 * time.Minute}, 1, 3},
		{"past the threshold in the same window", anticheat.IllegalIntents{Threshold: 3, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, 1, 2, 3, 4, 5, 6}, 1, 3},
		{"earlier intents expire with the window", anticheat.IllegalIntents{Threshold: 3, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, time.Minute, 11 * time.Minute, 12 * time.Minute}, 0, 0},
		{"a later window flags again", anticheat.IllegalIntents{Threshold: 2, Window: 10 * time.Minute, Weight: 5}, []time.Duration{0, time.Minute, 20 * time.Minute, 21 * time.Minute}, 2, 2},
		{"weight 0 keeps the evidence", anticheat.IllegalIntents{Threshold: 2, Window: 10 * time.Minute}, []time.Duration{0, time.Second}, 1, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var account string
			if err := pool.QueryRow(ctx, `insert into accounts (created_at) values (now()) returning id`).Scan(&account); err != nil {
				t.Fatal(err)
			}
			clock := &stoppedClock{now: start}
			illegal := NewIllegalIntents(store, func() anticheat.IllegalIntents { return test.rule }, clock, slog.New(slog.NewTextHandler(io.Discard, nil)))
			for _, at := range test.at {
				clock.now = start.Add(at)
				illegal.burst.Observe(ctx, account)
			}
			signals, err := store.Signals(ctx, account)
			if err != nil || len(signals) != test.signals {
				t.Fatalf("%d signals, want %d: %v", len(signals), test.signals, err)
			}
			if test.signals == 0 {
				return
			}
			first := signals[0]
			if first.Kind != KindIllegalIntent || first.Weight != test.rule.Weight || first.Evidence["intents"] != float64(test.count) || first.Evidence["window_seconds"] != test.rule.Window.Seconds() {
				t.Fatalf("signal %+v", first)
			}
		})
	}
}

func TestTurningTheIllegalIntentSignalOffForgetsTheCounts(t *testing.T) {
	rule := anticheat.IllegalIntents{Threshold: 3, Window: 10 * time.Minute, Weight: 5}
	illegal := NewIllegalIntents(nil, func() anticheat.IllegalIntents { return rule }, &stoppedClock{now: time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	illegal.burst.Observe(context.Background(), "ayla")
	illegal.burst.Observe(context.Background(), "bora")
	if illegal.burst.Tracked() != 2 {
		t.Fatalf("%d accounts tracked, want 2", illegal.burst.Tracked())
	}
	rule.Threshold = 0
	illegal.burst.Observe(context.Background(), "ayla")
	if illegal.burst.Tracked() != 0 {
		t.Fatalf("%d accounts still tracked after the signal was turned off", illegal.burst.Tracked())
	}
}
