package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var (
	ErrSeatTaken     = errors.New("room: seat already taken")
	ErrNotSeated     = errors.New("room: seat is empty")
	ErrNotStarted    = errors.New("room: the match has not started")
	ErrFinished      = errors.New("room: the match is over")
	ErrSeatByBot     = errors.New("room: the seat is played by the bot")
	ErrStaleSequence = errors.New("room: sequence already handled")
)

type State uint8

const (
	StateWaiting State = iota
	StateOpening
	StatePlaying
	StateFinished
)

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type Message struct {
	View     View
	Rejected bool
	Sequence uint64
	Reason   error
}

type Sink interface {
	Deliver(Message)
}

type Options struct {
	Mode     rules.Mode
	Config   rules.Config
	Bot      bot.Config
	BotLevel bot.Level
	Seed     uint64
	Clock    Clock
}

type seatState struct {
	sink  Sink
	acked uint64
}

type turnKey struct {
	phase   rules.Phase
	current rules.Seat
	hands   int
}

type command func(*Room)

type Room struct {
	options  Options
	commands chan command
	state    State
	seats    [6]seatState
	recorder *replay.Recorder
	presence rules.Presence
	bots     [6]*bot.Player
	armed    turnKey
	deadline time.Time
	timer    <-chan time.Time
	moves    []rules.Move
	done     chan struct{}
}

func New(options Options) *Room {
	r := &Room{
		options:  options,
		commands: make(chan command, 64),
		presence: rules.NewPresence(options.Config),
		armed:    turnKey{hands: -1},
		moves:    make([]rules.Move, 0, 4*rules.SetSize),
		done:     make(chan struct{}),
	}
	random := rules.NewSeededRandom(options.Seed ^ 0x9e3779b97f4a7c15)
	for _, seat := range options.Mode.Seats() {
		r.bots[seat] = bot.New(options.BotLevel, options.Bot, random)
	}
	return r
}

func (r *Room) Run(ctx context.Context) {
	defer close(r.done)
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-r.commands:
			cmd(r)
		case <-r.timer:
			r.onTimer()
		}
	}
}

func (r *Room) Done() <-chan struct{} {
	return r.done
}

func (r *Room) enqueue(cmd command) {
	select {
	case r.commands <- cmd:
	case <-r.done:
	}
}

func (r *Room) Join(seat rules.Seat, sink Sink) {
	r.enqueue(func(r *Room) { r.join(seat, sink) })
}

func (r *Room) Reconnect(seat rules.Seat, sink Sink) {
	r.enqueue(func(r *Room) { r.reconnect(seat, sink) })
}

func (r *Room) Leave(seat rules.Seat) {
	r.enqueue(func(r *Room) { r.leave(seat) })
}

func (r *Room) Intent(seat rules.Seat, sequence uint64, intent replay.Intent) {
	r.enqueue(func(r *Room) { r.intent(seat, sequence, intent) })
}

func (r *Room) Snapshot(seat rules.Seat, reply chan<- View) {
	r.enqueue(func(r *Room) { reply <- r.view(seat) })
}

func (r *Room) Record(reply chan<- replay.Record) {
	r.enqueue(func(r *Room) {
		if r.recorder == nil {
			reply <- replay.Record{}
			return
		}
		reply <- r.recorder.Record()
	})
}

func (r *Room) join(seat rules.Seat, sink Sink) {
	if !r.seated(seat) || r.seats[seat].sink != nil {
		sink.Deliver(Message{Rejected: true, Reason: ErrSeatTaken})
		return
	}
	r.seats[seat].sink = sink
	if r.state == StateWaiting && r.full() {
		r.recorder = replay.NewRecorder(r.options.Mode, r.options.Config, r.options.Seed)
	}
	r.settle()
}

func (r *Room) seated(seat rules.Seat) bool {
	for _, candidate := range r.options.Mode.Seats() {
		if candidate == seat {
			return true
		}
	}
	return false
}

func (r *Room) full() bool {
	for _, seat := range r.options.Mode.Seats() {
		if r.seats[seat].sink == nil {
			return false
		}
	}
	return true
}

