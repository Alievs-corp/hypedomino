package matchstore_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/partnerfit"
	"github.com/Alievs-corp/hypedomino/server/internal/profile"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

var start = time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

func accounts(t testing.TB, pool *pgxpool.Pool, n int) []string {
	t.Helper()
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var id string
		if err := pool.QueryRow(context.Background(), "insert into accounts default values returning id").Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func started(id string, kind string, players []string) matchstore.Started {
	m := matchstore.Started{ID: id, StartedAt: start, Mode: rules.Mode2v2, Kind: kind, Region: "eu", ConfigVersion: 3, Seed: 42}
	for index, seat := range rules.Mode2v2.Seats() {
		p := matchstore.Participant{Seat: seat}
		if index < len(players) {
			p.Account = players[index]
		} else {
			p.Bot = true
		}
		m.Participants = append(m.Participants, p)
	}
	return m
}

func finished(id string, kind string) matchstore.Finished {
	payload := map[string]any{"matchId": id, "kind": kind}
	return matchstore.Finished{
		ID: id, StartedAt: start, EndedAt: start.Add(9 * time.Minute), HandsPlayed: 4,
		Scores: rules.TeamPoints{370, 215}, Winner: rules.TeamA, HasWinner: true,
		Replay: []byte{'H', 'D', 'R', 1},
		Results: map[rules.Seat]string{
			rules.SeatNorth: matchstore.ResultWin, rules.SeatSouth: matchstore.ResultWin,
			rules.SeatEast: matchstore.ResultLoss, rules.SeatWest: matchstore.ResultLoss,
		},
		Events: []matchstore.Event{
			{Seq: 0, Kind: "hand_end", Payload: map[string]any{"hand": 1, "scores": []int{95, 40}}, At: start.Add(2 * time.Minute)},
			{Seq: 1, Kind: "timeout", Seat: rules.SeatEast, HasSeat: true, Payload: map[string]any{"consecutive": 1}, At: start.Add(3 * time.Minute)},
			{Seq: 2, Kind: "match_end", Payload: map[string]any{"winner": "A"}},
		},
		Outbox: matchstore.MatchMessages(kind, "match:"+id, payload),
	}
}

func count(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestBeginWritesTheMatchAndItsParticipantsOnce(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 3)
	id := uuid.NewString()
	if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
		t.Fatal(err)
	}
	if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
		t.Fatalf("second begin: %v", err)
	}
	if n := count(t, pool, "select count(*) from matches where id = $1", id); n != 1 {
		t.Fatalf("%d match rows", n)
	}
	if n := count(t, pool, "select count(*) from match_participants where match_id = $1", id); n != 4 {
		t.Fatalf("%d participants", n)
	}
	if n := count(t, pool, "select count(*) from match_participants where match_id = $1 and bot and account_id is null", id); n != 1 {
		t.Fatalf("%d bot seats", n)
	}
	bad := started(uuid.NewString(), matchstore.KindRanked, players)
	bad.Participants = bad.Participants[:2]
	if err := store.Begin(ctx, bad); !errors.Is(err, matchstore.ErrInvalid) {
		t.Fatalf("a 2v2 with two participants accepted: %v", err)
	}
}

