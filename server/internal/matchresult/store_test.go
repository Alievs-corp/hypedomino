package matchresult_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/matchresult"
	"github.com/Alievs-corp/hypedomino/server/internal/matchreward"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

var (
	start = time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	end   = start.Add(9*time.Minute + 123456789*time.Nanosecond)
)

const retryAfter = 2 * time.Second

type fixture struct {
	pool   *pgxpool.Pool
	store  *matchresult.Store
	winner string
	loser  string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	_, pool := dbtest.Fresh(t)
	f := fixture{pool: pool, store: matchresult.NewStore(pool, retryAfter)}
	f.winner, f.loser = f.account(t), f.account(t)
	return f
}

func (f fixture) account(t *testing.T) string {
	t.Helper()
	var id string
	if err := f.pool.QueryRow(context.Background(), `insert into accounts (created_at) values ($1) returning id`, start.Add(-time.Hour)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func (f fixture) begin(t *testing.T, id, kind string) matchstore.Started {
	t.Helper()
	started := matchstore.Started{ID: id, StartedAt: start, Mode: rules.Mode1v1, Kind: kind, Seed: 7, Participants: []matchstore.Participant{{Seat: rules.SeatX, Account: f.winner}, {Seat: rules.SeatY, Account: f.loser}}}
	if err := matchstore.New(f.pool).Begin(context.Background(), started); err != nil {
		t.Fatal(err)
	}
	return started
}

func (f fixture) finish(t *testing.T, started matchstore.Started) {
	t.Helper()
	outcome := matchstore.Outcome{Match: started.ID, Mode: started.Mode, Kind: started.Kind, EndedAt: end, Players: []matchstore.PlayerOutcome{{Account: f.winner, Team: 0, Won: true}, {Account: f.loser, Team: 1}}}
	finished := matchstore.Finished{ID: started.ID, StartedAt: start, EndedAt: end, HandsPlayed: 3, Scores: rules.TeamPoints{150, 80}, Winner: rules.TeamA, HasWinner: true,
		Results: map[rules.Seat]string{rules.SeatX: matchstore.ResultWin, rules.SeatY: matchstore.ResultLoss}, Outbox: matchstore.MatchMessages(started.Kind, "match:"+started.ID, outcome)}
	if err := matchstore.New(f.pool).Finish(context.Background(), finished); err != nil {
		t.Fatal(err)
	}
}

func (f fixture) relay(t *testing.T, topic string, publisher matchstore.Publisher) {
	t.Helper()
	if published, err := matchstore.NewRelay(matchstore.New(f.pool), publisher, topic, 0, nil).Once(context.Background()); err != nil || published != 1 {
		t.Fatalf("relay %s published %d, %v", topic, published, err)
	}
}

func (f fixture) exec(sql string, args ...any) matchstore.PublisherFunc {
	return func(ctx context.Context, _ matchstore.Pending) error {
		_, err := f.pool.Exec(ctx, sql, args...)
		return err
	}
}

func (f fixture) read(t *testing.T, account, id string) matchresult.Result {
	t.Helper()
	result, err := f.store.Read(context.Background(), account, id)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

const ranked = "0b9f6c1e-4d2a-4f61-9a57-3c8e2d7b5a10"

func TestOnlyAPlayerOfTheMatchReadsItsResult(t *testing.T) {
	f := newFixture(t)
	f.begin(t, ranked, matchstore.KindRanked)
	stranger := f.account(t)
	for _, c := range []struct{ account, id string }{{stranger, ranked}, {f.winner, "6d1c7a52-0e8b-4c43-b1f5-5f0a9b2c4e77"}, {f.winner, "not-a-match"}, {f.winner, ""}} {
		if _, err := f.store.Read(context.Background(), c.account, c.id); !errors.Is(err, matchresult.ErrNotPlayed) {
			t.Fatalf("%s reading %q: %v", c.account, c.id, err)
		}
	}
}

func TestAMatchThatHasNotEndedIsPendingInEveryPartItFeeds(t *testing.T) {
	f := newFixture(t)
	f.begin(t, ranked, matchstore.KindRanked)
	casual := "6d1c7a52-0e8b-4c43-b1f5-5f0a9b2c4e77"
	f.begin(t, casual, matchstore.KindCasual)
	got := f.read(t, f.winner, ranked)
	want := matchresult.Result{Match: ranked, Kind: matchstore.KindRanked, Pending: []string{"rating", "pass_xp", "missions", "coins"}, RetryAfterMs: 2000, Missions: []matchresult.Mission{}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unfinished ranked %+v, want %+v", got, want)
	}
	if got := f.read(t, f.winner, casual); !reflect.DeepEqual(got.Pending, []string{"pass_xp", "missions", "coins"}) || got.Result != "" {
		t.Fatalf("unfinished casual %+v", got)
	}
}

func TestEachPartArrivesWhenItsRelayFinishes(t *testing.T) {
	f := newFixture(t)
	f.finish(t, f.begin(t, ranked, matchstore.KindRanked))
	got := f.read(t, f.winner, ranked)
	if got.Result != matchstore.ResultWin || !reflect.DeepEqual(got.Pending, []string{"rating", "pass_xp", "missions", "coins"}) || got.RetryAfterMs != 2000 || got.Coins != (matchresult.Coins{}) || got.Rating != nil {
		t.Fatalf("before the relays %+v", got)
	}

	rewards := matchreward.NewStore(f.pool, matchreward.NewWatcher(f.pool, slog.New(slog.NewTextHandler(io.Discard, nil))), wallet.NewStore(f.pool))
	f.relay(t, matchstore.TopicEconomy, rewards)
	got = f.read(t, f.winner, ranked)
	if got.Coins != (matchresult.Coins{Match: 30, FirstWin: 100}) || !reflect.DeepEqual(got.Pending, []string{"rating", "pass_xp", "missions"}) {
		t.Fatalf("after the economy relay %+v", got)
	}
	if lost := f.read(t, f.loser, ranked); lost.Result != matchstore.ResultLoss || lost.Coins != (matchresult.Coins{Match: 10}) {
		t.Fatalf("the loser reads %+v", lost)
	}

	f.relay(t, matchstore.TopicXp, f.exec(`insert into pass_xp_events (account_id, season_id, amount, source_type, source_ref, idempotency_key, match_xp, economy_day, created_at)
		values ($1, 'season.test', 40, 'match', $2, $3, true, $4, $5)`, f.winner, ranked, "match:"+ranked, start, end))
	f.relay(t, matchstore.TopicMissions, f.exec(`with assigned as (
			insert into mission_assignments (account_id, mission_id, period, mission_type, slot, progress, assigned_at, completed_at) values
			($1, 'mission.win_3', '2026-11-03', 'daily', 1, 3, $2, $3),
			($1, 'mission.chain_1', 'once', 'achievement', 0, 1, $2, $3),
			($1, 'mission.play_10', '2026-W45', 'weekly', 1, 4, $2, null),
			($1, 'mission.win_1', '2026-11-03', 'daily', 2, 1, $2, $2))
		insert into reward_grants (account_id, bundle_id, revision, source_type, source_ref, idempotency_key, config_version, report, created_at) values
			($1, 'bundle.daily_mission_slot_1', 1, 'mission', 'mission.win_3', 'mission:mission.win_3:2026-11-03', 1, '{"coins": 50, "pass_xp": 20, "medals": []}', $3),
			($1, 'bundle.chain_1', 1, 'mission', 'mission.chain_1', 'mission:mission.chain_1:once', 1, '{"coins": 0, "pass_xp": 0, "medals": ["medal.chain_1"]}', $3)`, f.winner, start, end))
	f.relay(t, matchstore.TopicRating, f.exec(`with applied as (insert into rating_results (match_id, mode, config_version, applied_at) values ($1, '1v1', 1, $4))
		insert into rating_changes (match_id, account_id, won, offender, placement, k, expected, mmr_before, mmr_after, lp_delta, rank_before, lp_before, rank_after, lp_after) values
			($1, $2, true, false, false, 32, 0.5, 1500, 1516, 21, 13, 90, 14, 11),
			($1, $3, false, false, true, 40, 0.5, 1500, 1484, 0, 0, 0, 0, 0)`, ranked, f.winner, f.loser, end))

	got = f.read(t, f.winner, ranked)
	want := matchresult.Result{
		Match: ranked, Kind: matchstore.KindRanked, Result: matchstore.ResultWin, Pending: []string{},
		Coins:  matchresult.Coins{Match: 30, FirstWin: 100},
		PassXP: 40,
		Missions: []matchresult.Mission{
			{Mission: "mission.chain_1", Period: "once", Medals: []string{"medal.chain_1"}},
			{Mission: "mission.win_3", Period: "2026-11-03", Coins: 50, PassXP: 20, Medals: []string{}},
		},
		Rating: &matchresult.RatingChange{LPDelta: 21, RankBefore: "Gold III", LPBefore: 90, RankAfter: "Gold II", LPAfter: 11},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("after every relay\n got %+v %+v\nwant %+v %+v", got, got.Rating, want, want.Rating)
	}
	if lost := f.read(t, f.loser, ranked); lost.PassXP != 0 || len(lost.Missions) != 0 || !reflect.DeepEqual(lost.Rating, &matchresult.RatingChange{Placement: true}) {
		t.Fatalf("the loser in placement reads %+v %+v", lost, lost.Rating)
	}
}

func TestARevokedMissionGrantShowsTheMissionWithoutItsReward(t *testing.T) {
	f := newFixture(t)
	f.finish(t, f.begin(t, ranked, matchstore.KindCasual))
	if _, err := f.pool.Exec(context.Background(), `insert into mission_assignments (account_id, mission_id, period, mission_type, slot, progress, assigned_at, completed_at)
		values ($1, 'mission.win_3', '2026-11-03', 'daily', 1, 3, $2, $3)`, f.winner, start, end); err != nil {
		t.Fatal(err)
	}
	if _, err := f.pool.Exec(context.Background(), `insert into reward_grants (account_id, bundle_id, revision, source_type, source_ref, idempotency_key, config_version, report, created_at, revoked_at, revoke_key)
		values ($1, 'bundle.daily_mission_slot_1', 1, 'mission', 'mission.win_3', 'mission:mission.win_3:2026-11-03', 1, '{"coins": 50, "pass_xp": 20, "medals": []}', $2, $2, 'revoke:1')`, f.winner, end); err != nil {
		t.Fatal(err)
	}
	got := f.read(t, f.winner, ranked)
	if !reflect.DeepEqual(got.Missions, []matchresult.Mission{{Mission: "mission.win_3", Period: "2026-11-03", Medals: []string{}}}) || got.Rating != nil {
		t.Fatalf("revoked grant %+v", got)
	}
}
