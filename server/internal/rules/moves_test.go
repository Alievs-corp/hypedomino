package rules

import (
	"errors"
	"testing"
)

func tiles(texts ...string) []Tile {
	tiles := make([]Tile, 0, len(texts))
	for _, text := range texts {
		tile, err := ParseTile(text)
		if err != nil {
			panic(err)
		}
		tiles = append(tiles, tile)
	}
	return tiles
}

func exampleSpinnerBoard(t *testing.T) Board {
	t.Helper()
	board := NewBoard(MustTile(2, 3))
	mustPlace(t, &board, "3-3", EndRight)
	mustPlace(t, &board, "2-4", EndLeft)
	mustPlace(t, &board, "1-3", EndRight)
	return board
}

func TestLegalMovesListsEveryTileAndEnd(t *testing.T) {
	board := exampleSpinnerBoard(t)
	var buffer [16]Move
	moves := board.LegalMoves(tiles("4-1", "3-6", "0-0"), buffer[:0])
	want := []Move{
		{MustTile(1, 4), EndLeft},
		{MustTile(1, 4), EndRight},
		{MustTile(3, 6), EndTop},
		{MustTile(3, 6), EndBottom},
	}
	if len(moves) != len(want) {
		t.Fatalf("moves %v, want %v", moves, want)
	}
	for i := range want {
		if moves[i] != want[i] {
			t.Fatalf("move %d = %v, want %v", i, moves[i], want[i])
		}
	}
	if !board.HasLegalMove(tiles("0-0", "3-6")) || board.HasLegalMove(tiles("0-0", "6-6")) {
		t.Fatal("HasLegalMove disagrees with LegalMoves")
	}
	if len(board.LegalMoves(tiles("0-0", "6-6"), buffer[:0])) != 0 {
		t.Fatal("moves listed for a hand that fits nowhere")
	}
}

func TestValidateChecksHandAndBoard(t *testing.T) {
	board := exampleSpinnerBoard(t)
	held := tiles("4-1", "0-0")
	if err := board.Validate(held, MustTile(1, 4), EndLeft); err != nil {
		t.Fatalf("legal move refused: %v", err)
	}
	if err := board.Validate(held, MustTile(3, 6), EndTop); !errors.Is(err, ErrTileNotHeld) {
		t.Fatalf("unheld tile: %v", err)
	}
	if err := board.Validate(held, MustTile(0, 0), EndLeft); !errors.Is(err, ErrIllegalPlacement) {
		t.Fatalf("mismatched tile: %v", err)
	}
	if err := board.Validate(held, MustTile(1, 4), EndTop); !errors.Is(err, ErrIllegalPlacement) {
		t.Fatalf("1-4 on an arm with pip 3: %v", err)
	}
}

func TestDrawUntilPlayable(t *testing.T) {
	board := NewBoard(MustTile(3, 6))
	mustPlace(t, &board, "3-2", EndLeft)
	cases := []struct {
		name     string
		mode     Mode
		hand     Hand
		boneyard Boneyard
		drawn    []Tile
		outcome  TurnOutcome
		handSize int
		left     int
	}{
		{"must place without drawing", Mode1v1, tiles("1-1", "2-4"), tiles("0-5", "1-6"), nil, OutcomePlace, 2, 2},
		{"draws until a fit", Mode1v1, tiles("1-1", "3-4", "4-5"), tiles("1-4", "0-5", "1-6", "3-3"), tiles("1-4", "0-5", "1-6"), OutcomePlace, 6, 1},
		{"last tile is never drawn", Mode1v1, tiles("1-1"), tiles("5-6"), nil, OutcomePass, 1, 1},
		{"stops before the last tile", Mode1v1, tiles("1-1"), tiles("1-4", "0-5", "1-6"), tiles("1-4", "0-5"), OutcomePass, 3, 1},
		{"2v2 passes without drawing", Mode2v2, tiles("1-1", "3-4"), nil, nil, OutcomePass, 2, 0},
		{"2v2 must place", Mode2v2, tiles("1-1", "1-2"), nil, nil, OutcomePlace, 2, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			held := append(Hand{}, c.hand...)
			boneyard := append(Boneyard{}, c.boneyard...)
			var buffer [SetSize]Tile
			drawn, outcome := DrawUntilPlayable(c.mode, &board, &held, &boneyard, buffer[:0])
			if outcome != c.outcome || !sameTiles(drawn, c.drawn) {
				t.Fatalf("drawn %v outcome %v, want %v %v", drawn, outcome, c.drawn, c.outcome)
			}
			if len(held) != c.handSize || len(boneyard) != c.left {
				t.Fatalf("hand %d boneyard %d, want %d %d", len(held), len(boneyard), c.handSize, c.left)
			}
			if (outcome == OutcomePlace) != board.HasLegalMove(held) {
				t.Fatalf("outcome %v but HasLegalMove %v", outcome, board.HasLegalMove(held))
			}
		})
	}
}

