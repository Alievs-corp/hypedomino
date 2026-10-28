package loadbot

import (
	"sync"
	"time"
)

const (
	bucketWidth = time.Millisecond
	bucketCount = 1000
)

type Histogram struct {
	mu      sync.Mutex
	buckets [bucketCount + 1]uint32
	count   uint64
	total   time.Duration
	max     time.Duration
}

func (h *Histogram) Add(d time.Duration) {
	index := int(d / bucketWidth)
	if index < 0 {
		index = 0
	}
	if index > bucketCount {
		index = bucketCount
	}
	h.mu.Lock()
	h.buckets[index]++
	h.count++
	h.total += d
	h.max = max(h.max, d)
	h.mu.Unlock()
}

func (h *Histogram) Count() uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

func (h *Histogram) Max() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.max
}

func (h *Histogram) Mean() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	return h.total / time.Duration(h.count)
}

func (h *Histogram) Percentile(p float64) time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	rank := uint64(p / 100 * float64(h.count))
	if rank < 1 {
		rank = 1
	}
	seen := uint64(0)
	for index, n := range h.buckets[:bucketCount] {
		seen += uint64(n)
		if seen >= rank {
			return time.Duration(index+1) * bucketWidth
		}
	}
	return h.max
}

type Windows struct {
	mu     sync.Mutex
	start  time.Time
	length time.Duration
	slots  []*Histogram
}

func NewWindows(start time.Time, length time.Duration) *Windows {
	return &Windows{start: start, length: length}
}

func (w *Windows) At(now time.Time) *Histogram {
	index := int(now.Sub(w.start) / w.length)
	if index < 0 {
		index = 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for len(w.slots) <= index {
		w.slots = append(w.slots, &Histogram{})
	}
	return w.slots[index]
}

func (w *Windows) Slots() []*Histogram {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]*Histogram(nil), w.slots...)
}
