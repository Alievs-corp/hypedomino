package gateway

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

func roomServer(t *testing.T, mode rules.Mode, seed uint64) (*httptest.Server, *room.Room) {
	t.Helper()
	r := room.New(room.Options{Mode: mode, Config: rules.DefaultConfig(), Bot: bot.DefaultConfig(), BotLevel: bot.Hard, Seed: seed, Clock: realClock{}})
	ctx, cancel := context.WithCancel(context.Background())
	go r.Run(ctx)
	t.Cleanup(func() { cancel(); <-r.Done() })
	handler := HandlerFunc(func(ctx context.Context, conn *Conn, hello *pb.ClientHello) (Session, error) {
		kind, name, _ := strings.Cut(hello.GetSessionToken(), ":")
		seat, err := rules.ParseSeat(name)
		if err != nil {
			return nil, ErrUnauthenticated
		}
		if kind == "resume" {
			return ResumeRoom(r, seat, conn, hello.GetLastAckedSequence()), nil
		}
		return JoinRoom(r, seat, conn), nil
	})
	config := testConfig()
	config.IntentsPerSecond = 1000
	config.IntentBurst = 1000
	server := httptest.NewServer(New(Options{Config: config, Handler: handler}))
	t.Cleanup(server.Close)
	return server, r
}

type player struct {
	t        *testing.T
	seat     pb.Seat
	socket   *websocket.Conn
	ctx      context.Context
	sequence uint64
	acked    uint64
	sent     int
	bytes    int
	buffer   []byte
	last     *pb.MatchView
}

func join(t *testing.T, server *httptest.Server, seat rules.Seat, resume bool) *player {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	socket, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{CompressionMode: websocket.CompressionNoContextTakeover})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { socket.CloseNow() })
	p := &player{t: t, seat: seatMessage(seat), socket: socket, ctx: ctx}
	token := "join:" + seat.String()
	if resume {
		token = "resume:" + seat.String()
	}
	frame := hello(token, "1.0.0", 1)
	p.write(frame)
	reply := p.read()
	if reply.GetServerHello().GetStatus() != pb.HelloStatus_HELLO_STATUS_ACCEPTED {
		t.Fatalf("%s: hello answered %v", seat, reply)
	}
	if !resume {
		p.write(&pb.Envelope{ProtocolVersion: ProtocolVersion, Payload: &pb.Envelope_LoadingIntent{LoadingIntent: &pb.LoadingIntent{Progress: 100, Ready: true}}})
	}
	return p
}

func (p *player) write(frame *pb.Envelope) {
	p.t.Helper()
	data, err := proto.MarshalOptions{}.MarshalAppend(p.buffer[:0], frame)
	if err != nil {
		p.t.Fatal(err)
	}
	p.buffer = data
	if err := p.socket.Write(p.ctx, websocket.MessageBinary, data); err != nil {
		p.t.Fatal(err)
	}
}

func (p *player) read() *pb.Envelope {
	p.t.Helper()
	_, data, err := p.socket.Read(p.ctx)
	if err != nil {
		p.t.Fatalf("%v: read: %v", p.seat, err)
	}
	frame := &pb.Envelope{}
	if err := proto.Unmarshal(data, frame); err != nil {
		p.t.Fatal(err)
	}
	return frame
}

func (p *player) intent(frame *pb.Envelope) {
	p.sequence++
	frame = intentFrame(p.sequence, frame)
	p.write(frame)
	p.sent++
	p.bytes += proto.Size(frame)
}

func (p *player) step() bool {
	frame := p.read()
	if frame.GetError() != nil {
		p.t.Fatalf("%v: rejected %v", p.seat, frame.GetError())
	}
	view := frame.GetMatchView()
	if view == nil {
		p.t.Fatalf("%v: unexpected frame %v", p.seat, frame)
	}
	p.acked = frame.GetAckedSequence()
	p.last = view
	if view.GetState() == pb.MatchState_MATCH_STATE_FINISHED {
		return false
	}
	if view.GetCurrent() != p.seat {
		return true
	}
	p.act(view)
	return true
}

