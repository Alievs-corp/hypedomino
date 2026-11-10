package adminapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

func weekendPack(opens, closes time.Time) map[string]any {
	return map[string]any{
		"id": "offer.weekend_pack", "kind": "bundle", "section": "featured", "priceCrystals": 90,
		"bundle":  map[string]any{"id": "bundle.weekend_pack", "revision": 1, "cosmetics": []string{"avatar.comet_fox"}},
		"window":  map[string]any{"start": opens, "end": closes},
		"nameKey": "shop.offer.weekend_pack",
	}
}

func cometFox() map[string]any {
	return map[string]any{
		"id": "offer.avatar_comet_fox", "kind": "cosmetic", "section": "daily_rotation", "priceCoins": 400, "limitPerAccount": 1,
		"bundle":  map[string]any{"id": "bundle.avatar_comet_fox", "revision": 1, "cosmetics": []string{"avatar.comet_fox"}},
		"window":  map[string]any{"start": "2026-01-01T00:00:00Z"},
		"nameKey": "shop.offer.avatar_comet_fox",
	}
}

func catalog(offers ...map[string]any) map[string]any {
	return map[string]any{
		"schema": 1, "offers": offers,
		"rotations": []map[string]any{{"period": "daily", "start": "2026-11-01T00:00:00Z", "offers": []string{"offer.avatar_comet_fox"}}},
	}
}

func (a *api) publishShopItems() {
	a.t.Helper()
	raw, err := os.ReadFile("../../../docs/cosmetics/examples/catalog.json")
	if err != nil {
		a.t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		a.t.Fatal(err)
	}
	for _, item := range manifest["items"].([]any) {
		if entry := item.(map[string]any); entry["id"] == "avatar.comet_fox" {
			entry["sources"] = append(entry["sources"].([]any), map[string]any{"type": "shop", "ref": "offer.weekend_pack"})
		}
	}
	if raw, err = json.Marshal(manifest); err != nil {
		a.t.Fatal(err)
	}
	if _, err := snapshot.Publish(context.Background(), a.pool, cosmetics.Domain, raw); err != nil {
		a.t.Fatal(err)
	}
}

