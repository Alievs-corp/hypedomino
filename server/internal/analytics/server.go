package analytics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	MissionComplete  = "mission_complete"
	PassLevel        = "pass_level"
	RewardGrant      = "reward_grant"
	PurchaseComplete = "purchase_complete"
)

func RecordServerIn(ctx context.Context, tx pgx.Tx, name string, count int, at time.Time) error {
	_, err := tx.Exec(ctx, `insert into analytics_events (name, label, value, received_at) select $1, '', 0, $2 from generate_series(1, $3)`, name, at, count)
	return err
}
