package matchresult_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/matchresult"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

type fakeVerifier struct{}

func (fakeVerifier) Verify(access string) (token.Claims, error) {
	switch {
	case access == "expired":
		return token.Claims{}, token.ErrExpired
	case strings.HasPrefix(access, "access:"):
		return token.Claims{Account: strings.TrimPrefix(access, "access:")}, nil
	}
	return token.Claims{}, token.ErrMalformed
}

type fakeReader struct {
	result  matchresult.Result
	err     error
	account string
	match   string
}

func (r *fakeReader) Read(_ context.Context, account, matchID string) (matchresult.Result, error) {
	r.account, r.match = account, matchID
	return r.result, r.err
}

func call(t *testing.T, reader matchresult.Reader, path, access string) (int, string) {
	t.Helper()
	handler := matchresult.NewHandler(fakeVerifier{}, reader, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if access != "" {
		request.Header.Set("Authorization", "Bearer "+access)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code, strings.TrimSpace(recorder.Body.String())
}

func TestTheResultRouteAnswersTheCallersOwnSeat(t *testing.T) {
	reader := &fakeReader{result: matchresult.Result{
		Match: ranked, Kind: "ranked", Result: "win", Pending: []string{"rating"}, RetryAfterMs: 2000,
		Coins: matchresult.Coins{Match: 30, FirstWin: 100}, PassXP: 40,
		Missions: []matchresult.Mission{{Mission: "mission.win_3", Period: "2026-11-03", Coins: 50, PassXP: 20, Medals: []string{}}},
	}}
	status, body := call(t, reader, "/matches/"+ranked+"/result", "access:acc-1")
	want := `{"match":"` + ranked + `","kind":"ranked","result":"win","pending":["rating"],"retry_after_ms":2000,"coins":{"match":30,"first_win":100},"pass_xp":40,"missions":[{"mission":"mission.win_3","period":"2026-11-03","coins":50,"pass_xp":20,"medals":[]}],"rating":null}`
	if status != http.StatusOK || body != want || reader.account != "acc-1" || reader.match != ranked {
		t.Fatalf("result %d %s (asked %s for %s)", status, body, reader.account, reader.match)
	}
	reader.result = matchresult.Result{Match: ranked, Kind: "ranked", Result: "loss", Pending: []string{}, Missions: []matchresult.Mission{}, Rating: &matchresult.RatingChange{LPDelta: -18, RankBefore: "Gold II", LPBefore: 10, RankAfter: "Gold III", LPAfter: 92}}
	_, body = call(t, reader, "/matches/"+ranked+"/result", "access:acc-1")
	if want := `"pending":[],"coins":{"match":0,"first_win":0},"pass_xp":0,"missions":[],"rating":{"placement":false,"lp_delta":-18,"rank_before":"Gold II","lp_before":10,"rank_after":"Gold III","lp_after":92}}`; !strings.HasSuffix(body, want) || strings.Contains(body, "retry_after_ms") {
		t.Fatalf("final result %s", body)
	}
}

func TestTheResultRouteRefusesStrangersAndBadTokens(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		access string
		status int
		body   string
	}{
		{"not played", matchresult.ErrNotPlayed, "access:acc-1", http.StatusNotFound, `{"error":"match.not_played"}`},
		{"store failure", errors.New("db down"), "access:acc-1", http.StatusInternalServerError, `{"error":"internal"}`},
		{"expired", nil, "expired", http.StatusUnauthorized, `{"error":"access_expired"}`},
		{"no token", nil, "", http.StatusUnauthorized, `{"error":"access_rejected"}`},
	}
	for _, c := range cases {
		status, body := call(t, &fakeReader{err: c.err}, "/matches/"+ranked+"/result", c.access)
		if status != c.status || body != c.body {
			t.Errorf("%s: %d %s", c.name, status, body)
		}
	}
	if status, _ := call(t, &fakeReader{}, "/matches/"+ranked, "access:acc-1"); status != http.StatusNotFound {
		t.Fatalf("a path without /result answered %d", status)
	}
}
