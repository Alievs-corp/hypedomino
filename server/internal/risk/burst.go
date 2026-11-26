package risk

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type burstRule struct {
	kind      string
	countName string
	threshold int
	window    time.Duration
	weight    int
}

type Burst struct {
	store   *Store
	rule    func() burstRule
	clock   Clock
	logger  *slog.Logger
	mu      sync.Mutex
	recent  map[string][]time.Time
	sweepAt time.Time
}

func NewReplays(store *Store, config Config, clock Clock, logger *slog.Logger) *Burst {
	return newBurst(store, fixedRule(burstRule{kind: KindReplay, countName: "rejections", threshold: config.ReplayThreshold, window: config.ReplayWindow, weight: config.ReplayWeight}), clock, logger)
}

func NewMacFailures(store *Store, config Config, clock Clock, logger *slog.Logger) *Burst {
	return newBurst(store, fixedRule(burstRule{kind: KindMacFailure, countName: "failures", threshold: config.MacFailureThreshold, window: config.MacFailureWindow, weight: config.MacFailureWeight}), clock, logger)
}

func NewUnownedEquips(store *Store, config Config, clock Clock, logger *slog.Logger) *Burst {
	return newBurst(store, fixedRule(burstRule{kind: KindUnownedEquip, countName: "attempts", threshold: config.UnownedEquipThreshold, window: config.UnownedEquipWindow, weight: config.UnownedEquipWeight}), clock, logger)
}

func fixedRule(rule burstRule) func() burstRule {
	return func() burstRule { return rule }
}

func newBurst(store *Store, rule func() burstRule, clock Clock, logger *slog.Logger) *Burst {
	return &Burst{store: store, rule: rule, clock: clock, logger: logger, recent: map[string][]time.Time{}}
}

func (b *Burst) Observe(ctx context.Context, account string) {
	rule := b.rule()
	if rule.threshold < 1 || rule.window <= 0 {
		b.forget()
		return
	}
	now := b.clock.Now().UTC()
	count, reached := b.count(account, now, rule)
	if !reached {
		return
	}
	signal := Signal{
		Account: account, Kind: rule.kind, Weight: rule.weight, Dedupe: now.Truncate(rule.window).Format(time.RFC3339), At: now,
		Evidence: map[string]any{rule.countName: count, "window_seconds": int64(rule.window.Seconds())},
	}
	if _, err := b.store.Record(ctx, signal); err != nil {
		b.logger.Error("risk signal not recorded", "kind", rule.kind, "account", account, "error", err)
	}
}

func (b *Burst) forget() {
	b.mu.Lock()
	defer b.mu.Unlock()
	clear(b.recent)
}

func (b *Burst) count(account string, now time.Time, rule burstRule) (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	since := now.Add(-rule.window)
	b.sweep(now, since, rule.window)
	kept := within(b.recent[account], since)
	kept = append(kept, now)
	if len(kept) < rule.threshold {
		b.recent[account] = kept
		return len(kept), false
	}
	delete(b.recent, account)
	return len(kept), true
}

func (b *Burst) sweep(now, since time.Time, window time.Duration) {
	if now.Before(b.sweepAt) {
		return
	}
	b.sweepAt = now.Add(window)
	for account, times := range b.recent {
		if kept := within(times, since); len(kept) > 0 {
			b.recent[account] = kept
		} else {
			delete(b.recent, account)
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

func (b *Burst) Tracked() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.recent)
}