func TestBeginWritesEveryParticipantColumnInOneStatement(t *testing.T) {
	dsn, pool := dbtest.Fresh(t)
	ctx := context.Background()
	statements := &statementCounter{}
	store := matchstore.New(tracedPool(t, dsn, statements))
	players := accounts(t, pool, 3)
	id := uuid.NewString()
	m := started(id, matchstore.KindRanked, players)
	m.Participants[0].Loadout = profile.Loadout{Avatar: "avatar.fox", Tile: "tile.ivory", Emojis: []string{"emoji.laugh"}}
	if err := store.Begin(ctx, m); err != nil {
		t.Fatal(err)
	}
	if n := statements.matching("insert into match_participants"); n != 1 {
		t.Fatalf("%d participant insert statements for %d seats, want 1", n, len(m.Participants))
	}
	rows, err := pool.Query(ctx, "select seat, team, account_id::text, bot, loadout from match_participants where match_id = $1 order by seat", id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var seat, team int16
		var account *string
		var bot bool
		var loadout profile.Loadout
		if err := rows.Scan(&seat, &team, &account, &bot, &loadout); err != nil {
			t.Fatal(err)
		}
		want := m.Participants[seen]
		if rules.Seat(seat) != want.Seat || rules.Team(team) != want.Seat.Team() || bot != want.Bot {
			t.Fatalf("row %d: seat %d team %d bot %v, want %+v", seen, seat, team, bot, want)
		}
		if want.Bot != (account == nil) || (account != nil && *account != want.Account) {
			t.Fatalf("row %d: account %v, want %q", seen, account, want.Account)
		}
		if loadout.Avatar != want.Loadout.Avatar || loadout.Tile != want.Loadout.Tile || len(loadout.Emojis) != len(want.Loadout.Emojis) {
			t.Fatalf("row %d: loadout %+v, want %+v", seen, loadout, want.Loadout)
		}
		seen++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if seen != len(m.Participants) {
		t.Fatalf("%d participant rows, want %d", seen, len(m.Participants))
	}
}

type statementCounter struct {
	mu  sync.Mutex
	sql []string
}

func (c *statementCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sql = append(c.sql, data.SQL)
	return ctx
}

func (c *statementCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (c *statementCounter) matching(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, sql := range c.sql {
		if strings.HasPrefix(sql, prefix) {
			n++
		}
	}
	return n
}

func tracedPool(t testing.TB, dsn string, tracer pgx.QueryTracer) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestFinishIsAtomicAndAppliedOnce(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	id := uuid.NewString()
	if err := store.Finish(ctx, finished(id, matchstore.KindRanked)); !errors.Is(err, matchstore.ErrUnknownMatch) {
		t.Fatalf("finish before begin: %v", err)
	}
	if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
		t.Fatal(err)
	}
	broken := finished(id, matchstore.KindRanked)
	broken.Results[rules.SeatX] = matchstore.ResultWin
	if err := store.Finish(ctx, broken); !errors.Is(err, matchstore.ErrInvalid) {
		t.Fatalf("a result for a seat outside the match: %v", err)
	}
	if n := count(t, pool, "select count(*) from matches where id = $1 and ended_at is not null", id); n != 0 {
		t.Fatal("a failed finish left the match ended")
	}
	if n := count(t, pool, "select count(*) from outbox"); n != 0 {
		t.Fatalf("a failed finish left %d outbox rows", n)
	}
	if err := store.Finish(ctx, finished(id, matchstore.KindRanked)); err != nil {
		t.Fatal(err)
	}
	if err := store.Finish(ctx, finished(id, matchstore.KindRanked)); !errors.Is(err, matchstore.ErrAlreadyFinished) {
		t.Fatalf("second finish: %v", err)
	}
	var handsPlayed, scoreA, scoreB int
	var winner *int16
	var replay []byte
	if err := pool.QueryRow(ctx, "select hands_played, score_a, score_b, winner_team, replay from matches where id = $1", id).Scan(&handsPlayed, &scoreA, &scoreB, &winner, &replay); err != nil {
		t.Fatal(err)
	}
	if handsPlayed != 4 || scoreA != 370 || scoreB != 215 || winner == nil || *winner != 0 || string(replay) != "HDR\x01" {
		t.Fatalf("stored match %d %d %d %v %q", handsPlayed, scoreA, scoreB, winner, replay)
	}
	if n := count(t, pool, "select count(*) from match_participants where match_id = $1 and result = 'win'", id); n != 2 {
		t.Fatalf("%d winners", n)
	}
	if n := count(t, pool, "select count(*) from match_events where match_id = $1", id); n != 3 {
		t.Fatalf("%d events", n)
	}
	if n := count(t, pool, "select count(*) from match_events where match_id = $1 and seat = 1 and kind = 'timeout'", id); n != 1 {
		t.Fatal("the seat event lost its seat")
	}
	if n := count(t, pool, "select count(*) from match_events where match_id = $1 and seq = 2 and at = $2", id, start.Add(9*time.Minute)); n != 1 {
		t.Fatal("an event without a time does not take the end time")
	}
	if n := count(t, pool, "select count(*) from outbox where key = $1", "match:"+id); n != 5 {
		t.Fatalf("%d outbox rows for a ranked match", n)
	}
	if n := count(t, pool, "select count(*) from outbox where key = $1 and topic = 'rating'", "match:"+id); n != 1 {
		t.Fatal("a ranked match emits a rating message")
	}
}

func TestMatchMessagesFollowTheKind(t *testing.T) {
	topics := func(kind string) []string {
		var out []string
		for _, m := range matchstore.MatchMessages(kind, "k", nil) {
			out = append(out, m.Topic)
		}
		return out
	}
	cases := map[string][]string{
		matchstore.KindRanked: {"rating", "xp", "missions", "economy", "analytics"},
		matchstore.KindCasual: {"xp", "missions", "economy", "analytics"},
		matchstore.KindCustom: {"missions", "analytics"},
	}
	for kind, want := range cases {
		got := topics(kind)
		if len(got) != len(want) {
			t.Fatalf("%s: %v", kind, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: %v, want %v", kind, got, want)
			}
		}
	}
}

func TestRelayPublishesPendingMessagesAtLeastOnce(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	first, second := uuid.NewString(), uuid.NewString()
	for _, id := range []string{first, second} {
		if err := store.Begin(ctx, started(id, matchstore.KindCasual, players)); err != nil {
			t.Fatal(err)
		}
		if err := store.Finish(ctx, finished(id, matchstore.KindCasual)); err != nil {
			t.Fatal(err)
		}
	}
	var seen []string
	fail := true
	publisher := matchstore.PublisherFunc(func(ctx context.Context, m matchstore.Pending) error {
		if fail {
			return errors.New("consumer down")
		}
		seen = append(seen, m.Topic+":"+m.Key)
		return nil
	})
	now := start.Add(time.Hour)
	relay := matchstore.NewRelay(store, publisher, matchstore.TopicXp, 10, func() time.Time { return now })
	if n, err := relay.Once(ctx); err == nil || n != 0 {
		t.Fatalf("a failing publisher published %d, %v", n, err)
	}
	if n := count(t, pool, "select count(*) from outbox where attempts = 1 and published_at is null"); n != 1 {
		t.Fatalf("%d rows marked as attempted", n)
	}
	fail = false
	if n, err := relay.Once(ctx); err != nil || n != 2 {
		t.Fatalf("published %d, %v", n, err)
	}
	if len(seen) != 2 || seen[0] != "xp:match:"+first || seen[1] != "xp:match:"+second {
		t.Fatalf("seen %v", seen)
	}
	if n, err := relay.Once(ctx); err != nil || n != 0 {
		t.Fatalf("published again %d, %v", n, err)
	}
	if n := count(t, pool, "select count(*) from outbox where topic = 'xp' and published_at = $1", now); n != 2 {
		t.Fatalf("%d xp rows published", n)
	}
	pending, err := store.Pending(ctx, "", 100)
	if err != nil || len(pending) != 6 {
		t.Fatalf("other topics pending %d, %v", len(pending), err)
	}
	all := matchstore.NewRelay(store, publisher, "", 0, nil)
	if n, err := all.Once(ctx); err != nil || n != 6 {
		t.Fatalf("all topics published %d, %v", n, err)
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		all.Run(runCtx, time.Millisecond, nil)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop")
	}
}

func TestEventsArePartitionedWithTheMatch(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	id := uuid.NewString()
	m := started(id, matchstore.KindRanked, players)
	m.StartedAt = time.Now().UTC()
	if err := store.Begin(ctx, m); err != nil {
		t.Fatal(err)
	}
	f := finished(id, matchstore.KindRanked)
	f.StartedAt, f.EndedAt = m.StartedAt, m.StartedAt.Add(time.Minute)
	if err := store.Finish(ctx, f); err != nil {
		t.Fatal(err)
	}
	var partition string
	if err := pool.QueryRow(ctx, "select tableoid::regclass::text from match_events where match_id = $1 limit 1", id).Scan(&partition); err != nil {
		t.Fatal(err)
	}
	if partition != "match_events_"+m.StartedAt.Format("200601") {
		t.Fatalf("events landed in %s", partition)
	}
}

func TestFinishKeepsThePartnerFitOfEverySeat(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	id := uuid.NewString()
	if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
		t.Fatal(err)
	}
	impossible := finished(id, matchstore.KindRanked)
	impossible.PartnerFit = map[rules.Seat]partnerfit.Count{rules.SeatNorth: {Decisions: 3, Fits: 4}}
	if err := store.Finish(ctx, impossible); err == nil {
		t.Fatal("more fits than decisions were stored")
	}
	outside := finished(id, matchstore.KindRanked)
	outside.PartnerFit = map[rules.Seat]partnerfit.Count{rules.SeatX: {Decisions: 1}}
	if err := store.Finish(ctx, outside); !errors.Is(err, matchstore.ErrInvalid) {
		t.Fatalf("a fit for a seat outside the match: %v", err)
	}
	done := finished(id, matchstore.KindRanked)
	done.PartnerFit = map[rules.Seat]partnerfit.Count{rules.SeatNorth: {Decisions: 12, Fits: 11}, rules.SeatEast: {Decisions: 9, Fits: 4}}
	if err := store.Finish(ctx, done); err != nil {
		t.Fatal(err)
	}
	if n := count(t, pool, "select count(*) from match_participants where match_id = $1 and seat = $2 and partner_decisions = 12 and partner_fits = 11", id, int16(rules.SeatNorth)); n != 1 {
		t.Fatal("north fit not stored")
	}
	if n := count(t, pool, "select count(*) from match_participants where match_id = $1 and partner_decisions is null", id); n != 2 {
		t.Fatalf("%d seats without a measurement, want 2", n)
	}
}

