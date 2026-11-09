package rating

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrShowcase       = errors.New("ranked.showcase_invalid")
	ErrEmblemNotOwned = errors.New("ranked.emblem_not_owned")
)

func emblemID(mode, season, board string) string {
	return mode + ":" + season + ":" + board
}

func (s *Store) Showcase(ctx context.Context, account string) ([]string, error) {
	var chosen []string
	err := s.pool.QueryRow(ctx, `select emblems from emblem_showcases where account_id = $1`, account).Scan(&chosen)
	if errors.Is(err, pgx.ErrNoRows) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	owned, err := s.Emblems(ctx, account)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(chosen, func(id string) bool {
		return !slices.ContainsFunc(owned, func(e Emblem) bool { return e.ID == id })
	}), nil
}

func (s *Store) SetShowcase(ctx context.Context, account string, chosen []string, now time.Time) ([]string, error) {
	current, _, ok := s.configs.Current()
	if !ok {
		return nil, ErrNoConfig
	}
	if len(chosen) > current.ProfileEmblemSlots {
		return nil, ErrShowcase
	}
	for i, id := range chosen {
		if slices.Contains(chosen[:i], id) {
			return nil, ErrShowcase
		}
	}
	owned, err := s.Emblems(ctx, account)
	if err != nil {
		return nil, err
	}
	for _, id := range chosen {
		if !slices.ContainsFunc(owned, func(e Emblem) bool { return e.ID == id }) {
			return nil, ErrEmblemNotOwned
		}
	}
	showcase := append([]string{}, chosen...)
	if _, err := s.pool.Exec(ctx, `insert into emblem_showcases (account_id, emblems, updated_at) values ($1, $2, $3)
		on conflict (account_id) do update set emblems = excluded.emblems, updated_at = excluded.updated_at`, account, showcase, now); err != nil {
		return nil, err
	}
	return showcase, nil
}
