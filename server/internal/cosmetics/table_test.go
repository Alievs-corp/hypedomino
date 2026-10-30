package cosmetics

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestTheDefaultOrdinalsMatchTheDesignTokens(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/design-tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens struct {
		Rarity struct {
			Tiers map[string]struct {
				Ordinal int `json:"ordinal"`
			} `json:"tiers"`
		} `json:"rarity"`
	}
	if err := json.Unmarshal(raw, &tokens); err != nil {
		t.Fatal(err)
	}
	defaults := DefaultRarities()
	if len(tokens.Rarity.Tiers) != len(defaults) {
		t.Fatalf("design tokens list %d tiers, defaults %d", len(tokens.Rarity.Tiers), len(defaults))
	}
	for id, tier := range tokens.Rarity.Tiers {
		if defaults[id] != tier.Ordinal {
			t.Fatalf("%s ordinal %d in design tokens, %d in defaults", id, tier.Ordinal, defaults[id])
		}
	}
	if err := defaults.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRarityOrdinalsParseAndAreGapless(t *testing.T) {
	parsed, err := ParseRarities(" common:1, rare:2 ,epic:3")
	if err != nil || len(parsed) != 3 || parsed["rare"] != 2 {
		t.Fatalf("parsed %v %v", parsed, err)
	}
	for spec, reason := range map[string]string{
		"":                  "not id:ordinal",
		"common":            "not id:ordinal",
		":1":                "not id:ordinal",
		"common:x":          "not id:ordinal",
		"common:1,common:2": "listed twice",
		"common:1,rare:3":   "outside 1..2",
		"common:0,rare:1":   "outside",
		"common:1,rare:1":   "share ordinal 1",
	} {
		if _, err := ParseRarities(spec); !errors.Is(err, ErrBadRarities) || !strings.Contains(err.Error(), reason) {
			t.Fatalf("%q: %v", spec, err)
		}
	}
	if err := (Rarities{}).Validate(); !errors.Is(err, ErrBadRarities) {
		t.Fatalf("empty table: %v", err)
	}
}

func draw(t *testing.T, first8 string) Draw {
	t.Helper()
	var d Draw
	decoded, err := hex.DecodeString(first8)
	if err != nil || len(decoded) != 8 {
		t.Fatal("bad hex", first8, err)
	}
	copy(d[:], decoded)
	for i := 8; i < len(d); i++ {
		d[i] = 0xAB
	}
	return d
}

func TestPickTableFollowsTheWorkedExamples(t *testing.T) {
	cases := []struct {
		ordinals []int
		first8   string
		want     int
	}{
		{[]int{4, 1, 2, 3}, "0000000000000000", 0},
		{[]int{3, 5, 1, 5}, "0000000000000003", 3},
		{[]int{2, 2, 2, 2}, "FFFFFFFFFFFFFFFF", 3},
		{[]int{1, 1}, "000000000000000A", 0},
		{[]int{1, 1}, "000000000000000B", 1},
		{[]int{0, 0, 0}, "0000000000000004", 1},
	}
	for _, c := range cases {
		if got := PickTable(c.ordinals, draw(t, c.first8)); got != c.want {
			t.Fatalf("PickTable(%v, %s) = %d, want %d", c.ordinals, c.first8, got, c.want)
		}
	}
	tail := draw(t, "0000000000000001")
	tail[9], tail[31] = 0x11, 0x22
	if PickTable([]int{1, 1}, tail) != 1 {
		t.Fatal("bytes after the first eight influenced the choice")
	}
}

func TestTiedSeatsAreChosenUniformly(t *testing.T) {
	for _, k := range []int{2, 3, 4} {
		const samples = 60000
		counts := make([]int, k)
		ordinals := make([]int, k)
		for range samples {
			counts[PickTable(ordinals, NewDraw())]++
		}
		expected := samples / k
		for seat, count := range counts {
			if count < expected*97/100 || count > expected*103/100 {
				t.Fatalf("k=%d seat %d chosen %d times, expected about %d: %v", k, seat, count, expected, counts)
			}
		}
	}
}

func TestTheCatalogResolvesEquippedTablesBeforePicking(t *testing.T) {
	catalog, err := Load("../../../docs/cosmetics/examples/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	rarities := DefaultRarities()
	night, ok := catalog.Item("table.night_market")
	if !ok || rarities[night.Rarity] <= rarities["common"] {
		t.Fatalf("night market %+v", night)
	}
	if got := catalog.Table([]string{"table.starter", "table.night_market"}, rarities, draw(t, "0000000000000000")); got != "table.night_market" {
		t.Fatalf("highest rarity: %s", got)
	}
	if got := catalog.Table([]string{"", "table.nope", "avatar.starter", "frame.withdrawn"}, rarities, draw(t, "0000000000000002")); got != "table.starter" {
		t.Fatalf("missing, unknown, wrong type and disabled tables: %s", got)
	}
	if got := catalog.Table([]string{"table.starter", "table.starter"}, rarities, draw(t, "0000000000000001")); got != "table.starter" {
		t.Fatalf("tied defaults: %s", got)
	}
	if got := catalog.Table([]string{"table.night_market"}, Rarities{"common": 1}, draw(t, "0000000000000000")); got != "table.night_market" {
		t.Fatalf("unknown rarity still returns the only table: %s", got)
	}
}
