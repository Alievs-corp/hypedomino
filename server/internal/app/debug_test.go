package app

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestMovesOfALiveMatchReachTheMetrics(t *testing.T) {
	r := start(t, DefaultConfig())
	r.openRoom(rules.Mode1v1)
	x, y := r.dial(rules.SeatX), r.dial(rules.SeatY)
	finished := make(chan error, 2)
	for _, c := range []*botclient.Client{x, y} {
		go func() {
			_, err := c.PlayUntil(botclient.Finished)
			finished <- err
		}()
	}
	for range 2 {
		if err := <-finished; err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	r.app.debugHandler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	count := regexp.MustCompile(`(?m)^move_roundtrip_seconds_count (\d+)$`).FindStringSubmatch(w.Body.String())
	if w.Code != http.StatusOK || count == nil || count[1] == "0" {
		t.Fatalf("status %d metrics:\n%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `move_roundtrip_seconds_bucket{le="0.1"}`) {
		t.Fatal("no bucket at the 100 ms budget")
	}
	for _, gauge := range []string{`(?m)^go_heap_live_bytes [1-9]`, `(?m)^go_goroutines [1-9]`, `(?m)^hd_rooms \d+$`, `(?m)^hd_connections \d+$`} {
		if !regexp.MustCompile(gauge).MatchString(w.Body.String()) {
			t.Fatalf("missing %s in\n%s", gauge, w.Body.String())
		}
	}
	profile := httptest.NewRecorder()
	r.app.debugHandler().ServeHTTP(profile, httptest.NewRequest(http.MethodGet, "/debug/pprof/heap", nil))
	if profile.Code != http.StatusOK {
		t.Fatalf("heap profile status %d", profile.Code)
	}
	if status, _ := r.get("/metrics"); status != http.StatusNotFound {
		t.Fatalf("metrics on the public listener: %d", status)
	}
}

func TestRunFailsOnABadDebugAddress(t *testing.T) {
	config := DefaultConfig()
	config.ListenAddr = "127.0.0.1:0"
	config.DebugAddr = "256.0.0.1:1"
	if err := New(config, NewLogger(&strings.Builder{}, "info"), &seatHandler{}).Run(t.Context()); err == nil {
		t.Fatal("Run started without its debug listener")
	}
}
