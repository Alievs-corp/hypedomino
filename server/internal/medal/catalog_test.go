package medal

import (
	"errors"
	"os"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
)

func example(t *testing.T) *Catalog {
	t.Helper()
	raw, err := os.ReadFile("../../../docs/pass/examples/medals.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestTheExampleHoldsTheSampleMedals(t *testing.T) {
	catalog := example(t)
	if len(catalog.Medals) != 10 || catalog.ProfileMedalSlots != 3 {
		t.Fatalf("%d medals, %d slots", len(catalog.Medals), catalog.ProfileMedalSlots)
	}
	medal, ok := catalog.Medal("medal.wins_5000")
	if !ok || medal.RarityID != "legendary" || !medal.Lists(medal.Sources[0]) {
		t.Fatalf("medal.wins_5000 = %+v", medal)
	}
}

func TestPublishRulesRefuseBrokenMedals(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Medal)
	}{
		{"bad id", func(m *Medal) { m.ID = "badge.wins_50" }},
		{"revision zero", func(m *Medal) { m.Revision = 0 }},
		{"name key of another slug", func(m *Medal) { m.NameKey = "medal.other.name" }},
		{"description key missing", func(m *Medal) { m.DescriptionKey = "" }},
		{"static address of another slug", func(m *Medal) { m.StaticAddress = "medal/other/static" }},
		{"unknown rarity", func(m *Medal) { m.RarityID = "shiny" }},
		{"unknown availability", func(m *Medal) { m.Availability = "hidden" }},
		{"no source", func(m *Medal) { m.Sources = nil }},
		{"shop source", func(m *Medal) { m.Sources = []cosmetics.Source{{Type: "shop", Ref: "offer.x"}} }},
		{"pass source", func(m *Medal) { m.Sources = []cosmetics.Source{{Type: "pass", Ref: "pass.season_1.premium.level_10"}} }},
		{"source without ref", func(m *Medal) { m.Sources = []cosmetics.Source{{Type: SourceMission}} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			catalog := example(t)
			c.change(&catalog.Medals[0])
			if err := catalog.Validate(cosmetics.DefaultRarities()); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("Validate() = %v", err)
			}
		})
	}
}

func TestAMedalIdIsListedOnce(t *testing.T) {
	catalog := example(t)
	catalog.Medals = append(catalog.Medals, catalog.Medals[0])
	if err := catalog.Validate(cosmetics.DefaultRarities()); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("Validate() = %v", err)
	}
}

func TestNegativeSlotsAreRefused(t *testing.T) {
	catalog := example(t)
	catalog.ProfileMedalSlots = -1
	if err := catalog.Validate(cosmetics.DefaultRarities()); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("Validate() = %v", err)
	}
}
