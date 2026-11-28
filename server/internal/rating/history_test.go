package rating_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type historyMailbox struct {
	rating.Mailbox
	history rating.SeasonHistory
	err     error
	account string
}

func (m *historyMailbox) SeasonHistory(_ context.Context, account string) (rating.SeasonHistory, error) {
	m.account = account
	return m.history, m.err
}

func callHistory(t *testing.T, mailbox rating.Mailbox, access string) (int, string) {
	t.Helper()
	handler := rating.NewHandler(fakeVerifier{}, mailbox, slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now })
	request := httptest.NewRequest(http.MethodGet, "/ranked/seasons/history", nil)
	request.Header.Set("Authorization", "Bearer "+access)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Body.String()
}

func TestTheSeasonHistoryRouteAnswersTheViewersOwnSeasons(t *testing.T) {
	mailbox := &historyMailbox{history: rating.SeasonHistory{Seasons: []rating.PastSeason{{
		SeasonClock: rating.SeasonClock{ID: "season.2026_08", NameKey: "season.2026_08.name", StartsAt: now.AddDate(0, -3, 0), EndsAt: now},
		Modes:       []rating.PastMode{{Mode: "1v1", Rank: "Gold II", Tier: "Gold", LP: 42, Position: 17, Reward: rating.PastReward{Granted: true, Coins: 700, BundleID: "bundle.rank_gold", Emblems: []string{}}}},
	}}}}
	status, body := callHistory(t, mailbox, "access:acc-1")
	if status != http.StatusOK || mailbox.account != "acc-1" {
		t.Fatalf("history %d %s %s", status, body, mailbox.account)
	}
	if strings.Contains(strings.ToLower(body), "mmr") {
		t.Fatalf("the history carries MMR: %s", body)
	}
	var reply struct {
		Seasons []struct {
			ID      string `json:"id"`
			NameKey string `json:"name_key"`
			EndsAt  string `json:"ends_at"`
			Modes   []struct {
				Mode     string `json:"mode"`
				Rank     string `json:"rank"`
				LP       int    `json:"lp"`
				Position int    `json:"position"`
				Reward   struct {
					Granted bool  `json:"granted"`
					Coins   int64 `json:"coins"`
				} `json:"reward"`
			} `json:"modes"`
		} `json:"seasons"`
	}
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatal(err)
	}
	if len(reply.Seasons) != 1 || reply.Seasons[0].ID != "season.2026_08" || reply.Seasons[0].NameKey != "season.2026_08.name" || reply.Seasons[0].EndsAt != now.Format(time.RFC3339) {
		t.Fatalf("seasons %+v", reply.Seasons)
	}
	mode := reply.Seasons[0].Modes[0]
	if mode.Mode != "1v1" || mode.Rank != "Gold II" || mode.LP != 42 || mode.Position != 17 || !mode.Reward.Granted || mode.Reward.Coins != 700 {
		t.Fatalf("mode %+v", mode)
	}
}

func TestTheSeasonHistoryRouteRefusesWithoutAccess(t *testing.T) {
	if status, body := callHistory(t, &historyMailbox{err: errors.New("down")}, "access:acc-1"); status != http.StatusInternalServerError || !strings.Contains(body, `"internal"`) {
		t.Fatalf("failure %d %s", status, body)
	}
	if status, body := callHistory(t, &historyMailbox{}, "expired"); status != http.StatusUnauthorized || !strings.Contains(body, `"access_expired"`) {
		t.Fatalf("expired %d %s", status, body)
	}
	if status, body := callHistory(t, &historyMailbox{}, "nope"); status != http.StatusUnauthorized || !strings.Contains(body, `"access_rejected"`) {
		t.Fatalf("rejected %d %s", status, body)
	}
}