func (a *api) featuredAt(at time.Time) []string {
	a.t.Helper()
	watcher := snapshot.New(a.pool, shop.Domain, shop.Parser(shop.DefaultConfig()), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := watcher.Load(context.Background()); err != nil && !errors.Is(err, snapshot.ErrNone) {
		a.t.Fatal(err)
	}
	current, _, _ := watcher.Current()
	ids := []string{}
	if current == nil {
		return ids
	}
	for _, section := range current.Sections(shop.DefaultConfig(), at) {
		if section.Name == shop.SectionFeatured {
			for _, offer := range section.Offers {
				ids = append(ids, offer.ID)
			}
		}
	}
	return ids
}

func TestAScheduledOfferOpensAndClosesOnServerTimeAndCanBeRolledBack(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps := a.signIn("leyla")
	opens, closes := start.Add(time.Hour), start.Add(3*time.Hour)

	status, body := a.call("GET", "/admin/config/shop", liveOps, nil)
	if status != http.StatusOK || body["config_version"] != float64(0) || body["catalog"] != nil {
		t.Fatalf("empty %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": catalog(cometFox()), "reason": "too early"})
	if status != http.StatusBadRequest || !strings.Contains(body["detail"].(string), "offer.avatar_comet_fox grants avatar.comet_fox, whose sources do not list it") {
		t.Fatalf("offer before its item %d %v", status, body)
	}
	a.publishShopItems()
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": catalog(cometFox(), weekendPack(opens, closes)), "reason": "weekend sale"})
	if status != http.StatusOK || body["config_version"] != float64(2) || len(body["catalog"].(map[string]any)["offers"].([]any)) != 2 {
		t.Fatalf("publish %d %v", status, body)
	}
	for at, want := range map[time.Time]int{start: 0, opens: 1, closes.Add(-time.Second): 1, closes: 0} {
		if got := a.featuredAt(at); len(got) != want {
			t.Fatalf("featured at %s: %v", at.Format(time.RFC3339), got)
		}
	}

	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": catalog(cometFox())})
	if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), "offer.weekend_pack") {
		t.Fatalf("removed offer %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": catalog(cometFox(), weekendPack(opens, opens.Add(time.Hour)))})
	if status != http.StatusOK || body["config_version"] != float64(3) {
		t.Fatalf("shortened %d %v", status, body)
	}
	if got := a.featuredAt(opens.Add(90 * time.Minute)); len(got) != 0 {
		t.Fatalf("featured after the shortened window: %v", got)
	}

	a.now = start.Add(11 * time.Minute)
	status, body = a.call("POST", "/admin/config/shop/rollback", liveOps, map[string]any{"version": 2, "reason": "sale extended"})
	expect(t, status, body, http.StatusForbidden, "step_up_required")
	a.call("POST", "/admin/session/step-up", liveOps, map[string]string{"id_token": "leyla", "nonce": "nonce"})
	status, body = a.call("POST", "/admin/config/shop/rollback", liveOps, map[string]any{"version": 7, "reason": "sale extended"})
	expect(t, status, body, http.StatusNotFound, "not_found")
	status, body = a.call("POST", "/admin/config/shop/rollback", liveOps, map[string]any{"version": 2, "reason": "sale extended"})
	if status != http.StatusOK || body["config_version"] != float64(4) {
		t.Fatalf("rollback %d %v", status, body)
	}
	if got := a.featuredAt(opens.Add(90 * time.Minute)); len(got) != 1 {
		t.Fatalf("featured after rollback: %v", got)
	}

	status, body = a.call("GET", "/admin/config/shop/history", liveOps, nil)
	versions := body["versions"].([]any)
	if status != http.StatusOK || len(versions) != 3 || versions[0].(map[string]any)["config_version"] != float64(4) {
		t.Fatalf("history %d %v", status, body)
	}
	entries := a.entries(liveOps, "object_type=shop_offers&object_id=shop")
	if len(entries) != 3 {
		t.Fatalf("entries %v", entries)
	}
	rollback, shortened, first := entries[0], entries[1], entries[2]
	if rollback["action"] != "rollback" || rollback["reason"] != "sale extended" || rollback["config_version"] != float64(4) || rollback["before"] == nil {
		t.Fatalf("rollback entry %v", rollback)
	}
	if shortened["action"] != "publish" || first["before"] != nil || first["reason"] != "weekend sale" {
		t.Fatalf("publish entries %v %v", shortened, first)
	}
}

func TestACatalogIsCheckedLikeTheGameServerAndFollowsTheRoleMatrix(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	liveOps, content, support := a.signIn("leyla"), a.signIn("ayla"), a.signIn("orxan")
	bothPrices := weekendPack(start, start.Add(time.Hour))
	bothPrices["productId"] = "offer_starter"
	lateRotation := catalog(cometFox())
	lateRotation["rotations"] = []map[string]any{{"period": "daily", "start": "2026-11-01T05:00:00Z", "offers": []string{"offer.avatar_comet_fox"}}}
	for name, request := range map[string]map[string]any{
		"no catalog":                 {"reason": "x"},
		"another schema":             {"catalog": map[string]any{"schema": 2, "offers": []any{}}},
		"coins and a product":        {"catalog": catalog(cometFox(), bothPrices)},
		"a rotation off its period":  {"catalog": lateRotation},
		"an offer listed twice":      {"catalog": catalog(cometFox(), cometFox())},
		"a catalog that is a string": {"catalog": "shop"},
	} {
		status, body := a.call("PUT", "/admin/config/shop", liveOps, request)
		if status != http.StatusBadRequest || body["error"] != "invalid" || body["detail"] == "" {
			t.Fatalf("%s: %d %v", name, status, body)
		}
	}
	huge := catalog(cometFox())
	huge["padding"] = strings.Repeat("x", 1<<20)
	status, body := a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": huge})
	expect(t, status, body, http.StatusRequestEntityTooLarge, "too_large")
	for _, token := range []string{content, support} {
		status, body := a.call("PUT", "/admin/config/shop", token, map[string]any{"catalog": catalog(cometFox())})
		expect(t, status, body, http.StatusForbidden, "forbidden")
		status, body = a.call("GET", "/admin/config/shop", token, nil)
		expect(t, status, body, http.StatusOK, "")
	}
	if entries := a.entries(a.signIn("root"), "object_type=shop_offers"); len(entries) != 0 {
		t.Fatalf("refused publishes left entries: %v", entries)
	}
	if got := a.featuredAt(start); len(got) != 0 {
		t.Fatalf("featured %v", got)
	}
}