func TestTurnOutcomeText(t *testing.T) {
	for _, text := range []string{"place", "pass"} {
		var outcome TurnOutcome
		if err := outcome.UnmarshalText([]byte(text)); err != nil || outcome.String() != text {
			t.Fatalf("%q = %v, %v", text, outcome, err)
		}
	}
	var outcome TurnOutcome
	if err := outcome.UnmarshalText([]byte("draw")); err == nil {
		t.Error("invalid outcome accepted")
	}
}

func TestLegalMovesDoNotAllocate(t *testing.T) {
	board := exampleSpinnerBoard(t)
	held := tiles("4-1", "3-6", "0-0", "5-5", "2-2", "1-1", "6-6")
	var buffer [SetSize * int(endCount)]Move
	allocations := testing.AllocsPerRun(100, func() {
		board.LegalMoves(held, buffer[:0])
		board.HasLegalMove(held)
		_ = board.Validate(held, held[0], EndLeft)
	})
	if allocations != 0 {
		t.Fatalf("listing moves allocates %.0f times", allocations)
	}
}

func FuzzRandomPlay(f *testing.F) {
	f.Add(uint64(1), uint8(1))
	f.Add(uint64(2), uint8(2))
	f.Fuzz(func(t *testing.T, seed uint64, modeByte uint8) {
		mode := Mode(modeByte%2 + 1)
		random := seeded(seed)
		config := DefaultConfig()
		deal := NewDeal(mode, config, random)
		seats := mode.Seats()
		hands := make([]Hand, len(seats))
		for i, seat := range seats {
			hands[i] = append(Hand{}, deal.Hands[seat]...)
		}
		boneyard := deal.Boneyard
		board := NewBoard(hands[0][0])
		hands[0].Remove(hands[0][0])
		var moveBuffer [SetSize * int(endCount)]Move
		var drawBuffer [SetSize]Tile
		passes := 0
		for turn := 1; turn < 200 && passes < len(seats); turn++ {
			player := turn % len(seats)
			if len(hands[player]) == 0 {
				return
			}
			drawn, outcome := DrawUntilPlayable(mode, &board, &hands[player], &boneyard, drawBuffer[:0])
			if mode == Mode2v2 && len(drawn) > 0 {
				t.Fatal("drew in 2v2")
			}
			moves := board.LegalMoves(hands[player], moveBuffer[:0])
			if (outcome == OutcomePlace) != (len(moves) > 0) {
				t.Fatalf("outcome %v with %d legal moves", outcome, len(moves))
			}
			if outcome == OutcomePass {
				if mode == Mode1v1 && len(boneyard) > 1 {
					t.Fatal("passed while the boneyard still had drawable tiles")
				}
				passes++
				continue
			}
			passes = 0
			assertOnlyListedMovesValidate(t, &board, hands[player], moves)
			move := moves[random.IntN(len(moves))]
			before := board
			if err := board.Place(move.Tile, move.End); err != nil {
				t.Fatalf("listed move refused: %v", err)
			}
			if board == before {
				t.Fatal("placement changed nothing")
			}
			hands[player].Remove(move.Tile)
		}
	})
}

func assertOnlyListedMovesValidate(t *testing.T, board *Board, held Hand, listed []Move) {
	t.Helper()
	set := FullSet()
	for _, tile := range set {
		for end := EndLeft; end < endCount; end++ {
			err := board.Validate(held, tile, end)
			expected := false
			for _, move := range listed {
				if move.Tile == tile && move.End == end {
					expected = true
				}
			}
			if (err == nil) != expected {
				t.Fatalf("Validate(%s, %s) = %v, listed %v", tile, end, err, expected)
			}
		}
	}
}

func BenchmarkLegalMoves(b *testing.B) {
	board := NewBoard(MustTile(2, 3))
	held := tiles("4-1", "3-6", "0-0", "5-5", "2-2", "1-1", "6-6")
	var buffer [SetSize * int(endCount)]Move
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		board.LegalMoves(held, buffer[:0])
	}
}
