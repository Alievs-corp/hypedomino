package custom

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type counter struct{ next uint64 }

func (c *counter) Uint64() uint64 {
	c.next++
	return c.next * 0x9e3779b97f4a7c15
}

type sameCode struct{}

func (sameCode) Uint64() uint64 { return 7 }

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

type chanSink chan room.Message

func (s chanSink) Deliver(m room.Message) { s <- m }

func opener(t *testing.T) (Opener, *[]*room.Room) {
	t.Helper()
	var opened []*room.Room
	return func(options room.Options) (*room.Room, error) {
		options.Config, options.Bot, options.BotLevel, options.Clock = rules.DefaultConfig(), bot.DefaultConfig(), bot.Hard, realClock{}
		r := room.New(options)
		ctx, cancel := context.WithCancel(context.Background())
		go r.Run(ctx)
		t.Cleanup(func() { cancel(); <-r.Done() })
		opened = append(opened, r)
		return r, nil
	}, &opened
}

func accounts(state State) []string {
	var out []string
	for _, seat := range state.Seats {
		out = append(out, seat.Account)
	}
	return out
}

func TestCodesAreShortUnambiguousAndUnique(t *testing.T) {
	open, _ := opener(t)
	lobby := NewLobby(&counter{}, open)
	seen := map[string]bool{}
	for index := 0; index < 500; index++ {
		r, err := lobby.Create("creator", rules.Mode1v1)
		if err != nil {
			t.Fatal(err)
		}
		code := r.Code()
		if len(code) != CodeLength || strings.ContainsAny(code, "01IO") || strings.ToUpper(code) != code || seen[code] {
			t.Fatalf("code %q", code)
		}
		seen[code] = true
		if found, err := lobby.Find(code); err != nil || found != r {
			t.Fatalf("find %q: %v", code, err)
		}
	}
	if _, err := lobby.Find("NOPE42"); !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("unknown code: %v", err)
	}
	if _, err := lobby.Create("creator", rules.Mode(9)); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("bad mode: %v", err)
	}
	if lobby.Len() != 500 {
		t.Fatalf("lobby holds %d rooms", lobby.Len())
	}
}

func TestCreateRetriesACollidingCode(t *testing.T) {
	open, _ := opener(t)
	lobby := NewLobby(sameCode{}, open)
	first, _ := lobby.Create("a", rules.Mode1v1)
	first.Close()
	second, err := lobby.Create("b", rules.Mode1v1)
	if err != nil || second.Code() != first.Code() || lobby.Len() != 1 {
		t.Fatalf("after closing, the code was not reused: %v", err)
	}
}

