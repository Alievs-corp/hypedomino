package dbtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Alievs-corp/hypedomino/server/internal/db"
)

const (
	Image  = "postgres:16-alpine"
	EnvURL = "HD_TEST_DATABASE_URL"
)

var (
	mu        sync.Mutex
	adminURL  string
	container *postgres.PostgresContainer
	counter   atomic.Uint64
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
	adminURL = ""
}

func admin(t testing.TB) string {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	if adminURL != "" {
		return adminURL
	}
	if url := os.Getenv(EnvURL); url != "" {
		adminURL = url
		return adminURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	c, err := postgres.Run(ctx, Image,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword(rand.Text()),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("postgres container: %v", err)
	}
	url, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		c.Terminate(ctx)
		t.Fatalf("postgres connection string: %v", err)
	}
	container, adminURL = c, url
	return adminURL
}

func Empty(t testing.TB) string {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped under -short")
	}
	base := admin(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	config, err := pgx.ParseConfig(base)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect %s: %v", EnvURL, err)
	}
	name := fmt.Sprintf("hd_test_%d_%d", os.Getpid(), counter.Add(1))
	if _, err := conn.Exec(ctx, "create database "+pgx.Identifier{name}.Sanitize()); err != nil {
		conn.Close(ctx)
		t.Fatal(err)
	}
	conn.Close(ctx)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			return
		}
		defer conn.Close(ctx)
		conn.Exec(ctx, "drop database "+pgx.Identifier{name}.Sanitize()+" with (force)")
	})
	parsed, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + name
	return parsed.String()
}

func Fresh(t testing.TB) (string, *pgxpool.Pool) {
	t.Helper()
	dsn := Empty(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	migrator, err := db.NewMigrator(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrator.Up(ctx); err != nil {
		migrator.Close()
		t.Fatalf("migrate: %v", err)
	}
	if err := migrator.Close(); err != nil {
		t.Fatal(err)
	}
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return dsn, pool
}
