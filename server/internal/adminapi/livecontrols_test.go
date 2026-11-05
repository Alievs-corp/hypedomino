package adminapi_test

import (
	"context"
	"fmt"
	"net/http"
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