func TestRunningSeesAnUnfinishedMatchUntilItEndsOrGoesStale(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	store := matchstore.New(pool)
	players := accounts(t, pool, 3)
	id := uuid.NewString()
	if err := store.Begin(ctx, started(id, matchstore.KindRanked, players[:2])); err != nil {
		t.Fatal(err)
	}
	running := func(account string, startedAfter time.Time) bool {
		t.Helper()
		got, err := store.Running(ctx, account, startedAfter)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	if !running(players[0], start.Add(-2*time.Hour)) {
		t.Fatal("a seated player of an unfinished match is not running")
	}
	if running(players[2], start.Add(-2*time.Hour)) {
		t.Fatal("a player outside the match is running")
	}
	if running(players[0], start.Add(time.Minute)) {
		t.Fatal("a match started before the stale bound still counts")
	}
	if err := store.Finish(ctx, finished(id, matchstore.KindRanked)); err != nil {
		t.Fatal(err)
	}
	if running(players[0], start.Add(-2*time.Hour)) {
		t.Fatal("a finished match still counts")
	}
}

func BenchmarkBeginA2v2Match(b *testing.B) {
	_, pool := dbtest.Fresh(b)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(b, pool, 3)
	b.ReportAllocs()
	for b.Loop() {
		if err := store.Begin(ctx, started(uuid.NewString(), matchstore.KindRanked, players)); err != nil {
			b.Fatal(err)
		}
	}
}

func TestFinishWritesEverySeatOutcomeInOneStatement(t *testing.T) {
	fit := func(decisions, fits int) partnerfit.Count { return partnerfit.Count{Decisions: decisions, Fits: fits} }
	cases := []struct {
		name       string
		results    map[rules.Seat]string
		fits       map[rules.Seat]partnerfit.Count
		statements int
		want       map[rules.Seat]seatOutcome
	}{
		{
			name:       "results and partner fit of every seat",
			results:    map[rules.Seat]string{rules.SeatNorth: matchstore.ResultWin, rules.SeatSouth: matchstore.ResultWin, rules.SeatEast: matchstore.ResultLoss, rules.SeatWest: matchstore.ResultLoss},
			fits:       map[rules.Seat]partnerfit.Count{rules.SeatNorth: fit(12, 11), rules.SeatEast: fit(9, 4), rules.SeatSouth: fit(0, 0), rules.SeatWest: fit(3, 3)},
			statements: 1,
			want: map[rules.Seat]seatOutcome{
				rules.SeatNorth: {result: "win", decisions: 12, fits: 11, measured: true},
				rules.SeatEast:  {result: "loss", decisions: 9, fits: 4, measured: true},
				rules.SeatSouth: {result: "win", measured: true},
				rules.SeatWest:  {result: "loss", decisions: 3, fits: 3, measured: true},
			},
		},
		{
			name:       "partner fit of some seats keeps the others unmeasured",
			results:    map[rules.Seat]string{rules.SeatNorth: matchstore.ResultCancelled, rules.SeatSouth: matchstore.ResultCancelled, rules.SeatEast: matchstore.ResultCancelled, rules.SeatWest: matchstore.ResultCancelled},
			fits:       map[rules.Seat]partnerfit.Count{rules.SeatWest: fit(5, 2)},
			statements: 1,
			want: map[rules.Seat]seatOutcome{
				rules.SeatNorth: {result: "cancelled"},
				rules.SeatEast:  {result: "cancelled"},
				rules.SeatSouth: {result: "cancelled"},
				rules.SeatWest:  {result: "cancelled", decisions: 5, fits: 2, measured: true},
			},
		},
		{
			name:       "partner fit without a result",
			fits:       map[rules.Seat]partnerfit.Count{rules.SeatEast: fit(1, 1)},
			statements: 1,
			want: map[rules.Seat]seatOutcome{
				rules.SeatNorth: {},
				rules.SeatEast:  {decisions: 1, fits: 1, measured: true},
				rules.SeatSouth: {},
				rules.SeatWest:  {},
			},
		},
		{
			name:       "no seat outcome sends no statement",
			statements: 0,
			want:       map[rules.Seat]seatOutcome{rules.SeatNorth: {}, rules.SeatEast: {}, rules.SeatSouth: {}, rules.SeatWest: {}},
		},
	}
	dsn, pool := dbtest.Fresh(t)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statements := &statementCounter{}
			store := matchstore.New(tracedPool(t, dsn, statements))
			id := uuid.NewString()
			if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
				t.Fatal(err)
			}
			f := finished(id, matchstore.KindRanked)
			f.Results, f.PartnerFit = tc.results, tc.fits
			if err := store.Finish(ctx, f); err != nil {
				t.Fatal(err)
			}
			if n := statements.matching("update match_participants"); n != tc.statements {
				t.Fatalf("%d participant update statements, want %d", n, tc.statements)
			}
			got := seatOutcomes(t, pool, id)
			if len(got) != len(tc.want) {
				t.Fatalf("%d seats, want %d", len(got), len(tc.want))
			}
			for seat, want := range tc.want {
				if got[seat] != want {
					t.Errorf("seat %s: %+v, want %+v", seat, got[seat], want)
				}
			}
		})
	}
}

