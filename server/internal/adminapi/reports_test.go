package adminapi_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/report"
)

func (a *api) fileReport(reporter, target, reason string, at time.Time) {
	a.t.Helper()
	if _, err := a.pool.Exec(context.Background(), `insert into player_reports (reporter, target, reason, name_text, created_at, updated_at) values ($1, $2, $3, 'Rude', $4, $4)`, reporter, target, reason, at); err != nil {
		a.t.Fatal(err)
	}
}

func TestSupportUpholdsOrRejectsTheOpenReportsOfAPlayer(t *testing.T) {
	a := newAPI(t)
	a.now = time.Now().UTC()
	a.create("orxan@example.com", rbac.Support)
	support, content, root := a.signIn("orxan"), a.signIn("ayla"), a.signIn("root")
	rude := a.player("g-1", "rude@mail.example", "Rude")
	calm := a.player("g-2", "calm@mail.example", "Calm")
	first := a.player("g-3", "first@mail.example", "First")
	second := a.player("g-4", "second@mail.example", "Second")
	a.fileReport(first, rude, report.ReasonNickname, a.now.Add(-2*time.Hour))
	a.fileReport(second, rude, report.ReasonImpersonation, a.now.Add(-time.Hour))
	a.fileReport(first, calm, report.ReasonOther, a.now.Add(-30*time.Minute))

	status, body := a.call("GET", "/admin/reports", content, nil)
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("GET", "/admin/reports", support, nil)
	targets := body["targets"].([]any)
	if status != http.StatusOK || len(targets) != 2 || targets[0].(map[string]any)["account_id"] != rude || len(targets[0].(map[string]any)["reports"].([]any)) != 2 {
		t.Fatalf("queue %d %v", status, body)
	}

	status, body = a.call("POST", "/admin/players/"+rude+"/reports/reject", support, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", "/admin/players/"+rude+"/reports/reject", support, map[string]any{"reason": "name is fine"})
	if status != http.StatusOK || body["closed"] != float64(2) || body["state"] != report.StateRejected {
		t.Fatalf("reject %d %v", status, body)
	}
	status, body = a.call("POST", "/admin/players/"+rude+"/reports/uphold", support, map[string]any{"reason": "again"})
	expect(t, status, body, http.StatusConflict, "no_open_reports")
	status, body = a.call("POST", "/admin/players/"+calm+"/reports/uphold", support, map[string]any{"reason": "confirmed"})
	if status != http.StatusOK || body["closed"] != float64(1) {
		t.Fatalf("uphold %d %v", status, body)
	}
	if _, body := a.call("GET", "/admin/reports", support, nil); len(body["targets"].([]any)) != 0 {
		t.Fatalf("queue after decisions %v", body)
	}
	rejected := a.entries(root, "object_type=players&action="+adminapi.RejectReportsAction)
	upheld := a.entries(root, "object_type=players&action="+adminapi.UpholdReportsAction)
	if len(rejected) != 1 || len(upheld) != 1 {
		t.Fatalf("entries %v %v", rejected, upheld)
	}
}
