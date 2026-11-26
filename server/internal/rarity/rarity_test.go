package rarity

import (
	"encoding/json"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

func defaultsWith(t *testing.T, change func(table map[string]any)) []byte {
	t.Helper()
	var table map[string]any
	if err := json.Unmarshal(DefaultJSON(), &table); err != nil {
		t.Fatal(err)
	}
	change(table)
	raw, err := json.Marshal(table)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func tier(table map[string]any, id string) map[string]any {
	return table["tiers"].(map[string]any)[id].(map[string]any)
}

func TestTheBuiltinDefaultsAreAValidTable(t *testing.T) {
	table := Default()
	if table.DiamondMinSize != 72 || table.FlowPeriodMs != 2400 || len(table.Tiers) != 5 {
		t.Fatalf("defaults %+v", table)
	}
	if mythic := table.Tiers["mythic"]; mythic.Ordinal != 5 || mythic.Pips != 5 || mythic.Glow != "intense" || !mythic.Flow || mythic.Fill != "#FF4530" {
		t.Fatalf("mythic %+v", mythic)
	}
}

func TestGlowLevelsRunFromNoneToIntense(t *testing.T) {
	levels := GlowLevels()
	want := []string{"none", "soft", "strong", "intense"}
	if len(levels) != len(want) {
		t.Fatalf("levels %v", levels)
	}
	for at := range want {
		if levels[at] != want[at] {
			t.Fatalf("levels %v, want %v", levels, want)
		}
	}
}

func TestParseRefusesATableThatBreaksASectionSevenRule(t *testing.T) {
	cases := []struct {
		name   string
		change func(table map[string]any)
	}{
		{"not positive diamond size", func(table map[string]any) { table["diamondMinSize"] = 0 }},
		{"not positive flow period", func(table map[string]any) { table["flowPeriodMs"] = -1 }},
		{"unknown field", func(table map[string]any) { table["colour"] = "red" }},
		{"unknown tier field", func(table map[string]any) { tier(table, "rare")["shine"] = true }},
		{"no tiers", func(table map[string]any) { table["tiers"] = map[string]any{} }},
		{"bad id", func(table map[string]any) {
			tiers := table["tiers"].(map[string]any)
			tiers["Cosmic"] = map[string]any{"ordinal": 6, "nameKey": "rarity.Cosmic.name", "fill": "#FF4530", "on": "#0B0605", "text": "#FF9C8F", "pips": 6, "glow": "intense", "flow": true}
		}},
		{"name key of another tier", func(table map[string]any) { tier(table, "epic")["nameKey"] = "rarity.rare.name" }},
		{"fill not a hex colour", func(table map[string]any) { tier(table, "epic")["fill"] = "purple" }},
		{"short hex", func(table map[string]any) { tier(table, "epic")["text"] = "#FFF" }},
		{"no pips", func(table map[string]any) { tier(table, "common")["pips"] = 0 }},
		{"seven pips", func(table map[string]any) { tier(table, "mythic")["pips"] = 7 }},
		{"unknown glow", func(table map[string]any) { tier(table, "rare")["glow"] = "blinding" }},
		{"fill too dark on the ground", func(table map[string]any) { tier(table, "rare")["fill"] = "#202040" }},
		{"text too dark on the ground", func(table map[string]any) { tier(table, "rare")["text"] = "#5C78FF" }},
		{"pips too close to the fill", func(table map[string]any) { tier(table, "legendary")["on"] = "#FFD98A" }},
		{"ordinal gap", func(table map[string]any) { tier(table, "mythic")["ordinal"] = 6 }},
		{"ordinal repeat", func(table map[string]any) { tier(table, "mythic")["ordinal"] = 4 }},
		{"shared pips", func(table map[string]any) { tier(table, "mythic")["pips"] = 4 }},
		{"glow weaker than a lower tier", func(table map[string]any) { tier(table, "legendary")["glow"] = "soft" }},
		{"builtin tier removed", func(table map[string]any) {
			tiers := table["tiers"].(map[string]any)
			delete(tiers, "mythic")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse(defaultsWith(t, c.change)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("parse error %v, want ErrInvalid", err)
			}
		})
	}
}

func TestParseRefusesSomethingThatIsNotATable(t *testing.T) {
	for _, raw := range []string{``, `[]`, `{"tiers": 3}`, `null x`} {
		if _, err := Parse([]byte(raw)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%q parsed with error %v", raw, err)
		}
	}
}

func TestParseAcceptsATunedTableAndASixthTier(t *testing.T) {
	cases := []struct {
		name   string
		change func(table map[string]any)
		check  func(t *testing.T, table Table)
	}{
		{"larger diamond", func(table map[string]any) { table["diamondMinSize"] = 96 }, func(t *testing.T, table Table) {
			if table.DiamondMinSize != 96 {
				t.Fatalf("diamond %d", table.DiamondMinSize)
			}
		}},
		{"lowercase hex", func(table map[string]any) { tier(table, "rare")["fill"] = "#5c78ff" }, func(t *testing.T, table Table) {
			if table.Tiers["rare"].Fill != "#5c78ff" {
				t.Fatalf("rare %+v", table.Tiers["rare"])
			}
		}},
		{"sixth tier", func(table map[string]any) {
			tiers := table["tiers"].(map[string]any)
			tiers["cosmic"] = map[string]any{"ordinal": 6, "nameKey": "rarity.cosmic.name", "fill": "#FF4530", "on": "#0B0605", "text": "#FF9C8F", "pips": 6, "glow": "intense", "flow": true}
		}, func(t *testing.T, table Table) {
			if len(table.Tiers) != 6 || table.Tiers["cosmic"].Ordinal != 6 {
				t.Fatalf("tiers %+v", table.Tiers)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			table, err := Parse(defaultsWith(t, c.change))
			if err != nil {
				t.Fatal(err)
			}
			c.check(t, table)
		})
	}
}

func TestHelloListsTiersByOrdinalWithTheirGlowValues(t *testing.T) {
	got := Default().Hello()
	want := &pb.RarityTableConfig{DiamondMinSize: 72, FlowPeriodMs: 2400, Tiers: []*pb.RarityTierConfig{
		{Id: "common", Ordinal: 1, NameKey: "rarity.common.name", Fill: "#A3A6BF", On: "#0B0605", Text: "#C9CBDB", Pips: 1, GlowAlpha: 0, GlowScale: 1},
		{Id: "rare", Ordinal: 2, NameKey: "rarity.rare.name", Fill: "#5C78FF", On: "#0B0605", Text: "#A3B3FF", Pips: 2, GlowAlpha: 0.35, GlowScale: 1.15},
		{Id: "epic", Ordinal: 3, NameKey: "rarity.epic.name", Fill: "#B45CFF", On: "#0B0605", Text: "#D3A3FF", Pips: 3, GlowAlpha: 0.55, GlowScale: 1.3},
		{Id: "legendary", Ordinal: 4, NameKey: "rarity.legendary.name", Fill: "#FFC23D", On: "#0B0605", Text: "#FFD98A", Pips: 4, GlowAlpha: 0.8, GlowScale: 1.5},
		{Id: "mythic", Ordinal: 5, NameKey: "rarity.mythic.name", Fill: "#FF4530", On: "#0B0605", Text: "#FF9C8F", Pips: 5, GlowAlpha: 0.8, GlowScale: 1.5, Flow: true},
	}}
	if !proto.Equal(got, want) {
		t.Fatalf("hello table %v", got)
	}
}

func TestContrastTruncatesLikeTheTokenChecker(t *testing.T) {
	cases := []struct {
		first, second string
		want          float64
	}{
		{"#A3A6BF", "#0B0605", 8.40},
		{"#5C78FF", "#4B1D0F", 3.76},
		{"#0B0605", "#FFC23D", 12.50},
		{"#FFFFFF", "#000000", 21},
	}
	for _, c := range cases {
		if got := contrast(c.first, c.second); got != c.want {
			t.Fatalf("contrast %s on %s = %v, want %v", c.first, c.second, got, c.want)
		}
	}
}
