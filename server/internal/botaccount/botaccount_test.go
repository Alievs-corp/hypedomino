package botaccount_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/botaccount"
	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

type accessTokens struct{}

func (accessTokens) Verify(access string) (token.Claims, error) {
	account, ok := strings.CutPrefix(access, "access:")
	if !ok {
		return token.Claims{}, token.ErrSignature
	}
	return token.Claims{Account: account}, nil
}

func TestIsReadsTheBotFlagOfTheAccount(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	ctx := context.Background()
	var player, bot string
	if err := pool.QueryRow(ctx, `insert into accounts default values returning id`).Scan(&player); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `insert into accounts (is_bot) values (true) returning id`).Scan(&bot); err != nil {
		t.Fatal(err)
	}
	for account, want := range map[string]bool{player: false, bot: true, "00000000-0000-0000-0000-000000000000": false} {
		if got, err := botaccount.Is(ctx, pool, account); err != nil || got != want {
			t.Fatalf("account %s: bot %v, %v", account, got, err)
		}
	}
}

func TestCasualOnlyRefusesABotBeforeTheEconomyRouteRuns(t *testing.T) {
	lookupFailed := errors.New("database down")
	isBot := func(_ context.Context, account string) (bool, error) {
		if account == "broken" {
			return false, lookupFailed
		}
		return account == "staging-bot", nil
	}
	served := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served++
		w.WriteHeader(http.StatusOK)
	})
	gate := botaccount.CasualOnly(accessTokens{}, isBot, slog.New(slog.NewTextHandler(io.Discard, nil)), next)
	cases := []struct {
		name   string
		bearer string
		status int
		reason string
		served bool
	}{
		{"bot", "access:staging-bot", http.StatusForbidden, botaccount.RefusalCode, false},
		{"player", "access:kamal", http.StatusOK, "", true},
		{"no token", "", http.StatusOK, "", true},
		{"lookup failed", "access:broken", http.StatusInternalServerError, "internal", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			served = 0
			request := httptest.NewRequest(http.MethodPost, "/shop/buy", strings.NewReader(`{}`))
			if c.bearer != "" {
				request.Header.Set("Authorization", "Bearer "+c.bearer)
			}
			recorder := httptest.NewRecorder()
			gate.ServeHTTP(recorder, request)
			if recorder.Code != c.status || (served == 1) != c.served {
				t.Fatalf("answered %d, served %d", recorder.Code, served)
			}
			if c.reason == "" {
				return
			}
			var body map[string]string
			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil || body["error"] != c.reason {
				t.Fatalf("body %v, %v", body, err)
			}
		})
	}
}
