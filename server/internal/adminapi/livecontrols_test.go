package adminapi_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/livecontrol"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

func (a *api) publishedControls() livecontrol.Controls {
	a.t.Helper()
	_, raw, err := snapshot.Latest(context.Background(), a.pool, livecontrol.Domain)
	if err != nil {
		a.t.Fatal(err)
	}
	controls, err := livecontrol.Parse(raw)
	if err != nil {
		a.t.Fatal(err)
	}
	return controls
}

func controlsBody(minimum string, purchasesDisabled bool) map[string]any {
	return map[string]any{
		"controls": map[string]any{"minimum_app_version": minimum, "purchases_disabled": purchasesDisabled},
		"reason":   "incident", "ref": "OPS-1",
	}
}

func TestStagingSwitchesLiveControlsAtOnceWithAReason(t *testing.T) {
	a := newAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps, root, content := a.signIn("leyla"), a.signIn("root"), a.signIn("ayla")

	status, body := a.call("GET", "/admin/live-controls", root, nil)
	if status != http.StatusOK || body["config_version"] != float64(0) || body["controls"] != nil {
		t.Fatalf("empty controls %d %v", status, body)
	}
	status, body = a.call("PUT", "/admin/live-controls", content, controlsBody("1.4.0", true))
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("PUT", "/admin/live-controls", liveOps, map[string]any{"controls": map[string]any{"minimum_app_version": "1.4.0"}})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("PUT", "/admin/live-controls", liveOps, controlsBody("latest", true))
	expect(t, status, body, http.StatusBadRequest, "invalid")

	a.now = start.Add(11 * time.Minute)
	status, body = a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.4.0", true))
	expect(t, status, body, http.StatusForbidden, "step_up_required")
	a.call("POST", "/admin/session/step-up", liveOps, map[string]string{"id_token": "leyla", "nonce": "nonce"})

	status, body = a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.4.0", true))
	if status != http.StatusOK || body["config_version"] != float64(1) {
		t.Fatalf("switch %d %v", status, body)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.4.0" || !controls.PurchasesDisabled {
		t.Fatalf("published %+v", controls)
	}
	entries := a.entries(a.signIn("root"), "object_type=live_controls&action="+adminapi.SwitchLiveControlsAction)
	if len(entries) != 1 || entries[0]["reason"] != "incident" || entries[0]["ref"] != "OPS-1" {
		t.Fatalf("entries %v", entries)
	}
}

func TestProductionLiveControlsNeedASecondAdmin(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	proposer, second := a.signIn("leyla"), a.signIn("root")

	status, body := a.call("PUT", "/admin/live-controls", proposer, controlsBody("1.4.0", true))
	if status != http.StatusAccepted || body["approval_id"] == nil {
		t.Fatalf("propose %d %v", status, body)
	}
	id := fmt.Sprintf("%.0f", body["approval_id"].(float64))
	if status, body := a.call("GET", "/admin/live-controls", proposer, nil); status != http.StatusOK || body["config_version"] != float64(0) {
		t.Fatalf("a proposal changed the controls: %d %v", status, body)
	}
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", proposer, map[string]any{"reason": "mine"})
	expect(t, status, body, http.StatusForbidden, "self_decision")
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", second, map[string]any{"reason": "checked"})
	if status != http.StatusOK || body["state"] != "approved" {
		t.Fatalf("approve %d %v", status, body)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.4.0" || !controls.PurchasesDisabled {
		t.Fatalf("published %+v", controls)
	}
	entries := a.entries(second, "object_type=live_controls&action="+adminapi.SwitchLiveControlsAction)
	if len(entries) != 1 || entries[0]["approval_id"] == nil {
		t.Fatalf("entries %v", entries)
	}
}

func TestAnApprovalCannotUndoANewerLiveControlChange(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps, root := a.signIn("leyla"), a.signIn("root")

	_, first := a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.4.0", false))
	_, second := a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.5.0", true))
	approve := func(body map[string]any) (int, map[string]any) {
		return a.call("POST", fmt.Sprintf("/admin/approvals/%.0f/approve", body["approval_id"].(float64)), root, map[string]any{"reason": "checked"})
	}
	if status, body := approve(second); status != http.StatusOK {
		t.Fatalf("approve %d %v", status, body)
	}
	status, body := approve(first)
	expect(t, status, body, http.StatusConflict, "stale_proposal")
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.5.0" || !controls.PurchasesDisabled {
		t.Fatalf("published %+v", controls)
	}
}

