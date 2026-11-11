package panics

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func exposed(t *testing.T) string {
	t.Helper()
	var out strings.Builder
	if err := WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestSuperviseRestartsAPanickingJobAndCountsIt(t *testing.T) {
	logs := captureLogs(t)
	ctx, cancel := context.WithCancel(t.Context())
	runs := 0
	Supervise(ctx, "test job", time.Millisecond, func(ctx context.Context, every time.Duration) {
		runs++
		if runs < 3 {
			panic("job broke")
		}
		cancel()
	})
	if runs != 3 {
		t.Fatalf("runs %d, want 3", runs)
	}
	if !strings.Contains(exposed(t), `hd_panics_total{where="test job"} 2`) {
		t.Fatalf("metrics:\n%s", exposed(t))
	}
	for _, field := range []string{`"level":"ERROR"`, `"where":"test job"`, `"panic":"job broke"`, `"stack":"goroutine `} {
		if !strings.Contains(logs.String(), field) {
			t.Fatalf("missing %s in %s", field, logs.String())
		}
	}
}

func TestSuperviseStopsWhenTheJobReturns(t *testing.T) {
	runs := 0
	Supervise(t.Context(), "returning job", time.Millisecond, func(context.Context, time.Duration) { runs++ })
	if runs != 1 || strings.Contains(exposed(t), "returning job") {
		t.Fatalf("runs %d metrics:\n%s", runs, exposed(t))
	}
}

func TestCaughtIgnoresNoPanic(t *testing.T) {
	if Caught("calm", nil) || strings.Contains(exposed(t), "calm") {
		t.Fatal("counted a nil recover")
	}
}
