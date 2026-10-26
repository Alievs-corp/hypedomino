package rules

import (
	"errors"
	"testing"
)

func lineBoard(t *testing.T, texts ...string) Board {
	t.Helper()
	first, err := ParseOriented(texts[0])
	if err != nil {
		t.Fatal(err)
	}
	board := NewBoard(first.Tile())
	board.line[board.head] = first
	for _, text := range texts[1:] {
		next, err := ParseOriented(text)
		if err != nil {
			t.Fatal(err)
		}
		mustPlace(t, &board, next.Tile().String(), EndRight)
	}
	return board
}

func TestTurnExample3(t *testing.T) {
	board := lineBoard(t, "0-2", "2-5")
	turn := BeginTurn(&board, DefaultConfig())
	first, err := turn.Place(MustTile(0, 0), EndLeft)
	if err != nil || first != (Placement{Total: 5, Score: 5}) {
		t.Fatalf("first = %v, %v", first, err)
	}
	if !turn.CanContinue() {
		t.Fatal("a scoring double does not open a continuation")
	}
	second, err := turn.Place(MustTile(5, 5), EndRight)
	if err != nil || second != (Placement{Total: 10, Score: 10}) {
		t.Fatalf("second = %v, %v", second, err)
	}
	if _, err := turn.Place(MustTile(1, 1), EndLeft); !errors.Is(err, ErrIllegalPlacement) {
		t.Fatalf("1-1 without a 1 end: %v", err)
	}
	if turn.Score() != 15 || turn.Placements() != 2 || board.TileCount() != 4 {
		t.Fatalf("score %d placements %d tiles %d", turn.Score(), turn.Placements(), board.TileCount())
	}
}

func TestTurnRefusesContinuationsThatDoNotScore(t *testing.T) {
	board := lineBoard(t, "2-4", "4-6")
	turn := BeginTurn(&board, DefaultConfig())
	if _, err := turn.Place(MustTile(2, 2), EndLeft); err != nil {
		t.Fatal(err)
	}
	before := board
	if _, err := turn.Place(MustTile(6, 6), EndRight); !errors.Is(err, ErrIllegalPlacement) {
		t.Fatalf("6-6 making 16: %v", err)
	}
	if board != before {
		t.Fatal("a refused continuation changed the board")
	}
	if turn.Score() != 10 {
		t.Fatalf("score %d", turn.Score())
	}
}

func TestTurnEndsAfterNonDoubleOrNonScoringDouble(t *testing.T) {
	config := DefaultConfig()
	board := lineBoard(t, "5-1", "1-4")
	turn := BeginTurn(&board, config)
	if _, err := turn.Place(MustTile(0, 4), EndRight); err != nil {
		t.Fatal(err)
	}
	if turn.CanContinue() {
		t.Fatal("a non-double continues the turn")
	}
	if _, err := turn.Place(MustTile(0, 0), EndRight); !errors.Is(err, ErrTurnOver) {
		t.Fatalf("after a non-double: %v", err)
	}
	board = lineBoard(t, "2-5", "5-3")
	turn = BeginTurn(&board, config)
	placement, err := turn.Place(MustTile(3, 3), EndRight)
	if err != nil || placement != (Placement{Total: 8, Score: 0}) {
		t.Fatalf("3-3 = %v, %v", placement, err)
	}
	if _, err := turn.Place(MustTile(2, 2), EndLeft); !errors.Is(err, ErrTurnOver) {
		t.Fatalf("after a non-scoring double: %v", err)
	}
	board = lineBoard(t, "1-3", "3-3", "3-2")
	turn = BeginTurn(&board, config)
	if _, err := turn.Place(MustTile(2, 2), EndRight); err != nil {
		t.Fatal(err)
	}
	if _, err := turn.Place(MustTile(1, 6), EndLeft); !errors.Is(err, ErrIllegalPlacement) {
		t.Fatalf("a non-double as continuation: %v", err)
	}
}

func TestTurnLegalMovesListsOnlyScoringDoubles(t *testing.T) {
	board := lineBoard(t, "0-2", "2-5")
	turn := BeginTurn(&board, DefaultConfig())
	held := tiles("0-0", "5-5", "1-1", "2-5")
	var buffer [SetSize * int(endCount)]Move
	if got := len(turn.LegalMoves(held, buffer[:0])); got != 3 {
		t.Fatalf("first placement lists %d moves", got)
	}
	if _, err := turn.Place(MustTile(0, 0), EndLeft); err != nil {
		t.Fatal(err)
	}
	moves := turn.LegalMoves(tiles("5-5", "1-1", "2-5"), buffer[:0])
	if len(moves) != 1 || moves[0] != (Move{MustTile(5, 5), EndRight}) {
		t.Fatalf("continuations %v", moves)
	}
	if _, err := turn.Place(MustTile(5, 5), EndRight); err != nil {
		t.Fatal(err)
	}
	if got := len(turn.LegalMoves(tiles("1-1", "2-5"), buffer[:0])); got != 0 {
		t.Fatalf("%d continuations after the ends closed", got)
	}
	board = lineBoard(t, "5-1", "1-4")
	turn = BeginTurn(&board, DefaultConfig())
	if _, err := turn.Place(MustTile(0, 4), EndRight); err != nil {
		t.Fatal(err)
	}
	if got := len(turn.LegalMoves(tiles("0-0"), buffer[:0])); got != 0 {
		t.Fatalf("%d continuations after a non-double", got)
	}
}

func TestTurnDoesNotAllocate(t *testing.T) {
	board := lineBoard(t, "0-2", "2-5")
	held := tiles("0-0", "5-5", "1-1")
	var buffer [SetSize * int(endCount)]Move
	allocations := testing.AllocsPerRun(100, func() {
		copy := board
		turn := BeginTurn(&copy, DefaultConfig())
		_, _ = turn.Place(MustTile(0, 0), EndLeft)
		turn.LegalMoves(held, buffer[:0])
		_, _ = turn.Place(MustTile(5, 5), EndRight)
	})
	if allocations != 0 {
		t.Fatalf("a turn allocates %.0f times", allocations)
	}
}

func FuzzTurn(f *testing.F) {
	f.Add(uint64(3), []byte{0, 5, 9, 13})
	f.Fuzz(func(t *testing.T, seed uint64, moves []byte) {
		random := seeded(seed)
		set := FullSet()
		Shuffle(set[:], random)
		board := NewBoard(set[0])
		held := Hand(set[1:8])
		turn := BeginTurn(&board, DefaultConfig())
		for _, move := range moves {
			tile := held[int(move>>2)%len(held)]
			end := End(move & 3)
			before := board
			placedBefore := turn.Placements()
			placement, err := turn.Place(tile, end)
			if err != nil {
				if board != before {
					t.Fatal("a refused placement changed the board")
				}
				continue
			}
			if placedBefore > 0 && (!tile.IsDouble() || placement.Score == 0) {
				t.Fatalf("continuation %s accepted with score %d", tile, placement.Score)
			}
			if placement.Total != board.OpenEnds().Total() {
				t.Fatal("placement total disagrees with the board")
			}
			held.Remove(tile)
			if len(held) == 0 {
				return
			}
		}
	})
}
