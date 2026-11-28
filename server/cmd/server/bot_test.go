package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/botaccount"
	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	"github.com/Alievs-corp/hypedomino/server/internal/bottoken"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

func TestAStagingBotSignsInAndPlaysCasualOnly(t *testing.T) {
	secret := make([]byte, bottoken.MinSecretBytes)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	h := startRedTeamWith(t, map[string]string{"HD_ENVIRONMENT": "staging", "HD_BOT_TOKEN_SECRET": base64.StdEncoding.EncodeToString(secret)})
	ctx := context.Background()
	minted, err := bottoken.Mint(secret, "staging-bot-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	session, err := botclient.SignIn(ctx, http.DefaultClient, h.base, minted, redTeamDevice)
	if err != nil {
		t.Fatal(err)
	}
	if flagged, err := botaccount.Is(ctx, h.pool, session.AccountID); err != nil || !flagged {
		t.Fatalf("signed-in account is a bot: %v, %v", flagged, err)
	}
	requestKey, err := base64.StdEncoding.DecodeString(session.RequestKey)
	if err != nil {
		t.Fatal(err)
	}
	bot := &attacker{account: session.AccountID, pair: token.Pair{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, RequestKey: requestKey}, address: "198.51.100.251"}
	player := h.attacker(t)
	for _, route := range []struct {
		method, path string
	}{
		{http.MethodGet, "/wallet"},
		{http.MethodGet, "/shop"},
		{http.MethodPost, "/shop/buy"},
		{http.MethodPost, "/shop/complete"},
		{http.MethodGet, "/purchases"},
		{http.MethodPost, "/daily-login/claim"},
		{http.MethodGet, "/pass"},
		{http.MethodGet, "/missions"},
		{http.MethodPost, "/promo/redeem"},
	} {
		if answer := h.signed(t, bot, route.method, route.path, []byte(`{}`)); answer.status != http.StatusForbidden || answer.code() != botaccount.RefusalCode {
			t.Errorf("%s %s for a bot answered %d %v", route.method, route.path, answer.status, answer.body)
		}
		if answer := h.signed(t, player, route.method, route.path, []byte(`{}`)); answer.code() == botaccount.RefusalCode {
			t.Errorf("%s %s for a player answered %d %v", route.method, route.path, answer.status, answer.body)
		}
	}
	c := h.connect(t, bot)
	send(t, c, h.frame(c, &pb.Envelope{Payload: &pb.Envelope_JoinQueueIntent{JoinQueueIntent: &pb.JoinQueueIntent{Mode: pb.Mode_MODE_1V1, Queue: pb.QueueKind_QUEUE_KIND_RANKED}}}))
	if state := nextQueueState(t, c); state.GetStatus() != pb.QueueStatus_QUEUE_STATUS_REFUSED || state.GetRefusal() != pb.QueueRefusal_QUEUE_REFUSAL_BOT_CASUAL_ONLY {
		t.Fatalf("ranked queue for a bot answered %v", state)
	}
	send(t, c, h.frame(c, &pb.Envelope{Payload: &pb.Envelope_JoinQueueIntent{JoinQueueIntent: &pb.JoinQueueIntent{Mode: pb.Mode_MODE_1V1, Queue: pb.QueueKind_QUEUE_KIND_CASUAL}}}))
	if state := nextQueueState(t, c); state.GetStatus() != pb.QueueStatus_QUEUE_STATUS_SEARCHING {
		t.Fatalf("casual queue for a bot answered %v", state)
	}
	foreign := make([]byte, bottoken.MinSecretBytes)
	if _, err := rand.Read(foreign); err != nil {
		t.Fatal(err)
	}
	forged, err := bottoken.Mint(foreign, "staging-bot-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := botclient.SignIn(ctx, http.DefaultClient, h.base, forged, redTeamDevice); !errors.Is(err, botclient.ErrSignInRefused) {
		t.Fatalf("a token from another secret signed in: %v", err)
	}
}
