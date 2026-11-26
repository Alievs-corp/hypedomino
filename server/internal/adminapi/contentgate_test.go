package adminapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

const gatePath = "/admin/content-gate/reports"

var launchAt = map[string]any{"launch_at": "2027-01-15T12:00:00Z"}

func gateCheck(t *testing.T, report map[string]any, id string) map[string]any {
	t.Helper()
	for _, check := range report["checks"].([]any) {
		if found := check.(map[string]any); found["id"] == id {
			return found
		}
	}
	t.Fatalf("no check %s in %v", id, report["checks"])
	return nil
}

func gateFailures(check map[string]any) []string {
	var rules []string
	for _, failure := range check["failures"].([]any) {
		rules = append(rules, failure.(map[string]any)["rule"].(string))
	}
	return rules
}

func TestTheContentGateRunsOnlyForTheRunnerRolesAndEveryRoleViews(t *testing.T) {
	a := newAPI(t)
	a.create("vusal@example.com", rbac.Support)
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Analyst)
	for _, who := range []string{"vusal", "orxan"} {
		token := a.signIn(who)
		status, body := a.call("POST", gatePath, token, launchAt)
		expect(t, status, body, http.StatusForbidden, "forbidden")
		status, body = a.call("GET", gatePath, token, nil)
		expect(t, status, body, http.StatusOK, "")
	}
	for _, who := range []string{"root", "ayla", "leyla"} {
		status, body := a.call("POST", gatePath, a.signIn(who), launchAt)
		expect(t, status, body, http.StatusCreated, "")
	}
	status, body := a.call("GET", gatePath, a.signIn("orxan"), nil)
	if reports := body["reports"].([]any); status != http.StatusOK || len(reports) != 3 {
		t.Fatalf("history %d %v", status, body)
	}
}

func TestAContentGateRunIsStoredAuditedAndRedOnAnEmptyCatalogue(t *testing.T) {
	a := newAPI(t)
	content := a.signIn("ayla")
	status, body := a.call("POST", gatePath, content, launchAt)
	expect(t, status, body, http.StatusCreated, "")
	report := body["report"].(map[string]any)
	if body["stale"] != false || report["verdict"] != "red" || report["environment"] != "staging" || report["launch_at"] != "2027-01-15T12:00:00Z" || len(report["checks"].([]any)) != 12 {
		t.Fatalf("report %v", body)
	}
	if check := gateCheck(t, report, "CG-13"); check["result"] != "pass" {
		t.Fatalf("default thresholds %v", check)
	}
	if check := gateCheck(t, report, "CG-7"); check["result"] != "fail" || !slices.Contains(gateFailures(check), "MR-1") {
		t.Fatalf("medals %v", check)
	}
	if check := gateCheck(t, report, "CG-12"); check["result"] != "fail" || !strings.Contains(check["failures"].([]any)[0].(map[string]any)["detail"].(string), "HD_PROMO_KEY") {
		t.Fatalf("promo without a key %v", check)
	}
	if check := gateCheck(t, report, "CG-8"); check["result"] != "fail" || !slices.Contains(gateFailures(check), "CR-5") || len(check["unchecked"].([]any)) == 0 {
		t.Fatalf("builtin items without uploads %v", check)
	}
	thresholds := report["thresholds"].([]any)
	if len(thresholds) != 6 || thresholds[2].(map[string]any)["key"] != "ContentGateMinFrames" || thresholds[2].(map[string]any)["value"] != float64(20) || thresholds[2].(map[string]any)["config_version"] != nil {
		t.Fatalf("thresholds %v", thresholds)
	}
	id := report["id"].(string)
	entries := a.entries(a.signIn("root"), "object_type=content_gate_reports&object_id="+id)
	if len(entries) != 1 || entries[0]["action"] != "create" || entries[0]["after"].(map[string]any)["verdict"] != "red" {
		t.Fatalf("entries %v", entries)
	}
	if _, err := a.pool.Exec(context.Background(), `update content_gate_reports set verdict = 'green'`); err == nil {
		t.Fatal("a stored report was changed")
	}
	status, body = a.call("GET", gatePath+"/"+id, content, nil)
	if status != http.StatusOK || body["report"].(map[string]any)["id"] != id || body["stale"] != false {
		t.Fatalf("stored report %d %v", status, body)
	}
}

