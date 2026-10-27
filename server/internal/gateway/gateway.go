package gateway

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/coder/websocket"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
)

const ProtocolVersion = 1

var (
	ErrUnauthenticated = errors.New("gateway: session token rejected")
	ErrSlowConsumer    = errors.New("gateway: write queue full")
	ErrClosed          = errors.New("gateway: connection closed")
)

type Config struct {
	MinimumAppVersion    string
	CatalogVersion       string
	MaxFrameBytes        int
	WriteQueue           int
	HelloTimeout         time.Duration
	WriteTimeout         time.Duration
	CloseTimeout         time.Duration
	PingInterval         time.Duration
	PongTimeout          time.Duration
	CompressionThreshold int
	IntentsPerSecond     int
	IntentBurst          int
}

func DefaultConfig() Config {
	return Config{
		MinimumAppVersion:    "1.0.0",
		MaxFrameBytes:        1024,
		WriteQueue:           32,
		HelloTimeout:         5 * time.Second,
		WriteTimeout:         5 * time.Second,
		CloseTimeout:         2 * time.Second,
		PingInterval:         15 * time.Second,
		PongTimeout:          5 * time.Second,
		CompressionThreshold: 256,
		IntentsPerSecond:     10,
		IntentBurst:          20,
	}
}

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

type Session interface {
	Acked() uint64
	Attach()
	Intent(sequence uint64, intent replay.Intent)
	Close()
}

type Handler interface {
	Connect(ctx context.Context, conn *Conn, hello *pb.ClientHello) (Session, error)
}

type HandlerFunc func(ctx context.Context, conn *Conn, hello *pb.ClientHello) (Session, error)

func (f HandlerFunc) Connect(ctx context.Context, conn *Conn, hello *pb.ClientHello) (Session, error) {
	return f(ctx, conn, hello)
}

type Options struct {
	Config  Config
	Handler Handler
	Clock   Clock
	Origins []string
}

type Gateway struct {
	options Options
}

func New(options Options) *Gateway {
	if options.Clock == nil {
		options.Clock = systemClock{}
	}
	return &Gateway{options: options}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:       g.options.Origins,
		CompressionMode:      websocket.CompressionNoContextTakeover,
		CompressionThreshold: g.options.Config.CompressionThreshold,
	})
	if err != nil {
		return
	}
	conn := newConn(socket, g.options)
	conn.serve(r.Context())
}
