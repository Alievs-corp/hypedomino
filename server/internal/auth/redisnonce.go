package auth

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisNonces struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisNonces(client redis.UniversalClient, prefix string) *RedisNonces {
	return &RedisNonces{client: client, prefix: prefix + "nonce:"}
}

func (n *RedisNonces) Save(ctx context.Context, nonce string, _ time.Time, ttl time.Duration) error {
	return n.client.Set(ctx, n.prefix+nonce, "1", ttl).Err()
}

func (n *RedisNonces) Consume(ctx context.Context, nonce string, _ time.Time) (bool, error) {
	err := n.client.GetDel(ctx, n.prefix+nonce).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return err == nil, err
}
