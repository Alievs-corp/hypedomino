package adminapi_test

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func exampleCatalog(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("../../../docs/pass/examples/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]any
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	return catalog
}

func listed(catalog map[string]any, key, id string) map[string]any {
	for _, entry := range catalog[key].([]any) {
		if object := entry.(map[string]any); object["id"] == id {
			return object
		}
	}
	return nil
}

func without(catalog map[string]any, key, id string) map[string]any {
	kept := []any{}
	for _, entry := range catalog[key].([]any) {
		if entry.(map[string]any)["id"] != id {
			kept = append(kept, entry)
		}
	}
	catalog[key] = kept
	return catalog
}

func TestMissionsArePublishedWithServerStampsAndNeverDropped(t *testing.T) {
	a := newAPI(t)
	a.publishExampleCatalog()
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps, content := a.signIn("leyla"), a.signIn("ayla")

	status, body := a.call("GET", "/admin/config/missions", liveOps, nil)
	rules := body["rules"].(map[string]any)
	if status != http.StatusOK || body["catalog"] != nil || len(rules["events"].([]any)) != 6 || len(rules["types"].([]any)) != 4 || len(rules["modes"].([]any)) != 2 {
		t.Fatalf("empty %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/missions", liveOps, map[string]any{"catalog": exampleCatalog(t, "missions")})
	if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), "medal.wins_50") {
		t.Fatalf("a medal grant before the medal is published: %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/medals", content, map[string]any{"catalog": exampleCatalog(t, "medals"), "reason": "launch medals"})
	if status != http.StatusOK {
		t.Fatalf("medals %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/missions", liveOps, map[string]any{"catalog": exampleCatalog(t, "missions"), "reason": "launch missions"})
	first, _ := body["config_version"].(float64)
	if status != http.StatusOK || first < 1 {
		t.Fatalf("missions %d %v", status, body)
	}

	edited := exampleCatalog(t, "missions")
	win50 := listed(edited, "missions", "mission.win_50")
	win50["target"], win50["revision"], win50["nameKey"] = 60, 9, "typed by hand"
	edited["missions"] = append(edited["missions"].([]any), map[string]any{
		"id": "mission.hands_3", "type": "daily", "event": "hand_won", "target": 3, "filters": map[string]any{"queue": "casual", "mode": "2v2"},
		"slot": 1, "weight": 2, "availability": "live",
	})
	status, body = a.call("PUT", "/admin/config/missions", liveOps, map[string]any{"catalog": edited})
	published := body["catalog"].(map[string]any)
	stamped, added, untouched := listed(published, "missions", "mission.win_50"), listed(published, "missions", "mission.hands_3"), listed(published, "missions", "mission.win_250")
	if status != http.StatusOK || body["config_version"] != first+1 {
		t.Fatalf("edit %d %v", status, body)
	}
	if stamped["revision"] != float64(2) || stamped["nameKey"] != "mission.win_50.name" || stamped["target"] != float64(60) {
		t.Fatalf("stamped %v", stamped)
	}
	if added["revision"] != float64(1) || added["nameKey"] != "mission.hands_3.name" || untouched["revision"] != float64(1) {
		t.Fatalf("added %v untouched %v", added, untouched)
	}

	for name, catalog := range map[string]map[string]any{
		"a published mission dropped": without(exampleCatalog(t, "missions"), "missions", "mission.win_5000"),
		"an unknown progress event": func() map[string]any {
			c := exampleCatalog(t, "missions")
			listed(c, "missions", "mission.win_50")["event"] = "tile_flipped"
			return c
		}(),
		"a medal not listing it": func() map[string]any {
			c := exampleCatalog(t, "missions")
			listed(c, "missions", "mission.win_250")["bundle"].(map[string]any)["medals"] = []string{"medal.wins_50"}
			return c
		}(),
	} {
		status, body := a.call("PUT", "/admin/config/missions", liveOps, map[string]any{"catalog": catalog})
		if status != http.StatusBadRequest || body["error"] != "invalid" || body["detail"] == "" {
			t.Fatalf("%s: %d %v", name, status, body)
		}
	}

	status, body = a.call("PUT", "/admin/config/missions", content, map[string]any{"catalog": edited})
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("PUT", "/admin/config/medals", liveOps, map[string]any{"catalog": exampleCatalog(t, "medals")})
	expect(t, status, body, http.StatusForbidden, "forbidden")

	a.now = start.Add(11 * time.Minute)
	a.call("POST", "/admin/session/step-up", liveOps, map[string]string{"id_token": "leyla", "nonce": "nonce"})
	status, body = a.call("POST", "/admin/config/missions/rollback", liveOps, map[string]any{"version": first, "reason": "target too high"})
	restored := body["catalog"].(map[string]any)
	if status != http.StatusOK || body["config_version"] != first+2 {
		t.Fatalf("rollback %d %v", status, body)
	}
	back, kept := listed(restored, "missions", "mission.win_50"), listed(restored, "missions", "mission.hands_3")
	if back["target"] != float64(50) || back["revision"] != float64(3) {
		t.Fatalf("rolled back %v", back)
	}
	if kept == nil || kept["availability"] != "retired" || kept["revision"] != float64(2) {
		t.Fatalf("a mission added after the target version is kept as retired: %v", kept)
	}

	status, body = a.call("GET", "/admin/config/missions/history", liveOps, nil)
	if status != http.StatusOK || len(body["versions"].([]any)) != 3 {
		t.Fatalf("history %d %v", status, body)
	}
	entries := a.entries(a.signIn("root"), "object_type=missions&object_id=missions")
	if len(entries) != 3 || entries[0]["action"] != "rollback" || entries[2]["reason"] != "launch missions" || entries[2]["before"] != nil {
		t.Fatalf("entries %v", entries)
	}
}

func TestAMedalKeepsTheSourcesOfPublishedMissionsAndIsNeverDropped(t *testing.T) {
	a := newAPI(t)
	a.publishExampleCatalog()
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	liveOps, content, support := a.signIn("leyla"), a.signIn("ayla"), a.signIn("orxan")

	status, body := a.call("GET", "/admin/config/medals", support, nil)
	rules := body["rules"].(map[string]any)
	if status != http.StatusOK || rules["rarities"].([]any)[0] != "common" || len(rules["source_types"].([]any)) != 2 {
		t.Fatalf("empty %d %v", status, body)
	}
	medals := exampleCatalog(t, "medals")
	wins50 := listed(medals, "medals", "medal.wins_50")
	wins50["nameKey"], wins50["staticAddress"], wins50["revision"] = "x", "y", 4
	status, body = a.call("PUT", "/admin/config/medals", content, map[string]any{"catalog": medals})
	stamped := listed(body["catalog"].(map[string]any), "medals", "medal.wins_50")
	if status != http.StatusOK || stamped["nameKey"] != "medal.wins_50.name" || stamped["descriptionKey"] != "medal.wins_50.description" || stamped["staticAddress"] != "medal/wins_50/static" || stamped["revision"] != float64(1) {
		t.Fatalf("stamped %d %v", status, stamped)
	}
	status, body = a.call("PUT", "/admin/config/missions", liveOps, map[string]any{"catalog": exampleCatalog(t, "missions")})
	expect(t, status, body, http.StatusOK, "")

	unlisted := exampleCatalog(t, "medals")
	listed(unlisted, "medals", "medal.wins_250")["sources"] = []map[string]string{{"type": "rank_reward", "ref": "rank.season_1.gold"}}
	status, body = a.call("PUT", "/admin/config/medals", content, map[string]any{"catalog": unlisted})
	if status != http.StatusBadRequest || body["error"] != "invalid" || !strings.Contains(body["detail"].(string), "medal.wins_250") {
		t.Fatalf("a source a published mission needs: %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/config/medals", content, map[string]any{"catalog": without(exampleCatalog(t, "medals"), "medals", "medal.chain_100")})
	if status != http.StatusBadRequest || !strings.Contains(body["detail"].(string), "medal.chain_100") {
		t.Fatalf("a dropped medal: %d %v", status, body)
	}
	retired := exampleCatalog(t, "medals")
	listed(retired, "medals", "medal.chain_100")["availability"] = "retired"
	status, body = a.call("PUT", "/admin/config/medals", content, map[string]any{"catalog": retired})
	if status != http.StatusOK || listed(body["catalog"].(map[string]any), "medals", "medal.chain_100")["revision"] != float64(2) {
		t.Fatalf("retired %d %v", status, body)
	}
}
