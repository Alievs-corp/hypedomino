package cosmetics

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrBadRarities = errors.New("cosmetics: bad rarity ordinals")

type Rarities map[string]int

func DefaultRarities() Rarities {
	return Rarities{"common": 1, "rare": 2, "epic": 3, "legendary": 4, "mythic": 5}
}

func ParseRarities(spec string) (Rarities, error) {
	rarities := Rarities{}
	for _, field := range strings.Split(spec, ",") {
		id, ordinal, ok := strings.Cut(strings.TrimSpace(field), ":")
		value, err := strconv.Atoi(ordinal)
		if !ok || id == "" || err != nil {
			return nil, fmt.Errorf("%w: %q is not id:ordinal", ErrBadRarities, field)
		}
		if _, dup := rarities[id]; dup {
			return nil, fmt.Errorf("%w: %s listed twice", ErrBadRarities, id)
		}
		rarities[id] = value
	}
	return rarities, rarities.Validate()
}

func (r Rarities) Validate() error {
	seen := make(map[int]string, len(r))
	for id, ordinal := range r {
		if ordinal < 1 || ordinal > len(r) {
			return fmt.Errorf("%w: %s has ordinal %d outside 1..%d", ErrBadRarities, id, ordinal, len(r))
		}
		if other, dup := seen[ordinal]; dup {
			return fmt.Errorf("%w: %s and %s share ordinal %d", ErrBadRarities, other, id, ordinal)
		}
		seen[ordinal] = id
	}
	if len(r) == 0 {
		return fmt.Errorf("%w: empty", ErrBadRarities)
	}
	return nil
}

type Draw [32]byte

func NewDraw() Draw {
	var draw Draw
	if _, err := rand.Read(draw[:]); err != nil {
		panic(err)
	}
	return draw
}

func PickTable(ordinals []int, draw Draw) int {
	highest := -1
	var tied []int
	for seat, ordinal := range ordinals {
		switch {
		case ordinal > highest:
			highest = ordinal
			tied = append(tied[:0], seat)
		case ordinal == highest:
			tied = append(tied, seat)
		}
	}
	if len(tied) == 1 {
		return tied[0]
	}
	return tied[int(binary.BigEndian.Uint64(draw[:8])%uint64(len(tied)))]
}

func (c *Catalog) Table(equipped []string, rarities Rarities, draw Draw) string {
	tables := make([]string, len(equipped))
	ordinals := make([]int, len(equipped))
	for seat, id := range equipped {
		item, ok := c.Item(id)
		if !ok || item.Type != TypeTable || !item.Equippable() {
			item, _ = c.Default(TypeTable)
		}
		tables[seat] = item.ID
		ordinals[seat] = rarities[item.Rarity]
	}
	return tables[PickTable(ordinals, draw)]
}
