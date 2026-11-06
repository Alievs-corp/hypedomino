package adminapi_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func bundleRows(t *testing.T, a *api, token, id string) []map[string]any {
	t.Helper()
	status, body := a.call("GET", "/admin/reward-bundles", token, nil)
	expect(t, status, body, http.StatusOK, "")
	var rows []map[string]any
	for _, row := range body["bundles"].([]any) {
		if entry := row.(map[string]any); entry["id"] == id {
			rows = append(rows, entry)
		}
	}
	return rows
}

func grantingMission(id string, coins int, cosmetics ...any) map[string]any {
	return map[string]any{
		"id": id, "type": "achievement", "nameKey": "mission." + id + ".name", "event": "match_won", "target": 10, "filters": map[string]any{},
		"availability": "live",
		"bundle":       map[string]any{"id": "bundle.wins", "revision": 1, "coins": coins, "cosmetics": cosmetics, "medals": []any{}, "passXp": 0},
	}
}

func publishMissions(t *testing.T, a *api, token string, missions ...any) {
	t.Helper()
	catalog := map[string]any{"dailySlots": []any{}, "weeklySlots": []any{}, "missions": missions}
	status, body := a.call("PUT", "/admin/config/missions", token, map[string]any{"catalog": catalog, "reason": "grants"})
	expect(t, status, body, http.StatusOK, "")
}

func TestRewardBundlesListTheDefaultsAnyRoleCanSee(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.Analyst)
	analyst := a.signIn("leyla")

	daily := bundleRows(t, a, analyst, "bundle.daily_login_day_1")
	if len(daily) != 1 || daily[0]["guideValue"].(float64) != 30 {
		t.Fatalf("daily %v", daily)
	}
	gold := bundleRows(t, a, analyst, "bundle.rank_gold")
	if len(gold) != 1 || gold[0]["differs"] != false || len(gold[0]["usedBy"].([]any)) != 2 || gold[0]["guideValue"].(float64) != 700 {
		t.Fatalf("rank gold %v", gold)
	}
}

func TestRewardBundlesTrackUsesValueAndDifferingContent(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	content, liveOps := a.signIn("ayla"), a.signIn("leyla")

	status, body := a.call("GET", "/admin/config/cosmetics", liveOps, nil)
	expect(t, status, body, http.StatusOK, "")
	builtin := body["rules"].(map[string]any)["builtin"].(map[string]any)
	fox := foxAvatar()
	fox["sources"] = []any{
		map[string]any{"type": "mission", "ref": "mission.first"},
		map[string]any{"type": "mission", "ref": "mission.second"},
		map[string]any{"type": "mission", "ref": "mission.third"},
	}
	catalog := map[string]any{"items": append(append([]any{}, builtin["items"].([]any)...), fox)}
	status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": catalog, "reason": "fox"})
	expect(t, status, body, http.StatusOK, "")

	publishMissions(t, a, liveOps, grantingMission("mission.first", 100, "avatar.fox"), grantingMission("mission.second", 100, "avatar.fox"))
	rows := bundleRows(t, a, liveOps, "bundle.wins")
	if len(rows) != 1 || rows[0]["differs"] != false || len(rows[0]["usedBy"].([]any)) != 2 {
		t.Fatalf("same content %v", rows)
	}
	if rows[0]["guideValue"].(float64) != 1600 {
		t.Fatalf("value %v, want the coins plus the rare avatar guide price", rows[0]["guideValue"])
	}
	use := rows[0]["usedBy"].([]any)[0].(map[string]any)
	if use["domain"] != "missions" || use["object"] != "mission.first" {
		t.Fatalf("use %v", use)
	}

	publishMissions(t, a, liveOps, grantingMission("mission.first", 100, "avatar.fox"), grantingMission("mission.second", 100, "avatar.fox"), grantingMission("mission.third", 250, "avatar.fox"))
	rows = bundleRows(t, a, liveOps, "bundle.wins")
	if len(rows) != 2 || rows[0]["differs"] != true || rows[1]["differs"] != true {
		t.Fatalf("differing content %v", rows)
	}
}

func TestRewardBundlesCoverTheShopThePassAndPromoCampaigns(t *testing.T) {
	a := newPromoAPI(t)
	a.publishExampleCatalog()
	liveOps := a.signIn("leyla")

	status, body := a.call("PUT", "/admin/config/shop", liveOps, map[string]any{"catalog": catalog(weekendPack(start, start.Add(time.Hour)), cometFox()), "reason": "weekend"})
	expect(t, status, body, http.StatusOK, "")
	status, body = a.call("PUT", "/admin/config/pass", liveOps, map[string]any{"catalog": passWithListedCosmetics(t), "reason": "season"})
	expect(t, status, body, http.StatusOK, "")
	for _, id := range []string{"promo.launch", "promo.launch_two"} {
		status, body = a.call("POST", "/admin/promo-codes", liveOps, promoBody(id, 300, nil))
		expect(t, status, body, http.StatusCreated, "")
	}

	offer := bundleRows(t, a, liveOps, "bundle.weekend_pack")
	if len(offer) != 1 || offer[0]["usedBy"].([]any)[0].(map[string]any)["object"] != "offer.weekend_pack" {
		t.Fatalf("offer %v", offer)
	}
	codes := bundleRows(t, a, liveOps, "bundle.launch_coins")
	if len(codes) != 1 || len(codes[0]["usedBy"].([]any)) != 1 || codes[0]["usedBy"].([]any)[0].(map[string]any)["object"] != "campaign.launch_week" {
		t.Fatalf("a campaign is one use however many codes it has: %v", codes)
	}
	status, body = a.call("GET", "/admin/reward-bundles", liveOps, nil)
	expect(t, status, body, http.StatusOK, "")
	if !grantsFrom(body, "pass", "premium.30") {
		t.Fatalf("no pass level use in %v", body)
	}
}

func grantsFrom(body map[string]any, domain, detail string) bool {
	for _, row := range body["bundles"].([]any) {
		for _, use := range row.(map[string]any)["usedBy"].([]any) {
			if entry := use.(map[string]any); entry["domain"] == domain && entry["detail"] == detail {
				return true
			}
		}
	}
	return false
}
