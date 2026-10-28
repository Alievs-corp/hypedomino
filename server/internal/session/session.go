package session

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var (
	ErrUnknownSession = errors.New("session: unknown token")
	ErrAlreadySeated  = errors.New("session: already seated in a room")
)

type Config struct {
	ReconnectGrace time.Duration
}

func DefaultConfig() Config {
	return Config{ReconnectGrace: 60 * time.Second}
}

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type Random interface {
	Uint64() uint64
}

type Session struct {
	token      string
	account    string
	room       *room.Room
	seat       rules.Seat
	joined     bool
	acked      uint64
	active     *connection
	generation uint64
}

type Router interface {
	Claim(ctx context.Context, token string) error
	Release(ctx context.Context, token string) error
}

type Manager struct {
	config   Config
	clock    Clock
	random   Random
	router   Router
	mu       sync.Mutex
	sessions map[string]*Session
}

func New(config Config, clock Clock, random Random) *Manager {
	return &Manager{config: config, clock: clock, random: random, sessions: map[string]*Session{}}
}

func (m *Manager) Route(router Router) {
	m.router = router
}

func (m *Manager) claim(token string) error {
	if m.router == nil {
		return nil
	}
	return m.router.Claim(context.Background(), token)
}

func (m *Manager) release(token string) {
	if m.router != nil {
		m.router.Release(context.Background(), token)
	}
}

func (m *Manager) Create(account string) string {
	var raw [16]byte
	for offset := 0; offset < len(raw); offset += 8 {
		value := m.random.Uint64()
		for index := 0; index < 8; index++ {
			raw[offset+index] = byte(value >> (8 * index))
		}
	}
	token := hex.EncodeToString(raw[:])
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = &Session{token: token, account: account}
	return token
}

func (m *Manager) Seat(token string, r *room.Room, seat rules.Seat) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	switch {
	case !ok:
		return ErrUnknownSession
	case s.room != nil:
		return ErrAlreadySeated
	}
	if err := m.claim(token); err != nil {
		return err
	}
	s.room, s.seat, s.joined, s.acked = r, seat, false, 0
	if s.active != nil {
		s.active.attach(s)
	}
	return nil
}

func (m *Manager) Revoke(token string) {
	m.mu.Lock()
	s, ok := m.sessions[token]
	delete(m.sessions, token)
	m.mu.Unlock()
	if ok && s.room != nil {
		m.release(token)
	}
	if ok && s.active != nil {
		s.active.conn.Close(websocket.StatusPolicyViolation, "session revoked")
	}
}

func (m *Manager) Connect(ctx context.Context, conn *gateway.Conn, hello *pb.ClientHello) (gateway.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[hello.GetSessionToken()]
	if !ok {
		return nil, gateway.ErrUnauthenticated
	}
	if previous := s.active; previous != nil {
		previous.conn.Close(websocket.StatusPolicyViolation, "session replaced")
	}
	s.generation++
	c := &connection{manager: m, session: s, conn: conn}
	s.active = c
	return c, nil
}

type connection struct {
	manager *Manager
	session *Session
	conn    *gateway.Conn
	mu      sync.Mutex
	seat    *gateway.RoomSession
}

func (c *connection) Acked() uint64 {
	c.manager.mu.Lock()
	defer c.manager.mu.Unlock()
	return c.session.acked
}

func (c *connection) Attach() {
	c.manager.mu.Lock()
	defer c.manager.mu.Unlock()
	if c.session.room != nil {
		c.attach(c.session)
	}
}

func (c *connection) attach(s *Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.joined {
		c.seat = gateway.ResumeRoom(s.room, s.seat, c.conn, s.acked)
	} else {
		c.seat = gateway.JoinRoom(s.room, s.seat, c.conn)
		s.joined = true
	}
	c.seat.Attach()
}

func (c *connection) Intent(sequence uint64, intent replay.Intent) {
	c.mu.Lock()
	seat := c.seat
	c.mu.Unlock()
	if seat == nil {
		c.conn.Send(gateway.ErrorFrame(pb.ErrorCode_ERROR_CODE_ILLEGAL_INTENT, sequence, c.Acked()))
		return
	}
	seat.Intent(sequence, intent)
}

func (c *connection) Close() {
	c.mu.Lock()
	seat := c.seat
	c.mu.Unlock()
	if seat != nil {
		seat.Close()
	}
	c.manager.mu.Lock()
	defer c.manager.mu.Unlock()
	s := c.session
	if s.active != c {
		return
	}
	s.active = nil
	if seat != nil {
		s.acked = seat.Acked()
	}
	if s.room != nil {
		go c.manager.expire(s, s.generation, c.manager.clock.After(c.manager.config.ReconnectGrace))
	}
}

func (m *Manager) Connected(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	return ok && s.active != nil
}

func (m *Manager) expire(s *Session, generation uint64, due <-chan time.Time) {
	<-due
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.generation != generation || s.active != nil || s.room == nil {
		return
	}
	s.room.Leave(s.seat)
	m.release(s.token)
}
