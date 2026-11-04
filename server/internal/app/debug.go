package app

import (
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/metrics"
)

var moveBounds = []time.Duration{
	5 * time.Millisecond, 10 * time.Millisecond, 25 * time.Millisecond, 50 * time.Millisecond, 75 * time.Millisecond,
	100 * time.Millisecond, 150 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, time.Second,
}

func newMoveRoundTrip() *metrics.Histogram {
	return metrics.NewHistogram("move_roundtrip_seconds", "Accepted move intent from receive to the resulting broadcast queued for every seat.", moveBounds)
}

func (a *App) debugHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		a.moves.WriteTo(w)
	})
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

func (a *App) serveDebug() (func(), error) {
	if a.config.DebugAddr == "" {
		return func() {}, nil
	}
	listener, err := net.Listen("tcp", a.config.DebugAddr)
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: a.debugHandler(), ReadHeaderTimeout: 10 * time.Second}
	go server.Serve(listener)
	a.logger.Info("debug listening", "addr", listener.Addr().String())
	return func() { server.Close() }, nil
}
