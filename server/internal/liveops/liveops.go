package liveops

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const Domain = "liveops"

var ErrInvalid = errors.New("liveops: invalid liveops snapshot")

type Key struct {
	Name    string
	Min     int
	Default int
}

var keys = []Key{
	{Name: "ContentGateMinAvatarsAnimated", Default: 30},
	{Name: "ContentGateMinAvatarsStatic", Default: 60},
	{Name: "ContentGateMinFrames", Default: 20},
	{Name: "ContentGateMinTables", Default: 15},
	{Name: "ContentGateMinEmoji", Default: 40},
	{Name: "ContentGateMinMedals", Default: 40},
}

func Keys() []Key {
	return keys
}

type Values map[string]int

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
		values[name] = number
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
