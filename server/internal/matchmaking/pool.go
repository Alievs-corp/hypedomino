package matchmaking

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var (
	ErrAlreadyQueued = errors.New("matchmaking: account already queued")
	ErrInvalidEntry  = errors.New("matchmaking: invalid entry")
)

type Pool struct {
	config  Config
	random  *rand.Rand
	mu      sync.Mutex
	entries map[string]Entry
}

func NewPool(config Config, random *rand.Rand) *Pool {
	return &Pool{config: config, random: random, entries: map[string]Entry{}}
}

func (p *Pool) Enqueue(entry Entry) error {
	if err := validate(entry); err != nil {
		return err
	}
	if entry.SearchFrom.IsZero() {
		entry.SearchFrom = entry.EnqueuedAt
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, queued := p.entries[entry.Account]; queued {
		return ErrAlreadyQueued
	}
	p.entries[entry.Account] = entry
	return nil
}

func validate(entry Entry) error {
	switch {
	case entry.Account == "":
		return fmt.Errorf("%w: empty account", ErrInvalidEntry)
	case entry.Mode != rules.Mode1v1 && entry.Mode != rules.Mode2v2:
		return fmt.Errorf("%w: mode %s", ErrInvalidEntry, entry.Mode)
	case entry.Queue != QueueRanked && entry.Queue != QueueCasual:
		return fmt.Errorf("%w: queue %s", ErrInvalidEntry, entry.Queue)
	case entry.EnqueuedAt.IsZero():
		return fmt.Errorf("%w: no enqueue time", ErrInvalidEntry)
	}
	return nil
}

func (p *Pool) Leave(account string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, queued := p.entries[account]
	delete(p.entries, account)
	return queued
}

func (p *Pool) Queued(account string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, queued := p.entries[account]
	return queued
}

func (p *Pool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.entries)
}

func (p *Pool) Tick(now time.Time) []Match {
	p.mu.Lock()
	defer p.mu.Unlock()
	entries := make([]Entry, 0, len(p.entries))
	for _, entry := range p.entries {
		entries = append(entries, entry)
	}
	matches, remaining := Form(p.config, entries, now, p.random)
	p.entries = make(map[string]Entry, len(remaining))
	for _, entry := range remaining {
		p.entries[entry.Account] = entry
	}
	return matches
}