func (r *Room) reconnect(seat rules.Seat, sink Sink) {
	if !r.seated(seat) {
		sink.Deliver(Message{Rejected: true, Reason: ErrNotSeated})
		return
	}
	r.seats[seat].sink = sink
	r.presence.Reconnect(seat)
	r.settle()
}

func (r *Room) leave(seat rules.Seat) {
	if !r.seated(seat) {
		return
	}
	r.seats[seat].sink = nil
	r.presence.Leave(seat)
	r.settle()
}

func (r *Room) intent(seat rules.Seat, sequence uint64, intent replay.Intent) {
	if err := r.admit(seat, sequence); err != nil {
		if !errors.Is(err, ErrStaleSequence) {
			r.reject(seat, sequence, err)
		}
		return
	}
	intent.Seat = seat
	if err := replay.Apply(r.recorder, intent); err != nil {
		r.reject(seat, sequence, err)
		return
	}
	r.seats[seat].acked = sequence
	r.presence.Move(seat)
	r.settle()
}

func (r *Room) admit(seat rules.Seat, sequence uint64) error {
	switch {
	case !r.seated(seat):
		return ErrNotSeated
	case r.recorder == nil:
		return ErrNotStarted
	case r.state == StateFinished:
		return ErrFinished
	case r.presence.Controller(seat) == rules.ControllerBot:
		return ErrSeatByBot
	case sequence <= r.seats[seat].acked:
		return ErrStaleSequence
	}
	return nil
}

func (r *Room) match() *rules.Match {
	return r.recorder.Match
}

func (r *Room) key() turnKey {
	m := r.match()
	return turnKey{phase: m.Phase(), current: m.Current(), hands: m.HandsPlayed()}
}

func (r *Room) actingSeat() (rules.Seat, bool) {
	m := r.match()
	switch m.Phase() {
	case rules.PhasePlaying:
		return m.Current(), true
	case rules.PhaseOpening:
		if opening, forced := m.ForcedOpening(); forced {
			return opening.Seat, true
		}
		var buffer [2]rules.Seat
		return r.options.Mode.TeamSeats(m.OpeningTeam(), buffer[:0])[0], true
	}
	return 0, false
}

func (r *Room) settle() {
	if r.recorder != nil {
		for guard := 0; guard < 8*rules.SetSize; guard++ {
			r.state = stateOf(r.match().Phase())
			if r.state == StateFinished {
				r.timer = nil
				break
			}
			if key := r.key(); key != r.armed {
				r.armed = key
				limit := rules.TurnLimit(r.options.Config)
				r.deadline = r.options.Clock.Now().Add(limit)
				r.timer = r.options.Clock.After(limit)
			}
			seat, acting := r.actingSeat()
			if !acting || r.presence.Controller(seat) != rules.ControllerBot {
				break
			}
			if _, err := r.bots[seat].Act(r.recorder, seat); err != nil {
				break
			}
			r.presence.Move(seat)
		}
	}
	r.broadcast()
}

func stateOf(phase rules.Phase) State {
	switch phase {
	case rules.PhaseOpening:
		return StateOpening
	case rules.PhasePlaying:
		return StatePlaying
	}
	return StateFinished
}

func (r *Room) onTimer() {
	if r.recorder == nil || r.state == StateFinished {
		return
	}
	seat, acting := r.actingSeat()
	if !acting {
		return
	}
	r.presence.Timeout(seat)
	key := r.key()
	for guard := 0; guard < rules.SetSize && r.key() == key; guard++ {
		if _, err := r.bots[seat].Act(r.recorder, seat); err != nil {
			break
		}
	}
	r.settle()
}

func (r *Room) reject(seat rules.Seat, sequence uint64, reason error) {
	if sink := r.seats[seat].sink; sink != nil {
		sink.Deliver(Message{View: r.view(seat), Rejected: true, Sequence: sequence, Reason: fmt.Errorf("rejected intent %d: %w", sequence, reason)})
	}
}

func (r *Room) broadcast() {
	for _, seat := range r.options.Mode.Seats() {
		if sink := r.seats[seat].sink; sink != nil {
			sink.Deliver(Message{View: r.view(seat)})
		}
	}
}