func avatarChest() map[string]any {
	return map[string]any{
		"id": "offer.chest_avatar", "kind": "chest", "section": "chests", "priceCoins": 500,
		"bundle":  map[string]any{"id": "bundle.chest_avatar", "revision": 1, "cosmetics": []string{"avatar.paper_kite", "avatar.moon_owl"}},
		"window":  map[string]any{"start": "2026-01-01T00:00:00Z"},
		"nameKey": "shop.offer.chest_avatar",
	}
}

func TestAChestPoolMustHoldPublishedCommonOrRareItemsOfTheChest(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps := a.signIn("leyla")
	withChest := catalog(cometFox(), avatarChest())
	withChest["chests"] = map[string]any{"rarityPercent": map[string]int{"common": 85, "rare": 15}, "rarePityCount": 10}

	status, body := a.call("GET", "/admin/cosmetics/items", liveOps, nil)
	if status != http.StatusOK || len(body["items"].([]any)) != 5 {
		t.Fatalf("built-in items %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": withChest})
	if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), "unknown item avatar.paper_kite") {
		t.Fatalf("chest before its items %d %v", status, body)
	}
	raw, err := os.ReadFile("../../../docs/cosmetics/examples/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Publish(context.Background(), a.pool, cosmetics.Domain, raw); err != nil {
		t.Fatal(err)
	}
	status, body = a.call("GET", "/admin/cosmetics/items", liveOps, nil)
	items := body["items"].([]any)
	if status != http.StatusOK || len(items) != 15 {
		t.Fatalf("published items %d %v", status, body)
	}
	kite := items[6].(map[string]any)
	if kite["id"] != "avatar.paper_kite" || kite["rarity"] != "common" || kite["availability"] != "live" || kite["sources"].([]any)[0].(map[string]any)["type"] != "chest" {
		t.Fatalf("chest item %v", kite)
	}
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": withChest})
	if status != http.StatusOK || body["config_version"] != float64(2) {
		t.Fatalf("chest after its items %d %v", status, body)
	}
	withChest["offers"] = []map[string]any{cometFox(), avatarChest()}
	withChest["offers"].([]map[string]any)[1]["bundle"].(map[string]any)["cosmetics"] = []string{"avatar.paper_kite", "frame.season_1_top"}
	status, body = a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": withChest})
	if status != http.StatusBadRequest || !strings.Contains(body["detail"].(string), "rarity epic") {
		t.Fatalf("epic pool item %d %v", status, body)
	}
}

func TestTheShopCatalogCarriesTheRotationRulesOfTheServer(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Shop = shop.Config{DailySlots: 6, WeeklySlots: 3, ResetHourUTC: 4} })
	a.create("leyla@example.com", rbac.LiveOps)
	status, body := a.call("GET", "/admin/config/shop", a.signIn("leyla"), nil)
	want := map[string]any{"daily_slots": float64(6), "weekly_slots": float64(3), "reset_hour_utc": float64(4)}
	if status != http.StatusOK || !reflect.DeepEqual(body["rotation"], want) {
		t.Fatalf("got %d %v, want rotation %v", status, body["rotation"], want)
	}
}
