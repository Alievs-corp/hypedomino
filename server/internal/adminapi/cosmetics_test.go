package adminapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func cosmeticItem(catalog map[string]any, id string) map[string]any {
	return listed(catalog, "items", id)
}

func foxAvatar() map[string]any {
	return map[string]any{
		"id": "avatar.fox", "type": "avatar", "rarityId": "rare", "animated": true, "availability": "live",
		"sources": []any{map[string]any{"type": "shop", "ref": "offer.fox"}}, "artSource": "handmade", "culturalReview": "approved",
	}
}

func TestCosmeticsPublishStampsItemsAndKeepsThemForever(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	content, liveOps := a.signIn("ayla"), a.signIn("leyla")

	status, body := a.call("GET", "/admin/config/cosmetics", liveOps, nil)
	rules, _ := body["rules"].(map[string]any)
	if status != http.StatusOK || body["catalog"] != nil || len(rules["types"].([]any)) != 6 || len(rules["source_types"].([]any)) != 9 {
		t.Fatalf("empty %d %v", status, body)
	}
	builtin := rules["builtin"].(map[string]any)
	status, body = a.call("PUT", "/admin/config/cosmetics", liveOps, map[string]any{"catalog": builtin, "reason": "no"})
	expect(t, status, body, http.StatusForbidden, "forbidden")

	first := map[string]any{"items": append(append([]any{}, builtin["items"].([]any)...), foxAvatar())}
	status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": first, "reason": "fox avatar"})
	expect(t, status, body, http.StatusOK, "")
	published := body["catalog"].(map[string]any)
	fox := cosmeticItem(published, "avatar.fox")
	if published["manifestVersion"] != builtin["manifestVersion"].(float64)+1 || fox["revision"] != 1.0 || fox["nameKey"] != "cosmetic.avatar.fox.name" || fox["animationAddress"] != "cosmetic/avatar/fox/anim" {
		t.Fatalf("stamped %v", published)
	}
	starter := cosmeticItem(published, "avatar.starter")
	if starter["revision"] != cosmeticItem(builtin, "avatar.starter")["revision"] {
		t.Fatalf("an unchanged builtin item keeps its revision: %v", starter)
	}

	refused := []struct {
		name   string
		change func(map[string]any)
		detail string
	}{
		{"a price", func(c map[string]any) { cosmeticItem(c, "avatar.fox")["price"] = 5 }, "V-1"},
		{"a pending review", func(c map[string]any) { cosmeticItem(c, "avatar.fox")["culturalReview"] = "pending" }, "V-1"},
		{"an unknown rarity", func(c map[string]any) { cosmeticItem(c, "avatar.fox")["rarityId"] = "ultra" }, "V-14"},
		{"a new type", func(c map[string]any) { cosmeticItem(c, "avatar.fox")["type"] = "frame" }, "frame"},
		{"a removed item", func(c map[string]any) { c["items"] = without(c, "items", "avatar.fox")["items"] }, "V-12"},
		{"a builtin default removed", func(c map[string]any) { c["items"] = without(c, "items", "avatar.starter")["items"] }, "default"},
	}
	for _, c := range refused {
		var catalog map[string]any
		raw, _ := json.Marshal(published)
		if err := json.Unmarshal(raw, &catalog); err != nil {
			t.Fatal(err)
		}
		c.change(catalog)
		status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": catalog, "reason": "refused"})
		if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), c.detail) {
			t.Errorf("%s: %d %v", c.name, status, body)
		}
	}

	cosmeticItem(published, "avatar.fox")["rarityId"] = "epic"
	status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": published, "reason": "rarer fox"})
	expect(t, status, body, http.StatusOK, "")
	if got := cosmeticItem(body["catalog"].(map[string]any), "avatar.fox"); got["revision"] != 2.0 || got["rarityId"] != "epic" {
		t.Fatalf("edited %v", got)
	}

	status, body = a.call("GET", "/admin/config/cosmetics/history", content, nil)
	versions := body["versions"].([]any)
	if status != http.StatusOK || len(versions) != 2 {
		t.Fatalf("history %d %v", status, body)
	}
	earlier := versions[1].(map[string]any)["config_version"]
	status, body = a.call("POST", "/admin/config/cosmetics/rollback", content, map[string]any{"version": earlier, "reason": "back to rare"})
	expect(t, status, body, http.StatusOK, "")
	rolled := body["catalog"].(map[string]any)
	if got := cosmeticItem(rolled, "avatar.fox"); got["revision"] != 3.0 || got["rarityId"] != "rare" {
		t.Fatalf("rollback %v", got)
	}

	entries := a.entries(a.signIn("root"), "object_type=cosmetic_items")
	if len(entries) != 3 {
		t.Fatalf("audit %v", entries)
	}
}

