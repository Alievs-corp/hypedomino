package gateway

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/replay"
)

type Conn struct {
	socket  *websocket.Conn
	options Options
	queue   chan *pb.Envelope
	closed  chan struct{}
	once    sync.Once
	status  websocket.StatusCode
	reason  string
	session Session
	bucket  bucket
	stop    context.CancelFunc
	served  chan struct{}
}

func newConn(socket *websocket.Conn, options Options) *Conn {
	socket.SetReadLimit(int64(options.Config.MaxFrameBytes))
	return &Conn{
		socket:  socket,
		options: options,
		queue:   make(chan *pb.Envelope, options.Config.WriteQueue),
		closed:  make(chan struct{}),
		served:  make(chan struct{}),
		bucket:  newBucket(options.Config.IntentsPerSecond, options.Config.IntentBurst, options.Clock.Now()),
	}
}

func (c *Conn) Send(frame *pb.Envelope) error {
	frame.ProtocolVersion = ProtocolVersion
	select {
	case <-c.closed:
		return ErrClosed
	default:
	}
	select {
	case c.queue <- frame:
		return nil
	case <-c.closed:
		return ErrClosed
	default:
		c.Close(websocket.StatusPolicyViolation, "slow consumer")
		return ErrSlowConsumer
	}
}

func (c *Conn) Close(status websocket.StatusCode, reason string) {
	c.once.Do(func() {
		c.status = status
		c.reason = reason
		close(c.closed)
	})
}

func (c *Conn) Done() <-chan struct{} {
	return c.closed
}

func (c *Conn) serve(ctx context.Context) {
	defer close(c.served)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.stop = cancel
	var writer sync.WaitGroup
	writer.Add(1)
	go func() {
		defer writer.Done()
		c.writeLoop(ctx)
	}()
	socketClosed := make(chan struct{})
	go func() {
		defer close(socketClosed)
		<-c.closed
		writer.Wait()
		c.closeSocket()
	}()
	c.readLoop(ctx)
	c.Close(websocket.StatusNormalClosure, "")
	if c.session != nil {
		c.session.Close()
	}
	<-socketClosed
}

func (c *Conn) closeSocket() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.socket.Close(c.status, c.reason)
	}()
	select {
	case <-done:
	case <-time.After(c.options.Config.CloseTimeout):
		c.stop()
		<-done
	}
}

func (c *Conn) writeLoop(ctx context.Context) {
	var buffer []byte
	for {
		select {
		case <-c.closed:
			c.flush(ctx, buffer)
			return
		case frame := <-c.queue:
			data, ok := c.writeFrame(ctx, buffer, frame)
			if !ok {
				return
			}
			buffer = data
		}
	}
}

func (c *Conn) flush(ctx context.Context, buffer []byte) {
	for {
		select {
		case frame := <-c.queue:
			data, ok := c.writeFrame(ctx, buffer, frame)
			if !ok {
				return
			}
			buffer = data
		default:
			return
		}
	}
}

func (c *Conn) writeFrame(ctx context.Context, buffer []byte, frame *pb.Envelope) ([]byte, bool) {
	data, err := proto.MarshalOptions{}.MarshalAppend(buffer[:0], frame)
	if err != nil {
		c.Close(websocket.StatusInternalError, "encode")
		return buffer, false
	}
	if err := c.write(ctx, data); err != nil {
		c.Close(websocket.StatusAbnormalClosure, "write")
		return data, false
	}
	return data, true
}

func (c *Conn) write(ctx context.Context, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, c.options.Config.WriteTimeout)
	defer cancel()
	return c.socket.Write(ctx, websocket.MessageBinary, data)
}

func (c *Conn) readLoop(ctx context.Context) {
	var buffer bytes.Buffer
	frame := &pb.Envelope{}
	if !c.handshake(ctx, &buffer, frame) {
		return
	}
	for {
		if !c.readFrame(ctx, &buffer, frame) {
			return
		}
		if !c.dispatch(frame) {
			return
		}
	}
}

