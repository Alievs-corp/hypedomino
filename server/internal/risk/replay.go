package risk

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Replays struct {
	store   *Store
	config  Config
	clock   Clock
	logger  *slog.Logger
	mu      sync.Mutex
	recent  map[string][]time.Time
	sweepAt time.Time
}

func NewReplays(store *Store, config Config, clock Clock, logger *slog.Logger) *Replays {
	return &Replays{store: store, config: config, clock: clock, logger: logger, recent: map[string][]time.Time{}}
}

func (r *Replays) Observe(ctx context.Context, account string) {
	now := r.clock.Now().UTC()
	count, reached := r.count(account, now)
	if !reached {
		return
	}
	signal := Signal{
		Account: account, Kind: KindReplay, Weight: r.config.ReplayWeight, Dedupe: now.Truncate(r.config.ReplayWindow).Format(time.RFC3339), At: now,
		Evidence: map[string]any{"rejections": count, "window_seconds": int64(r.config.ReplayWindow.Seconds())},
	}
	if _, err := r.store.Record(ctx, signal); err != nil {
		r.logger.Error("replay signal not recorded", "account", account, "error", err)
	}
}

func (r *Replays) count(account string, now time.Time) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	since := now.Add(-r.config.ReplayWindow)
	r.sweep(now, since)
	kept := within(r.recent[account], since)
	kept = append(kept, now)
	if len(kept) < r.config.ReplayThreshold {
		r.recent[account] = kept
		return len(kept), false
	}
	delete(r.recent, account)
	return len(kept), true
}

func (r *Replays) sweep(now, since time.Time) {
	if now.Before(r.sweepAt) {
		return
	}
	r.sweepAt = now.Add(r.config.ReplayWindow)
	for account, times := range r.recent {
		if kept := within(times, since); len(kept) > 0 {
			r.recent[account] = kept
		} else {
			delete(r.recent, account)
		}
	}
}

func within(times []time.Time, since time.Time) []time.Time {
	for i, at := range times {
		if at.After(since) {
			return times[i:]
		}
	}
	return times[:0]
}

func (r *Replays) Tracked() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.recent)
}
