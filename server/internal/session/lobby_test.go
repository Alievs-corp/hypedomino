package session

import (
	"context"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/custom"
	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/secrand"
)

type friendList map[string][]string

func (f friendList) FriendIDs(ctx context.Context, account string) ([]string, error) {
	return f[account], nil
}

func newLobbyTable(t *testing.T) (*table, *custom.Lobby) {
	t.Helper()
	clock := &fakeClock{now: time.UnixMilli(1761480000000)}
	manager := New(DefaultConfig(), clock, secrand.Source{})
	lobby := custom.NewLobby(secrand.Source{}, func(options room.Options, seats []custom.SeatState) (*room.Room, error) {
		options.Config, options.Bot, options.BotLevel, options.Clock = rules.DefaultConfig(), bot.DefaultConfig(), bot.Hard, clock
		r := room.New(options)
		ctx, cancel := context.WithCancel(context.Background())
		go r.Run(ctx)
		t.Cleanup(func() { cancel(); <-r.Done() })
		return r, nil
	})
	members := func(ctx context.Context, account string) (custom.Member, error) {
		return custom.Member{Clips: []string{"idle"}, Emoji: []string{"emoji.laugh"}}, nil
	}
	manager.Host(lobby, members, friendList{"ana": {"ben"}})
	config := gateway.DefaultConfig()
	config.IntentsPerSecond, config.IntentBurst, config.CloseTimeout = 1000, 1000, 500*time.Millisecond
	server := httptest.NewServer(gateway.New(Options(config, manager)))
	t.Cleanup(server.Close)
	return &table{t: t, clock: clock, manager: manager, server: server}, lobby
}

func (tab *table) member(account string) *player {
	tab.t.Helper()
	return tab.connect(tab.manager.Create(account), true)
}

func (p *player) lobby(intent *pb.LobbyIntent) {
	p.write(&pb.Envelope{ProtocolVersion: gateway.ProtocolVersion, Payload: &pb.Envelope_LobbyIntent{LobbyIntent: intent}})
}

func (p *player) until(match func(*pb.Envelope) bool) *pb.Envelope {
	p.t.Helper()
	for {
		if frame := p.read(); match(frame) {
			return frame
		}
	}
}

func (p *player) lobbyState() *pb.LobbyState {
	p.t.Helper()
	return p.until(func(f *pb.Envelope) bool { return f.GetLobbyState() != nil }).GetLobbyState()
}

func (p *player) relay() *pb.LobbyRelay {
	p.t.Helper()
	return p.until(func(f *pb.Envelope) bool { return f.GetLobbyRelay() != nil }).GetLobbyRelay()
}

func (p *player) refusal(want pb.LobbyRefusal) {
	p.t.Helper()
	if got := p.lobbyState().GetRefusal(); got != want {
		p.t.Fatalf("refusal %v, want %v", got, want)
	}
}

func command(c pb.LobbyCommand) *pb.LobbyIntent {
	return &pb.LobbyIntent{Command: c}
}

