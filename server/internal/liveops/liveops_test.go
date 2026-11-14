package liveops

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func documentedDefaults(t *testing.T) map[string]string {
	t.Helper()
	doc, err := os.ReadFile("../../../docs/launch-content-gate.md")
	if err != nil {
		t.Fatal(err)
	}
	_, table, found := strings.Cut(string(doc), "| Key | Default | Section |\n|---|---|---|\n")
	if !found {
		t.Fatal("launch-content-gate.md has no config table")
	}
	rows := map[string]string{}
	for _, line := range strings.Split(table, "\n") {
		if !strings.HasPrefix(line, "| ") {
			break
		}
		cells := strings.Split(strings.Trim(line, "| "), " | ")
		rows[strings.Trim(cells[0], "`")] = cells[1]
	}
	return rows
}

func TestTheKeysMatchTheConfigTableOfTheContentGateSpec(t *testing.T) {
	documented := documentedDefaults(t)
	if len(documented) != len(Keys()) {
		t.Fatalf("launch-content-gate.md lists %d keys, code has %d", len(documented), len(Keys()))
	}
	for _, key := range Keys() {
		if got := documented[key.Name]; got != strconv.Itoa(key.Default) {
			t.Errorf("%s: spec default %q, code %d", key.Name, got, key.Default)
		}
		if key.Min != 0 || key.Default < key.Min {
			t.Errorf("%s: min %d, default %d, spec says integers >= 0", key.Name, key.Min, key.Default)
		}
	}
}

func TestSnapshotsAreCheckedAgainstTheKeys(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Values
	}{
		{"every key", `{"ContentGateMinAvatarsAnimated": 31, "ContentGateMinAvatarsStatic": 61, "ContentGateMinFrames": 21, "ContentGateMinTables": 16, "ContentGateMinEmoji": 41, "ContentGateMinMedals": 41}`, Values{"ContentGateMinAvatarsAnimated": 31, "ContentGateMinAvatarsStatic": 61, "ContentGateMinFrames": 21, "ContentGateMinTables": 16, "ContentGateMinEmoji": 41, "ContentGateMinMedals": 41}},
		{"one key", `{"ContentGateMinFrames": 25}`, Values{"ContentGateMinFrames": 25}},
		{"zero", `{"ContentGateMinEmoji": 0}`, Values{"ContentGateMinEmoji": 0}},
		{"below the default", `{"ContentGateMinMedals": 10}`, Values{"ContentGateMinMedals": 10}},
		{"large", `{"ContentGateMinTables": 1000}`, Values{"ContentGateMinTables": 1000}},
		{"no keys", `{}`, Values{}},
	}
	for _, test := range tests {
		got, err := Parse([]byte(test.raw))
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: %v %v", test.name, got, err)
		}
	}
	for name, raw := range map[string]string{
		"unknown key": `{"ContentGateMinChests": 5}`,
		"rules key":   `{"TurnSeconds": 20}`,
		"negative":    `{"ContentGateMinFrames": -1}`,
		"fraction":    `{"ContentGateMinFrames": 20.5}`,
		"string":      `{"ContentGateMinFrames": "20"}`,
		"null value":  `{"ContentGateMinFrames": null}`,
		"boolean":     `{"ContentGateMinFrames": true}`,
		"array":       `[20]`,
		"null":        `null`,
		"not JSON":    `ContentGateMinFrames=20`,
	} {
		if _, err := Parse([]byte(raw)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s accepted: %v", name, err)
		}
	}
}
