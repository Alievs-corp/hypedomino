package gateway

import (
	"sync/atomic"

	"github.com/coder/websocket"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type RoomSession struct {
	room   *room.Room
	seat   rules.Seat
	conn   *Conn
	resume bool
	acked  atomic.Uint64
}

func JoinRoom(r *room.Room, seat rules.Seat, conn *Conn) *RoomSession {
	return &RoomSession{room: r, seat: seat, conn: conn}
}

func ResumeRoom(r *room.Room, seat rules.Seat, conn *Conn, acked uint64) *RoomSession {
	s := &RoomSession{room: r, seat: seat, conn: conn, resume: true}
	s.acked.Store(acked)
	return s
}

func (s *RoomSession) Acked() uint64 {
	return s.acked.Load()
}

func (s *RoomSession) Attach() {
	if s.resume {
		s.room.Reconnect(s.seat, s)
		return
	}
	s.room.Join(s.seat, s)
}

func (s *RoomSession) Intent(sequence uint64, intent replay.Intent) {
	s.room.Intent(s.seat, sequence, intent)
}

func (s *RoomSession) Close() {
	s.room.Disconnect(s.seat, s)
}

func (s *RoomSession) Deliver(message room.Message) {
	if message.Rejected && message.Sequence == 0 {
		s.conn.Close(websocket.StatusPolicyViolation, message.Reason.Error())
		return
	}
	s.acked.Store(message.View.Acked)
	if message.Rejected {
		s.conn.Send(errorFrame(pb.ErrorCode_ERROR_CODE_ILLEGAL_INTENT, message.Sequence, message.View.Acked))
	}
	s.conn.Send(&pb.Envelope{
		AckedSequence: message.View.Acked,
		Payload:       &pb.Envelope_MatchView{MatchView: ViewMessage(message.View)},
	})
}
