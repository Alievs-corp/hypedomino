package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
)

func parseLevel(text string) (slog.Level, bool) {
	switch strings.ToLower(text) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	}
	return slog.LevelInfo, false
}

func NewLogger(out io.Writer, level string) *slog.Logger {
	parsed, _ := parseLevel(level)
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: parsed}))
}

type App struct {
	config   Config
	logger   *slog.Logger
	gateway  *gateway.Gateway
	rooms    *Rooms
	mux      *http.ServeMux
	ready    atomic.Bool
	listener net.Listener
	addr     chan string
}

func New(config Config, logger *slog.Logger, handler gateway.Handler) *App {
	a := &App{
		config:  config,
		logger:  logger,
		gateway: gateway.New(gateway.Options{Config: config.Gateway, Handler: handler, Origins: config.Origins}),
		rooms:   NewRooms(),
		mux:     http.NewServeMux(),
		addr:    make(chan string, 1),
	}
	a.mux.HandleFunc("GET /healthz", a.healthz)
	a.mux.HandleFunc("GET /readyz", a.readyz)
	a.mux.HandleFunc("GET /ws", a.ws)
	return a
}

func (a *App) Rooms() *Rooms {
	return a.rooms
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) Addr() <-chan string {
	return a.addr
}

func (a *App) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "ok\n")
}

func (a *App) readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if !a.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "draining\n")
		return
	}
	io.WriteString(w, "ready\n")
}

func (a *App) ws(w http.ResponseWriter, r *http.Request) {
	if !a.ready.Load() {
		http.Error(w, "draining", http.StatusServiceUnavailable)
		return
	}
	a.gateway.ServeHTTP(w, r)
}

func (a *App) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", a.config.ListenAddr)
	if err != nil {
		return err
	}
	a.listener = listener
	a.ready.Store(true)
	server := &http.Server{Handler: a.mux, ReadHeaderTimeout: 10 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	a.logger.Info("listening", "addr", listener.Addr().String())
	a.addr <- listener.Addr().String()
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	return a.shutdown(server)
}

func (a *App) shutdown(server *http.Server) error {
	a.ready.Store(false)
	a.rooms.Close()
	a.logger.Info("draining", "rooms", a.rooms.Len(), "connections", a.gateway.Connections())
	drainCtx, cancelDrain := context.WithTimeout(context.Background(), a.config.DrainTimeout)
	defer cancelDrain()
	if err := a.rooms.Drain(drainCtx); err != nil {
		a.logger.Warn("drain timeout, cancelling rooms", "rooms", a.rooms.Len())
	}
	a.rooms.Stop()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer cancelShutdown()
	if err := a.gateway.Shutdown(shutdownCtx); err != nil {
		a.logger.Warn("connections did not close in time", "connections", a.gateway.Connections())
	}
	err := server.Shutdown(shutdownCtx)
	a.logger.Info("stopped")
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}
