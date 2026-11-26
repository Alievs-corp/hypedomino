package anticheat

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const Domain = "anticheat"

var ErrInvalid = errors.New("anticheat: invalid anti-cheat snapshot")

const (
	IllegalIntentThreshold     = "IllegalIntentThreshold"
	IllegalIntentWindowMinutes = "IllegalIntentWindowMinutes"
	IllegalIntentWeight        = "IllegalIntentWeight"
	MaxWindowMinutes           = 30 * 24 * 60
)

type Key struct {
	Name    string
	Unit    string
	Min     int
	Max     *int
	Default int
}

var keys = []Key{
	{Name: IllegalIntentThreshold, Unit: "intents"},
	{Name: IllegalIntentWindowMinutes, Unit: "minutes", Max: bound(MaxWindowMinutes)},
	{Name: IllegalIntentWeight, Unit: "points"},
}

func bound(limit int) *int {
	return &limit
}

func Keys() []Key {
	return keys
}

type Values map[string]int

type IllegalIntents struct {
	Threshold int
	Window    time.Duration
	Weight    int
}

func (r IllegalIntents) Enabled() bool {
	return r.Threshold > 0 && r.Window > 0
}

func (v Values) IllegalIntents() IllegalIntents {
	return IllegalIntents{
		Threshold: v.value(IllegalIntentThreshold),
		Window:    time.Duration(v.value(IllegalIntentWindowMinutes)) * time.Minute,
		Weight:    v.value(IllegalIntentWeight),
	}
}

func (v Values) value(name string) int {
	if value, set := v[name]; set {
		return value
	}
	key, _ := lookup(name)
	return key.Default
}

func Parse(raw []byte) (Values, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("%w: not a JSON object: %v", ErrInvalid, err)
	}
	values := Values{}
	for name, field := range fields {
		key, known := lookup(name)
		if !known {
			return nil, fmt.Errorf("%w: unknown key %s", ErrInvalid, name)
		}
		number, err := strconv.Atoi(string(field))
		if err != nil {
			return nil, fmt.Errorf("%w: %s is not an integer: %s", ErrInvalid, name, field)
		}
		if number < key.Min {
			return nil, fmt.Errorf("%w: %s %d below %d", ErrInvalid, name, number, key.Min)
		}
		if key.Max != nil && number > *key.Max {
			return nil, fmt.Errorf("%w: %s %d above %d", ErrInvalid, name, number, *key.Max)
		}
		values[name] = number
	}
	if values.value(IllegalIntentThreshold) > 0 && values.value(IllegalIntentWindowMinutes) < 1 {
		return nil, fmt.Errorf("%w: %s needs %s of at least 1", ErrInvalid, IllegalIntentThreshold, IllegalIntentWindowMinutes)
	}
	return values, nil
}

func lookup(name string) (Key, bool) {
	for _, key := range keys {
		if key.Name == name {
			return key, true
		}
	}
	return Key{}, false
}