func TestFriendsPickSeatsAndTeamsAndTheCreatorStarts(t *testing.T) {
	open, opened := opener(t)
	lobby := NewLobby(&counter{}, open)
	r, _ := lobby.Create("ana", rules.Mode2v2)
	for _, friend := range []string{"ben", "cem", "dan"} {
		if _, err := r.Join(friend); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.Join("eve"); !errors.Is(err, ErrFull) {
		t.Fatalf("fifth player: %v", err)
	}
	if _, err := r.Join("ben"); !errors.Is(err, ErrAlreadyInRoom) {
		t.Fatalf("joining twice: %v", err)
	}
	if err := r.Sit("ben", rules.SeatSouth); !errors.Is(err, ErrSeatTaken) {
		t.Fatalf("sitting on a taken seat: %v", err)
	}
	if err := r.Sit("eve", rules.SeatSouth); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("outsider sitting: %v", err)
	}
	if err := r.Sit("ben", rules.SeatX); !errors.Is(err, ErrInvalidSeat) {
		t.Fatalf("seat outside the mode: %v", err)
	}
	if err := r.Leave("dan"); err != nil {
		t.Fatal(err)
	}
	if err := r.Sit("ben", rules.SeatWest); err != nil {
		t.Fatal(err)
	}
	if err := r.Sit("cem", rules.SeatEast); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Join("dan"); err != nil {
		t.Fatal(err)
	}
	state := r.State()
	if got := strings.Join(accounts(state), ","); got != "ana,cem,dan,ben" {
		t.Fatalf("seating %s", got)
	}
	if rules.SeatNorth.Team() != rules.SeatSouth.Team() || rules.SeatEast.Team() != rules.SeatWest.Team() {
		t.Fatal("team layout assumption broken")
	}
	if _, err := r.Start("ana", room.Options{}); !errors.Is(err, ErrNotReady) {
		t.Fatalf("start before ready: %v", err)
	}
	for _, account := range []string{"ana", "ben", "cem", "dan"} {
		if err := r.Ready(account, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Ready("eve", true); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("outsider ready: %v", err)
	}
	if _, err := r.Start("ben", room.Options{}); !errors.Is(err, ErrNotCreator) {
		t.Fatalf("a friend started: %v", err)
	}
	table, err := r.Start("ana", room.Options{Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	if table.Kind() != room.KindCustom || len(*opened) != 1 || r.Table() != table || !r.State().Started {
		t.Fatalf("started table kind %s, opened %d", table.Kind(), len(*opened))
	}
	if _, err := r.Start("ana", room.Options{}); !errors.Is(err, ErrStarted) {
		t.Fatalf("second start: %v", err)
	}
	if _, err := r.Join("eve"); !errors.Is(err, ErrStarted) {
		t.Fatalf("joining a running room: %v", err)
	}
	if err := r.Sit("ben", rules.SeatNorth); !errors.Is(err, ErrStarted) {
		t.Fatalf("moving during the match: %v", err)
	}
}

func TestTheCreatorRemovesAndAddsPlayersDuringTheMatch(t *testing.T) {
	open, _ := opener(t)
	lobby := NewLobby(&counter{}, open)
	r, _ := lobby.Create("ana", rules.Mode1v1)
	r.Join("ben")
	r.Ready("ana", true)
	r.Ready("ben", true)
	table, err := r.Start("ana", room.Options{Seed: 4})
	if err != nil {
		t.Fatal(err)
	}
	sinks := map[string]chanSink{"ana": make(chanSink, 64), "ben": make(chanSink, 64)}
	table.Join(rules.SeatX, sinks["ana"])
	table.Join(rules.SeatY, sinks["ben"])
	for _, sink := range sinks {
		if m := <-sink; m.Rejected {
			t.Fatalf("join rejected: %v", m.Reason)
		}
	}
	if err := r.Remove("ben", "ana"); !errors.Is(err, ErrNotCreator) {
		t.Fatalf("a friend removed the creator: %v", err)
	}
	if err := r.Remove("ana", "ana"); !errors.Is(err, ErrNotCreator) {
		t.Fatalf("the creator removed themselves: %v", err)
	}
	if err := r.Remove("ana", "zed"); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("removing an outsider: %v", err)
	}
	if err := r.Remove("ana", "ben"); err != nil {
		t.Fatal(err)
	}
	if got := accounts(r.State()); got[1] != "" {
		t.Fatalf("seat Y still held by %q", got[1])
	}
	view := lastView(t, sinks["ana"])
	if view.Controllers[rules.SeatY] != rules.ControllerPlayer && view.Controllers[rules.SeatY] != rules.ControllerBot {
		t.Fatalf("view %+v", view)
	}
	reply := make(chan room.View, 1)
	table.Snapshot(rules.SeatX, reply)
	if v := <-reply; v.Controllers[rules.SeatY] != rules.ControllerBot {
		t.Fatalf("the removed seat is not played by the bot: %+v", v)
	}
	back := make(chanSink, 16)
	table.Reconnect(rules.SeatY, back)
	if m := <-back; !m.Rejected || !errors.Is(m.Reason, room.ErrBarred) {
		t.Fatalf("the removed player came back on their own: %+v", m)
	}
	if err := r.Add("ana", "cem", rules.SeatX); !errors.Is(err, ErrSeatTaken) {
		t.Fatalf("adding to a taken seat: %v", err)
	}
	if err := r.Add("ana", "ana", rules.SeatY); !errors.Is(err, ErrAlreadyInRoom) {
		t.Fatalf("adding a seated account: %v", err)
	}
	if err := r.Add("ben", "cem", rules.SeatY); !errors.Is(err, ErrNotCreator) {
		t.Fatalf("a friend added: %v", err)
	}
	if err := r.Add("ana", "cem", rules.SeatY); err != nil {
		t.Fatal(err)
	}
	table.Join(rules.SeatY, back)
	if m := <-back; m.Rejected || m.View.Controllers[rules.SeatY] != rules.ControllerPlayer {
		t.Fatalf("the added player could not take the seat: %+v", m)
	}
	if err := r.Leave("cem"); err != nil {
		t.Fatal(err)
	}
	table.Snapshot(rules.SeatX, reply)
	if v := <-reply; v.Controllers[rules.SeatY] != rules.ControllerBot || accounts(r.State())[1] != "cem" {
		t.Fatalf("leaving during the match: %+v seats %v", v, accounts(r.State()))
	}
}

func lastView(t *testing.T, sink chanSink) room.View {
	t.Helper()
	var view room.View
	received := false
	for {
		select {
		case m := <-sink:
			view, received = m.View, true
		case <-time.After(200 * time.Millisecond):
			if !received {
				t.Fatal("no view received")
			}
			return view
		}
	}
}

func TestTheCreatorLeavingBeforeTheStartDissolvesTheRoom(t *testing.T) {
	open, _ := opener(t)
	lobby := NewLobby(&counter{}, open)
	r, _ := lobby.Create("ana", rules.Mode1v1)
	r.Join("ben")
	if err := r.Leave("zed"); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("outsider leaving: %v", err)
	}
	if err := r.Leave("ana"); err != nil {
		t.Fatal(err)
	}
	if !r.State().Closed || lobby.Len() != 0 {
		t.Fatalf("room still open: %+v", r.State())
	}
	if _, err := r.Join("cem"); !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("joining a dissolved room: %v", err)
	}
	if err := r.Add("ana", "cem", rules.SeatY); !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("adding to a dissolved room: %v", err)
	}
	if _, err := lobby.Find(r.Code()); !errors.Is(err, ErrUnknownCode) {
		t.Fatal("dissolved room still findable")
	}
}

