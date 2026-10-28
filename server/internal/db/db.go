package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var embedded embed.FS

var ErrMigrationFile = errors.New("db: malformed migration file")

func Migrations() fs.FS {
	sub, err := fs.Sub(embedded, "migrations")
	if err != nil {
		panic(err)
	}
	return sub
}

type Migration struct {
	Version int64
	Name    string
}

func ListMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(Migrations(), ".")
	if err != nil {
		return nil, err
	}
	var migrations []Migration
	seen := map[int64]string{}
	for _, entry := range entries {
		name := entry.Name()
		prefix, _, ok := strings.Cut(name, "_")
		version, parseErr := strconv.ParseInt(prefix, 10, 64)
		if !ok || parseErr != nil || version <= 0 || !strings.HasSuffix(name, ".sql") {
			return nil, fmt.Errorf("%w: %s", ErrMigrationFile, name)
		}
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("%w: version %d used by %s and %s", ErrMigrationFile, version, other, name)
		}
		seen[version] = name
		data, err := fs.ReadFile(Migrations(), name)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(string(data), "-- +goose Up") || !strings.Contains(string(data), "-- +goose Down") {
			return nil, fmt.Errorf("%w: %s lacks an Up or a Down section", ErrMigrationFile, name)
		}
		migrations = append(migrations, Migration{Version: version, Name: name})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	return migrations, nil
}

type Migrator struct {
	db       *sql.DB
	provider *goose.Provider
}

func NewMigrator(url string) (*Migrator, error) {
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	provider, err := goose.NewProvider(database.DialectPostgres, sqlDB, Migrations(), goose.WithSessionLocker(locker))
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	return &Migrator{db: sqlDB, provider: provider}, nil
}

func (m *Migrator) Up(ctx context.Context) (int64, error) {
	if _, err := m.provider.Up(ctx); err != nil {
		return 0, err
	}
	return m.provider.GetDBVersion(ctx)
}

func (m *Migrator) DownTo(ctx context.Context, version int64) (int64, error) {
	if _, err := m.provider.DownTo(ctx, version); err != nil {
		return 0, err
	}
	return m.provider.GetDBVersion(ctx)
}

func (m *Migrator) Version(ctx context.Context) (int64, error) {
	return m.provider.GetDBVersion(ctx)
}

type Status struct {
	Version int64
	Name    string
	Applied bool
}

func (m *Migrator) Status(ctx context.Context) ([]Status, error) {
	statuses, err := m.provider.Status(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Status, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, Status{Version: s.Source.Version, Name: s.Source.Path, Applied: s.State == goose.StateApplied})
	}
	return out, nil
}

func (m *Migrator) Close() error {
	return errors.Join(m.provider.Close(), m.db.Close())
}

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

var _ = stdlib.GetDefaultDriver
