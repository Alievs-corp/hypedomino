package cluster_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster/redistest"
)

func TestMain(m *testing.M) {
	os.Exit(redistest.Main(m))
}

type loadSource struct {
	mu   sync.Mutex
	load cluster.Load
}

func (s *loadSource) set(load cluster.Load) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load = load
}

func (s *loadSource) get() cluster.Load {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load
}

func config(prefix, id string) cluster.Config {
	c := cluster.DefaultConfig()
	c.NodeID, c.Addr, c.Capacity, c.KeyPrefix = id, id+".internal:8080", 10, prefix
	c.Heartbeat, c.NodeTTL, c.RouteTTL = 50*time.Millisecond, 10*time.Second, time.Minute
	return c
}

func node(t *testing.T, client *redis.Client, prefix, id string) (*cluster.Cluster, *loadSource) {
	t.Helper()
	c, err := cluster.New(config(prefix, id), client, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	source := &loadSource{}
	c.Observe(source.get)
	return c, source
}

func ids(nodes []cluster.Node) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.ID)
	}
	return out
}

func TestConfigValidation(t *testing.T) {
	good := config("hd:", "a")
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*cluster.Config){
		"empty node id":         func(c *cluster.Config) { c.NodeID = " " },
		"zero capacity":         func(c *cluster.Config) { c.Capacity = 0 },
		"ttl below heartbeat":   func(c *cluster.Config) { c.NodeTTL = c.Heartbeat },
		"zero route ttl":        func(c *cluster.Config) { c.RouteTTL = 0 },
		"inverted thresholds":   func(c *cluster.Config) { c.ScaleOutAt, c.ScaleInAt = 0.2, 0.5 },
		"threshold above one":   func(c *cluster.Config) { c.ScaleOutAt = 1.5 },
		"negative in threshold": func(c *cluster.Config) { c.ScaleInAt = -0.1 },
	}
	for name, mutate := range cases {
		bad := good
		mutate(&bad)
		if err := bad.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := cluster.New(cluster.Config{}, nil, nil); err == nil {
		t.Fatal("an empty config built a cluster")
	}
}

func TestAnnounceRegistersAndLeaveRemovesTheNode(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, loadA := node(t, client, prefix, "a")
	b, _ := node(t, client, prefix, "b")
	loadA.set(cluster.Load{Rooms: 3, Connections: 7})
	for _, c := range []*cluster.Cluster{a, b} {
		if err := c.Announce(ctx); err != nil {
			t.Fatal(err)
		}
	}
	nodes, err := a.Nodes(ctx)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("nodes %v, %v", ids(nodes), err)
	}
	got, err := b.Node(ctx, "a")
	if err != nil || got.Addr != "a.internal:8080" || got.Load.Rooms != 3 || got.Load.Connections != 7 || got.Capacity != 10 || got.Free() != 7 {
		t.Fatalf("node a seen as %+v, %v", got, err)
	}
	if err := a.Leave(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Node(ctx, "a"); !errors.Is(err, cluster.ErrNodeGone) {
		t.Fatalf("left node still found: %v", err)
	}
	nodes, err = b.Nodes(ctx)
	if err != nil || len(nodes) != 1 || nodes[0].ID != "b" {
		t.Fatalf("nodes after leave %v, %v", ids(nodes), err)
	}
}

