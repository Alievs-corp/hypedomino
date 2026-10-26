package rules

import (
	"errors"
	"testing"
)

type matchSnapshot struct {
	phase       Phase
	scores      TeamPoints
	current     Seat
	handsPlayed int
	carryOver   int
	tiles       int
	boneyard    int
	placements  int
	handSizes   [seatCount]int
}

func snapshot(m *Match) matchSnapshot {
	s := matchSnapshot{
		phase:       m.Phase(),
		scores:      m.Scores(),
		current:     m.Current(),
		handsPlayed: m.HandsPlayed(),
		carryOver:   m.CarryOver(),
		tiles:       m.Board().TileCount(),
		boneyard:    m.BoneyardSize(),
		placements:  m.Turn().Placements(),
	}
	for _, seat := range m.Mode().Seats() {
		s.handSizes[seat] = len(m.Hand(seat))
	}
	return s
}

func applyFuzzIntent(m *Match, op, pick byte) error {
	seats := m.Mode().Seats()
	seat := seats[int(pick)%len(seats)]
	hand := m.Hand(seat)
	tile := MustTile(pick%7, (pick/7)%7)
	if len(hand) > 0 && op%2 == 0 {
		tile = hand[int(pick)%len(hand)]
	}
	end := End(pick % 4)
	switch op % 6 {
	case 0:
		_, err := m.Open(seat, tile)
		return err
	case 1:
		_, err := m.Place(seat, tile, end)
		return err
	case 2:
		return m.EndTurn(seat)
	case 3:
		_, err := m.Draw(seat)
		return err
	case 4:
		return m.Pass(seat)
	}
	var buffer [8]Move
	moves := m.LegalMoves(seat, buffer[:0])
	if len(moves) == 0 {
		return m.Pass(seat)
	}
	move := moves[int(pick)%len(moves)]
	_, err := m.Place(seat, move.Tile, move.End)
	return err
}

func FuzzMatchIntents(f *testing.F) {
	f.Add(uint64(1), []byte{0, 5, 1, 9, 2, 3, 4})
	f.Add(uint64(7), []byte{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5})
	f.Add(uint64(3), []byte{0, 0, 1, 1, 3, 3, 4, 4})
	f.Fuzz(func(t *testing.T, seed uint64, script []byte) {
		for _, mode := range []Mode{Mode1v1, Mode2v2} {
			m := NewMatch(mode, DefaultConfig(), seeded(seed))
			for i := 0; i+1 < len(script); i += 2 {
				before := snapshot(m)
				err := applyFuzzIntent(m, script[i], script[i+1])
				if err != nil && snapshot(m) != before {
					t.Fatalf("%s intent %d: %v changed the match from %+v to %+v", mode, i/2, err, before, snapshot(m))
				}
				if err == nil && m.Phase() == PhasePlaying && snapshot(m) == before {
					t.Fatalf("%s intent %d: accepted but changed nothing", mode, i/2)
				}
				checkTileConservation(t, m, mode, seed)
			}
		}
	})
}

func TestMatchRefusesIntentsOutOfPhaseAndOutOfTurn(t *testing.T) {
	m := NewMatch(Mode1v1, DefaultConfig(), seeded(11))
	if m.Config() != DefaultConfig() || m.CarryOver() != 0 || m.Mode() != Mode1v1 {
		t.Fatal("accessors do not report the construction values")
	}
	opening, _ := m.ForcedOpening()
	other := opening.Seat.Next()
	var buffer [8]Move
	if moves := m.LegalMoves(opening.Seat, buffer[:0]); len(moves) != 0 {
		t.Fatal("legal moves before the opening")
	}
	for name, err := range map[string]error{
		"end turn": m.EndTurn(opening.Seat),
		"pass":     m.Pass(opening.Seat),
	} {
		if !errors.Is(err, ErrWrongPhase) {
			t.Errorf("%s before the opening: %v", name, err)
		}
	}
	if _, err := m.Draw(opening.Seat); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("draw before the opening: %v", err)
	}
	if _, err := m.Open(opening.Seat, opening.Tile); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Open(other, m.Hand(other)[0]); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("second opening: %v", err)
	}
	if moves := m.LegalMoves(opening.Seat, buffer[:0]); len(moves) != 0 {
		t.Fatal("legal moves for the seat that is not on turn")
	}
	if err := m.EndTurn(opening.Seat); !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("end turn out of turn: %v", err)
	}
	if _, err := m.Draw(opening.Seat); !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("draw out of turn: %v", err)
	}
	if err := m.Pass(opening.Seat); !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("pass out of turn: %v", err)
	}
}

func TestMatchLaterHandOpeningBelongsToTheOpeningTeam(t *testing.T) {
	config := DefaultConfig()
	config.TargetScore = 10000
	m := NewMatch(Mode2v2, config, seeded(5))
	random := seeded(50)
	var buffer [4 * SetSize]Move
	for m.HandsPlayed() == 0 {
		if err := simulateStep(m, random, buffer[:0]); err != nil {
			t.Fatal(err)
		}
	}
	if m.Phase() != PhaseOpening {
		t.Fatalf("phase after the first hand: %d", m.Phase())
	}
	if _, ok := m.ForcedOpening(); ok {
		t.Fatal("a later hand has a forced opener")
	}
	wrong := Mode2v2.TeamSeats(m.OpeningTeam().Other(), make([]Seat, 0, 2))[0]
	if _, err := m.Open(wrong, m.Hand(wrong)[0]); !errors.Is(err, ErrWrongOpener) {
		t.Fatalf("other team opened: %v", err)
	}
	right := Mode2v2.TeamSeats(m.OpeningTeam(), make([]Seat, 0, 2))[1]
	if _, err := m.Open(right, m.Hand(wrong)[0]); !errors.Is(err, ErrTileNotHeld) {
		t.Fatalf("opened with a tile of another hand: %v", err)
	}
	if _, err := m.Open(right, m.Hand(right)[0]); err != nil || m.Phase() != PhasePlaying || m.Current() != right.Next() {
		t.Fatalf("opening by the opening team: %v phase %d current %s", err, m.Phase(), m.Current())
	}
}

func TestMatchOneVoneDrawsBeforePassing(t *testing.T) {
	random := seeded(21)
	var buffer [4 * SetSize]Move
	for seed := uint64(1); seed < 200; seed++ {
		m := NewMatch(Mode1v1, DefaultConfig(), seeded(seed))
		for m.Phase() != PhaseFinished {
			seat := m.Current()
			if m.Phase() == PhasePlaying && m.Turn().Placements() == 0 && len(m.LegalMoves(seat, buffer[:0])) == 0 && m.BoneyardSize() > 1 {
				if err := m.Pass(seat); !errors.Is(err, ErrMustDraw) {
					t.Fatalf("seed %d: pass with a drawable boneyard: %v", seed, err)
				}
				tile, err := m.Draw(seat)
				if err != nil || !m.Hand(seat).Contains(tile) {
					t.Fatalf("seed %d: draw: %v", seed, err)
				}
				if m.Board().HasLegalMove(m.Hand(seat)) {
					if _, err := m.Draw(seat); !errors.Is(err, ErrMustPlace) {
						t.Fatalf("seed %d: drew with a legal move: %v", seed, err)
					}
				}
				return
			}
			if err := simulateStep(m, random, buffer[:0]); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
	}
	t.Fatal("no 1v1 game reached a draw")
}
