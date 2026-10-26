package rules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	MaxPip  = 6
	SetSize = 28
)

var ErrInvalidTile = errors.New("rules: invalid tile")

type Tile struct {
	Low  uint8
	High uint8
}

func NewTile(a, b uint8) (Tile, error) {
	if a > MaxPip || b > MaxPip {
		return Tile{}, fmt.Errorf("%w: pips %d-%d", ErrInvalidTile, a, b)
	}
	if a > b {
		a, b = b, a
	}
	return Tile{Low: a, High: b}, nil
}

func MustTile(a, b uint8) Tile {
	tile, err := NewTile(a, b)
	if err != nil {
		panic(err)
	}
	return tile
}

func ParseTile(text string) (Tile, error) {
	halves := strings.Split(text, "-")
	if len(halves) != 2 {
		return Tile{}, fmt.Errorf("%w: %q", ErrInvalidTile, text)
	}
	low, err := parsePip(halves[0])
	if err != nil {
		return Tile{}, fmt.Errorf("%w: %q", ErrInvalidTile, text)
	}
	high, err := parsePip(halves[1])
	if err != nil {
		return Tile{}, fmt.Errorf("%w: %q", ErrInvalidTile, text)
	}
	return NewTile(low, high)
}

func parsePip(text string) (uint8, error) {
	value, err := strconv.ParseUint(text, 10, 8)
	if err != nil || value > MaxPip {
		return 0, ErrInvalidTile
	}
	return uint8(value), nil
}

func (t Tile) String() string {
	return strconv.Itoa(int(t.Low)) + "-" + strconv.Itoa(int(t.High))
}

func (t Tile) IsDouble() bool {
	return t.Low == t.High
}

func (t Tile) Pips() int {
	return int(t.Low) + int(t.High)
}

func (t Tile) Has(pip uint8) bool {
	return t.Low == pip || t.High == pip
}

func (t Tile) Valid() bool {
	return t.Low <= t.High && t.High <= MaxPip
}

func (t Tile) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Tile) UnmarshalText(text []byte) error {
	parsed, err := ParseTile(string(text))
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

func FullSet() [SetSize]Tile {
	var set [SetSize]Tile
	index := 0
	for low := uint8(0); low <= MaxPip; low++ {
		for high := low; high <= MaxPip; high++ {
			set[index] = Tile{Low: low, High: high}
			index++
		}
	}
	return set
}
