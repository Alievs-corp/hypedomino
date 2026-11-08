package rating_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestTheRatingRelayAppliesARankedResultOnceAndSkipsACancelledMatch(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	winner, loser := f.account(t), f.account(t)
	f.set(t, winner, rules.Mode1v1, 1300, 30, rank(t, "Silver I", 40))
	f.set(t, loser, rules.Mode1v1, 1250, 30, rank(t, "Silver I", 0))
	for id, cancelled := range map[string]bool{"11111111-1111-1111-1111-111111111111": false, "22222222-2222-2222-2222-222222222222": true} {
		payload, _ := json.Marshal(matchstore.Outcome{Match: id, Mode: rules.Mode1v1, Kind: matchstore.KindRanked, Cancelled: cancelled, EndedAt: now, Players: []matchstore.PlayerOutcome{{Account: loser, Team: 1}, {Account: winner, Team: 0, Won: !cancelled}}})
		if _, err := f.pool.Exec(ctx, `insert into outbox (topic, key, payload) values ($1, $2, $3)`, matchstore.TopicRating, "match:"+id, payload); err != nil {
			t.Fatal(err)
		}
	}
	relay := matchstore.NewRelay(matchstore.New(f.pool), f.store, matchstore.TopicRating, 0, nil)
	for range 2 {
		if published, err := relay.Once(ctx); err != nil || published != 2 {
			t.Fatalf("published %d, %v", published, err)
		}
		if _, err := f.pool.Exec(ctx, `update outbox set published_at = null where topic = $1`, matchstore.TopicRating); err != nil {
			t.Fatal(err)
		}
	}
	if won, lost := f.rating(t, winner, rules.Mode1v1), f.rating(t, loser, rules.Mode1v1); won.Matches != 31 || won.MMR <= 1300 || lost.Matches != 31 || lost.MMR >= 1250 {
		t.Fatalf("winner %+v, loser %+v", won, lost)
	}
	var results int
	if err := f.pool.QueryRow(ctx, `select count(*) from rating_results`).Scan(&results); err != nil || results != 1 {
		t.Fatalf("results %d, %v", results, err)
	}
}
