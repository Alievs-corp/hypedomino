package rules

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

func seeded(seed uint64) *rand.Rand {
	var key [32]byte
	for i := range 8 {
		key[i] = byte(seed >> (8 * i))
	}
	return rand.New(rand.NewChaCha8(key))
}

func TestNewDealShape(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		mode     Mode
		boneyard int
	}{
		{Mode2v2, 0},
		{Mode1v1, 14},
	}
	for _, c := range cases {
		deal := NewDeal(c.mode, config, seeded(1))
		if err := ValidateDeal(c.mode, config, deal); err != nil {
			t.Fatalf("%s: %v", c.mode, err)
		}
		if len(deal.Boneyard) != c.boneyard {
			t.Fatalf("%s: boneyard %d, want %d", c.mode, len(deal.Boneyard), c.boneyard)
		}
		for _, seat := range c.mode.Seats() {
			if len(deal.Hands[seat]) != config.HandSize {
				t.Fatalf("%s: seat %s holds %d", c.mode, seat, len(deal.Hands[seat]))
			}
		}
	}
}

func TestNewDealIsDeterministicPerSeed(t *testing.T) {
	first := NewDeal(Mode2v2, DefaultConfig(), seeded(7))
	second := NewDeal(Mode2v2, DefaultConfig(), seeded(7))
	other := NewDeal(Mode2v2, DefaultConfig(), seeded(8))
	for _, seat := range Mode2v2.Seats() {
		if !sameTiles(first.Hands[seat], second.Hands[seat]) {
			t.Fatalf("seat %s differs for the same seed", seat)
		}
	}
	if sameTiles(first.Hands[SeatNorth], other.Hands[SeatNorth]) {
		t.Fatal("different seeds dealt the same hand to North")
	}
}

func sameTiles(a, b []Tile) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestValidateDealRejectsBrokenDeals(t *testing.T) {
	config := DefaultConfig()
	valid := NewDeal(Mode1v1, config, seeded(3))
	if err := ValidateDeal(Mode1v1, config, valid); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDeal(Mode2v2, config, valid); !errors.Is(err, ErrInvalidDeal) {
		t.Fatalf("1v1 deal accepted as 2v2: %v", err)
	}
	if err := ValidateDeal(Mode(0), config, valid); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("invalid mode accepted: %v", err)
	}
	duplicated := NewDeal(Mode1v1, config, seeded(3))
	duplicated.Hands[SeatX][0] = duplicated.Hands[SeatY][0]
	if err := ValidateDeal(Mode1v1, config, duplicated); !errors.Is(err, ErrInvalidDeal) {
		t.Fatalf("duplicate tile accepted: %v", err)
	}
	corrupt := NewDeal(Mode1v1, config, seeded(3))
	corrupt.Boneyard[0] = Tile{Low: 7, High: 7}
	if err := ValidateDeal(Mode1v1, config, corrupt); !errors.Is(err, ErrInvalidTile) {
		t.Fatalf("invalid tile accepted: %v", err)
	}
	renamed := NewDeal(Mode1v1, config, seeded(3))
	renamed.Hands[SeatNorth] = renamed.Hands[SeatX]
	delete(renamed.Hands, SeatX)
	if err := ValidateDeal(Mode1v1, config, renamed); !errors.Is(err, ErrInvalidDeal) {
		t.Fatalf("wrong seat accepted: %v", err)
	}
}

func TestShuffleIsUniform(t *testing.T) {
	const rounds = 28 * 400
	random := seeded(2026)
	var counts [SetSize][SetSize]int
	for round := 0; round < rounds; round++ {
		set := FullSet()
		Shuffle(set[:], random)
		for position, tile := range set {
			counts[tileIndex(tile)][position]++
		}
	}
	expected := float64(rounds) / SetSize
	chiSquare := 0.0
	for _, row := range counts {
		for _, count := range row {
			delta := float64(count) - expected
			chiSquare += delta * delta / expected
		}
	}
	degrees := float64((SetSize - 1) * (SetSize - 1))
	limit := degrees + 3.09*math.Sqrt(2*degrees)
	if chiSquare > limit {
		t.Fatalf("chi-square %.1f exceeds %.1f for %.0f degrees of freedom", chiSquare, limit, degrees)
	}
}

func FuzzNewDeal(f *testing.F) {
	f.Add(uint64(0), uint8(1))
	f.Add(uint64(1), uint8(2))
	f.Add(uint64(math.MaxUint64), uint8(2))
	f.Fuzz(func(t *testing.T, seed uint64, modeByte uint8) {
		mode := Mode(modeByte%2 + 1)
		config := DefaultConfig()
		deal := NewDeal(mode, config, seeded(seed))
		if err := ValidateDeal(mode, config, deal); err != nil {
			t.Fatalf("seed %d %s: %v", seed, mode, err)
		}
		seen := map[Tile]bool{}
		for _, seat := range mode.Seats() {
			for _, tile := range deal.Hands[seat] {
				seen[tile] = true
			}
		}
		for _, tile := range deal.Boneyard {
			seen[tile] = true
		}
		if len(seen) != SetSize {
			t.Fatalf("seed %d %s: %d distinct tiles", seed, mode, len(seen))
		}
	})
}

func BenchmarkNewDeal(b *testing.B) {
	random := seeded(1)
	config := DefaultConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewDeal(Mode2v2, config, random)
	}
}