func TestExpiredHeartbeatDropsTheNode(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	short := config(prefix, "a")
	short.NodeTTL = 300 * time.Millisecond
	a, err := cluster.New(short, client, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := node(t, client, prefix, "b")
	if err := a.Announce(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Announce(ctx); err != nil {
		t.Fatal(err)
	}
	if err := b.Leave(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(400 * time.Millisecond)
	nodes, err := b.Nodes(ctx)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("nodes after the ttl %v, %v", ids(nodes), err)
	}
	if n, err := client.SCard(ctx, prefix+"nodes").Result(); err != nil || n != 0 {
		t.Fatalf("stale members kept: %d, %v", n, err)
	}
}

func TestRunHeartbeatsUntilStoppedAndPublishesMembership(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, _ := node(t, client, prefix, "a")
	b, _ := node(t, client, prefix, "b")
	events := make(chan string, 8)
	subCtx, stopSub := context.WithCancel(ctx)
	defer stopSub()
	subDone := make(chan struct{})
	go func() {
		defer close(subDone)
		b.Subscribe(subCtx, func(id string) { events <- id })
	}()
	time.Sleep(100 * time.Millisecond)
	runCtx, stop := context.WithCancel(ctx)
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(runCtx) }()
	select {
	case id := <-events:
		if id != "a" {
			t.Fatalf("membership event %q", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no membership event after Run")
	}
	time.Sleep(500 * time.Millisecond)
	if _, err := b.Node(ctx, "a"); err != nil {
		t.Fatalf("node a not kept alive by the heartbeat: %v", err)
	}
	stop()
	if err := <-runDone; err != nil {
		t.Fatalf("Run returned %v", err)
	}
	if _, err := b.Node(ctx, "a"); !errors.Is(err, cluster.ErrNodeGone) {
		t.Fatalf("node a still registered after Run stopped: %v", err)
	}
	select {
	case id := <-events:
		if id != "a" {
			t.Fatalf("leave event %q", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no membership event after Leave")
	}
	stopSub()
	<-subDone
}

func TestTwoNodesBalanceRoomsByFreeCapacity(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, loadA := node(t, client, prefix, "a")
	b, loadB := node(t, client, prefix, "b")
	rooms := map[string]int{}
	sources := map[string]*loadSource{"a": loadA, "b": loadB}
	announce := func() {
		for _, c := range []*cluster.Cluster{a, b} {
			if err := c.Announce(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
	announce()
	for i := 0; i < 10; i++ {
		chosen, ok, err := a.Least(ctx)
		if err != nil || !ok {
			t.Fatalf("round %d: no node, %v", i, err)
		}
		rooms[chosen.ID]++
		sources[chosen.ID].set(cluster.Load{Rooms: rooms[chosen.ID]})
		announce()
	}
	if rooms["a"] != 5 || rooms["b"] != 5 {
		t.Fatalf("rooms split %v", rooms)
	}
	loadA.set(cluster.Load{Rooms: 5, Draining: true})
	announce()
	chosen, ok, err := b.Least(ctx)
	if err != nil || !ok || chosen.ID != "b" {
		t.Fatalf("draining node chosen: %+v %v %v", chosen, ok, err)
	}
	loadB.set(cluster.Load{Rooms: 10})
	announce()
	if _, ok, err := b.Least(ctx); err != nil || ok {
		t.Fatalf("a full node was chosen, %v", err)
	}
}

func TestSummaryAdvisesScaling(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, loadA := node(t, client, prefix, "a")
	b, loadB := node(t, client, prefix, "b")
	announce := func() {
		for _, c := range []*cluster.Cluster{a, b} {
			if err := c.Announce(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
	summary, err := a.Summarize(ctx)
	if err != nil || summary.Nodes != 0 || summary.Advice != cluster.AdviceOut {
		t.Fatalf("empty cluster %+v, %v", summary, err)
	}
	loadA.set(cluster.Load{Rooms: 1, Connections: 2})
	loadB.set(cluster.Load{Rooms: 2, Connections: 4})
	announce()
	summary, err = a.Summarize(ctx)
	if err != nil || summary.Nodes != 2 || summary.Rooms != 3 || summary.Connections != 6 || summary.Capacity != 20 || summary.Advice != cluster.AdviceIn {
		t.Fatalf("light cluster %+v, %v", summary, err)
	}
	loadA.set(cluster.Load{Rooms: 7})
	loadB.set(cluster.Load{Rooms: 8})
	announce()
	summary, err = a.Summarize(ctx)
	if err != nil || summary.Utilization != 0.75 || summary.Advice != cluster.AdviceOut {
		t.Fatalf("heavy cluster %+v, %v", summary, err)
	}
	loadA.set(cluster.Load{Rooms: 5})
	loadB.set(cluster.Load{Rooms: 5, Draining: true})
	announce()
	summary, err = a.Summarize(ctx)
	if err != nil || summary.Nodes != 1 || summary.Rooms != 5 || summary.Advice != cluster.AdviceHold {
		t.Fatalf("half cluster with a draining node %+v, %v", summary, err)
	}
}

func TestRoutesBelongToTheClaimingNode(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, _ := node(t, client, prefix, "a")
	b, _ := node(t, client, prefix, "b")
	if err := a.Announce(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Claim(ctx, "token-1"); err != nil {
		t.Fatal(err)
	}
	if err := a.Claim(ctx, "token-1"); err != nil {
		t.Fatalf("the owner cannot renew its route: %v", err)
	}
	if err := b.Claim(ctx, "token-1"); !errors.Is(err, cluster.ErrNotOwner) {
		t.Fatalf("another node took the route: %v", err)
	}
	found, err := b.Lookup(ctx, "token-1")
	if err != nil || found.ID != "a" || found.Addr != "a.internal:8080" {
		t.Fatalf("lookup from b: %+v, %v", found, err)
	}
	self, err := a.Lookup(ctx, "token-1")
	if err != nil || self.ID != "a" {
		t.Fatalf("lookup from the owner: %+v, %v", self, err)
	}
	if _, err := b.Lookup(ctx, "token-2"); !errors.Is(err, cluster.ErrNoRoute) {
		t.Fatalf("unknown token: %v", err)
	}
	if err := b.Release(ctx, "token-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Lookup(ctx, "token-1"); err != nil {
		t.Fatalf("a release by another node removed the route: %v", err)
	}
	if err := a.Release(ctx, "token-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Lookup(ctx, "token-1"); !errors.Is(err, cluster.ErrNoRoute) {
		t.Fatalf("released route still found: %v", err)
	}
	if err := b.Claim(ctx, "token-1"); err != nil {
		t.Fatalf("a released route cannot be claimed by another node: %v", err)
	}
	if keys, err := client.Keys(ctx, prefix+"route:*").Result(); err != nil || len(keys) != 1 || len(keys[0]) != len(prefix+"route:")+32 {
		t.Fatalf("route keys %v, %v", keys, err)
	}
	if ttl, err := client.PTTL(ctx, func() string { keys, _ := client.Keys(ctx, prefix+"route:*").Result(); return keys[0] }()).Result(); err != nil || ttl <= 0 || ttl > time.Minute {
		t.Fatalf("route ttl %v, %v", ttl, err)
	}
}

func TestRouteToAGoneNodeIsDropped(t *testing.T) {
	client := redistest.Client(t)
	prefix := redistest.Prefix(t)
	ctx := context.Background()
	a, _ := node(t, client, prefix, "a")
	b, _ := node(t, client, prefix, "b")
	if err := a.Announce(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Claim(ctx, "token-1"); err != nil {
		t.Fatal(err)
	}
	if err := a.Leave(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Lookup(ctx, "token-1"); !errors.Is(err, cluster.ErrNodeGone) {
		t.Fatalf("route to a gone node: %v", err)
	}
	if _, err := b.Lookup(ctx, "token-1"); !errors.Is(err, cluster.ErrNoRoute) {
		t.Fatalf("route to a gone node kept: %v", err)
	}
	if err := b.Claim(ctx, "token-1"); err != nil {
		t.Fatalf("the dropped route cannot be reclaimed: %v", err)
	}
}