func breakGlassBody(minimum, severity, incident string) map[string]any {
	body := controlsBody(minimum, true)
	body["break_glass"] = map[string]any{"severity": severity, "id": incident}
	return body
}

func (a *api) unconfirmedBreakGlass(token string) []any {
	a.t.Helper()
	status, body := a.call("GET", "/admin/live-controls/break-glass", token, nil)
	if status != http.StatusOK {
		a.t.Fatalf("break-glass list %d %v", status, body)
	}
	return body["proposals"].([]any)
}

func TestBreakGlassRefusesWithoutSuperAdminOrADeclaredIncident(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	liveOps, support, root := a.signIn("leyla"), a.signIn("orxan"), a.signIn("root")
	cases := []struct {
		name       string
		token      string
		body       map[string]any
		wantStatus int
		wantError  string
	}{
		{"live_ops role", liveOps, breakGlassBody("1.4.0", "SEV1", "INC-12"), http.StatusForbidden, "forbidden"},
		{"support role", support, breakGlassBody("1.4.0", "SEV1", "INC-12"), http.StatusForbidden, "forbidden"},
		{"sev3", root, breakGlassBody("1.4.0", "SEV3", "INC-12"), http.StatusBadRequest, "incident_required"},
		{"no severity", root, breakGlassBody("1.4.0", "", "INC-12"), http.StatusBadRequest, "incident_required"},
		{"no incident id", root, breakGlassBody("1.4.0", "SEV1", ""), http.StatusBadRequest, "incident_required"},
		{"bad controls", root, breakGlassBody("latest", "SEV1", "INC-12"), http.StatusBadRequest, "invalid"},
		{"no reason", root, map[string]any{"controls": map[string]any{"minimum_app_version": "1.4.0"}, "break_glass": map[string]any{"severity": "SEV1", "id": "INC-12"}}, http.StatusBadRequest, "reason_required"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body := a.call("PUT", "/admin/live-controls", c.token, c.body)
			expect(t, status, body, c.wantStatus, c.wantError)
		})
	}
	status, body := a.call("GET", "/admin/live-controls", liveOps, nil)
	if status != http.StatusOK || body["config_version"] != float64(0) || len(a.unconfirmedBreakGlass(liveOps)) != 0 {
		t.Fatalf("a refused break-glass changed something: %d %v", status, body)
	}
	if entries := a.entries(root, "object_type=live_controls"); len(entries) != 0 {
		t.Fatalf("a refused break-glass was audited: %v", entries)
	}
}