func TestRemovingBeforeTheStartFreesTheSeatAndReadyResetsOnMove(t *testing.T) {
	open, _ := opener(t)
	lobby := NewLobby(&counter{}, open)
	r, _ := lobby.Create("ana", rules.Mode2v2)
	r.Join("ben")
	r.Ready("ben", true)
	if err := r.Sit("ben", rules.SeatSouth); err != nil {
		t.Fatal(err)
	}
	if state := r.State(); state.Seats[2].Account != "ben" || state.Seats[2].Ready {
		t.Fatalf("after moving: %+v", state.Seats)
	}
	if err := r.Sit("ben", rules.SeatSouth); err != nil {
		t.Fatalf("sitting where one already sits: %v", err)
	}
	if err := r.Remove("ana", "ben"); err != nil {
		t.Fatal(err)
	}
	if state := r.State(); state.Seats[2].Account != "" {
		t.Fatalf("seat not freed: %+v", state.Seats)
	}
	if err := r.Add("ana", "cem", rules.SeatWest); err != nil || r.State().Seats[3].Account != "cem" {
		t.Fatalf("add before start: %v %+v", err, r.State().Seats)
	}
	if err := r.Add("ana", "dan", rules.SeatX); !errors.Is(err, ErrInvalidSeat) {
		t.Fatalf("adding outside the mode: %v", err)
	}
}