func (c *Conn) handshake(ctx context.Context, buffer *bytes.Buffer, frame *pb.Envelope) bool {
	timer := time.AfterFunc(c.options.Config.HelloTimeout, func() { c.Close(websocket.StatusPolicyViolation, "hello timeout") })
	defer timer.Stop()
	if !c.readFrame(ctx, buffer, frame) {
		return false
	}
	hello := frame.GetClientHello()
	switch {
	case hello == nil:
		c.Close(websocket.StatusPolicyViolation, "hello expected")
		return false
	case frame.GetProtocolVersion() != ProtocolVersion:
		c.Send(ErrorFrame(pb.ErrorCode_ERROR_CODE_UNSUPPORTED_PROTOCOL, frame.GetSequence(), 0))
		c.Close(websocket.StatusProtocolError, "unsupported protocol")
		return false
	case !VersionAtLeast(hello.GetAppVersion(), c.options.Config.MinimumAppVersion):
		c.Send(c.serverHello(pb.HelloStatus_HELLO_STATUS_UPDATE_REQUIRED, 0))
		c.Close(websocket.StatusNormalClosure, "update required")
		return false
	}
	session, err := c.options.Handler.Connect(ctx, c, hello)
	switch {
	case errors.Is(err, ErrUnauthenticated):
		c.Send(ErrorFrame(pb.ErrorCode_ERROR_CODE_UNAUTHENTICATED, frame.GetSequence(), 0))
		c.Close(websocket.StatusPolicyViolation, "unauthenticated")
		return false
	case err != nil:
		c.Close(websocket.StatusInternalError, "connect")
		return false
	}
	c.session = session
	if c.Send(c.serverHello(pb.HelloStatus_HELLO_STATUS_ACCEPTED, session.Acked())) != nil {
		return false
	}
	session.Attach()
	go c.pingLoop(ctx)
	return true
}

func (c *Conn) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(c.options.Config.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.closed:
			return
		case <-ticker.C:
			if !c.ping(ctx) {
				c.Close(websocket.StatusGoingAway, "no pong")
				return
			}
		}
	}
}

func (c *Conn) ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, c.options.Config.PongTimeout)
	defer cancel()
	return c.socket.Ping(ctx) == nil
}

func (c *Conn) readFrame(ctx context.Context, buffer *bytes.Buffer, frame *pb.Envelope) bool {
	kind, reader, err := c.socket.Reader(ctx)
	if err != nil {
		c.Close(websocket.StatusNormalClosure, "")
		return false
	}
	if kind != websocket.MessageBinary {
		c.Close(websocket.StatusUnsupportedData, "binary frames only")
		return false
	}
	buffer.Reset()
	if _, err := io.Copy(buffer, reader); err != nil {
		c.Close(websocket.StatusMessageTooBig, "frame too big")
		return false
	}
	if err := (proto.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(buffer.Bytes(), frame); err != nil {
		c.Close(websocket.StatusPolicyViolation, "malformed frame")
		return false
	}
	return true
}

func (c *Conn) dispatch(frame *pb.Envelope) bool {
	sequence := frame.GetSequence()
	intent, ok := DecodeIntent(frame)
	if !ok {
		return c.Send(ErrorFrame(pb.ErrorCode_ERROR_CODE_ILLEGAL_INTENT, sequence, c.session.Acked())) == nil
	}
	if !c.bucket.take(c.options.Clock.Now()) {
		return c.Send(ErrorFrame(pb.ErrorCode_ERROR_CODE_RATE_LIMITED, sequence, c.session.Acked())) == nil
	}
	c.session.Intent(sequence, intent)
	return true
}

func DecodeIntent(frame *pb.Envelope) (replay.Intent, bool) {
	switch payload := frame.GetPayload().(type) {
	case *pb.Envelope_PlaceTileIntent:
		return decodePlacement(payload.PlaceTileIntent)
	case *pb.Envelope_EndTurnIntent:
		return replay.Intent{Kind: replay.KindEndTurn}, true
	case *pb.Envelope_DrawIntent:
		return replay.Intent{Kind: replay.KindDraw}, true
	case *pb.Envelope_PassIntent:
		return replay.Intent{Kind: replay.KindPass}, true
	}
	return replay.Intent{}, false
}

func decodePlacement(intent *pb.PlaceTileIntent) (replay.Intent, bool) {
	tile, ok := tileOf(intent.GetTile())
	if !ok {
		return replay.Intent{}, false
	}
	if intent.GetEnd() == pb.BoardEnd_BOARD_END_UNSPECIFIED {
		return replay.Intent{Kind: replay.KindOpen, Tile: tile}, true
	}
	end, ok := endOf(intent.GetEnd())
	if !ok {
		return replay.Intent{}, false
	}
	return replay.Intent{Kind: replay.KindPlace, Tile: tile, End: end}, true
}

type bucket struct {
	tokens float64
	rate   float64
	burst  float64
	last   time.Time
}

func newBucket(perSecond, burst int, now time.Time) bucket {
	return bucket{tokens: float64(burst), rate: float64(perSecond), burst: float64(burst), last: now}
}

func (b *bucket) take(now time.Time) bool {
	if b.rate <= 0 {
		return true
	}
	b.tokens = min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