func TestBreakGlassSwitchesAtOnceAndWaitsForASecondAdmin(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("orxan@example.com", rbac.Support)
	liveOps, root, support := a.signIn("leyla"), a.signIn("root"), a.signIn("orxan")

	_, pending := a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.3.0", false))
	status, body := a.call("PUT", "/admin/live-controls", root, breakGlassBody("1.4.0", "SEV1", "INC-12"))
	if status != http.StatusOK || body["config_version"] != float64(1) || body["break_glass_id"] == nil {
		t.Fatalf("break-glass %d %v", status, body)
	}
	id := fmt.Sprintf("%.0f", body["break_glass_id"].(float64))
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.4.0" || !controls.PurchasesDisabled {
		t.Fatalf("published %+v", controls)
	}
	switched := a.entries(root, "object_type=live_controls&action="+adminapi.BreakGlassAction)
	if len(switched) != 1 || switched[0]["ref"] != "INC-12" || switched[0]["reason"] != "incident" || switched[0]["config_version"] != float64(1) {
		t.Fatalf("break-glass entries %v", switched)
	}
	if record := switched[0]["after"].(map[string]any); record["approval_id"] != body["break_glass_id"] ||
		record["incident"].(map[string]any)["severity"] != "SEV1" || record["controls"].(map[string]any)["minimum_app_version"] != "1.4.0" {
		t.Fatalf("break-glass record %v", record)
	}
	if listed := a.unconfirmedBreakGlass(support); len(listed) != 1 || listed[0].(map[string]any)["proposer_id"] == "" {
		t.Fatalf("unconfirmed %v", listed)
	}

	status, body = a.call("POST", fmt.Sprintf("/admin/approvals/%.0f/approve", pending["approval_id"].(float64)), root, map[string]any{"reason": "checked"})
	expect(t, status, body, http.StatusConflict, "stale_proposal")
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", root, map[string]any{"reason": "mine"})
	expect(t, status, body, http.StatusForbidden, "self_decision")
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", support, map[string]any{"reason": "seen"})
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", liveOps, map[string]any{"reason": "confirmed in the incident channel"})
	if status != http.StatusOK || body["state"] != "approved" {
		t.Fatalf("confirm %d %v", status, body)
	}
	if listed := a.unconfirmedBreakGlass(support); len(listed) != 0 {
		t.Fatalf("confirmed still listed %v", listed)
	}
	confirmed := a.entries(root, "object_type=live_controls&action=approve")
	if len(confirmed) != 1 || confirmed[0]["ref"] != "INC-12" || confirmed[0]["actor_id"] == switched[0]["actor_id"] ||
		confirmed[0]["approval_id"] != switched[0]["id"] || confirmed[0]["after"].(map[string]any)["approval_id"] != switched[0]["after"].(map[string]any)["approval_id"] {
		t.Fatalf("confirmation entries %v", confirmed)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.4.0" {
		t.Fatalf("confirmation changed the controls %+v", controls)
	}
}

func TestBreakGlassIsLimitedPerAdminPerHour(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) {
		c.Environment = "production"
		c.Limits.BreakGlassPerHour = 2
	})
	a.create("vusal@example.com", rbac.SuperAdmin)
	root, other := a.signIn("root"), a.signIn("vusal")
	for minute, minimum := range []string{"1.4.0", "1.5.0"} {
		a.now = start.Add(time.Duration(minute) * time.Minute)
		if status, body := a.call("PUT", "/admin/live-controls", root, breakGlassBody(minimum, "SEV1", "INC-12")); status != http.StatusOK {
			t.Fatalf("break-glass %d within the limit: %d %v", minute+1, status, body)
		}
	}

	a.now = start.Add(3 * time.Minute)
	refused := a.raw("PUT", "/admin/live-controls", root, `{"controls":{"minimum_app_version":"1.6.0"},"reason":"incident","break_glass":{"severity":"SEV1","id":"INC-12"}}`)
	if refused.Code != http.StatusTooManyRequests || !strings.Contains(refused.Body.String(), `"error":"break_glass_rate_limited"`) ||
		!strings.Contains(refused.Body.String(), `"retry_at":"2026-11-01T10:00:00Z"`) || refused.Header().Get("Retry-After") != "3420" {
		t.Fatalf("over the limit: %d %v %s", refused.Code, refused.Header(), refused.Body.String())
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.5.0" {
		t.Fatalf("a refused break-glass switched the controls %+v", controls)
	}
	if status, body := a.call("PUT", "/admin/live-controls", other, breakGlassBody("1.6.0", "SEV1", "INC-12")); status != http.StatusOK {
		t.Fatalf("another super_admin shares the limit: %d %v", status, body)
	}

	a.now = start.Add(time.Hour)
	root = a.signIn("root")
	if status, body := a.call("PUT", "/admin/live-controls", root, breakGlassBody("1.7.0", "SEV1", "INC-12")); status != http.StatusOK {
		t.Fatalf("the oldest switch left the hour: %d %v", status, body)
	}
}

