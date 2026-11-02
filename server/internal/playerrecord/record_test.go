package playerrecord_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
)

func TestARecordCarriesInventoryGrantsPurchasesAndMatches(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	at := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	var id string
	if err := pool.QueryRow(ctx, `insert into accounts (country) values ('AZ') returning id::text`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `insert into matches (id, started_at, mode, kind, config_version, seed, score_a, score_b, winner_team)
		values ('11111111-1111-1111-1111-111111111111', $1, '1v1', 'ranked', 3, 9, 101, 87, 0)`, at); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`insert into inventory_items (account_id, item_id, source_type, source_ref, granted_at, revoked_at) values ($1, 'avatar.comet_fox', 'shop', 'offer.avatar_comet_fox', $2, $2)`,
		`insert into reward_grants (account_id, bundle_id, revision, source_type, source_ref, idempotency_key, config_version, report, created_at)
			values ($1, 'bundle.starter', 2, 'shop', 'offer.starter', 'shop:r1', 4, '{}', $2)`,
		`insert into purchases (platform, transaction_id, account_id, offer_id, product_id, environment, state, purchased_at)
			values ('google', 'GPA.1', $1, 'offer.starter', 'offer_starter', 'production', 'refunded', $2)`,
		`insert into match_participants (match_id, started_at, seat, team, account_id, result)
			values ('11111111-1111-1111-1111-111111111111', $2, 1, 0, $1, 'win')`,
	} {
		if _, err := pool.Exec(ctx, statement, id, at); err != nil {
			t.Fatal(err)
		}
	}

	record, err := playerrecord.Load(ctx, pool, id, false)
	if err != nil {
		t.Fatal(err)
	}
	if record.Account.Country != "AZ" || record.Profile != nil || len(record.Account.Emails) != 0 || record.Coins != 0 || record.Crystals != 0 {
		t.Fatalf("account %+v", record)
	}
	if len(record.Inventory) != 1 || record.Inventory[0].RevokedAt == nil || len(record.Grants) != 1 || record.Grants[0].Revision != 2 {
		t.Fatalf("inventory %+v grants %+v", record.Inventory, record.Grants)
	}
	if len(record.Purchases) != 1 || record.Purchases[0].State != "refunded" || record.Purchases[0].GrantedAt != nil {
		t.Fatalf("purchases %+v", record.Purchases)
	}
	if len(record.Matches) != 1 || record.Matches[0].Result != "win" || record.Matches[0].ScoreA != 101 || record.Matches[0].Seat != 1 {
		t.Fatalf("matches %+v", record.Matches)
	}
	if _, err := playerrecord.Load(ctx, pool, "11111111-1111-1111-1111-111111111111", false); !errors.Is(err, playerrecord.ErrNotFound) {
		t.Fatalf("unknown account: %v", err)
	}
}
