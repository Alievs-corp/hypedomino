package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	"deal": runDealVector,
}

var pendingVectorKinds = map[string]bool{
	"seat":          true,
	"first-opener":  true,
	"opening-score": true,
	"place":         true,
	"turn":          true,
	"draw":          true,
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
