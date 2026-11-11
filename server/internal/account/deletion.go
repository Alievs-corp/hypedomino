package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

var ErrNoDeletionPending = errors.New("account: no deletion pending")

type erasure int

const (
	eraseRows erasure = iota + 1
	moveToTombstone
)

var erasures = map[string]erasure{
	"auth_identities":     eraseRows,
	"auth_sessions":       eraseRows,
	"app_attest_keys":     eraseRows,
	"push_devices":        eraseRows,
	"push_settings":       eraseRows,
	"profiles":            eraseRows,
	"nickname_history":    eraseRows,
	"inventory_items":     eraseRows,
	"wallets":             eraseRows,
	"ledger_entries":      eraseRows,
	"reward_grants":       eraseRows,
	"chest_opens":         eraseRows,
	"promo_targets":       eraseRows,
	"promo_redemptions":   eraseRows,
	"promo_attempts":      eraseRows,
	"pass_progress":       eraseRows,
	"pass_xp_events":      eraseRows,
	"medal_ownership":     eraseRows,
	"medal_showcases":     eraseRows,
	"mission_assignments": eraseRows,
	"mission_matches":     eraseRows,
	"daily_logins":        eraseRows,
	"friendships":         eraseRows,
	"friend_requests":     eraseRows,
	"blocks":              eraseRows,
	"friend_codes":        eraseRows,
	"ratings":             eraseRows,
	"match_participants":  moveToTombstone,
	"rating_changes":      moveToTombstone,
	"season_results":      moveToTombstone,
	"season_rewards":      moveToTombstone,
	"leaderboard_emblems": moveToTombstone,
	"emblem_showcases":    moveToTombstone,
	"rating_signals":      moveToTombstone,
	"purchases":           moveToTombstone,
	"penalty_records":     moveToTombstone,
	"risk_signals":        moveToTombstone,
	"risk_reviews":        moveToTombstone,
	"risk_actions":        moveToTombstone,
	"player_reports":      moveToTombstone,
	"data_exports":        moveToTombstone,
}

type reference struct {
	table  string
	column string
}

func (s *Store) Lookup(ctx context.Context, identity idtoken.Identity) (Account, error) {
	a, found, err := s.find(ctx, identity)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, ErrUnknown
	}
	return a, admitted(a, time.Now())
}

func (s *Store) RequestDeletion(ctx context.Context, id string, now time.Time) (time.Time, error) {
	var requested time.Time
	err := s.pool.QueryRow(ctx, `with requested as (update accounts set state = $2, deletion_requested_at = $3
		where id = $1 and state = $4 and (banned_until is null or banned_until <= $3) returning id, deletion_requested_at),
		unregistered as (delete from push_devices where account_id in (select id from requested))
		select deletion_requested_at from requested`,
		id, StateDeleting, now.UTC(), StateActive).Scan(&requested)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrBanned
	}
	return requested.UTC(), err
}

func (s *Store) CancelDeletion(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `update accounts set state = $2, deletion_requested_at = null where id = $1 and state = $3`, id, StateActive, StateDeleting)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoDeletionPending
	}
	return nil
}

func (s *Store) DeleteDue(ctx context.Context, requestedBefore, now time.Time) (int, error) {
	rows, err := s.pool.Query(ctx, `select id::text from accounts where state = $1 and deletion_requested_at <= $2 order by deletion_requested_at`, StateDeleting, requestedBefore.UTC())
	if err != nil {
		return 0, err
	}
	due, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, id := range due {
		done, err := s.erase(ctx, id, requestedBefore, now)
		if err != nil {
			return deleted, fmt.Errorf("erase %s: %w", id, err)
		}
		if done {
			deleted++
		}
	}
	return deleted, nil
}

func (s *Store) RunDeletions(ctx context.Context, logger *slog.Logger, every, grace time.Duration, now func() time.Time) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		at := now()
		deleted, err := s.DeleteDue(ctx, at.Add(-grace), at)
		if err != nil {
			logger.Error("account deletion failed", "error", err)
		}
		if deleted > 0 {
			logger.Info("accounts deleted", "count", deleted)
		}
	}
}

func (s *Store) erase(ctx context.Context, id string, requestedBefore, now time.Time) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var due bool
	err = tx.QueryRow(ctx, `select state = $2 and deletion_requested_at <= $3 from accounts where id = $1 for update`, id, StateDeleting, requestedBefore.UTC()).Scan(&due)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !due) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	references, err := accountReferences(ctx, tx)
	if err != nil {
		return false, err
	}
	var tombstone string
	if err := tx.QueryRow(ctx, `insert into accounts (state, created_at, deleted_at) values ($1, $2, $2) returning id::text`, StateDeleted, now.UTC()).Scan(&tombstone); err != nil {
		return false, err
	}
	if err := holdNickname(ctx, tx, id, tombstone, now); err != nil {
		return false, err
	}
	for _, r := range references {
		if err := apply(ctx, tx, r, id, tombstone); err != nil {
			return false, err
		}
	}
	if _, err := tx.Exec(ctx, `delete from accounts where id = $1`, id); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

func holdNickname(ctx context.Context, tx pgx.Tx, id, tombstone string, now time.Time) error {
	_, err := tx.Exec(ctx, `insert into nickname_history (account_id, nickname, nickname_key, replaced_at)
		select $2, nickname, nickname_key, $3 from profiles where account_id = $1`, id, tombstone, now.UTC())
	return err
}

func apply(ctx context.Context, tx pgx.Tx, r reference, id, tombstone string) error {
	table, column := pgx.Identifier{r.table}.Sanitize(), pgx.Identifier{r.column}.Sanitize()
	var err error
	switch erasures[r.table] {
	case eraseRows:
		_, err = tx.Exec(ctx, `delete from `+table+` where `+column+` = $1`, id)
	case moveToTombstone:
		_, err = tx.Exec(ctx, `update `+table+` set `+column+` = $2 where `+column+` = $1`, id, tombstone)
	default:
		err = fmt.Errorf("table %s references accounts and has no erasure rule", r.table)
	}
	return err
}

func accountReferences(ctx context.Context, tx pgx.Tx) ([]reference, error) {
	rows, err := tx.Query(ctx, `select c.conrelid::regclass::text, a.attname::text from pg_constraint c
		join pg_attribute a on a.attrelid = c.conrelid and a.attnum = any(c.conkey)
		where c.contype = 'f' and c.confrelid = 'accounts'::regclass and c.conparentid = 0
		order by 1, 2`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (reference, error) {
		var r reference
		err := row.Scan(&r.table, &r.column)
		return r, err
	})
}