func TestFriendsFormALobbySplitTeamsAndStartACustomMatch(t *testing.T) {
	tab, _ := newLobbyTable(t)
	ana, ben := tab.member("ana"), tab.member("ben")
	ana.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_CREATE))
	created := ana.lobbyState()
	if created.GetCode() == "" || created.GetLeaderAccountId() != "ana" || created.GetPhase() != pb.LobbyPhase_LOBBY_PHASE_LOBBY {
		t.Fatalf("created %v", created)
	}
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_JOIN, Code: created.GetCode()})
	for _, p := range []*player{ana, ben} {
		if state := p.lobbyState(); !slices.Equal(state.GetMemberAccountIds(), []string{"ana", "ben"}) {
			t.Fatalf("members %v", state.GetMemberAccountIds())
		}
	}
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_ACT, Action: "dance"})
	ben.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_UNKNOWN_ACTION)
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_ACT, Action: "idle"})
	ben.write(&pb.Envelope{ProtocolVersion: gateway.ProtocolVersion, Payload: &pb.Envelope_EmojiIntent{EmojiIntent: &pb.EmojiIntent{ItemId: "emoji.laugh"}}})
	for _, p := range []*player{ana, ben} {
		if action, emoji := p.relay(), p.relay(); action.GetAccountId() != "ben" || action.GetAction() != "idle" || emoji.GetItemId() != "emoji.laugh" {
			t.Fatalf("relayed %v %v", action, emoji)
		}
	}
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_TEAMS, Mode: pb.Mode_MODE_1V1})
	ben.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_NOT_LEADER)
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_TEAMS, Mode: pb.Mode_MODE_1V1})
	if state := ben.lobbyState(); state.GetPhase() != pb.LobbyPhase_LOBBY_PHASE_TEAMS || len(state.GetSeats()) != 2 {
		t.Fatalf("team setup %v", state)
	}
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_REQUEST_SEAT, Seat: pb.Seat_SEAT_Y})
	if request := ana.relay(); request.GetAccountId() != "ben" || request.GetRequestedSeat() != pb.Seat_SEAT_Y {
		t.Fatalf("seat request %v", request)
	}
	ana.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_START))
	ana.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_SEATS_OPEN)
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_PLACE, AccountId: "ana", Seat: pb.Seat_SEAT_X})
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_PLACE, AccountId: "ben", Seat: pb.Seat_SEAT_Y})
	ana.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_START))
	for seat, p := range map[pb.Seat]*player{pb.Seat_SEAT_X: ana, pb.Seat_SEAT_Y: ben} {
		if view := p.view(); view.GetSeat() != seat || view.GetMode() != pb.Mode_MODE_1V1 || view.GetState() != pb.MatchState_MATCH_STATE_LOADING {
			t.Fatalf("custom match view %v", view)
		}
	}
	if !tab.manager.InMatch("ana") || !tab.manager.InMatch("ben") {
		t.Fatal("the custom match did not seat its players")
	}
}

func TestInvitesReachOnlineFriendsAndLeavingOrRemovalClearsTheLobby(t *testing.T) {
	tab, lobby := newLobbyTable(t)
	ana, ben, cem := tab.member("ana"), tab.member("ben"), tab.member("cem")
	ana.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_CREATE))
	code := ana.lobbyState().GetCode()
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_INVITE, AccountId: "cem"})
	ana.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_NOT_FRIEND)
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_INVITE, AccountId: "ben"})
	invite := ben.until(func(f *pb.Envelope) bool { return f.GetLobbyInvite() != nil }).GetLobbyInvite()
	if invite.GetCode() != code || invite.GetFromAccountId() != "ana" {
		t.Fatalf("invite %v", invite)
	}
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_JOIN, Code: code})
	ben.lobbyState()
	ben.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_JOIN, Code: code})
	ben.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_ALREADY_IN_LOBBY)
	ana.lobbyState()
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_REMOVE, AccountId: "ben"})
	if state := ben.lobbyState(); state.GetCode() != "" {
		t.Fatalf("removed member still sees %v", state)
	}
	if state := ana.lobbyState(); !slices.Equal(state.GetMemberAccountIds(), []string{"ana"}) {
		t.Fatalf("after the removal %v", state)
	}
	cem.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_JOIN, Code: code})
	cem.lobbyState()
	ana.lobbyState()
	cem.socket.CloseNow()
	for {
		if state := ana.lobbyState(); slices.Equal(state.GetMemberAccountIds(), []string{"ana"}) {
			break
		}
	}
	ana.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_LEAVE))
	if state := ana.lobbyState(); state.GetCode() != "" || lobby.Len() != 0 {
		t.Fatalf("after the last leave %v, %d rooms", state, lobby.Len())
	}
	ana.lobby(&pb.LobbyIntent{Command: pb.LobbyCommand_LOBBY_COMMAND_JOIN, Code: code})
	ana.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_UNKNOWN_CODE)
}

func TestLobbyIntentsAreRefusedWithoutAHostedLobby(t *testing.T) {
	tab := newTable(t, rules.Mode1v1)
	p := tab.connect(tab.manager.Create("ana"), true)
	p.lobby(command(pb.LobbyCommand_LOBBY_COMMAND_CREATE))
	p.refusal(pb.LobbyRefusal_LOBBY_REFUSAL_UNAVAILABLE)
}