func (p *player) act(view *pb.MatchView) {
	switch {
	case view.GetState() == pb.MatchState_MATCH_STATE_OPENING && view.GetForcedOpening() != nil:
		p.intent(&pb.Envelope{Payload: &pb.Envelope_PlaceTileIntent{PlaceTileIntent: &pb.PlaceTileIntent{Tile: view.GetForcedOpening().GetTile()}}})
	case view.GetState() == pb.MatchState_MATCH_STATE_OPENING:
		p.intent(&pb.Envelope{Payload: &pb.Envelope_PlaceTileIntent{PlaceTileIntent: &pb.PlaceTileIntent{Tile: view.GetHand()[0]}}})
	case len(view.GetLegalMoves()) > 0:
		move := view.GetLegalMoves()[0]
		p.intent(&pb.Envelope{Payload: &pb.Envelope_PlaceTileIntent{PlaceTileIntent: &pb.PlaceTileIntent{Tile: move.GetTile(), End: move.GetEnd()}}})
	case view.GetTurnPlacements() > 0:
		p.intent(&pb.Envelope{Payload: &pb.Envelope_EndTurnIntent{EndTurnIntent: &pb.EndTurnIntent{}}})
	case view.GetMode() == pb.Mode_MODE_1V1 && view.GetBoneyardSize() > 1:
		p.intent(&pb.Envelope{Payload: &pb.Envelope_DrawIntent{DrawIntent: &pb.DrawIntent{}}})
	default:
		p.intent(&pb.Envelope{Payload: &pb.Envelope_PassIntent{PassIntent: &pb.PassIntent{}}})
	}
}

func TestPlayersFinishAMatchOverWebSocket(t *testing.T) {
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		server, r := roomServer(t, mode, 7)
		players := make([]*player, 0, mode.Players())
		for _, seat := range mode.Seats() {
			players = append(players, join(t, server, seat, false))
		}
		done := make(chan *player, len(players))
		for _, p := range players {
			go func() {
				for p.step() {
				}
				done <- p
			}()
		}
		sent, bytes := 0, 0
		for range players {
			select {
			case p := <-done:
				if p.last.GetWinner() == pb.Team_TEAM_UNSPECIFIED || p.acked != p.sequence {
					t.Fatalf("%s %v: final view winner %v acked %d sent %d", mode, p.seat, p.last.GetWinner(), p.acked, p.sequence)
				}
				sent += p.sent
				bytes += p.bytes
			case <-time.After(60 * time.Second):
				t.Fatalf("%s: match did not finish", mode)
			}
		}
		reply := make(chan replay.Record, 1)
		r.Record(reply)
		record := <-reply
		if err := replay.Verify(record); err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if len(record.Intents) != sent {
			t.Fatalf("%s: record holds %d intents, clients sent %d", mode, len(record.Intents), sent)
		}
		average := float64(bytes) / float64(sent)
		t.Logf("%s: %d intents, %.1f bytes per intent on average", mode, sent, average)
		if average >= 64 {
			t.Fatalf("%s: intents average %.1f bytes, must stay under 64", mode, average)
		}
	}
}

func TestSeatTakenAndResume(t *testing.T) {
	server, _ := roomServer(t, rules.Mode1v1, 3)
	first := join(t, server, rules.SeatX, false)
	if view := first.read().GetMatchView(); view.GetState() != pb.MatchState_MATCH_STATE_LOADING || view.GetSeat() != pb.Seat_SEAT_X || view.GetLoadingDeadlineUnixMs() == 0 {
		t.Fatalf("loading view %v", view)
	}
	second := join(t, server, rules.SeatX, false)
	_, _, err := second.socket.Read(second.ctx)
	if websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatalf("second connection for the same seat: %v", err)
	}
	first.socket.Close(websocket.StatusNormalClosure, "")
	resumed := join(t, server, rules.SeatX, false)
	if view := resumed.read().GetMatchView(); view.GetState() != pb.MatchState_MATCH_STATE_LOADING {
		t.Fatalf("after leaving, joining again: %v", view)
	}
	other := join(t, server, rules.SeatY, false)
	views := map[pb.Seat]*pb.MatchView{}
	for _, p := range []*player{resumed, other} {
		for views[p.seat] == nil || views[p.seat].GetState() == pb.MatchState_MATCH_STATE_LOADING {
			views[p.seat] = p.read().GetMatchView()
		}
	}
	if views[pb.Seat_SEAT_X].GetState() != pb.MatchState_MATCH_STATE_OPENING || len(views[pb.Seat_SEAT_X].GetHand()) != 7 {
		t.Fatalf("opening view %v", views[pb.Seat_SEAT_X])
	}
	resumed.socket.CloseNow()
	back := join(t, server, rules.SeatX, true)
	view := back.read().GetMatchView()
	if view.GetState() != pb.MatchState_MATCH_STATE_OPENING || !proto.Equal(view.GetBoard(), views[pb.Seat_SEAT_X].GetBoard()) || len(view.GetHand()) != 7 {
		t.Fatalf("resumed view %v", view)
	}
	if view.GetSeats()[0].GetController() != pb.Controller_CONTROLLER_PLAYER {
		t.Fatalf("resumed seat is controlled by %v", view.GetSeats()[0].GetController())
	}
}

