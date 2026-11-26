package anticheat_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/anticheat"
)

func TestTheShippedDefaultsLeaveTheIllegalIntentSignalOff(t *testing.T) {
	for _, key := range anticheat.Keys() {
		if key.Default != 0 || key.Min != 0 {
			t.Fatalf("%s has default %d and minimum %d, want 0 and 0 until the owner sets launch values", key.Name, key.Default, key.Min)
		}
	}
	values, err := anticheat.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if rule := values.IllegalIntents(); rule != (anticheat.IllegalIntents{}) || rule.Enabled() {
		t.Fatalf("an empty snapshot gives %+v", rule)
	}
	var unpublished anticheat.Values
	if unpublished.IllegalIntents().Enabled() {
		t.Fatal("a node without a published snapshot counts illegal intents")
	}
}

func TestAnAntiCheatSnapshotIsChecked(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
		rule anticheat.IllegalIntents
	}{
		{"enabled", `{"IllegalIntentThreshold": 3, "IllegalIntentWindowMinutes": 15, "IllegalIntentWeight": 4}`, nil, anticheat.IllegalIntents{Threshold: 3, Window: 15 * time.Minute, Weight: 4}},
		{"weight zero keeps evidence only", `{"IllegalIntentThreshold": 2, "IllegalIntentWindowMinutes": 1}`, nil, anticheat.IllegalIntents{Threshold: 2, Window: time.Minute}},
		{"disabled with a window", `{"IllegalIntentThreshold": 0, "IllegalIntentWindowMinutes": 10}`, nil, anticheat.IllegalIntents{Window: 10 * time.Minute}},
		{"longest window", `{"IllegalIntentThreshold": 1, "IllegalIntentWindowMinutes": 43200}`, nil, anticheat.IllegalIntents{Threshold: 1, Window: 30 * 24 * time.Hour}},
		{"threshold without a window", `{"IllegalIntentThreshold": 3}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"window above thirty days", `{"IllegalIntentThreshold": 1, "IllegalIntentWindowMinutes": 43201}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"negative threshold", `{"IllegalIntentThreshold": -1}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"negative weight", `{"IllegalIntentWeight": -1}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"fraction", `{"IllegalIntentWeight": 1.5}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"string", `{"IllegalIntentWeight": "5"}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"unknown key", `{"MacFailureThreshold": 5}`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"not an object", `[]`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
		{"null", `null`, anticheat.ErrInvalid, anticheat.IllegalIntents{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values, err := anticheat.Parse([]byte(test.raw))
			if !errors.Is(err, test.want) {
				t.Fatalf("error %v, want %v", err, test.want)
			}
			if err == nil && values.IllegalIntents() != test.rule {
				t.Fatalf("rule %+v, want %+v", values.IllegalIntents(), test.rule)
			}
		})
	}
}
