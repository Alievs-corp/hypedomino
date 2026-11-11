package account_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
)

func TestExportHoldsTheAccountsOwnDataAndWithholdsOthersAndDetectionInternals(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := account.New(pool)
	ctx := context.Background()
	ayla := signedUp(t, store, pool, "g-ayla", "Ayla")
	ben := signedUp(t, store, pool, "g-ben", "Ben")
	sanctioned := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	pair, own := []any{ayla.ID, ben.ID, sanctioned}, []any{ayla.ID, sanctioned}
	for _, seed := range []struct {
		statement string
		args      []any
	}{
		{`insert into friendships (account_id, friend_id, created_at) values ($1, $2, $3), ($2, $1, $3)`, pair},
		{`insert into blocks (account_id, blocked_id, created_at) values ($2, $1, $3)`, pair},
		{`insert into friend_requests (sender_id, recipient_id, created_at) values ($2, $1, $3)`, pair},
		{`insert into rating_signals (account_id, kind, counterpart, mode, matches, wins, offenses, window_hours, first_seen, updated_at) values ($1, 'win_trade', $2, '1v1', 3, 3, 0, 24, $3, $3)`, pair},
		{`insert into purchases (platform, transaction_id, account_id, offer_id, product_id, environment, state, grant_key, purchased_at) values ('google', 'tx-1', $1, 'offer', 'product', 'production', 'granted', 'secret-grant-key', $2)`, own},
		{`insert into risk_signals (account_id, kind, weight, dedupe_key, evidence, created_at) values ($1, 'clock_skew', 7, 'd-1', '{"detector":"hidden-evidence"}', $2::timestamptz - interval '1 day'), ($1, 'mac_failure', 5, 'd-2', '{}', $2::timestamptz + interval '1 day')`, own},
		{`insert into risk_actions (account_id, action, score, wave, created_at) values ($1, 'shadowban', 987654321, $2, $2)`, own},
	} {
		if _, err := pool.Exec(ctx, seed.statement, seed.args...); err != nil {
			t.Fatalf("%s: %v", seed.statement, err)
		}
	}
	generated := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	raw, err := store.Export(ctx, ayla.ID, generated)
	if err != nil {
		t.Fatal(err)
	}
	for _, hidden := range []string{ben.ID, `"subject"`, `"provider"`, "g-ben", "secret-grant-key", "hidden-evidence", "987654321", "weight", "mac_failure", "win_trade"} {
		if strings.Contains(string(raw), hidden) {
			t.Errorf("export reveals %q: %s", hidden, raw)
		}
	}
	var document struct {
		SchemaVersion int                          `json:"schema_version"`
		GeneratedAt   time.Time                    `json:"generated_at"`
		Sections      map[string][]json.RawMessage `json:"sections"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != account.ExportSchemaVersion || !document.GeneratedAt.Equal(generated) {
		t.Fatalf("header %d %v", document.SchemaVersion, document.GeneratedAt)
	}
	wants := map[string][]string{
		"account":          {`"id":"` + ayla.ID + `"`},
		"email":            {`"email":"g-ayla@example.com"`},
		"profile":          {`"nickname":"Ayla"`},
		"friends":          {`"nickname":"Ben"`},
		"friend_requests":  {`"direction":"received"`, `"nickname":"Ben"`},
		"purchases":        {`"transaction_id":"tx-1"`},
		"sanctions":        {`"action":"shadowban"`},
		"sanction_signals": {`"category":"clock_skew"`},
		"blocks":           nil,
	}
	for name, fragments := range wants {
		rows, present := document.Sections[name]
		want := 1
		if fragments == nil {
			want = 0
		}
		if !present || len(rows) != want {
			t.Errorf("section %s: %s", name, rows)
			continue
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(rows[0]), fragment) {
				t.Errorf("section %s lacks %s: %s", name, fragment, rows[0])
			}
		}
	}
}
