package adminapi_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func TestThePrivacyConfigIsPublishedBySuperAdminsOnlyAndValidated(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	root := a.signIn("root")
	path := "/admin/config/privacy"
	config := map[string]any{"PrivacyPolicyVersion": "2026-11-01", "TermsVersion": "2026-10-01", "ConsentRecordRetentionYears": 3, "PrivacyPolicyUrl": "https://example.com/privacy", "TermsOfUseUrl": "https://example.com/terms"}

	for _, token := range []string{a.signIn("ayla"), a.signIn("leyla"), a.signIn("orxan")} {
		status, body := a.call("GET", path, token, nil)
		expect(t, status, body, http.StatusForbidden, "forbidden")
		status, body = a.call("PUT", path, token, map[string]any{"catalog": config, "reason": "launch"})
		expect(t, status, body, http.StatusForbidden, "forbidden")
	}
	for _, invalid := range []map[string]any{
		{"PrivacyPolicyVersion": ""},
		{"TermsVersion": "two words"},
		{"ConsentRecordRetentionYears": 0},
		{"ConsentRecordRetentionYears": 11},
		{"PrivacyPolicyUrl": "http://example.com/privacy"},
		{"TermsUrl": "https://example.com/terms"},
	} {
		status, body := a.call("PUT", path, root, map[string]any{"catalog": invalid, "reason": "launch"})
		expect(t, status, body, http.StatusBadRequest, "invalid")
	}

	status, body := a.call("PUT", path, root, map[string]any{"catalog": config, "reason": "launch"})
	if status != http.StatusOK || body["config_version"] != float64(1) {
		t.Fatalf("publish %d %v", status, body)
	}
	status, body = a.call("PUT", path, root, map[string]any{"catalog": map[string]any{"PrivacyPolicyVersion": "2026-12-01", "TermsVersion": "2026-10-01"}, "reason": "new processor"})
	if status != http.StatusOK || body["catalog"].(map[string]any)["PrivacyPolicyVersion"] != "2026-12-01" {
		t.Fatalf("second publish %d %v", status, body)
	}

	a.now = start.Add(11 * time.Minute)
	status, body = a.call("POST", path+"/rollback", root, map[string]any{"version": 1, "reason": "processor withdrawn"})
	expect(t, status, body, http.StatusForbidden, "step_up_required")
	a.call("POST", "/admin/session/step-up", root, map[string]string{"id_token": "root", "nonce": "nonce"})
	status, body = a.call("POST", path+"/rollback", root, map[string]any{"version": 1, "reason": "processor withdrawn"})
	if status != http.StatusOK || body["config_version"] != float64(3) || body["catalog"].(map[string]any)["PrivacyPolicyVersion"] != "2026-11-01" {
		t.Fatalf("rollback %d %v", status, body)
	}

	status, body = a.call("GET", path+"/history", root, nil)
	if versions := body["versions"].([]any); status != http.StatusOK || len(versions) != 3 {
		t.Fatalf("history %d %v", status, body)
	}
	entries := a.entries(root, "object_type=privacy_config&object_id=privacy")
	if len(entries) != 3 || entries[0]["action"] != "rollback" || entries[2]["reason"] != "launch" {
		t.Fatalf("entries %v", entries)
	}
}