func TestTheSeasonHistoryReadsTheFinalRanksAndRewardsNewestFirst(t *testing.T) {
	f := newRewardFixture(t)
	ctx := context.Background()
	end := now.Add(60 * 24 * time.Hour)
	challenger := f.apex(t, 600, 40, now)
	gold, newcomer := f.account(t), f.account(t)
	f.set(t, gold, rules.Mode1v1, 1500, 50, rank(t, "Gold II", 50))
	f.set(t, gold, rules.Mode2v2, 1400, 30, rank(t, "Silver I", 12))
	oldStart, oldEnd := now.Add(-120*24*time.Hour), now.Add(-30*24*time.Hour)
	if _, err := f.pool.Exec(ctx, `insert into ranked_seasons (id, name_key, starts_at, ends_at, closed_1v1_at, closed_2v2_at, rewarded_1v1_at, rewarded_2v2_at, created_at) values ('season.old', 'season.old.name', $1, $2, $2, $2, $2, $2, $1)`, oldStart, oldEnd); err != nil {
		t.Fatal(err)
	}
	if _, err := f.pool.Exec(ctx, `insert into season_results (season_id, mode, account_id, position, rank_index, lp, wins, mmr, reset_index) values ('season.old', '1v1', $1, 9, 9, 77, 20, 1350, 5)`, gold); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		if _, err := f.store.CloseSeason(ctx, "season.test", mode, end); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.store.DistributeRewards(ctx, "season.test", rules.Mode1v1, end); err != nil {
		t.Fatal(err)
	}

	history, err := f.store.SeasonHistory(ctx, gold)
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Seasons) != 2 || history.Seasons[0].ID != "season.test" || history.Seasons[1].ID != "season.old" {
		t.Fatalf("seasons %+v", history.Seasons)
	}
	latest := history.Seasons[0]
	if latest.NameKey != "season.test.name" || !latest.EndsAt.Equal(end) || latest.EndsAt.Location() != time.UTC || len(latest.Modes) != 2 {
		t.Fatalf("latest %+v", latest)
	}
	solo, teams := latest.Modes[0], latest.Modes[1]
	if solo.Mode != "1v1" || solo.Rank != "Gold II" || solo.Tier != "Gold" || solo.LP != 50 || solo.Position != 2 {
		t.Fatalf("solo %+v", solo)
	}
	if !solo.Reward.Granted || solo.Reward.Coins != 700 || solo.Reward.BundleID != "bundle.rank_season_1_gold" || strings.Join(solo.Reward.Emblems, " ") != "global:2" {
		t.Fatalf("solo reward %+v", solo.Reward)
	}
	if teams.Mode != "2v2" || teams.Rank != "Silver I" || teams.LP != 12 || teams.Position != 1 || teams.Reward.Granted || teams.Reward.Coins != 0 || teams.Reward.Emblems == nil {
		t.Fatalf("teams before its reward run %+v", teams)
	}
	old := history.Seasons[1]
	if len(old.Modes) != 1 || old.Modes[0].Rank != "Silver III" || old.Modes[0].LP != 77 || old.Modes[0].Position != 9 || old.Modes[0].Reward.Granted {
		t.Fatalf("old season %+v", old)
	}

	apex, err := f.store.SeasonHistory(ctx, challenger)
	if err != nil || len(apex.Seasons) != 1 || apex.Seasons[0].Modes[0].Rank != "Challenger" || apex.Seasons[0].Modes[0].Tier != "Challenger" || apex.Seasons[0].Modes[0].Position != 1 {
		t.Fatalf("apex history %+v %v", apex, err)
	}
	empty, err := f.store.SeasonHistory(ctx, newcomer)
	if err != nil || empty.Seasons == nil || len(empty.Seasons) != 0 {
		t.Fatalf("no final rank %+v %v", empty, err)
	}
}

func TestTheSeasonHistoryKeepsTheNewestSeasons(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	veteran := f.account(t)
	first := now.Add(-30*24*time.Hour).AddDate(0, -rating.SeasonHistoryLimit-1, 0)
	for i := 0; i <= rating.SeasonHistoryLimit; i++ {
		id := fmt.Sprintf("season.old_%02d", i)
		start, finish := first.AddDate(0, i, 0), first.AddDate(0, i+1, 0)
		if _, err := f.pool.Exec(ctx, `insert into ranked_seasons (id, name_key, starts_at, ends_at, closed_1v1_at, closed_2v2_at, created_at) values ($1, $1 || '.name', $2, $3, $3, $3, $2)`, id, start, finish); err != nil {
			t.Fatal(err)
		}
		if _, err := f.pool.Exec(ctx, `insert into season_results (season_id, mode, account_id, position, rank_index, lp, wins, mmr, reset_index) values ($1, '1v1', $2, 1, 4, $3, 1, 1200, 2)`, id, veteran, i); err != nil {
			t.Fatal(err)
		}
	}
	history, err := f.store.SeasonHistory(ctx, veteran)
	if err != nil || len(history.Seasons) != rating.SeasonHistoryLimit {
		t.Fatalf("history %d %v", len(history.Seasons), err)
	}
	if history.Seasons[0].ID != fmt.Sprintf("season.old_%02d", rating.SeasonHistoryLimit) || history.Seasons[len(history.Seasons)-1].ID != "season.old_01" {
		t.Fatalf("kept %s to %s", history.Seasons[0].ID, history.Seasons[len(history.Seasons)-1].ID)
	}
}
