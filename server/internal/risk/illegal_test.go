package risk_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/anticheat"
	"github.com/Alievs-corp/hypedomino/server/internal/risk"
)

func TestObservingAnIllegalIntentNeverBlocksNorAllocates(t *testing.T) {
	illegal := risk.NewIllegalIntents(nil, func() anticheat.IllegalIntents { return anticheat.IllegalIntents{} }, &fixedClock{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	account := "0f5c7a52-6a3e-4c1e-9f57-0d7f1c9e2b11"
	if allocs := testing.AllocsPerRun(100, func() { illegal.Observe(account) }); allocs != 0 {
		t.Fatalf("observing an illegal intent allocated %v times", allocs)
	}
	done := make(chan struct{})
	go func() {
		for range 2000 {
			illegal.Observe(account)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("observing blocked with nothing draining the queue")
	}
	var metrics strings.Builder
	if _, err := illegal.WriteTo(&metrics); err != nil {
		t.Fatal(err)
	}
	if text := metrics.String(); !strings.Contains(text, "hd_illegal_intents_total 2101\n") || !strings.Contains(text, "hd_illegal_intents_dropped_total 1077\n") {
		t.Fatalf("metrics:\n%s", text)
	}
}

func BenchmarkObservingAnIllegalIntent(b *testing.B) {
	illegal := risk.NewIllegalIntents(nil, func() anticheat.IllegalIntents { return anticheat.IllegalIntents{} }, &fixedClock{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go illegal.Run(ctx, 0)
	account := "0f5c7a52-6a3e-4c1e-9f57-0d7f1c9e2b11"
	b.ReportAllocs()
	for b.Loop() {
		illegal.Observe(account)
	}
}