func TestRejectingABreakGlassRevertsTheControls(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	liveOps, root := a.signIn("leyla"), a.signIn("root")
	_, proposed := a.call("PUT", "/admin/live-controls", liveOps, controlsBody("1.3.0", false))
	if status, body := a.call("POST", fmt.Sprintf("/admin/approvals/%.0f/approve", proposed["approval_id"].(float64)), root, map[string]any{"reason": "checked"}); status != http.StatusOK {
		t.Fatalf("approve %d %v", status, body)
	}
	_, glass := a.call("PUT", "/admin/live-controls", root, breakGlassBody("1.4.0", "SEV1", "INC-12"))

	reject := fmt.Sprintf("/admin/approvals/%.0f/reject", glass["break_glass_id"].(float64))
	status, body := a.call("POST", reject, liveOps, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", reject, liveOps, map[string]any{"reason": "no incident was declared"})
	if status != http.StatusOK || body["state"] != "rejected" || body["revert_config_version"] != float64(3) {
		t.Fatalf("reject %d %v", status, body)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.3.0" || controls.PurchasesDisabled {
		t.Fatalf("not reverted %+v", controls)
	}
	if listed := a.unconfirmedBreakGlass(liveOps); len(listed) != 0 {
		t.Fatalf("rejected still listed %v", listed)
	}
	switched := a.entries(root, "object_type=live_controls&action="+adminapi.BreakGlassAction)
	rejected := a.entries(root, "object_type=live_controls&action=reject")
	reverted := a.entries(root, "object_type=live_controls&action="+adminapi.RevertBreakGlassAction)
	if len(rejected) != 1 || rejected[0]["approval_id"] != switched[0]["id"] || rejected[0]["ref"] != "INC-12" || rejected[0]["reason"] != "no incident was declared" {
		t.Fatalf("rejection entries %v", rejected)
	}
	if len(reverted) != 1 || reverted[0]["approval_id"] != rejected[0]["id"] || reverted[0]["reason"] != "no incident was declared" ||
		reverted[0]["ref"] != "INC-12" || reverted[0]["config_version"] != float64(3) || reverted[0]["actor_id"] != rejected[0]["actor_id"] ||
		reverted[0]["before"].(map[string]any)["minimum_app_version"] != "1.4.0" || reverted[0]["after"].(map[string]any)["minimum_app_version"] != "1.3.0" {
		t.Fatalf("revert entries %v", reverted)
	}
}

func TestARejectedBreakGlassKeepsANewerSwitchAndAnUnpublishedStart(t *testing.T) {
	a := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("vusal@example.com", rbac.SuperAdmin)
	liveOps, root, other := a.signIn("leyla"), a.signIn("root"), a.signIn("vusal")
	_, first := a.call("PUT", "/admin/live-controls", root, breakGlassBody("1.4.0", "SEV1", "INC-12"))
	_, second := a.call("PUT", "/admin/live-controls", other, breakGlassBody("1.5.0", "SEV1", "INC-13"))
	rejectFirst := fmt.Sprintf("/admin/approvals/%.0f/reject", first["break_glass_id"].(float64))
	status, body := a.call("POST", rejectFirst, liveOps, map[string]any{"reason": "disputed"})
	if status != http.StatusOK || body["state"] != "rejected" || body["revert_config_version"] != nil {
		t.Fatalf("reject under a newer switch %d %v", status, body)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.5.0" {
		t.Fatalf("a newer switch was undone %+v", controls)
	}
	rejectSecond := fmt.Sprintf("/admin/approvals/%.0f/reject", second["break_glass_id"].(float64))
	status, body = a.call("POST", rejectSecond, liveOps, map[string]any{"reason": "disputed"})
	if status != http.StatusOK || body["revert_config_version"] != float64(3) {
		t.Fatalf("reject the newest %d %v", status, body)
	}
	if controls := a.publishedControls(); controls.MinimumAppVersion != "1.4.0" {
		t.Fatalf("the newest switch was not reverted %+v", controls)
	}

	fresh := newAPIWith(t, func(c *adminapi.Config) { c.Environment = "production" })
	fresh.create("leyla@example.com", rbac.LiveOps)
	_, only := fresh.call("PUT", "/admin/live-controls", fresh.signIn("root"), breakGlassBody("1.4.0", "SEV1", "INC-12"))
	status, body = fresh.call("POST", fmt.Sprintf("/admin/approvals/%.0f/reject", only["break_glass_id"].(float64)), fresh.signIn("leyla"), map[string]any{"reason": "disputed"})
	if status != http.StatusOK || body["revert_config_version"] != nil {
		t.Fatalf("reject with nothing to restore %d %v", status, body)
	}
	if controls := fresh.publishedControls(); controls.MinimumAppVersion != "1.4.0" {
		t.Fatalf("controls without an earlier version changed %+v", controls)
	}
}
