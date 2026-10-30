package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/auth"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

type fakeVerifier struct {
	nonce      string
	nonceErr   error
	identities map[string]idtoken.Identity
	errs       map[string]error
	seen       []string
}

func (v *fakeVerifier) IssueNonce(context.Context) (string, error) {
	return v.nonce, v.nonceErr
}

func (v *fakeVerifier) Verify(_ context.Context, provider idtoken.Provider, token, nonce string) (idtoken.Identity, error) {
	v.seen = append(v.seen, provider.String()+":"+token+":"+nonce)
	if err := v.errs[token]; err != nil {
		return idtoken.Identity{}, err
	}
	identity, ok := v.identities[token]
	if !ok {
		return idtoken.Identity{}, idtoken.ErrSignature
	}
	identity.Provider = provider
	return identity, nil
}

type fakeAccounts struct {
	states map[string]error
	known  map[string]bool
	calls  int
}

func (a *fakeAccounts) SignIn(_ context.Context, identity idtoken.Identity) (account.Account, bool, error) {
	a.calls++
	id := identity.Provider.String() + "-" + identity.Subject
	if err := a.states[id]; err != nil {
		return account.Account{}, false, err
	}
	created := !a.known[id]
	a.known[id] = true
	return account.Account{ID: id, State: account.StateActive, Email: identity.Email}, created, nil
}

type fakeSessions struct {
	issued []string
}

func (s *fakeSessions) Create(account string) string {
	s.issued = append(s.issued, account)
	return "token-for-" + account
}

type fixture struct {
	t        *testing.T
	server   *httptest.Server
	verifier *fakeVerifier
	accounts *fakeAccounts
	sessions *fakeSessions
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		t:        t,
		verifier: &fakeVerifier{nonce: "nonce-1", identities: map[string]idtoken.Identity{"good": {Subject: "sub-1", Email: "a@example.com"}}, errs: map[string]error{}},
		accounts: &fakeAccounts{states: map[string]error{}, known: map[string]bool{}},
		sessions: &fakeSessions{},
	}
	handler := auth.New(f.verifier, f.accounts, f.sessions, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.Handle("POST /auth/", handler)
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fixture) post(path, body string) (int, map[string]any) {
	f.t.Helper()
	resp, err := http.Post(f.server.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		f.t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Cache-Control") != "no-store" || !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		f.t.Fatalf("headers %v", resp.Header)
	}
	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		f.t.Fatal(err)
	}
	return resp.StatusCode, decoded
}

func TestNonceIsIssuedByTheVerifier(t *testing.T) {
	f := newFixture(t)
	if status, body := f.post("/auth/nonce", ""); status != http.StatusOK || body["nonce"] != "nonce-1" {
		t.Fatalf("nonce answered %d %v", status, body)
	}
	f.verifier.nonceErr = errors.New("redis down")
	if status, body := f.post("/auth/nonce", ""); status != http.StatusServiceUnavailable || body["error"] != "unavailable" {
		t.Fatalf("nonce without a store answered %d %v", status, body)
	}
}

func TestSignInVerifiesTheTokenFindsTheAccountAndIssuesASession(t *testing.T) {
	f := newFixture(t)
	status, body := f.post("/auth/signin", `{"provider":"google","id_token":"good","nonce":"nonce-1"}`)
	if status != http.StatusOK || body["session_token"] != "token-for-google-sub-1" || body["account_id"] != "google-sub-1" || body["created"] != true {
		t.Fatalf("first sign-in answered %d %v", status, body)
	}
	status, body = f.post("/auth/signin", `{"provider":"apple","id_token":"good","nonce":"nonce-2"}`)
	if status != http.StatusOK || body["account_id"] != "apple-sub-1" || body["created"] != true {
		t.Fatalf("apple sign-in answered %d %v", status, body)
	}
	status, body = f.post("/auth/signin", `{"provider":"google","id_token":"good","nonce":"nonce-3"}`)
	if status != http.StatusOK || body["created"] != false {
		t.Fatalf("returning sign-in answered %d %v", status, body)
	}
	if want := []string{"google:good:nonce-1", "apple:good:nonce-2", "google:good:nonce-3"}; strings.Join(f.verifier.seen, " ") != strings.Join(want, " ") {
		t.Fatalf("verifier saw %v", f.verifier.seen)
	}
	if len(f.sessions.issued) != 3 {
		t.Fatalf("sessions issued %v", f.sessions.issued)
	}
}

func TestSignInRefusalsCarryAStatusAndAReason(t *testing.T) {
	f := newFixture(t)
	f.verifier.errs["expired"] = idtoken.ErrExpired
	f.verifier.errs["replayed"] = idtoken.ErrNonceUsed
	f.verifier.errs["offline"] = idtoken.ErrKeyFetch
	f.verifier.identities["banned"] = idtoken.Identity{Subject: "banned"}
	f.accounts.states["google-banned"] = account.ErrBanned
	f.verifier.identities["deleted"] = idtoken.Identity{Subject: "deleted"}
	f.accounts.states["google-deleted"] = account.ErrDeleted
	f.verifier.identities["broken"] = idtoken.Identity{Subject: "broken"}
	f.accounts.states["google-broken"] = errors.New("database down")
	cases := []struct {
		body   string
		status int
		reason string
	}{
		{`{"provider":"google","id_token":"good","nonce":"n"`, http.StatusBadRequest, "malformed"},
		{`{"provider":"facebook","id_token":"good","nonce":"n"}`, http.StatusBadRequest, "unknown_provider"},
		{`{"provider":"google","id_token":"forged","nonce":"n"}`, http.StatusUnauthorized, "token_rejected"},
		{`{"provider":"google","id_token":"expired","nonce":"n"}`, http.StatusUnauthorized, "token_expired"},
		{`{"provider":"google","id_token":"replayed","nonce":"n"}`, http.StatusUnauthorized, "nonce_rejected"},
		{`{"provider":"google","id_token":"offline","nonce":"n"}`, http.StatusServiceUnavailable, "provider_unavailable"},
		{`{"provider":"google","id_token":"banned","nonce":"n"}`, http.StatusForbidden, "banned"},
		{`{"provider":"google","id_token":"deleted","nonce":"n"}`, http.StatusGone, "deleted"},
		{`{"provider":"google","id_token":"broken","nonce":"n"}`, http.StatusInternalServerError, "internal"},
	}
	for _, c := range cases {
		if status, body := f.post("/auth/signin", c.body); status != c.status || body["error"] != c.reason {
			t.Fatalf("%s answered %d %v, want %d %s", c.body, status, body, c.status, c.reason)
		}
	}
	if len(f.sessions.issued) != 0 {
		t.Fatalf("a refused sign-in issued a session: %v", f.sessions.issued)
	}
}

func TestOversizedBodiesAndOtherMethodsAreRefused(t *testing.T) {
	f := newFixture(t)
	big := `{"provider":"google","id_token":"` + strings.Repeat("x", auth.MaxBodyBytes) + `","nonce":"n"}`
	if status, body := f.post("/auth/signin", big); status != http.StatusBadRequest || body["error"] != "malformed" {
		t.Fatalf("oversized body answered %d %v", status, body)
	}
	resp, err := http.Get(f.server.URL + "/auth/nonce")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET nonce answered %d", resp.StatusCode)
	}
	if f.accounts.calls != 0 {
		t.Fatalf("accounts were consulted %d times", f.accounts.calls)
	}
}