func TestViewMessageMirrorsTheRoomView(t *testing.T) {
	deadline := time.UnixMilli(1761480020000)
	view := room.View{
		State:         room.StatePlaying,
		Seat:          rules.SeatX,
		Mode:          rules.Mode1v1,
		Table:         "table.night_market",
		HandsPlayed:   1,
		Scores:        rules.TeamPoints{45, 20},
		CarryOver:     3,
		Hand:          []rules.Tile{rules.MustTile(0, 4), rules.MustTile(2, 6), rules.MustTile(5, 5)},
		HandSizes:     [6]int{4: 3, 5: 5},
		BoneyardSize:  6,
		Current:       rules.SeatX,
		OpeningTeam:   rules.TeamB,
		TurnDeadline:  deadline,
		Controllers:   [6]rules.Controller{5: rules.ControllerBot},
		LegalMoves:    []rules.Move{{Tile: rules.MustTile(0, 4), End: rules.EndLeft}},
		LastResult:    rules.HandResult{Pips: rules.TeamPoints{0, 23}, Winner: rules.TeamA, HasWinner: true, Awarded: rules.TeamPoints{25, 0}, CarryOver: 3, Scores: rules.TeamPoints{45, 20}, NextOpener: rules.TeamA},
		HasLastResult: true,
	}
	board := rules.NewBoard(rules.MustTile(6, 6))
	for _, move := range []rules.Move{{Tile: rules.MustTile(4, 6), End: rules.EndLeft}, {Tile: rules.MustTile(1, 6), End: rules.EndRight}, {Tile: rules.MustTile(3, 6), End: rules.EndTop}} {
		if err := board.Place(move.Tile, move.End); err != nil {
			t.Fatal(err)
		}
	}
	view.Board = room.BoardView{Line: board.Line(), Top: board.Top(), Bottom: board.Bottom(), OpenEnds: board.OpenEnds()}
	view.Board.Spinner, view.Board.HasSpinner = board.Spinner()
	got := ViewMessage(view)
	want := &pb.Envelope{}
	if err := proto.Unmarshal(mustRead(t, "../../../proto/testdata/frames/match_view.bin"), want); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, want.GetMatchView()) {
		t.Fatalf("view message differs\n got %v\nwant %v", got, want.GetMatchView())
	}
	waiting := ViewMessage(room.View{State: room.StateWaiting, Seat: rules.SeatNorth, Mode: rules.Mode2v2})
	if err := proto.Unmarshal(mustRead(t, "../../../proto/testdata/frames/match_view_waiting.bin"), want); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(waiting, want.GetMatchView()) {
		t.Fatalf("waiting view differs\n got %v\nwant %v", waiting, want.GetMatchView())
	}
}

func TestDecodeIntentRefusesBadTiles(t *testing.T) {
	bad := []*pb.Envelope{
		intentFrame(1, &pb.Envelope{Payload: &pb.Envelope_PlaceTileIntent{PlaceTileIntent: &pb.PlaceTileIntent{}}}),
		place(2, 3, 2, pb.BoardEnd_BOARD_END_LEFT),
		place(3, 0, 7, pb.BoardEnd_BOARD_END_LEFT),
		{ProtocolVersion: 1, Sequence: 4},
		intentFrame(5, &pb.Envelope{Payload: &pb.Envelope_ClientHello{ClientHello: &pb.ClientHello{}}}),
	}
	for _, frame := range bad {
		if intent, ok := DecodeIntent(frame); ok {
			t.Errorf("%v decoded as %+v", frame, intent)
		}
	}
}

func TestRoomSessionDeliversRejectionsAsErrors(t *testing.T) {
	_, r := roomServer(t, rules.Mode1v1, 1)
	conn := &Conn{queue: make(chan *pb.Envelope, 4), closed: make(chan struct{})}
	session := JoinRoom(r, rules.SeatX, conn)
	session.Deliver(room.Message{View: room.View{State: room.StatePlaying, Mode: rules.Mode1v1, Acked: 4}, Rejected: true, Sequence: 5, Reason: errors.New("no")})
	first, second := <-conn.queue, <-conn.queue
	if first.GetError().GetCode() != pb.ErrorCode_ERROR_CODE_ILLEGAL_INTENT || first.GetError().GetRejectedSequence() != 5 || first.GetAckedSequence() != 4 {
		t.Fatalf("rejection frame %v", first)
	}
	if second.GetMatchView() == nil || second.GetAckedSequence() != 4 || session.Acked() != 4 {
		t.Fatalf("view after the rejection %v", second)
	}
	session.Deliver(room.Message{Rejected: true, Reason: room.ErrSeatTaken})
	select {
	case <-conn.Done():
	default:
		t.Fatal("a join rejection must close the connection")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
