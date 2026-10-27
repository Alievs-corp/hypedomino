package idtoken

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errNoUsableKeys = errors.New("no usable RS256 keys")

type keyCache struct {
	url             string
	client          *http.Client
	minTTL          time.Duration
	maxTTL          time.Duration
	refetchInterval time.Duration
	fetchTimeout    time.Duration
	maxBytes        int64
	minBits         int

	mu          sync.Mutex
	keys        map[string]*rsa.PublicKey
	freshUntil  time.Time
	lastAttempt time.Time
	attempted   bool
}

func newKeyCache(url string, config Config, client *http.Client) *keyCache {
	return &keyCache{
		url:             url,
		client:          client,
		minTTL:          config.KeyCacheMinTTL,
		maxTTL:          config.KeyCacheMaxTTL,
		refetchInterval: config.KeyRefetchMinInterval,
		fetchTimeout:    config.KeyFetchTimeout,
		maxBytes:        config.MaxKeySetBytes,
		minBits:         config.MinRSAKeyBits,
	}
}

func (c *keyCache) lookup(ctx context.Context, keyID string, now time.Time) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key, known := c.keys[keyID]
	if known && now.Before(c.freshUntil) {
		return key, nil
	}
	if !c.mayFetch(now) {
		if known {
			return key, nil
		}
		return nil, ErrUnknownKey
	}
	fetchErr := c.refresh(ctx, now)
	if fresh, found := c.keys[keyID]; found {
		return fresh, nil
	}
	if fetchErr != nil {
		return nil, fetchErr
	}
	return nil, ErrUnknownKey
}

func (c *keyCache) mayFetch(now time.Time) bool {
	return !c.attempted || now.Sub(c.lastAttempt) >= c.refetchInterval
}

func (c *keyCache) refresh(ctx context.Context, now time.Time) error {
	c.attempted = true
	c.lastAttempt = now
	keys, maxAge, err := c.fetch(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrKeyFetch, err)
	}
	c.keys = keys
	c.freshUntil = now.Add(clampDuration(maxAge, c.minTTL, c.maxTTL))
	return nil
}

func (c *keyCache) fetch(ctx context.Context) (map[string]*rsa.PublicKey, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, c.fetchTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, 0, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, c.maxBytes+1))
	if err != nil {
		return nil, 0, err
	}
	if int64(len(body)) > c.maxBytes {
		return nil, 0, fmt.Errorf("key set larger than %d bytes", c.maxBytes)
	}
	keys, err := parseKeySet(body, c.minBits)
	if err != nil {
		return nil, 0, err
	}
	return keys, cacheMaxAge(response.Header.Get("Cache-Control")), nil
}

type jsonWebKey struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func parseKeySet(body []byte, minBits int) (map[string]*rsa.PublicKey, error) {
	var set struct {
		Keys []jsonWebKey `json:"keys"`
	}
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, candidate := range set.Keys {
		key, ok := candidate.signingKey(minBits)
		if ok {
			keys[candidate.KeyID] = key
		}
	}
	if len(keys) == 0 {
		return nil, errNoUsableKeys
	}
	return keys, nil
}

func (k jsonWebKey) signingKey(minBits int) (*rsa.PublicKey, bool) {
	if k.KeyType != "RSA" || k.KeyID == "" {
		return nil, false
	}
	if k.Use != "" && k.Use != "sig" {
		return nil, false
	}
	if k.Algorithm != "" && k.Algorithm != algorithmRS256 {
		return nil, false
	}
	modulusBytes, err := segmentEncoding.DecodeString(k.Modulus)
	if err != nil {
		return nil, false
	}
	exponentBytes, err := segmentEncoding.DecodeString(k.Exponent)
	if err != nil || len(exponentBytes) == 0 || len(exponentBytes) > 4 {
		return nil, false
	}
	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := int(new(big.Int).SetBytes(exponentBytes).Int64())
	if modulus.BitLen() < minBits || exponent < 3 || exponent%2 == 0 {
		return nil, false
	}
	return &rsa.PublicKey{N: modulus, E: exponent}, true
}

func cacheMaxAge(cacheControl string) time.Duration {
	for _, directive := range strings.Split(cacheControl, ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(directive), "=")
		if !strings.EqualFold(name, "max-age") {
			continue
		}
		seconds, err := strconv.ParseInt(strings.Trim(value, `"`), 10, 64)
		if err != nil || seconds < 0 {
			return 0
		}
		return time.Duration(min(seconds, int64(math.MaxInt64/time.Second))) * time.Second
	}
	return 0
}

func clampDuration(value, floor, ceiling time.Duration) time.Duration {
	return max(floor, min(value, ceiling))
}
