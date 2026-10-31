package gamestrings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUnknownKey = errors.New("gamestrings: unknown key")
	ErrNoBundle   = errors.New("gamestrings: no such bundle version")
)

type Published struct {
	Language Language
	Version  int
	Hash     string
	Content  []byte
	By       string
	At       time.Time
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) SaveEntry(ctx context.Context, entry Entry, by string, now time.Time) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	placeholders, err := json.Marshal(entry.placeholders())
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `insert into game_strings (key, note, max_length, placeholders, deprecated, updated_by, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (key) do update set note = excluded.note, max_length = excluded.max_length,
			placeholders = excluded.placeholders, deprecated = excluded.deprecated,
			updated_by = excluded.updated_by, updated_at = excluded.updated_at`,
		entry.Key, entry.Note, entry.MaxLength, placeholders, entry.Deprecated, by, now.UTC())
	return err
}

func (e Entry) placeholders() []Placeholder {
	if e.Placeholders == nil {
		return []Placeholder{}
	}
	return e.Placeholders
}

func (s *Store) SetValue(ctx context.Context, key string, language Language, value Value, by string, now time.Time) error {
	if !slices.Contains(Languages, language) {
		return fmt.Errorf("%w: unknown language %q", ErrInvalidValue, language)
	}
	if value.Status != Draft && value.Status != Approved {
		return fmt.Errorf("%w: status %q cannot be set", ErrInvalidValue, value.Status)
	}
	entry, err := s.entry(ctx, key)
	if err != nil {
		return err
	}
	if value.Status == Approved {
		if err := entry.CheckValue(language, value.Text); err != nil {
			return err
		}
	}
	return upsertValue(ctx, s.pool, key, language, value, by, now)
}

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func upsertValue(ctx context.Context, db executor, key string, language Language, value Value, by string, now time.Time) error {
	_, err := db.Exec(ctx, `insert into game_string_values (key, language, text, status, updated_by, updated_at)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (key, language) do update set text = excluded.text, status = excluded.status,
			updated_by = excluded.updated_by, updated_at = excluded.updated_at`,
		key, language, value.Text, value.Status, by, now.UTC())
	return err
}

const entryColumns = `key, note, max_length, placeholders, deprecated`

func scanEntry(row pgx.CollectableRow) (Entry, error) {
	var e Entry
	var placeholders []byte
	if err := row.Scan(&e.Key, &e.Note, &e.MaxLength, &placeholders, &e.Deprecated); err != nil {
		return Entry{}, err
	}
	if err := json.Unmarshal(placeholders, &e.Placeholders); err != nil {
		return Entry{}, err
	}
	e.Values = map[Language]Value{}
	return e, nil
}

func (s *Store) entry(ctx context.Context, key string) (Entry, error) {
	rows, err := s.pool.Query(ctx, `select `+entryColumns+` from game_strings where key = $1`, key)
	if err != nil {
		return Entry{}, err
	}
	entry, err := pgx.CollectExactlyOneRow(rows, scanEntry)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entry{}, fmt.Errorf("%w: %s", ErrUnknownKey, key)
	}
	return entry, err
}

func (s *Store) Entries(ctx context.Context) ([]Entry, error) {
	rows, err := s.pool.Query(ctx, `select `+entryColumns+` from game_strings order by key`)
	if err != nil {
		return nil, err
	}
	entries, err := pgx.CollectRows(rows, scanEntry)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		byKey[entry.Key] = entry
	}
	rows, err = s.pool.Query(ctx, `select key, language, text, status from game_string_values`)
	if err != nil {
		return nil, err
	}
	var key string
	var language Language
	var value Value
	_, err = pgx.ForEachRow(rows, []any{&key, &language, &value.Text, &value.Status}, func() error {
		byKey[key].Values[language] = value
		return nil
	})
	return entries, err
}

func (s *Store) Publish(ctx context.Context, language Language, by string, now time.Time) (Published, []Problem, error) {
	if !slices.Contains(Languages, language) {
		return Published{}, nil, fmt.Errorf("%w: unknown language %q", ErrInvalidValue, language)
	}
	entries, err := s.Entries(ctx)
	if err != nil {
		return Published{}, nil, err
	}
	bundle, problems := BuildBundle(entries, language)
	if problems != nil {
		return Published{}, problems, fmt.Errorf("%w: %d values fail their checks", ErrBlocked, len(problems))
	}
	content, hash, err := bundle.Encode()
	if err != nil {
		return Published{}, nil, err
	}
	published, err := s.store(ctx, Published{Language: language, Hash: hash, Content: content, By: by, At: now.UTC()})
	return published, nil, err
}

func (s *Store) store(ctx context.Context, bundle Published) (Published, error) {
	var stored Published
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtext('game_string_bundles:' || $1))`, bundle.Language); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into game_string_bundles (language, version, hash, content, published_by, published_at)
			select $1, coalesce(max(version), 0) + 1, $2, $3, $4, $5 from game_string_bundles where language = $1
			on conflict (language, hash) do nothing`,
			bundle.Language, bundle.Hash, bundle.Content, bundle.By, bundle.At); err != nil {
			return err
		}
		var err error
		stored, err = scanBundle(tx.QueryRow(ctx, `select `+bundleColumns+` from game_string_bundles where language = $1 and hash = $2`, bundle.Language, bundle.Hash))
		return err
	})
	return stored, err
}

func (s *Store) Export(ctx context.Context, language Language, format Format) ([]byte, error) {
	entries, err := s.Entries(ctx)
	if err != nil {
		return nil, err
	}
	transfer, err := BuildTransfer(entries, language)
	if err != nil {
		return nil, err
	}
	return transfer.Encode(format)
}

func (s *Store) Import(ctx context.Context, language Language, format Format, data []byte, by string, now time.Time) ([]Change, []Problem, error) {
	transfer, err := ParseTransfer(data, format, language)
	if err != nil {
		return nil, nil, err
	}
	entries, err := s.Entries(ctx)
	if err != nil {
		return nil, nil, err
	}
	changes, problems := PlanImport(entries, transfer)
	if problems != nil {
		return nil, problems, fmt.Errorf("%w: %d rows fail their checks", ErrImportBlocked, len(problems))
	}
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, change := range changes {
			if err := upsertValue(ctx, tx, change.Key, language, Value{Text: change.Text, Status: Draft}, by, now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return changes, nil, nil
}

const bundleColumns = `language, version, hash, content, published_by, published_at`

func scanBundle(row pgx.Row) (Published, error) {
	var p Published
	if err := row.Scan(&p.Language, &p.Version, &p.Hash, &p.Content, &p.By, &p.At); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Published{}, ErrNoBundle
		}
		return Published{}, err
	}
	p.At = p.At.UTC()
	return p, nil
}

func (s *Store) Bundle(ctx context.Context, language Language, version int) (Published, error) {
	return scanBundle(s.pool.QueryRow(ctx, `select `+bundleColumns+` from game_string_bundles where language = $1 and version = $2`, language, version))
}
