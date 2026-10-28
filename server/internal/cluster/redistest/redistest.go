package redistest

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	Image  = "redis:7-alpine"
	EnvURL = "HD_TEST_REDIS_URL"
)

var (
	mu        sync.Mutex
	url       string
	container *tcredis.RedisContainer
)

func Main(m *testing.M) int {
	code := m.Run()
	Stop()
	return code
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		container.Terminate(ctx)
		container = nil
	}
	url = ""
}

func URL(t testing.TB) string {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped under -short")
	}
	mu.Lock()
	defer mu.Unlock()
	if url != "" {
		return url
	}
	if env := os.Getenv(EnvURL); env != "" {
		url = env
		return url
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	c, err := tcredis.Run(ctx, Image)
	if err != nil {
		t.Fatalf("redis container: %v", err)
	}
	connection, err := c.ConnectionString(ctx)
	if err != nil {
		c.Terminate(ctx)
		t.Fatalf("redis connection string: %v", err)
	}
	container, url = c, connection
	return url
}

func Client(t testing.TB) *redis.Client {
	t.Helper()
	options, err := redis.ParseURL(URL(t))
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect %s: %v", EnvURL, err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func Prefix(t testing.TB) string {
	t.Helper()
	return "hd:test:" + t.Name() + ":" + time.Now().UTC().Format("150405.000000") + ":"
}
