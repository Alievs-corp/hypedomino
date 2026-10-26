package bot

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func seeded(seed uint64) *rand.Rand {
	var key [32]byte
	for i := range 8 {
		key[i] = byte(seed >> (8 * i))
	}
	return rand.New(rand.NewChaCha8(key))
}

type table struct {
	match   *rules.Match
	players map[rules.Seat]*Player
}

func newTable(mode rules.Mode, seed uint64, levelOf func(rules.Team) Level) table {
	random := seeded(seed)
	t := table{match: rules.NewMatch(mode, rules.DefaultConfig(), random), players: map[rules.Seat]*Player{}}
	for _, seat := range mode.Seats() {
		t.players[seat] = New(levelOf(seat.Team()), DefaultConfig(), random)
	}
	return t
}

func (t table) step() (Action, error) {
	m := t.match
	if m.Phase() == rules.PhaseOpening {
		if opening, forced := m.ForcedOpening(); forced {
			return t.players[opening.Seat].Act(m, opening.Seat)
		}
		seats := m.Mode().TeamSeats(m.OpeningTeam(), make([]rules.Seat, 0, 2))
		return t.players[seats[0]].Act(m, seats[0])
	}
	return t.players[m.Current()].Act(m, m.Current())
}

func (t table) play(tb testing.TB, seed uint64) rules.Team {
	tb.Helper()
	for step := 0; t.match.Phase() != rules.PhaseFinished; step++ {
		if step > 20000 {
			tb.Fatalf("seed %d: the match does not end", seed)
		}
		if _, err := t.step(); err != nil {
			tb.Fatalf("seed %d step %d: %v", seed, step, err)
		}
	}
	winner, _ := t.match.Winner()
	return winner
}

func TestLevelsNeverPlayAnIllegalMove(t *testing.T) {
	matches := 300
	if testing.Short() {
		matches = 30
	}
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		for _, level := range []Level{Easy, Medium, Hard} {
			for seed := uint64(1); seed <= uint64(matches); seed++ {
				newTable(mode, seed, func(rules.Team) Level { return level }).play(t, seed)
			}
		}
	}
}

func FuzzBotNeverPlaysAnIllegalMove(f *testing.F) {
	f.Add(uint64(1), uint8(2), uint8(0))
	f.Add(uint64(9), uint8(1), uint8(1))
	f.Fuzz(func(t *testing.T, seed uint64, levelA, levelB uint8) {
		for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
			tab := newTable(mode, seed, func(team rules.Team) Level {
				if team == rules.TeamA {
					return Level(levelA % uint8(levelCount))
				}
				return Level(levelB % uint8(levelCount))
			})
			tab.play(t, seed)
		}
	})
}

func winRate(t *testing.T, mode rules.Mode, strong, weak Level, matches int) float64 {
	t.Helper()
	wins := 0
	for seed := uint64(1); seed <= uint64(matches); seed++ {
		strongTeam := rules.TeamA
		if seed%2 == 0 {
			strongTeam = rules.TeamB
		}
		winner := newTable(mode, seed, func(team rules.Team) Level {
			if team == strongTeam {
				return strong
			}
			return weak
		}).play(t, seed)
		if winner == strongTeam {
			wins++
		}
	}
	return float64(wins) / float64(matches)
}

func TestHardBeatsRandomPlayAtLeastEightyPercent(t *testing.T) {
	matches := 400
	if testing.Short() {
		matches = 100
	}
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		rate := winRate(t, mode, Hard, Easy, matches)
		t.Logf("%s: hard beats easy in %.1f%% of %d matches", mode, 100*rate, matches)
		if rate < 0.8 {
			t.Fatalf("%s: hard wins only %.1f%%", mode, 100*rate)
		}
		if medium := winRate(t, mode, Medium, Easy, matches); medium < 0.6 {
			t.Fatalf("%s: medium wins only %.1f%%", mode, 100*medium)
		}
	}
}

func TestBotActsOnlyForItsSeat(t *testing.T) {
	tab := newTable(rules.Mode2v2, 4, func(rules.Team) Level { return Hard })
	opening, _ := tab.match.ForcedOpening()
	other := opening.Seat.Next()
	if _, err := tab.players[other].Act(tab.match, other); err != ErrNotActing {
		t.Fatalf("a seat that may not open acted: %v", err)
	}
	if action, err := tab.players[opening.Seat].Act(tab.match, opening.Seat); err != nil || action != ActionOpen {
		t.Fatalf("forced opener: %v %v", action, err)
	}
	if _, err := tab.players[opening.Seat].Act(tab.match, opening.Seat); err != ErrNotActing {
		t.Fatalf("acted out of turn: %v", err)
	}
	if Level(9).String() != "invalid" || Hard.String() != "hard" || New(Easy, DefaultConfig(), seeded(1)).Level() != Easy {
		t.Fatal("level names")
	}
}

func TestHardPrefersTheScoringMove(t *testing.T) {
	tab := newTable(rules.Mode2v2, 6, func(rules.Team) Level { return Hard })
	m := tab.match
	for step := 0; m.Phase() != rules.PhaseFinished; step++ {
		if m.Phase() == rules.PhasePlaying && m.Turn().Placements() == 0 {
			seat := m.Current()
			var buffer [64]rules.Move
			moves := m.LegalMoves(seat, buffer[:0])
			bestScore := 0
			for _, move := range moves {
				trial := *m.Board()
				if trial.Place(move.Tile, move.End) == nil {
					if score := rules.PlacementScore(trial.OpenEnds().Total(), m.Config()); score > bestScore {
						bestScore = score
					}
				}
			}
			before := m.Scores()[seat.Team()]
			if _, err := tab.players[seat].Act(m, seat); err != nil {
				t.Fatal(err)
			}
			if gained := m.Scores()[seat.Team()] - before; bestScore > 0 && gained == 0 && m.Turn().Placements() <= 1 {
				t.Fatalf("step %d: %s could score %d but scored nothing", step, seat, bestScore)
			}
			if bestScore > 0 {
				return
			}
			continue
		}
		if _, err := tab.step(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDecisionsStayUnderTheTimeLimit(t *testing.T) {
	tab := newTable(rules.Mode2v2, 8, func(rules.Team) Level { return Hard })
	slowest := time.Duration(0)
	for tab.match.Phase() != rules.PhaseFinished {
		start := time.Now()
		if _, err := tab.step(); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed > slowest {
			slowest = elapsed
		}
	}
	if slowest > 50*time.Millisecond {
		t.Fatalf("slowest decision took %s", slowest)
	}
}

func TestDecisionsDoNotAllocate(t *testing.T) {
	tab := newTable(rules.Mode2v2, 12, func(rules.Team) Level { return Hard })
	for tab.match.Phase() == rules.PhaseOpening {
		if _, err := tab.step(); err != nil {
			t.Fatal(err)
		}
	}
	allocations := testing.AllocsPerRun(10, func() {
		if tab.match.Phase() == rules.PhasePlaying && tab.match.HandsPlayed() == 0 {
			tab.step()
		}
	})
	if allocations != 0 && tab.match.HandsPlayed() == 0 {
		t.Fatalf("a decision allocates %.0f times", allocations)
	}
}

func BenchmarkHardMatch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		newTable(rules.Mode2v2, uint64(i)+1, func(rules.Team) Level { return Hard }).play(b, uint64(i)+1)
	}
}
