package shop

import (
	"strings"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/reward"
)

func TestOffersAndTheFirstPurchaseGrantOnlyItemsListingThem(t *testing.T) {
	items, err := cosmetics.Load("../../../docs/cosmetics/examples/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*Catalog){
		"shop offer.starter grants avatar.paper_kite": func(c *Catalog) {
			for i := range c.Offers {
				if c.Offers[i].ID == "offer.starter" {
					c.Offers[i].Bundle.Cosmetics = []string{"avatar.comet_fox", "avatar.paper_kite"}
				}
			}
		},
		"shop offer.table_night_market grants avatar.missing": func(c *Catalog) {
			for i := range c.Offers {
				if c.Offers[i].ID == "offer.table_night_market" {
					c.Offers[i].Bundle.Cosmetics = []string{"avatar.missing"}
				}
			}
		},
		"first_purchase bundle.first_purchase grants avatar.comet_fox": func(c *Catalog) {
			bonus := *c.FirstPurchase
			bonus.Cosmetics = []string{"avatar.comet_fox"}
			c.FirstPurchase = &bonus
		},
	}
	for reason, change := range cases {
		catalog := example(t)
		change(catalog)
		if err := catalog.CheckSources(items); err == nil || !strings.Contains(err.Error(), reason) {
			t.Fatalf("%s: %v", reason, err)
		}
	}
	catalog := example(t)
	catalog.FirstPurchase = &reward.Bundle{ID: "bundle.coins_only", Crystals: 100}
	if err := catalog.CheckSources(items); err != nil {
		t.Fatalf("a bonus without items: %v", err)
	}
}
