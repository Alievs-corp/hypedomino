package matchreward_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/matchreward"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

func TestTheEconomyRelayPaysEveryPlayerOfAFinishedMatchOnce(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	store := matchreward.NewStore(pool, matchreward.NewWatcher(pool, slog.New(slog.NewTextHandler(io.Discard, nil))), wallet.NewStore(pool))
	var winner, loser string
	for _, account := range []*string{&winner, &loser} {
		if err := pool.QueryRow(ctx, `insert into accounts (created_at) values ($1) returning id`, start.Add(-time.Hour)).Scan(account); err != nil {
			t.Fatal(err)
		}
	}
	outcome := matchstore.Outcome{Match: "m1", Mode: rules.Mode1v1, Kind: matchstore.KindRanked, EndedAt: start, Players: []matchstore.PlayerOutcome{{Account: winner, Won: true}, {Account: loser}}}
	payload, _ := json.Marshal(outcome)
	for _, topic := range []string{matchstore.TopicEconomy, matchstore.TopicXp} {
		if _, err := pool.Exec(ctx, `insert into outbox (topic, key, payload) values ($1, 'match:m1', $2)`, topic, payload); err != nil {
			t.Fatal(err)
		}
	}
	relay := matchstore.NewRelay(matchstore.New(pool), store, matchstore.TopicEconomy, 0, nil)
	if published, err := relay.Once(ctx); err != nil || published != 1 {
		t.Fatalf("published %d, %v", published, err)
	}
	if _, err := pool.Exec(ctx, `update outbox set published_at = null where topic = $1`, matchstore.TopicEconomy); err != nil {
		t.Fatal(err)
	}
	if published, err := relay.Once(ctx); err != nil || published != 1 {
		t.Fatalf("redelivery published %d, %v", published, err)
	}
	for account, want := range map[string]int64{winner: 130, loser: 10} {
		var got int64
		if err := pool.QueryRow(ctx, `select coalesce(sum(amount), 0) from ledger_entries where account_id = $1`, account).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("account paid %d, want %d", got, want)
		}
	}
}