func TestAPublishAfterARunMakesItStaleAndALoweredThresholdGivesRed(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps := a.signIn("leyla")
	status, body := a.call("POST", gatePath, liveOps, launchAt)
	expect(t, status, body, http.StatusCreated, "")
	first := body["report"].(map[string]any)["id"].(string)

	status, body = a.call("PUT", "/admin/config/liveops", liveOps, map[string]any{"values": map[string]int{"ContentGateMinMedals": 10}, "reason": "fewer medals"})
	expect(t, status, body, http.StatusOK, "")
	status, body = a.call("GET", gatePath+"/"+first, liveOps, nil)
	if status != http.StatusOK || body["stale"] != true {
		t.Fatalf("after a publish %d %v", status, body)
	}

	a.now = start.Add(time.Minute)
	status, body = a.call("POST", gatePath, liveOps, launchAt)
	expect(t, status, body, http.StatusCreated, "")
	report := body["report"].(map[string]any)
	check := gateCheck(t, report, "CG-13")
	if check["result"] != "fail" || report["verdict"] != "red" || !strings.Contains(check["failures"].([]any)[0].(map[string]any)["detail"].(string), "below the repo default 40") {
		t.Fatalf("lowered threshold %v", check)
	}
	if minimum := gateCheck(t, report, "CG-7")["minimum"]; minimum != float64(10) {
		t.Fatalf("medal minimum %v", minimum)
	}

	status, body = a.call("GET", gatePath, liveOps, nil)
	reports := body["reports"].([]any)
	if status != http.StatusOK || len(reports) != 2 || reports[0].(map[string]any)["stale"] != false || reports[1].(map[string]any)["stale"] != true || reports[1].(map[string]any)["id"] != first {
		t.Fatalf("history %d %v", status, body)
	}
	if results := reports[0].(map[string]any)["results"].([]any); len(results) != 12 || results[11].(map[string]any)["id"] != "CG-13" || results[11].(map[string]any)["result"] != "fail" {
		t.Fatalf("results %v", results)
	}
}

func TestAContentGateReportExportsAsAJSONAttachment(t *testing.T) {
	a := newAPI(t)
	content := a.signIn("ayla")
	_, body := a.call("POST", gatePath, content, launchAt)
	id := body["report"].(map[string]any)["id"].(string)
	recorder := a.raw("GET", gatePath+"/"+id+"/export", content, "")
	var record map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &record); err != nil || recorder.Code != http.StatusOK || record["id"] != id {
		t.Fatalf("export %d %s %v", recorder.Code, recorder.Body, err)
	}
	if disposition := recorder.Header().Get("Content-Disposition"); disposition != `attachment; filename="content-gate-`+id+`.json"` || recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("headers %v", recorder.Header())
	}
}

func TestAContentGateRunNeedsALaunchTimeAndAnUnknownReportIsNotFound(t *testing.T) {
	a := newAPI(t)
	content := a.signIn("ayla")
	status, body := a.call("POST", gatePath, content, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "invalid")
	status, body = a.call("POST", gatePath, content, map[string]any{"launch_at": "soon"})
	expect(t, status, body, http.StatusBadRequest, "malformed")
	for _, id := range []string{"unknown", "6f1c1d8e-2f7a-4c1e-9d55-0b6a5f2b9c11"} {
		status, body = a.call("GET", gatePath+"/"+id, content, nil)
		expect(t, status, body, http.StatusNotFound, "not_found")
		status, body = a.call("GET", gatePath+"/"+id+"/export", content, nil)
		expect(t, status, body, http.StatusNotFound, "not_found")
	}
}

func TestTheContentGateReadsTheLaunchRolesOfPromoTemplates(t *testing.T) {
	a := newPromoAPI(t)
	liveOps := a.signIn("leyla")
	status, body := a.call("POST", "/admin/promo-templates", liveOps, map[string]any{"id": "promo.welcome", "campaign": "campaign.launch", "kind": "multi_use", "role": "welcome", "reason": "launch"})
	expect(t, status, body, http.StatusCreated, "")
	status, body = a.call("POST", gatePath, liveOps, launchAt)
	expect(t, status, body, http.StatusCreated, "")
	check := gateCheck(t, body["report"].(map[string]any), "CG-12")
	var details []string
	for _, failure := range check["failures"].([]any) {
		details = append(details, failure.(map[string]any)["detail"].(string))
	}
	want := []string{"the welcome template grants exactly one cosmetic", "no template has the creator role", "no template has the compensation role"}
	if check["result"] != "fail" || !slices.Equal(details, want) {
		t.Fatalf("templates %v", check)
	}
}
