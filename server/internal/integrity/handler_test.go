package integrity

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

type stubVerifier struct{}

func (stubVerifier) Verify(access string) (token.Claims, error) {
	switch access {
	case "good":
		return token.Claims{Account: "acc-1", Session: "sid-1"}, nil
	case "old":
		return token.Claims{}, token.ErrExpired
	}
	return token.Claims{}, token.ErrSignature
}

func TestNonceEndpointBindsTheNonceToTheCaller(t *testing.T) {
	clock := newTestClock()
	store := NewMemoryNonceStore(clock)
	nonces, err := NewNonces(NonceConfig{TTL: 5 * time.Minute}, store, clock, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(stubVerifier{}, nonces, slog.New(slog.NewTextHandler(io.Discard, nil)))
	call := func(access, body string) (int, map[string]any) {
		r := httptest.NewRequest(http.MethodPost, "/integrity/nonce", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+access)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		var decoded map[string]any
		json.Unmarshal(w.Body.Bytes(), &decoded)
		return w.Code, decoded
	}
	for _, purpose := range []Purpose{PurposeSessionStart, PurposePurchase, PurposeRankedMatch, PurposeCasualMatch} {
		status, body := call("good", `{"purpose":"`+purpose.String()+`"}`)
		if status != http.StatusOK || body["expires_in"] != float64(300) {
			t.Fatalf("%v: %d %v", purpose, status, body)
		}
		stored, found, _ := store.Take(t.Context(), body["nonce"].(string))
		if want := (Binding{Account: "acc-1", Session: "sid-1", Purpose: purpose}); !found || stored.Binding != want || !stored.ExpiresAt.Equal(clock.Now().Add(5*time.Minute)) {
			t.Fatalf("%v stored %+v %v", purpose, stored, found)
		}
	}
	refusals := []struct {
		access, body string
		status       int
		code         string
	}{
		{"bad", `{"purpose":"purchase"}`, http.StatusUnauthorized, "access_rejected"},
		{"old", `{"purpose":"purchase"}`, http.StatusUnauthorized, "access_expired"},
		{"good", `{"purpose":"invalid"}`, http.StatusBadRequest, "unknown_purpose"},
		{"good", `{}`, http.StatusBadRequest, "unknown_purpose"},
		{"good", `not json`, http.StatusBadRequest, "bad_request"},
		{"good", `{"purpose":"` + strings.Repeat("x", MaxBodyBytes) + `"}`, http.StatusBadRequest, "bad_request"},
	}
	for _, c := range refusals {
		if status, body := call(c.access, c.body); status != c.status || body["error"] != c.code {
			t.Fatalf("%s %.40s: %d %v", c.access, c.body, status, body)
		}
	}
}
