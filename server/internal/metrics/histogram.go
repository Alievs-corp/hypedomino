package metrics

import (
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"
)

type Histogram struct {
	name    string
	help    string
	bounds  []time.Duration
	buckets []atomic.Uint64
	count   atomic.Uint64
	sum     atomic.Int64
}

func NewHistogram(name, help string, bounds []time.Duration) *Histogram {
	return &Histogram{name: name, help: help, bounds: bounds, buckets: make([]atomic.Uint64, len(bounds))}
}

func (h *Histogram) Observe(elapsed time.Duration) {
	for i, bound := range h.bounds {
		if elapsed <= bound {
			h.buckets[i].Add(1)
			break
		}
	}
	h.count.Add(1)
	h.sum.Add(int64(elapsed))
}

func (h *Histogram) WriteTo(w io.Writer) (int64, error) {
	var written int64
	write := func(format string, args ...any) error {
		n, err := fmt.Fprintf(w, format, args...)
		written += int64(n)
		return err
	}
	if err := write("# HELP %s %s\n# TYPE %s histogram\n", h.name, h.help, h.name); err != nil {
		return written, err
	}
	var cumulative uint64
	for i, bound := range h.bounds {
		cumulative += h.buckets[i].Load()
		if err := write("%s_bucket{le=\"%s\"} %d\n", h.name, seconds(bound), cumulative); err != nil {
			return written, err
		}
	}
	count := h.count.Load()
	if err := write("%s_bucket{le=\"+Inf\"} %d\n%s_sum %s\n%s_count %d\n", h.name, count, h.name, seconds(time.Duration(h.sum.Load())), h.name, count); err != nil {
		return written, err
	}
	return written, nil
}

func seconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'g', -1, 64)
}

func WriteGauge(w io.Writer, name, help string, value float64) error {
	_, err := fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s gauge\n%s %s\n", name, help, name, name, strconv.FormatFloat(value, 'f', -1, 64))
	return err
}
