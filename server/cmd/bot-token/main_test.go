package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	"github.com/Alievs-corp/hypedomino/server/internal/bottoken"
)

var (
	secret = bytes.Repeat([]byte{6}, bottoken.MinSecretBytes)
	now    = time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
)

func env(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func TestTokensAreMintedForDevAndStagingOnly(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(secret)
	for _, environment := range []string{"", "production"} {
		var out bytes.Buffer
		err := run(context.Background(), env(map[string]string{"HD_ENVIRONMENT": environment, "HD_BOT_TOKEN_SECRET": encoded}), options{bot: "staging-bot-1"}, now, &out)
		if !errors.Is(err, errNotTestEnvironment) || out.Len() != 0 {
			t.Fatalf("environment %q: %q, %v", environment, out.String(), err)
		}
	}
	for _, environment := range []string{"dev", "staging"} {
		var out bytes.Buffer
		if err := run(context.Background(), env(map[string]string{"HD_ENVIRONMENT": environment, "HD_BOT_TOKEN_SECRET": encoded}), options{bot: "staging-bot-1"}, now, &out); err != nil {
			t.Fatal(err)
		}
		if bot, err := bottoken.Verify(secret, strings.TrimSpace(out.String()), now); err != nil || bot != "staging-bot-1" {
			t.Fatalf("%s minted a token for %q, %v", environment, bot, err)
		}
	}
	if err := run(context.Background(), env(map[string]string{"HD_ENVIRONMENT": "staging"}), options{bot: "staging-bot-1"}, now, &bytes.Buffer{}); !errors.Is(err, bottoken.ErrWeakSecret) {
		t.Fatalf("without a secret: %v", err)
	}
}

func TestServerFlagSignsInAndPrintsTheSession(t *testing.T) {
	var seen map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/signin" {
			http.NotFound(w, r)
			return
		}
		json.NewDecoder(r.Body).Decode(&seen)
		json.NewEncoder(w).Encode(botclient.Session{AccessToken: "access-1", RefreshToken: "refresh-1", ExpiresIn: 900, AccountID: "account-1"})
	}))
	defer server.Close()
	var out bytes.Buffer
	lookup := env(map[string]string{"HD_ENVIRONMENT": "staging", "HD_BOT_TOKEN_SECRET": base64.StdEncoding.EncodeToString(secret)})
	if err := run(context.Background(), lookup, options{bot: "staging-bot-1", server: server.URL + "/"}, time.Now(), &out); err != nil {
		t.Fatal(err)
	}
	if seen["provider"] != "bot" || seen["device_id"] != "bot-token-staging-bot-1" {
		t.Fatalf("sign-in request %v", seen)
	}
	if bot, err := bottoken.Verify(secret, seen["id_token"], time.Now()); err != nil || bot != "staging-bot-1" {
		t.Fatalf("sent token for %q, %v", bot, err)
	}
	var session botclient.Session
	if err := json.Unmarshal(out.Bytes(), &session); err != nil || session.AccessToken != "access-1" || session.AccountID != "account-1" {
		t.Fatalf("printed %q, %v", out.String(), err)
	}
}