func TestFinishRefusesSeatOutcomesOutsideTheMatch(t *testing.T) {
	cases := []struct {
		name    string
		results map[rules.Seat]string
		fits    map[rules.Seat]partnerfit.Count
	}{
		{name: "a result outside the match", results: map[rules.Seat]string{rules.SeatNorth: matchstore.ResultWin, rules.SeatX: matchstore.ResultWin}},
		{name: "a partner fit outside the match", fits: map[rules.Seat]partnerfit.Count{rules.SeatY: {Decisions: 1}}},
		{name: "a fit outside beside a result inside", results: map[rules.Seat]string{rules.SeatEast: matchstore.ResultLoss}, fits: map[rules.Seat]partnerfit.Count{rules.SeatX: {Decisions: 2, Fits: 1}}},
	}
	_, pool := dbtest.Fresh(t)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(t, pool, 4)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.NewString()
			if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
				t.Fatal(err)
			}
			f := finished(id, matchstore.KindRanked)
			f.Results, f.PartnerFit = tc.results, tc.fits
			if err := store.Finish(ctx, f); !errors.Is(err, matchstore.ErrInvalid) {
				t.Fatalf("finish: %v, want ErrInvalid", err)
			}
			if n := count(t, pool, "select count(*) from match_participants where match_id = $1 and (result is not null or partner_decisions is not null)", id); n != 0 {
				t.Fatalf("a refused finish left %d seat outcomes", n)
			}
		})
	}
}

