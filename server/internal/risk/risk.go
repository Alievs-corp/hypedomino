package risk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const KindClockSkew = "clock_skew"

var ErrInvalidConfig = errors.New("risk: invalid config")

type Config struct {
	ClockSkewTolerance time.Duration
	ClockSkewWeight    int
}

func DefaultConfig() Config {
	return Config{ClockSkewTolerance: 5 * time.Minute, ClockSkewWeight: 1}
}

func (c Config) Validate() error {
	if c.ClockSkewTolerance <= 0 || c.ClockSkewWeight < 0 {
		return fmt.Errorf("%w: clock skew tolerance must be positive and its weight not negative", ErrInvalidConfig)
	}
	return nil
}

type Signal struct {
	Account  string
	Kind     string
	Weight   int
	Dedupe   string
	Evidence map[string]any
	At       time.Time
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Record(ctx context.Context, signal Signal) (bool, error) {
	evidence, err := json.Marshal(signal.Evidence)
	if err != nil {
		return false, err
	}
	tag, err := s.pool.Exec(ctx, `insert into risk_signals (account_id, kind, weight, dedupe_key, evidence, created_at) values ($1, $2, $3, $4, $5, $6)
		on conflict (account_id, kind, dedupe_key) do nothing`, signal.Account, signal.Kind, signal.Weight, signal.Dedupe, evidence, signal.At)
	return tag.RowsAffected() == 1, err
}

func (s *Store) Signals(ctx context.Context, account string) ([]Signal, error) {
	rows, err := s.pool.Query(ctx, `select account_id, kind, weight, dedupe_key, evidence, created_at from risk_signals where account_id = $1 order by created_at, id`, account)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Signal, error) {
		var signal Signal
		var evidence []byte
		if err := row.Scan(&signal.Account, &signal.Kind, &signal.Weight, &signal.Dedupe, &evidence, &signal.At); err != nil {
			return signal, err
		}
		return signal, json.Unmarshal(evidence, &signal.Evidence)
	})
}

type Clock interface {
	Now() time.Time
}

type ClockSkew struct {
	store  *Store
	config Config
	clock  Clock
	logger *slog.Logger
}

func NewClockSkew(store *Store, config Config, clock Clock, logger *slog.Logger) *ClockSkew {
	return &ClockSkew{store: store, config: config, clock: clock, logger: logger}
}

func (c *ClockSkew) Observe(ctx context.Context, account string, clientUnixMs uint64) {
	if clientUnixMs == 0 {
		return
	}
	now := c.clock.Now().UTC()
	client := time.UnixMilli(int64(clientUnixMs)).UTC()
	skew := client.Sub(now)
	if skew.Abs() <= c.config.ClockSkewTolerance {
		return
	}
	signal := Signal{
		Account: account, Kind: KindClockSkew, Weight: c.config.ClockSkewWeight, Dedupe: now.Format(time.DateOnly), At: now,
		Evidence: map[string]any{"client_time": client.Format(time.RFC3339Nano), "server_time": now.Format(time.RFC3339Nano), "skew_seconds": int64(skew.Seconds())},
	}
	if _, err := c.store.Record(ctx, signal); err != nil {
		c.logger.Error("clock skew signal not recorded", "account", account, "error", err)
	}
}
