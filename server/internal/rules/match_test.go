package rules

import (
	"errors"
	"math/rand/v2"
	"testing"
)

func TestMatchEndChecksOnlyAfterADomino(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		end     HandStatus
		a, b    int
		outcome MatchOutcome
		winner  Team
	}{
		{HandDomino, 370, 200, MatchWon, TeamA},
		{HandDomino, 365, 100, MatchWon, TeamA},
		{HandDomino, 364, 200, MatchNextHand, TeamA},
		{HandDomino, 240, 370, MatchWon, TeamB},
		{HandDomino, 370, 375, MatchWon, TeamB},
		{HandDomino, 370, 370, MatchNextHand, TeamA},
		{HandBlocked, 380, 370, MatchNextHand, TeamA},
		{HandContinues, 380, 100, MatchContinues, TeamA},
	}
	for _, c := range cases {
		outcome, winner := MatchEnd(config, c.end, TeamPoints{c.a, c.b})
		if outcome != c.outcome || (outcome == MatchWon && winner != c.winner) {
			t.Errorf("%s %d-%d: %s %s, want %s %s", c.end, c.a, c.b, outcome, winner, c.outcome, c.winner)
		}
	}
	config.TargetScore = 100
	if outcome, winner := MatchEnd(config, HandDomino, TeamPoints{100, 0}); outcome != MatchWon || winner != TeamA {
		t.Errorf("target 100: %s %s", outcome, winner)
	}
}

func TestMatchOutcomeText(t *testing.T) {
	for _, outcome := range []MatchOutcome{MatchContinues, MatchNextHand, MatchWon} {
		var parsed MatchOutcome
		if err := parsed.UnmarshalText([]byte(outcome.String())); err != nil || parsed != outcome {
			t.Errorf("%s does not round-trip", outcome)
		}
	}
	var parsed MatchOutcome
	if err := parsed.UnmarshalText([]byte("draw")); err == nil || MatchOutcome(7).String() != "invalid" {
		t.Error("invalid outcomes are accepted")
	}
}

func TestMatchFirstHandIsOpenedByTheForcedOpener(t *testing.T) {
	m := NewMatch(Mode2v2, DefaultConfig(), seeded(1))
	opening, ok := m.ForcedOpening()
	if !ok || opening.Tile != MustTile(2, 3) || m.Phase() != PhaseOpening {
		t.Fatalf("forced opening %+v %v phase %d", opening, ok, m.Phase())
	}
	if _, err := m.Open(opening.Seat.Next(), opening.Tile); !errors.Is(err, ErrWrongOpener) {
		t.Fatalf("another seat opened: %v", err)
	}
	if _, err := m.Place(opening.Seat, opening.Tile, EndLeft); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("placed before opening: %v", err)
	}
	score, err := m.Open(opening.Seat, opening.Tile)
	if err != nil || score != 0 {
		t.Fatalf("Open = %d, %v", score, err)
	}
	if m.Phase() != PhasePlaying || m.Current() != opening.Seat.Next() || m.Board().TileCount() != 1 || len(m.Hand(opening.Seat)) != 6 {
		t.Fatalf("state after opening: phase %d current %s tiles %d hand %d", m.Phase(), m.Current(), m.Board().TileCount(), len(m.Hand(opening.Seat)))
	}
	if _, ok := m.ForcedOpening(); ok {
		t.Fatal("forced opening survives the opening")
	}
	if _, err := m.Place(opening.Seat, MustTile(0, 0), EndLeft); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("opener moved again: %v", err)
	}
}

func TestMatchTurnRulesAreEnforced(t *testing.T) {
	m := NewMatch(Mode2v2, DefaultConfig(), seeded(2))
	opening, _ := m.ForcedOpening()
	if _, err := m.Open(opening.Seat, opening.Tile); err != nil {
		t.Fatal(err)
	}
	seat := m.Current()
	if err := m.EndTurn(seat); !errors.Is(err, ErrMustPlace) {
		t.Fatalf("turn ended without a placement: %v", err)
	}
	if _, err := m.Draw(seat); !errors.Is(err, ErrMustPlace) && !errors.Is(err, ErrCannotDraw) {
		t.Fatalf("2v2 draw: %v", err)
	}
	var buffer [8]Move
	moves := m.LegalMoves(seat, buffer[:0])
	if len(moves) == 0 {
		if err := m.Pass(seat); err != nil {
			t.Fatalf("pass without a move: %v", err)
		}
		return
	}
	if err := m.Pass(seat); !errors.Is(err, ErrMustPlace) {
		t.Fatalf("pass with a legal move: %v", err)
	}
	absent := MustTile(6, 6)
	for m.Hand(seat).Contains(absent) {
		absent = MustTile(absent.Low-1, absent.High-1)
	}
	if _, err := m.Place(seat, absent, moves[0].End); !errors.Is(err, ErrTileNotHeld) {
		t.Fatalf("placed a tile not in hand: %v", err)
	}
	if _, err := m.Place(seat, moves[0].Tile, moves[0].End); err != nil {
		t.Fatal(err)
	}
	if m.Current() == seat && m.Turn().Placements() == 0 {
		t.Fatal("turn did not advance")
	}
}