type seatOutcome struct {
	result    string
	decisions int
	fits      int
	measured  bool
}

func seatOutcomes(t *testing.T, pool *pgxpool.Pool, id string) map[rules.Seat]seatOutcome {
	t.Helper()
	rows, err := pool.Query(context.Background(), "select seat, coalesce(result, ''), partner_decisions, partner_fits from match_participants where match_id = $1", id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[rules.Seat]seatOutcome{}
	for rows.Next() {
		var seat int16
		var result string
		var decisions, fits *int32
		if err := rows.Scan(&seat, &result, &decisions, &fits); err != nil {
			t.Fatal(err)
		}
		o := seatOutcome{result: result, measured: decisions != nil}
		if decisions != nil {
			o.decisions = int(*decisions)
		}
		if fits != nil {
			o.fits = int(*fits)
		}
		out[rules.Seat(seat)] = o
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func BenchmarkFinishA2v2Match(b *testing.B) {
	_, pool := dbtest.Fresh(b)
	store := matchstore.New(pool)
	ctx := context.Background()
	players := accounts(b, pool, 4)
	fits := map[rules.Seat]partnerfit.Count{rules.SeatNorth: {Decisions: 12, Fits: 11}, rules.SeatSouth: {Decisions: 10, Fits: 7}, rules.SeatEast: {Decisions: 9, Fits: 4}, rules.SeatWest: {Decisions: 11, Fits: 9}}
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		id := uuid.NewString()
		if err := store.Begin(ctx, started(id, matchstore.KindRanked, players)); err != nil {
			b.Fatal(err)
		}
		f := finished(id, matchstore.KindRanked)
		f.PartnerFit = fits
		b.StartTimer()
		if err := store.Finish(ctx, f); err != nil {
			b.Fatal(err)
		}
	}
}
