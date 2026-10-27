package matchmaking

import (
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var start = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

func entry(account string, mode rules.Mode, queue Queue, mmr int, waited time.Duration) Entry {
	return Entry{Account: account, Mode: mode, Queue: queue, Region: "eu", Mmr: mmr, EnqueuedAt: start.Add(-waited), SearchFrom: start.Add(-waited)}
}

func seeded(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed))
}

func accounts(match Match) map[string]rules.Seat {
	out := map[string]rules.Seat{}
	for _, seat := range match.Seats {
		if !seat.Bot {
			out[seat.Account] = seat.Seat
		}
	}
	return out
}

func bots(match Match) int {
	count := 0
	for _, seat := range match.Seats {
		if seat.Bot {
			count++
		}
	}
	return count
}

func TestWindowWidensWithTheWaitUpToTheMaximum(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		waited time.Duration
		want   int
	}{
		{0, 100},
		{9 * time.Second, 100},
		{10 * time.Second, 150},
		{25 * time.Second, 200},
		{60 * time.Second, 400},
		{time.Hour, 400},
		{-time.Second, 100},
	}
	for _, tc := range cases {
		if got := Window(config, tc.waited); got != tc.want {
			t.Errorf("Window(%v) = %d, want %d", tc.waited, got, tc.want)
		}
	}
	config.MmrWindowStepSeconds = 0
	if got := Window(config, time.Hour); got != 100 {
		t.Errorf("without a step period the window must stay at the start, got %d", got)
	}
}

func TestPairsTheOldestEntryWithTheClosestWaitingPlayerInsideTheWindow(t *testing.T) {
	entries := []Entry{
		entry("late", rules.Mode1v1, QueueRanked, 1200, 1*time.Second),
		entry("oldest", rules.Mode1v1, QueueRanked, 1200, 5*time.Second),
		entry("far", rules.Mode1v1, QueueRanked, 1500, 4*time.Second),
		entry("near", rules.Mode1v1, QueueRanked, 1280, 3*time.Second),
	}
	matches, remaining := Form(DefaultConfig(), entries, start, seeded(1))
	if len(matches) != 1 || len(remaining) != 2 {
		t.Fatalf("formed %d matches, %d remaining", len(matches), len(remaining))
	}
	seats := accounts(matches[0])
	if _, ok := seats["oldest"]; !ok || len(seats) != 2 || bots(matches[0]) != 0 {
		t.Fatalf("match seats %+v", matches[0].Seats)
	}
	if _, ok := seats["near"]; !ok {
		t.Fatalf("the oldest player was paired with %+v, want near", seats)
	}
	if matches[0].Mode != rules.Mode1v1 || matches[0].Queue != QueueRanked || matches[0].Region != "eu" {
		t.Fatalf("match header %+v", matches[0])
	}
}

func TestPlayersOutsideEachOthersWindowWaitUntilItWidens(t *testing.T) {
	config := DefaultConfig()
	entries := []Entry{
		entry("a", rules.Mode1v1, QueueRanked, 1000, 0),
		entry("b", rules.Mode1v1, QueueRanked, 1300, 0),
	}
	if matches, remaining := Form(config, entries, start, seeded(1)); len(matches) != 0 || len(remaining) != 2 {
		t.Fatalf("a gap of 300 matched at once: %+v", matches)
	}
	later := start.Add(39 * time.Second)
	if matches, _ := Form(config, entries, later, seeded(1)); len(matches) != 0 {
		t.Fatalf("a gap of 300 matched at a window of 250: %+v", matches)
	}
	later = start.Add(40 * time.Second)
	if matches, remaining := Form(config, entries, later, seeded(1)); len(matches) != 1 || len(remaining) != 0 {
		t.Fatalf("a gap of 300 did not match at a window of 300: %+v", matches)
	}
}

func TestTheNarrowerWindowOfTheTwoPlayersDecides(t *testing.T) {
	entries := []Entry{
		entry("patient", rules.Mode1v1, QueueRanked, 1000, 2*time.Minute),
		entry("fresh", rules.Mode1v1, QueueRanked, 1250, 0),
	}
	if matches, _ := Form(DefaultConfig(), entries, start, seeded(1)); len(matches) != 0 {
		t.Fatalf("the fresh player's window of 100 was ignored: %+v", matches)
	}
}

func TestQueuesModesAndRegionsNeverMix(t *testing.T) {
	entries := []Entry{
		entry("r1", rules.Mode1v1, QueueRanked, 1200, time.Minute),
		entry("c1", rules.Mode1v1, QueueCasual, 1200, 5*time.Second),
		entry("t1", rules.Mode2v2, QueueRanked, 1200, time.Minute),
	}
	other := entry("na", rules.Mode1v1, QueueRanked, 1200, time.Minute)
	other.Region = "na"
	entries = append(entries, other)
	matches, remaining := Form(DefaultConfig(), entries, start, seeded(1))
	if len(matches) != 0 || len(remaining) != 4 {
		t.Fatalf("entries of different queues, modes or regions were matched: %+v", matches)
	}
}

