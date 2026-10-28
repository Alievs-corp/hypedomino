package app

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster/redistest"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestMain(m *testing.M) {
	os.Exit(redistest.Main(m))
}

func freeAddr(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	return addr
}

func startMember(t *testing.T, prefix, id string) (*running, *cluster.Cluster) {
	t.Helper()
	client := redistest.Client(t)
	config := DefaultConfig()
	config.Cluster.NodeID, config.Cluster.KeyPrefix = id, prefix
	config.Cluster.Heartbeat, config.Cluster.NodeTTL = 50*time.Millisecond, 2*time.Second
	config.Cluster.Capacity = 4
	config.DrainTimeout = 5 * time.Second
	config.ListenAddr = freeAddr(t)
	config.Cluster.Addr = config.ListenAddr
	members, err := cluster.New(config.Cluster, client, nil)
	if err != nil {
		t.Fatal(err)
	}
	logs := &strings.Builder{}
	handler := &seatHandler{}
	a := New(config, NewLogger(logs, "info"), handler)
	a.Join(members)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	var addr string
	select {
	case addr = <-a.Addr():
	case err := <-done:
		t.Fatal(err)
	}
	r := &running{t: t, app: a, handler: handler, url: "http://" + addr, stop: cancel, done: done}
	t.Cleanup(func() {
		cancel()
		if !r.ended {
			r.wait()
		}
	})
	return r, members
}

func (r *running) loadz() map[string]json.RawMessage {
	r.t.Helper()
	resp, err := http.Get(r.url + "/loadz")
	if err != nil {
		r.t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		r.t.Fatalf("loadz answered %d", resp.StatusCode)
	}
	var report map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		r.t.Fatal(err)
	}
	return report
}

func TestLoadEndpointReportsTheNodeAndTheCluster(t *testing.T) {
	prefix := redistest.Prefix(t)
	a, _ := startMember(t, prefix, "a")
	b, _ := startMember(t, prefix, "b")
	a.openRoom(rules.Mode1v1)
	a.dial(rules.SeatX)
	var report map[string]json.RawMessage
	summary := eventually(t, func() (cluster.Summary, bool) {
		report = b.loadz()
		var s cluster.Summary
		if err := json.Unmarshal(report["cluster"], &s); err != nil {
			t.Fatal(err)
		}
		return s, s.Nodes == 2 && s.Rooms == 1 && s.Connections == 1
	})
	if summary.Capacity != 8 || summary.Advice != cluster.AdviceIn {
		t.Fatalf("cluster summary %+v", summary)
	}
	var self cluster.Node
	if err := json.Unmarshal(report["node"], &self); err != nil {
		t.Fatal(err)
	}
	if self.ID != "b" || self.Load.Rooms != 0 || self.Load.Draining {
		t.Fatalf("node b reported %+v", self)
	}
	plain := start(t, DefaultConfig())
	if _, ok := plain.loadz()["cluster"]; ok {
		t.Fatal("a node outside a cluster reported a cluster")
	}
}

func TestADrainingNodeKeepsItsMembershipAndAcceptsRelayedConnections(t *testing.T) {
	prefix := redistest.Prefix(t)
	a, members := startMember(t, prefix, "a")
	a.openRoom(rules.Mode1v1)
	north := a.dial(rules.SeatX)
	if _, err := north.View(); err != nil {
		t.Fatal(err)
	}
	a.stop()
	deadline := time.Now().Add(5 * time.Second)
	for status, _ := a.get("/readyz"); status != http.StatusServiceUnavailable && time.Now().Before(deadline); status, _ = a.get("/readyz") {
		time.Sleep(10 * time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + strings.TrimPrefix(a.url, "http") + "/ws"
	if _, err := botclient.Dial(ctx, wsURL, botclient.Options{Token: rules.SeatY.String(), AppVersion: "1.0.0"}); err == nil {
		t.Fatal("a draining node accepted a public connection")
	}
	header := http.Header{cluster.RelayHeader: []string{"b"}}
	socket, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("a draining node refused a relayed connection: %v", err)
	}
	socket.CloseNow()
	self := eventually(t, func() (cluster.Node, bool) {
		node, err := members.Node(ctx, "a")
		return node, err == nil && node.Load.Draining
	})
	if self.Load.Rooms != 1 {
		t.Fatalf("draining node in the registry: %+v", self)
	}
	north.Drop()
	if err := a.wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := members.Node(ctx, "a"); err == nil {
		t.Fatal("a stopped node stayed in the registry")
	}
}

func eventually[T any](t *testing.T, probe func() (T, bool)) T {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		value, ok := probe()
		if ok {
			return value
		}
		if time.Now().After(deadline) {
			t.Fatalf("condition not met in time: %+v", value)
			return value
		}
		time.Sleep(20 * time.Millisecond)
	}
}