func TestMatchSimulation(t *testing.T) {
	matches := 5000
	if testing.Short() {
		matches = 200
	}
	for _, mode := range []Mode{Mode1v1, Mode2v2} {
		hands := 0
		for seed := uint64(1); seed <= uint64(matches); seed++ {
			hands += simulateMatch(t, mode, seed)
		}
		if hands < matches {
			t.Fatalf("%s: %d matches played only %d hands", mode, matches, hands)
		}
	}
}

func simulateMatch(t *testing.T, mode Mode, seed uint64) int {
	t.Helper()
	config := DefaultConfig()
	random := seeded(seed)
	m := NewMatch(mode, config, random)
	var buffer [4 * SetSize]Move
	previous := m.Scores()
	for step := 0; m.Phase() != PhaseFinished; step++ {
		if step > 20000 {
			t.Fatalf("%s seed %d: match does not end", mode, seed)
		}
		checkTileConservation(t, m, mode, seed)
		if err := simulateStep(m, random, buffer[:0]); err != nil {
			t.Fatalf("%s seed %d step %d: %v", mode, seed, step, err)
		}
		scores := m.Scores()
		if scores[TeamA] < previous[TeamA] || scores[TeamB] < previous[TeamB] {
			t.Fatalf("%s seed %d: scores went down from %v to %v", mode, seed, previous, scores)
		}
		previous = scores
	}
	winner, ok := m.Winner()
	scores := m.Scores()
	if !ok || scores[winner] < config.TargetScore || scores[winner] <= scores[winner.Other()] {
		t.Fatalf("%s seed %d: winner %s with %v", mode, seed, winner, scores)
	}
	result, ok := m.LastResult()
	if !ok || !result.HasWinner {
		t.Fatalf("%s seed %d: the match ended on a hand without a winner: %+v", mode, seed, result)
	}
	return m.HandsPlayed()
}

func simulateStep(m *Match, random *rand.Rand, buffer []Move) error {
	if m.Phase() == PhaseOpening {
		opening, ok := m.ForcedOpening()
		if !ok {
			seats := m.Mode().TeamSeats(m.OpeningTeam(), make([]Seat, 0, 2))
			opening.Seat = seats[random.IntN(len(seats))]
			hand := m.Hand(opening.Seat)
			opening.Tile = hand[random.IntN(len(hand))]
		}
		_, err := m.Open(opening.Seat, opening.Tile)
		return err
	}
	seat := m.Current()
	moves := m.LegalMoves(seat, buffer)
	if m.Turn().Placements() > 0 {
		if len(moves) == 0 {
			return errors.New("a continued turn has no scoring double")
		}
		if random.IntN(2) == 0 {
			return m.EndTurn(seat)
		}
		_, err := m.Place(seat, moves[0].Tile, moves[0].End)
		return err
	}
	if len(moves) > 0 {
		move := moves[random.IntN(len(moves))]
		_, err := m.Place(seat, move.Tile, move.End)
		return err
	}
	if _, err := m.Draw(seat); err == nil {
		return nil
	} else if !errors.Is(err, ErrCannotDraw) {
		return err
	}
	return m.Pass(seat)
}

func checkTileConservation(t *testing.T, m *Match, mode Mode, seed uint64) {
	t.Helper()
	count := m.Board().TileCount() + m.BoneyardSize()
	for _, seat := range mode.Seats() {
		count += len(m.Hand(seat))
	}
	if m.Phase() == PhaseOpening {
		count -= m.Board().TileCount()
	}
	if count != SetSize {
		t.Fatalf("%s seed %d: %d tiles in play", mode, seed, count)
	}
}

func TestMatchPlayingStepsDoNotAllocate(t *testing.T) {
	m := NewMatch(Mode2v2, DefaultConfig(), seeded(3))
	opening, _ := m.ForcedOpening()
	if _, err := m.Open(opening.Seat, opening.Tile); err != nil {
		t.Fatal(err)
	}
	var buffer [8]Move
	allocations := testing.AllocsPerRun(20, func() {
		seat := m.Current()
		moves := m.LegalMoves(seat, buffer[:0])
		if len(moves) == 0 {
			m.Pass(seat)
			return
		}
		m.Place(seat, moves[0].Tile, moves[0].End)
		if m.Current() == seat {
			m.EndTurn(seat)
		}
	})
	if allocations != 0 && m.HandsPlayed() == 0 {
		t.Fatalf("a playing step allocates %.0f times", allocations)
	}
}

func BenchmarkMatch(b *testing.B) {
	for _, mode := range []Mode{Mode1v1, Mode2v2} {
		b.Run(mode.String(), func(b *testing.B) {
			var buffer [4 * SetSize]Move
			steps := 0
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				random := seeded(uint64(i) + 1)
				m := NewMatch(mode, DefaultConfig(), random)
				for m.Phase() != PhaseFinished {
					if err := simulateStep(m, random, buffer[:0]); err != nil {
						b.Fatal(err)
					}
					steps++
				}
			}
			b.ReportMetric(float64(steps)/float64(b.N), "steps/match")
		})
	}
}