func TestRankedQueueNeverSeatsABot(t *testing.T) {
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		entries := []Entry{entry("alone", mode, QueueRanked, 1200, 24*time.Hour)}
		matches, remaining := Form(DefaultConfig(), entries, start, seeded(1))
		if len(matches) != 0 || len(remaining) != 1 {
			t.Fatalf("%s: a ranked seat went to a bot: %+v", mode, matches)
		}
	}
}

func TestCasualQueueFillsWithBotsAfterTheConfiguredWait(t *testing.T) {
	config := DefaultConfig()
	entries := []Entry{
		entry("one", rules.Mode2v2, QueueCasual, 1200, 29*time.Second),
		entry("two", rules.Mode2v2, QueueCasual, 1210, 5*time.Second),
	}
	if matches, _ := Form(config, entries, start, seeded(1)); len(matches) != 0 {
		t.Fatalf("bots were seated before the wait: %+v", matches)
	}
	matches, remaining := Form(config, entries, start.Add(time.Second), seeded(1))
	if len(matches) != 1 || len(remaining) != 0 || bots(matches[0]) != 2 || len(accounts(matches[0])) != 2 {
		t.Fatalf("after the wait: %+v remaining %+v", matches, remaining)
	}
	config.CasualBotFillSeconds = 0
	if matches, _ := Form(config, entries, start.Add(time.Hour), seeded(1)); len(matches) != 0 {
		t.Fatalf("bot fill disabled by config still seated bots: %+v", matches)
	}
}

func TestSearchStartsOnlyAfterTheQueueDelay(t *testing.T) {
	delayed := entry("delayed", rules.Mode1v1, QueueRanked, 1200, 10*time.Second)
	delayed.SearchFrom = start.Add(5 * time.Second)
	entries := []Entry{delayed, entry("ready", rules.Mode1v1, QueueRanked, 1200, 10*time.Second)}
	if matches, _ := Form(DefaultConfig(), entries, start, seeded(1)); len(matches) != 0 {
		t.Fatalf("a delayed entry was searched early: %+v", matches)
	}
	if matches, _ := Form(DefaultConfig(), entries, start.Add(5*time.Second), seeded(1)); len(matches) != 1 {
		t.Fatal("the delayed entry did not match once its delay passed")
	}
}

func TestTwoVersusTwoTeamsAreRandom(t *testing.T) {
	config := DefaultConfig()
	together := 0
	runs := 2000
	for seed := uint64(1); seed <= uint64(runs); seed++ {
		entries := []Entry{
			entry("p1", rules.Mode2v2, QueueRanked, 1200, 4*time.Second),
			entry("p2", rules.Mode2v2, QueueRanked, 1200, 3*time.Second),
			entry("p3", rules.Mode2v2, QueueRanked, 1200, 2*time.Second),
			entry("p4", rules.Mode2v2, QueueRanked, 1200, 1*time.Second),
		}
		matches, remaining := Form(config, entries, start, seeded(seed))
		if len(matches) != 1 || len(remaining) != 0 || len(accounts(matches[0])) != 4 {
			t.Fatalf("seed %d: %+v", seed, matches)
		}
		seats := accounts(matches[0])
		if seats["p1"].Team() == seats["p2"].Team() {
			together++
		}
	}
	share := float64(together) / float64(runs)
	if share < 0.28 || share > 0.39 {
		t.Fatalf("two given players shared a team in %.1f%% of matches, want about one third", 100*share)
	}
}

func TestPoolQueuesOncePerAccountAndFormsOnTick(t *testing.T) {
	pool := NewPool(DefaultConfig(), seeded(7))
	a := entry("a", rules.Mode1v1, QueueCasual, 1200, 0)
	a.SearchFrom = time.Time{}
	if err := pool.Enqueue(a); err != nil {
		t.Fatal(err)
	}
	if err := pool.Enqueue(a); !errors.Is(err, ErrAlreadyQueued) {
		t.Fatalf("second enqueue: %v", err)
	}
	for _, bad := range []Entry{
		{Mode: rules.Mode1v1, Queue: QueueRanked, EnqueuedAt: start},
		{Account: "x", Mode: 9, Queue: QueueRanked, EnqueuedAt: start},
		{Account: "x", Mode: rules.Mode1v1, Queue: 0, EnqueuedAt: start},
		{Account: "x", Mode: rules.Mode1v1, Queue: QueueRanked},
	} {
		if err := pool.Enqueue(bad); !errors.Is(err, ErrInvalidEntry) {
			t.Fatalf("%+v accepted: %v", bad, err)
		}
	}
	if err := pool.Enqueue(entry("b", rules.Mode1v1, QueueCasual, 1250, 0)); err != nil {
		t.Fatal(err)
	}
	if pool.Len() != 2 || !pool.Queued("a") || pool.Queued("zzz") {
		t.Fatalf("pool holds %d", pool.Len())
	}
	matches := pool.Tick(start)
	if len(matches) != 1 || pool.Len() != 0 || len(accounts(matches[0])) != 2 {
		t.Fatalf("tick formed %+v, pool holds %d", matches, pool.Len())
	}
	if err := pool.Enqueue(entry("c", rules.Mode1v1, QueueRanked, 1200, 0)); err != nil {
		t.Fatal(err)
	}
	if !pool.Leave("c") || pool.Leave("c") || pool.Len() != 0 {
		t.Fatal("leave did not remove the entry exactly once")
	}
}
