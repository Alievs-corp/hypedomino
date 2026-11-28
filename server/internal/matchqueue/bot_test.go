package matchqueue_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/botaccount"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster/redistest"
	"github.com/Alievs-corp/hypedomino/server/internal/matchmaking"
	"github.com/Alievs-corp/hypedomino/server/internal/matchqueue"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func botsNamed(names ...string) func(context.Context, string) (bool, error) {
	return func(_ context.Context, account string) (bool, error) {
		for _, name := range names {
			if name == account {
				return true, nil
			}
		}
		return false, nil
	}
}

func TestABotAccountIsRefusedTheRankedQueueAndPlaysCasual(t *testing.T) {
	client, prefix := redistest.Client(t), redistest.Prefix(t)
	ctx := context.Background()
	plain := func(context.Context, string, rules.Mode, matchmaking.Queue, time.Time) (matchmaking.Entry, error) {
		return matchmaking.Entry{}, nil
	}
	gate := matchqueue.NewGate(matchqueue.NewStore(client, prefix), "node-a", "eu", matchqueue.CasualOnlyBots(botsNamed("staging-bot"), plain), func() bool { return true }, func() bool { return true }, func() time.Time { return now }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ranked := &pb.JoinQueueIntent{Mode: pb.Mode_MODE_1V1, Queue: pb.QueueKind_QUEUE_KIND_RANKED}
	casual := &pb.JoinQueueIntent{Mode: pb.Mode_MODE_1V1, Queue: pb.QueueKind_QUEUE_KIND_CASUAL}
	if state := gate.Join(ctx, "staging-bot", "s1", ranked); state.GetStatus() != pb.QueueStatus_QUEUE_STATUS_REFUSED || state.GetRefusal() != pb.QueueRefusal_QUEUE_REFUSAL_BOT_CASUAL_ONLY {
		t.Fatalf("bot in ranked %v", state)
	}
	if state := gate.Join(ctx, "staging-bot", "s1", casual); state.GetStatus() != pb.QueueStatus_QUEUE_STATUS_SEARCHING {
		t.Fatalf("bot in casual %v", state)
	}
	if state := gate.Join(ctx, "kamal", "s2", ranked); state.GetStatus() != pb.QueueStatus_QUEUE_STATUS_SEARCHING {
		t.Fatalf("player in ranked %v", state)
	}
}

func TestCasualOnlyBotsAsksOnlyForRankedEntries(t *testing.T) {
	asked := 0
	lookupFails := func(context.Context, string) (bool, error) {
		asked++
		return false, errors.New("database down")
	}
	plain := func(context.Context, string, rules.Mode, matchmaking.Queue, time.Time) (matchmaking.Entry, error) {
		return matchmaking.Entry{Mmr: 1200}, nil
	}
	entry := matchqueue.CasualOnlyBots(lookupFails, plain)
	if got, err := entry(context.Background(), "kamal", rules.Mode1v1, matchmaking.QueueCasual, now); err != nil || got.Mmr != 1200 || asked != 0 {
		t.Fatalf("casual entry %+v, %v, asked %d", got, err, asked)
	}
	if _, err := entry(context.Background(), "kamal", rules.Mode1v1, matchmaking.QueueRanked, now); err == nil || errors.Is(err, botaccount.ErrCasualOnly) || asked != 1 {
		t.Fatalf("ranked entry with a failed lookup: %v, asked %d", err, asked)
	}
}
