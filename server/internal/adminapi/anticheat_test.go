package adminapi_test

import (
	"net/http"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func TestTheIllegalIntentKeysAreOffUntilASuperAdminPublishesThem(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	root, liveOps, support := a.signIn("root"), a.signIn("leyla"), a.signIn("orxan")
	path := "/admin/config/anticheat"

	status, body := a.call("GET", path, root, nil)
	keys := body["keys"].([]any)
	if status != http.StatusOK || body["config_version"] != float64(0) || len(keys) != 3 {
		t.Fatalf("defaults %d %v", status, body)
	}
	for _, key := range keys {
		view := key.(map[string]any)
		if view["default"] != float64(0) || view["value"] != float64(0) || view["min"] != float64(0) || view["set"] != false {
			t.Fatalf("key %v is not off by default", view)
		}
	}
	if window := keys[1].(map[string]any); window["key"] != "IllegalIntentWindowMinutes" || window["max"] != float64(43200) || window["unit"] != "minutes" {
		t.Fatalf("window key %v", window)
	}

	for _, values := range []map[string]any{
		{"IllegalIntentThreshold": 3},
		{"IllegalIntentThreshold": 3, "IllegalIntentWindowMinutes": 43201},
		{"IllegalIntentWeight": -1},
		{"IllegalIntentWeight": "5"},
		{"TurnSeconds": 20},
	} {
		status, body := a.call("PUT", path, root, map[string]any{"values": values, "reason": "tune"})
		if status != http.StatusBadRequest || body["error"] != "invalid" || body["detail"] == "" {
			t.Fatalf("%v: %d %v", values, status, body)
		}
	}
	for _, token := range []string{liveOps, support} {
		status, body := a.call("GET", path, token, nil)
		expect(t, status, body, http.StatusForbidden, "forbidden")
		status, body = a.call("PUT", path, token, map[string]any{"values": map[string]int{"IllegalIntentThreshold": 1, "IllegalIntentWindowMinutes": 1}, "reason": "tune"})
		expect(t, status, body, http.StatusForbidden, "forbidden")
	}

	status, body = a.call("PUT", path, root, map[string]any{"values": map[string]int{"IllegalIntentThreshold": 4, "IllegalIntentWindowMinutes": 15, "IllegalIntentWeight": 2}, "reason": "owner values"})
	if value, set := ruleValue(t, body, "IllegalIntentThreshold"); status != http.StatusOK || body["config_version"] != float64(1) || value != 4 || !set {
		t.Fatalf("publish %d %v", status, body)
	}
	status, body = a.call("GET", path+"/history", root, nil)
	if versions := body["versions"].([]any); status != http.StatusOK || len(versions) != 1 {
		t.Fatalf("history %d %v", status, body)
	}
	entries := a.entries(root, "object_type=anti_cheat_config&object_id=anticheat")
	if len(entries) != 1 || entries[0]["action"] != "publish" || entries[0]["reason"] != "owner values" {
		t.Fatalf("entries %v", entries)
	}
}
