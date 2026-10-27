package custom

import (
	"errors"
	"sync"

	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var (
	ErrUnknownCode   = errors.New("custom: unknown room code")
	ErrNotCreator    = errors.New("custom: only the creator may do that")
	ErrSeatTaken     = errors.New("custom: seat taken")
	ErrNotInRoom     = errors.New("custom: account is not in the room")
	ErrAlreadyInRoom = errors.New("custom: account is already in the room")
	ErrFull          = errors.New("custom: no free seat")
	ErrNotReady      = errors.New("custom: every seat must be filled and ready")
	ErrStarted       = errors.New("custom: the match has started")
	ErrInvalidSeat   = errors.New("custom: seat outside the mode")
	ErrInvalidMode   = errors.New("custom: invalid mode")
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const CodeLength = 6

type Random interface {
	Uint64() uint64
}

type Opener func(options room.Options) (*room.Room, error)

type Lobby struct {
	random Random
	opener Opener
	mu     sync.Mutex
	rooms  map[string]*Room
}

func NewLobby(random Random, opener Opener) *Lobby {
	return &Lobby{random: random, opener: opener, rooms: map[string]*Room{}}
}

func (l *Lobby) Create(creator string, mode rules.Mode) (*Room, error) {
	if mode != rules.Mode1v1 && mode != rules.Mode2v2 {
		return nil, ErrInvalidMode
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	code := l.code()
	for l.rooms[code] != nil {
		code = l.code()
	}
	r := &Room{lobby: l, code: code, creator: creator, mode: mode}
	r.seats[mode.Seats()[0]] = creator
	l.rooms[code] = r
	return r, nil
}

func (l *Lobby) code() string {
	code := make([]byte, CodeLength)
	value := l.random.Uint64()
	for index := range code {
		code[index] = codeAlphabet[value%uint64(len(codeAlphabet))]
		value /= uint64(len(codeAlphabet))
	}
	return string(code)
}

func (l *Lobby) Find(code string) (*Room, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.rooms[code]
	if !ok {
		return nil, ErrUnknownCode
	}
	return r, nil
}

func (l *Lobby) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.rooms)
}

func (l *Lobby) close(code string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.rooms, code)
}

type Room struct {
	lobby   *Lobby
	code    string
	creator string
	mode    rules.Mode
	mu      sync.Mutex
	seats   [6]string
	ready   [6]bool
	table   *room.Room
	closed  bool
}

type SeatState struct {
	Seat    rules.Seat
	Account string
	Ready   bool
}

type State struct {
	Code    string
	Creator string
	Mode    rules.Mode
	Seats   []SeatState
	Started bool
	Closed  bool
}

func (r *Room) Code() string {
	return r.code
}

func (r *Room) Table() *room.Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.table
}

func (r *Room) State() State {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := State{Code: r.code, Creator: r.creator, Mode: r.mode, Started: r.table != nil, Closed: r.closed}
	for _, seat := range r.mode.Seats() {
		state.Seats = append(state.Seats, SeatState{Seat: seat, Account: r.seats[seat], Ready: r.ready[seat]})
	}
	return state
}

func (r *Room) seatOf(account string) (rules.Seat, bool) {
	for _, seat := range r.mode.Seats() {
		if r.seats[seat] == account {
			return seat, true
		}
	}
	return 0, false
}

func (r *Room) valid(seat rules.Seat) bool {
	for _, candidate := range r.mode.Seats() {
		if candidate == seat {
			return true
		}
	}
	return false
}

func (r *Room) freeSeat() (rules.Seat, bool) {
	for _, seat := range r.mode.Seats() {
		if r.seats[seat] == "" {
			return seat, true
		}
	}
	return 0, false
}

func (r *Room) Join(account string) (rules.Seat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.closed:
		return 0, ErrUnknownCode
	case r.table != nil:
		return 0, ErrStarted
	}
	if _, seated := r.seatOf(account); seated {
		return 0, ErrAlreadyInRoom
	}
	seat, ok := r.freeSeat()
	if !ok {
		return 0, ErrFull
	}
	r.seats[seat] = account
	return seat, nil
}

func (r *Room) Sit(account string, seat rules.Seat) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.table != nil:
		return ErrStarted
	case !r.valid(seat):
		return ErrInvalidSeat
	}
	from, seated := r.seatOf(account)
	switch {
	case !seated:
		return ErrNotInRoom
	case from == seat:
		return nil
	case r.seats[seat] != "":
		return ErrSeatTaken
	}
	r.seats[from], r.ready[from] = "", false
	r.seats[seat], r.ready[seat] = account, false
	return nil
}

func (r *Room) Ready(account string, ready bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.table != nil {
		return ErrStarted
	}
	seat, seated := r.seatOf(account)
	if !seated {
		return ErrNotInRoom
	}
	r.ready[seat] = ready
	return nil
}

func (r *Room) Leave(account string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	seat, seated := r.seatOf(account)
	if !seated {
		return ErrNotInRoom
	}
	if r.table != nil {
		r.table.Leave(seat)
		return nil
	}
	if account == r.creator {
		r.dissolve()
		return nil
	}
	r.seats[seat], r.ready[seat] = "", false
	return nil
}

func (r *Room) dissolve() {
	r.closed = true
	r.seats, r.ready = [6]string{}, [6]bool{}
	r.lobby.close(r.code)
}

func (r *Room) Remove(creator, account string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if creator != r.creator {
		return ErrNotCreator
	}
	seat, seated := r.seatOf(account)
	switch {
	case !seated:
		return ErrNotInRoom
	case account == r.creator:
		return ErrNotCreator
	}
	r.seats[seat], r.ready[seat] = "", false
	if r.table != nil {
		r.table.Kick(seat)
	}
	return nil
}

func (r *Room) Add(creator, account string, seat rules.Seat) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case creator != r.creator:
		return ErrNotCreator
	case r.closed:
		return ErrUnknownCode
	case !r.valid(seat):
		return ErrInvalidSeat
	case r.seats[seat] != "":
		return ErrSeatTaken
	}
	if _, seated := r.seatOf(account); seated {
		return ErrAlreadyInRoom
	}
	r.seats[seat] = account
	if r.table != nil {
		r.table.Admit(seat)
	}
	return nil
}

func (r *Room) Start(creator string, options room.Options) (*room.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case creator != r.creator:
		return nil, ErrNotCreator
	case r.table != nil:
		return nil, ErrStarted
	}
	for _, seat := range r.mode.Seats() {
		if r.seats[seat] == "" || !r.ready[seat] {
			return nil, ErrNotReady
		}
	}
	options.Kind, options.Mode = room.KindCustom, r.mode
	table, err := r.lobby.opener(options)
	if err != nil {
		return nil, err
	}
	r.table = table
	return table, nil
}

func (r *Room) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dissolve()
}