func TestCosmeticsPublishKeepsTheSourcesOtherDomainsGrantFrom(t *testing.T) {
	a := newAPI(t)
	a.publishExampleCatalog()
	a.create("leyla@example.com", rbac.LiveOps)
	content, liveOps := a.signIn("ayla"), a.signIn("leyla")

	status, body := a.call("PUT", "/admin/config/rating", liveOps, map[string]any{"catalog": exampleAt(t, "ranking/examples/rating"), "reason": "rank rewards"})
	expect(t, status, body, http.StatusOK, "")

	status, body = a.call("GET", "/admin/config/cosmetics", content, nil)
	expect(t, status, body, http.StatusOK, "")
	catalog := body["catalog"].(map[string]any)
	frame := cosmeticItem(catalog, "frame.season_1_top")
	frame["sources"] = []any{map[string]any{"type": "event", "ref": "event.other"}}
	status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": catalog, "reason": "drop the rank source"})
	if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), "bundle.rank_season_1_gold") {
		t.Fatalf("dropped source %d %v", status, body)
	}
}

func (a *api) uploadedBy(email, itemID string) {
	a.t.Helper()
	account, err := a.accounts.ByEmail(context.Background(), email)
	if err != nil {
		a.t.Fatal(err)
	}
	_, err = a.pool.Exec(context.Background(), `insert into cosmetic_assets (id, item_id, slot, address, object_key, status, uploaded_by, created_at)
		values (gen_random_uuid(), $1, 'static', $2, $2 || '/x.png', 'awaiting_file', $3, $4)`, itemID, "cosmetic/"+itemID+"/static", account.ID, start)
	if err != nil {
		a.t.Fatal(err)
	}
}

func TestTheUploaderOfANewItemsArtCannotPublishItsReview(t *testing.T) {
	a := newAPI(t)
	content, root := a.signIn("ayla"), a.signIn("root")
	a.uploadedBy("ayla@example.com", "avatar.fox")
	_, body := a.call("GET", "/admin/config/cosmetics", content, nil)
	builtin := body["rules"].(map[string]any)["builtin"].(map[string]any)
	catalog := map[string]any{"items": append(append([]any{}, builtin["items"].([]any)...), foxAvatar())}

	status, body := a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": catalog, "reason": "my own fox"})
	expect(t, status, body, http.StatusForbidden, "own_upload")
	status, body = a.call("PUT", "/admin/config/cosmetics", root, map[string]any{"catalog": catalog, "reason": "reviewed fox"})
	expect(t, status, body, http.StatusOK, "")
	published := body["catalog"].(map[string]any)
	cosmeticItem(published, "avatar.fox")["rarityId"] = "epic"
	status, body = a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": published, "reason": "rarer fox"})
	expect(t, status, body, http.StatusOK, "")
}

func TestLiveOpsSwitchesOffAnItemForAnIncident(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	content, liveOps := a.signIn("ayla"), a.signIn("leyla")
	_, body := a.call("GET", "/admin/config/cosmetics", content, nil)
	builtin := body["rules"].(map[string]any)["builtin"].(map[string]any)
	catalog := map[string]any{"items": append(append([]any{}, builtin["items"].([]any)...), foxAvatar())}
	status, body := a.call("PUT", "/admin/config/cosmetics", content, map[string]any{"catalog": catalog, "reason": "fox avatar"})
	expect(t, status, body, http.StatusOK, "")

	path := "/admin/cosmetics/items/avatar.fox/switch-off"
	incident := map[string]any{"reason": "offensive art reported", "ref": "SEV2-17"}
	status, body = a.call("POST", path, content, incident)
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("POST", path, liveOps, map[string]any{"reason": "offensive art reported"})
	expect(t, status, body, http.StatusBadRequest, "invalid")
	status, body = a.call("POST", path, liveOps, map[string]any{"ref": "SEV2-17"})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", "/admin/cosmetics/items/avatar.nobody/switch-off", liveOps, incident)
	expect(t, status, body, http.StatusNotFound, "not_found")

	status, body = a.call("POST", path, liveOps, incident)
	expect(t, status, body, http.StatusOK, "")
	if fox := cosmeticItem(body["catalog"].(map[string]any), "avatar.fox"); fox["availability"] != "disabled" || fox["revision"] != 2.0 {
		t.Fatalf("switched off %v", fox)
	}
	entries := a.entries(a.signIn("root"), "object_type=cosmetic_items&action=retire")
	if len(entries) != 1 || entries[0]["ref"] != "SEV2-17" || entries[0]["reason"] != "offensive art reported" {
		t.Fatalf("audit %v", entries)
	}
	a.now = start.Add(15 * time.Minute)
	status, body = a.call("POST", path, liveOps, incident)
	expect(t, status, body, http.StatusForbidden, "step_up_required")
}
