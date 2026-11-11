package panics

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"slices"
	"sync"
	"time"
)

var counts = struct {
	sync.Mutex
	by map[string]uint64
}{by: map[string]uint64{}}

func Caught(where string, value any, attrs ...any) bool {
	if value == nil {
		return false
	}
	counts.Lock()
	counts.by[where]++
	counts.Unlock()
	slog.Error("panic recovered", append([]any{"where", where, "panic", fmt.Sprint(value), "stack", string(debug.Stack())}, attrs...)...)
	return true
}

func Supervise(ctx context.Context, where string, every time.Duration, run func(context.Context, time.Duration)) {
	for !returned(ctx, where, every, run) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(every):
		}
	}
}

func returned(ctx context.Context, where string, every time.Duration, run func(context.Context, time.Duration)) bool {
	defer func() { Caught(where, recover()) }()
	run(ctx, every)
	return true
}

func WriteTo(w io.Writer) error {
	counts.Lock()
	defer counts.Unlock()
	if _, err := fmt.Fprint(w, "# HELP hd_panics_total Panics recovered in long-lived goroutines by site.\n# TYPE hd_panics_total counter\n"); err != nil {
		return err
	}
	sites := make([]string, 0, len(counts.by))
	for where := range counts.by {
		sites = append(sites, where)
	}
	slices.Sort(sites)
	for _, where := range sites {
		if _, err := fmt.Fprintf(w, "hd_panics_total{where=%q} %d\n", where, counts.by[where]); err != nil {
			return err
		}
	}
	return nil
}
