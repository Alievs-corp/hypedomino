package risk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/anticheat"
)

const illegalIntentQueue = 1024

type IllegalIntents struct {
	burst    *Burst
	queue    chan string
	observed atomic.Int64
	dropped  atomic.Int64
}

func NewIllegalIntents(store *Store, rule func() anticheat.IllegalIntents, clock Clock, logger *slog.Logger) *IllegalIntents {
	current := func() burstRule {
		r := rule()
		return burstRule{kind: KindIllegalIntent, countName: "intents", threshold: r.Threshold, window: r.Window, weight: r.Weight}
	}
	return &IllegalIntents{burst: newBurst(store, current, clock, logger), queue: make(chan string, illegalIntentQueue)}
}

func (i *IllegalIntents) Observe(account string) {
	i.observed.Add(1)
	select {
	case i.queue <- account:
	default:
		i.dropped.Add(1)
	}
}

func (i *IllegalIntents) Run(ctx context.Context, _ time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case account := <-i.queue:
			i.burst.Observe(ctx, account)
		}
	}
}

func (i *IllegalIntents) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "# HELP hd_illegal_intents_total Match intents refused as illegal.\n# TYPE hd_illegal_intents_total counter\nhd_illegal_intents_total %d\n"+
		"# HELP hd_illegal_intents_dropped_total Illegal intents not counted toward a risk signal because the queue was full.\n# TYPE hd_illegal_intents_dropped_total counter\nhd_illegal_intents_dropped_total %d\n",
		i.observed.Load(), i.dropped.Load())
	return int64(n), err
}
