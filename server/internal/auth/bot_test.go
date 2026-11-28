package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/bottoken"
)

var (
	botSecret = bytes.Repeat([]byte{3}, bottoken.MinSecretBytes)
	botNow    = time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
)

type fakeBots struct {
	known  map[string]bool
	humans map[string]bool
}

func (b *fakeBots) SignInBot(_ context.Context, bot string) (account.Account, bool, error) {
	if b.humans[bot] {
		return account.Account{}, false, account.ErrNotBot
	}
	created := !b.known[bot]
	b.known[bot] = true
	return account.Account{ID: "bot-" + bot, State: account.StateActive, Bot: true}, created, nil
}

func botFixture(t *testing.T) (*fixture, *fakeBots) {
	t.Helper()
	f := newFixture(t)
	bots := &fakeBots{known: map[string]bool{}, humans: map[string]bool{}}
	f.handler.EnableBotSignIn(botSecret, bots, func() time.Time { return botNow })
	return f, bots
}

func mintBot(t *testing.T, secret []byte, bot string, at time.Time) string {
	t.Helper()
	minted, err := bottoken.Mint(secret, bot, at)
	if err != nil {
		t.Fatal(err)
	}
	return minted
}

func botBody(token, device string) string {
	body := `{"provider":"bot","id_token":"` + token + `"`
	if device != "" {
		body += `,"device_id":"` + device + `"`
	}
	return body + "}"
}

func TestBotSignInIsAnUnknownProviderUntilEnabled(t *testing.T) {
	f := newFixture(t)
	status, body := f.post("/auth/signin", botBody(mintBot(t, botSecret, "staging-bot-1", botNow), "loadbot-1"))
	if status != http.StatusBadRequest || body["error"] != "unknown_provider" || len(f.tokens.issued) != 0 {
		t.Fatalf("bot sign-in without the secret answered %d %v, issued %v", status, body, f.tokens.issued)
	}
}

func TestBotSignInIssuesOrdinaryTokensForTheBotAccount(t *testing.T) {
	f, _ := botFixture(t)
	minted := mintBot(t, botSecret, "staging-bot-1", botNow)
	status, body := f.post("/auth/signin", botBody(minted, "loadbot-1"))
	if status != http.StatusOK || body["account_id"] != "bot-staging-bot-1" || body["access_token"] != "access:bot-staging-bot-1:loadbot-1" || body["created"] != true {
		t.Fatalf("first bot sign-in answered %d %v", status, body)
	}
	status, body = f.post("/auth/signin", botBody(minted, "loadbot-2"))
	if status != http.StatusOK || body["created"] != nil {
		t.Fatalf("returning bot sign-in answered %d %v", status, body)
	}
	if f.accounts.calls != 0 || len(f.verifier.seen) != 0 {
		t.Fatalf("bot sign-in reached the player path: %d accounts calls, verifier saw %v", f.accounts.calls, f.verifier.seen)
	}
	if want := "bot-staging-bot-1@loadbot-1 bot-staging-bot-1@loadbot-2"; strings.Join(f.tokens.issued, " ") != want {
		t.Fatalf("tokens issued %v", f.tokens.issued)
	}
	var out bytes.Buffer
	f.handler.SignIns().WriteTo(&out)
	if !strings.Contains(out.String(), `hd_signins_total{provider="bot",result="ok"} 2`) {
		t.Fatalf("metrics %s", out.String())
	}
}

func TestBotSignInRefusesForeignExpiredAndHumanTokens(t *testing.T) {
	f, bots := botFixture(t)
	bots.humans["taken-by-a-player"] = true
	other := bytes.Repeat([]byte{4}, bottoken.MinSecretBytes)
	cases := []struct {
		name   string
		body   string
		status int
		reason string
	}{
		{"no device", botBody(mintBot(t, botSecret, "staging-bot-1", botNow), ""), http.StatusBadRequest, "device_required"},
		{"other secret", botBody(mintBot(t, other, "staging-bot-1", botNow), "d"), http.StatusUnauthorized, "token_rejected"},
		{"expired", botBody(mintBot(t, botSecret, "staging-bot-1", botNow.Add(-2*bottoken.Lifetime)), "d"), http.StatusUnauthorized, "token_expired"},
		{"an access token", botBody("access:bot-staging-bot-1:d", "d"), http.StatusUnauthorized, "token_rejected"},
		{"not a bot account", botBody(mintBot(t, botSecret, "taken-by-a-player", botNow), "d"), http.StatusForbidden, "not_bot"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if status, body := f.post("/auth/signin", c.body); status != c.status || body["error"] != c.reason {
				t.Fatalf("answered %d %v, want %d %s", status, body, c.status, c.reason)
			}
		})
	}
	if len(f.tokens.issued) != 0 {
		t.Fatalf("tokens issued %v", f.tokens.issued)
	}
}
