package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
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
	cluster  *cluster.Cluster
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
	a.mux.HandleFunc("GET /loadz", a.loadz)
	return a
}

func (a *App) Handle(pattern string, handler http.Handler) {
	a.mux.Handle(pattern, handler)
}

func (a *App) Join(c *cluster.Cluster) {
	a.cluster = c
	c.Observe(a.load)
}

func (a *App) load() cluster.Load {
	return cluster.Load{Rooms: a.rooms.Len(), Connections: a.gateway.Connections(), Draining: !a.ready.Load()}
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
	relayed := a.cluster != nil && r.Header.Get(cluster.RelayHeader) != ""
	if !a.ready.Load() && !relayed {
		http.Error(w, "draining", http.StatusServiceUnavailable)
		return
	}
	if relayed {
		r = r.WithContext(cluster.MarkRelayed(r.Context()))
	}
	a.gateway.ServeHTTP(w, r)
}

func (a *App) loadz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	report := map[string]any{"load": a.load()}
	if a.cluster != nil {
		report["node"] = a.cluster.Self()
		summary, err := a.cluster.Summarize(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		report["cluster"] = summary
	}
	json.NewEncoder(w).Encode(report)
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
	stopCluster := a.announce()
	defer stopCluster()
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

func (a *App) announce() func() {
	if a.cluster == nil {
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := a.cluster.Run(ctx); err != nil && ctx.Err() == nil {
			a.logger.Error("cluster membership failed", "node", a.cluster.NodeID(), "error", err)
		}
	}()
	return func() {
		cancel()
		<-done
	}
}
