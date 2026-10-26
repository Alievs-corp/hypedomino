package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const vectorsDir = "../../../docs/rules-vectors"

type vector struct {
	ID     string
	Kind   string
	Mode   Mode
	Given  json.RawMessage
	Expect json.RawMessage
}

type vectorFile struct {
	Vectors []vector
}

var vectorRunners = map[string]func(*testing.T, vector){
	"deal":  runDealVector,
	"place": runPlaceVector,
	"draw":  runDrawVector,
}

var pendingVectorKinds = map[string]bool{
	"seat":          true,
	"first-opener":  true,
	"opening-score": true,
	"turn":          true,
	"hand-status":   true,
	"round":         true,
	"hand-end":      true,
	"match-end":     true,
	"turn-timer":    true,
	"afk":           true,
}

func TestRuleVectors(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(vectorsDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	ran := 0
	for _, path := range paths {
		if filepath.Base(path) == "schema.json" {
			continue
		}
		for _, v := range loadVectors(t, path) {
			t.Run(v.ID, func(t *testing.T) {
				run, ok := vectorRunners[v.Kind]
				if !ok && pendingVectorKinds[v.Kind] {
					t.Skipf("kind %q has no engine function yet", v.Kind)
				}
				if !ok {
					t.Fatalf("unknown vector kind %q", v.Kind)
				}
				run(t, v)
			})
			ran++
		}
	}
	if ran == 0 {
		t.Fatalf("no vectors found in %s", vectorsDir)
	}
}

func loadVectors(t *testing.T, path string) []vector {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file vectorFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return file.Vectors
}

func decodeVector(t *testing.T, raw json.RawMessage, into any) {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
}

func runDealVector(t *testing.T, v vector) {
	var given struct {
		Hands    map[Seat]Hand
		Boneyard Boneyard
	}
	var expect struct {
		Valid bool
	}
	decodeVector(t, v.Given, &given)
	decodeVector(t, v.Expect, &expect)
	err := ValidateDeal(v.Mode, DefaultConfig(), Deal{Hands: given.Hands, Boneyard: given.Boneyard})
	if (err == nil) != expect.Valid {
		t.Fatalf("valid = %v (%v), want %v", err == nil, err, expect.Valid)
	}
}

type vectorBoard struct {
	Line    []Oriented
	Spinner *Oriented
	Top     []Oriented
	Bottom  []Oriented
}

func boardFromVector(t *testing.T, vb vectorBoard) Board {
	t.Helper()
	board := Board{head: SetSize, tail: SetSize}
	for _, tile := range vb.Line {
		board.line[board.tail] = tile
		board.tail++
	}
	for _, tile := range vb.Top {
		board.top.push(tile)
	}
	for _, tile := range vb.Bottom {
		board.bottom.push(tile)
	}
	if vb.Spinner != nil {
		board.spinner = *vb.Spinner
		board.hasSpinner = true
	}
	return board
}

func boardToVector(board *Board) vectorBoard {
	vb := vectorBoard{
		Line:   append([]Oriented{}, board.Line()...),
		Top:    append([]Oriented{}, board.Top()...),
		Bottom: append([]Oriented{}, board.Bottom()...),
	}
	if board.hasSpinner {
		spinner := board.spinner
		vb.Spinner = &spinner
	}
	return vb
}

func openEndsToVector(ends OpenEnds) map[End]int {
	values := map[End]int{}
	for end := EndLeft; end < endCount; end++ {
		if value, open := ends.Value(end); open {
			values[end] = value
		}
	}
	return values
}

func runPlaceVector(t *testing.T, v vector) {
	var given struct {
		Board                vectorBoard
		Tile                 Tile
		End                  End
		TeamScoreAtHandStart int
	}
	var expect struct {
		Legal    bool
		Board    vectorBoard
		OpenEnds map[End]int
		Total    int
		Score    int
	}
	decodeVector(t, v.Given, &given)
	decodeVector(t, v.Expect, &expect)
	board := boardFromVector(t, given.Board)
	before := board
	err := board.Place(given.Tile, given.End)
	if (err == nil) != expect.Legal {
		t.Fatalf("legal = %v (%v), want %v", err == nil, err, expect.Legal)
	}
	if !expect.Legal {
		if board != before {
			t.Fatal("a refused placement changed the board")
		}
		return
	}
	if got := boardToVector(&board); !reflect.DeepEqual(got, expect.Board) {
		t.Fatalf("board = %+v, want %+v", got, expect.Board)
	}
	ends := board.OpenEnds()
	if got := openEndsToVector(ends); !reflect.DeepEqual(got, expect.OpenEnds) {
		t.Fatalf("openEnds = %v, want %v", got, expect.OpenEnds)
	}
	if ends.Total() != expect.Total {
		t.Fatalf("total = %d, want %d", ends.Total(), expect.Total)
	}
}

func runDrawVector(t *testing.T, v vector) {
	var given struct {
		Board    vectorBoard
		Hand     Hand
		Boneyard Boneyard
	}
	var expect struct {
		Drawn   []Tile
		Outcome TurnOutcome
	}
	decodeVector(t, v.Given, &given)
	decodeVector(t, v.Expect, &expect)
	board := boardFromVector(t, given.Board)
	var buffer [SetSize]Tile
	drawn, outcome := DrawUntilPlayable(v.Mode, &board, &given.Hand, &given.Boneyard, buffer[:0])
	if outcome != expect.Outcome {
		t.Fatalf("outcome = %v, want %v", outcome, expect.Outcome)
	}
	if !sameTiles(drawn, expect.Drawn) {
		t.Fatalf("drawn = %v, want %v", drawn, expect.Drawn)
	}
	if (outcome == OutcomePlace) != board.HasLegalMove(given.Hand) {
		t.Fatalf("outcome %v but HasLegalMove %v", outcome, board.HasLegalMove(given.Hand))
	}
}
