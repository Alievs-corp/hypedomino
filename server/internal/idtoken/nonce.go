package idtoken

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type NonceStore interface {
	Save(ctx context.Context, nonce string, now time.Time, ttl time.Duration) error
	Consume(ctx context.Context, nonce string, now time.Time) (bool, error)
}

func (f NonceForm) encode(raw string) (string, bool) {
	switch f {
	case NonceRaw:
		return raw, true
	case NonceSHA256Hex:
		digest := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(digest[:]), true
	}
	return "", false
}

type MemoryNonces struct {
	mu        sync.Mutex
	expiries  map[string]time.Time
	nextPrune time.Time
}

func NewMemoryNonces() *MemoryNonces {
	return &MemoryNonces{expiries: map[string]time.Time{}}
}

func (m *MemoryNonces) Save(_ context.Context, nonce string, now time.Time, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpired(now, ttl)
	m.expiries[nonce] = now.Add(ttl)
	return nil
}

func (m *MemoryNonces) Consume(_ context.Context, nonce string, now time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	expiresAt, ok := m.expiries[nonce]
	delete(m.expiries, nonce)
	return ok && now.Before(expiresAt), nil
}

func (m *MemoryNonces) pruneExpired(now time.Time, interval time.Duration) {
	if now.Before(m.nextPrune) {
		return
	}
	for nonce, expiresAt := range m.expiries {
		if !now.Before(expiresAt) {
			delete(m.expiries, nonce)
		}
	}
	m.nextPrune = now.Add(interval)
}
