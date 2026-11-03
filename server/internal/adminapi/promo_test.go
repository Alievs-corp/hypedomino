package adminapi_test

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

const promoKey = "0123456789abcdef0123456789abcdef"

func newPromoAPI(t *testing.T) *api {
	t.Helper()
	a := newAPIWith(t, func(c *adminapi.Config) {
		c.PromoKey, c.Promo, c.Cosmetics = promoKey, promo.DefaultConfig(), cosmetics.Builtin()
	})
	a.create("leyla@example.com", rbac.LiveOps)
	a.create("vusal@example.com", rbac.LiveOps)
	return a
}

func promoBody(id string, coins int, extra map[string]any) map[string]any {
	body := map[string]any{
		"id": id, "campaign": "campaign.launch_week", "kind": "multi_use", "cap": 50,
		"bundle": map[string]any{"id": "bundle.launch_coins", "revision": 1, "coins": coins},
		"start":  start, "end": start.Add(7 * 24 * time.Hour), "reason": "launch week",
	}
	for key, value := range extra {
		body[key] = value
	}
	return body
}

func (a *api) batch(token string, body map[string]any) *httptest.ResponseRecorder {
	a.t.Helper()
	encoded, _ := json.Marshal(body)
	return a.raw("POST", "/admin/promo-codes/batch", token, string(encoded))
}

func (a *api) promoCodes(token, query string) map[string]any {
	a.t.Helper()
	status, body := a.call("GET", "/admin/promo-codes"+query, token, nil)
	if status != http.StatusOK {
		a.t.Fatalf("list %d %v", status, body)
	}
	return body
}

func (a *api) promoState(id string) string {
	a.t.Helper()
	var state string
	if err := a.pool.QueryRow(context.Background(), `select state from promo_codes where promo_id = $1`, id).Scan(&state); err != nil {
		a.t.Fatal(err)
	}
	return state
}

func (a *api) storedCode(plaintext string) string {
	a.t.Helper()
	hasher, _ := promo.NewHasher(promoKey)
	normalized, err := promo.Normalize(plaintext)
	if err != nil {
		a.t.Fatal(err)
	}
	var id string
	if err := a.pool.QueryRow(context.Background(), `select promo_id from promo_codes where code_hash = $1`, hasher.Hash(normalized)).Scan(&id); err != nil {
		a.t.Fatalf("code %s not stored: %v", plaintext, err)
	}
	return id
}

func TestPromoCodesNeedTheirKey(t *testing.T) {
	a := newAPI(t)
	status, body := a.call("GET", "/admin/promo-codes", a.signIn("root"), nil)
	expect(t, status, body, http.StatusConflict, "not_configured")
}

func TestLiveOpsCreatesAPromoCodeAndSeesItOnce(t *testing.T) {
	a := newPromoAPI(t)
	liveOps, content, root := a.signIn("leyla"), a.signIn("ayla"), a.signIn("root")

	status, body := a.call("POST", "/admin/promo-codes", liveOps, promoBody("promo.launch", 300, map[string]any{"reason": " "}))
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", "/admin/promo-codes", content, promoBody("promo.launch", 300, nil))
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("POST", "/admin/promo-codes", liveOps, promoBody("promo.launch", 300, map[string]any{"cap": 0}))
	expect(t, status, body, http.StatusBadRequest, "invalid")

	status, body = a.call("POST", "/admin/promo-codes", liveOps, promoBody("promo.launch", 300, map[string]any{"code": "hype 2345"}))
	if status != http.StatusCreated || body["plaintext"] != "HYPE2345" || body["approval_id"] != nil {
		t.Fatalf("create %d %v", status, body)
	}
	if code := body["code"].(map[string]any); code["state"] != "active" || code["cap"] != float64(50) || code["creator"] == "" {
		t.Fatalf("code %v", code)
	}
	if a.storedCode("HYPE2345") != "promo.launch" {
		t.Fatal("code not stored under its hash")
	}
	status, body = a.call("POST", "/admin/promo-codes", liveOps, promoBody("promo.launch", 300, nil))
	expect(t, status, body, http.StatusConflict, "duplicate")

	listed := a.promoCodes(liveOps, "?campaign=campaign.launch_week")
	codes, campaigns := listed["codes"].([]any), listed["campaigns"].([]any)
	if len(codes) != 1 || codes[0].(map[string]any)["id"] != "promo.launch" {
		t.Fatalf("codes %v", codes)
	}
	usage := campaigns[0].(map[string]any)
	if usage["campaign"] != "campaign.launch_week" || usage["codes"] != float64(1) || usage["active"] != float64(1) || usage["cap"] != float64(50) || usage["redemptions"] != float64(0) {
		t.Fatalf("usage %v", usage)
	}
	if listed := a.promoCodes(liveOps, "?campaign=campaign.other"); len(listed["codes"].([]any)) != 0 {
		t.Fatalf("filtered %v", listed)
	}

	entries := a.entries(root, "object_type=promo_codes&action=create_promo")
	if len(entries) != 1 || entries[0]["object_id"] != "promo.launch" || entries[0]["reason"] != "launch week" {
		t.Fatalf("entries %v", entries)
	}
	if raw, _ := json.Marshal(entries[0]); strings.Contains(string(raw), "HYPE2345") {
		t.Fatalf("plaintext in audit: %s", raw)
	}
}

func TestAPromoAboveTheThresholdWaitsForASecondLiveOps(t *testing.T) {
	a := newPromoAPI(t)
	proposer, second, root := a.signIn("leyla"), a.signIn("vusal"), a.signIn("root")

	status, body := a.call("POST", "/admin/promo-codes", proposer, promoBody("promo.big", 2001, nil))
	if status != http.StatusCreated || body["code"].(map[string]any)["state"] != "pending_approval" || body["approval_id"] == nil {
		t.Fatalf("create %d %v", status, body)
	}
	id := strconv.FormatFloat(body["approval_id"].(float64), 'f', 0, 64)
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", proposer, map[string]any{})
	expect(t, status, body, http.StatusForbidden, "self_decision")
	status, body = a.call("POST", "/admin/approvals/"+id+"/approve", second, map[string]any{"reason": "checked"})
	if status != http.StatusOK || body["state"] != "approved" {
		t.Fatalf("approve %d %v", status, body)
	}
	if state := a.promoState("promo.big"); state != "active" {
		t.Fatalf("state %s", state)
	}
	entries := a.entries(root, "object_type=promo_codes&action=activate_promo")
	if len(entries) != 1 || entries[0]["approval_id"] == nil || entries[0]["after"].(map[string]any)["codes"] != float64(1) {
		t.Fatalf("entries %v", entries)
	}
}

func TestABatchIsHandedOutOnceAndDecidedAsOne(t *testing.T) {
	a := newPromoAPI(t)
	proposer, second := a.signIn("leyla"), a.signIn("vusal")

	for _, count := range []int{0, adminapi.MaxPromoBatchCount + 1} {
		status, body := a.call("POST", "/admin/promo-codes/batch", proposer, promoBody("promo.gift", 300, map[string]any{"count": count, "cap": 0}))
		expect(t, status, body, http.StatusBadRequest, "invalid")
	}

	response := a.batch(proposer, promoBody("promo.gift", 2001, map[string]any{"count": 3, "cap": 0}))
	if response.Code != http.StatusCreated || response.Header().Get("Hd-Promo-State") != "pending_approval" || !strings.HasPrefix(response.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("batch %d %v %s", response.Code, response.Header(), response.Body)
	}
	rows, err := csv.NewReader(response.Body).ReadAll()
	if err != nil || len(rows) != 4 || rows[0][0] != "promo_id" || rows[1][0] != "promo.gift_1" {
		t.Fatalf("csv %v %v", rows, err)
	}
	for _, row := range rows[1:] {
		if a.storedCode(row[1]) != row[0] {
			t.Fatalf("row %v", row)
		}
	}
	id := response.Header().Get("Hd-Approval-Id")
	status, body := a.call("POST", "/admin/approvals/"+id+"/reject", second, map[string]any{"reason": "wrong bundle"})
	if status != http.StatusOK || body["state"] != "rejected" {
		t.Fatalf("reject %d %v", status, body)
	}
	if state := a.promoState("promo.gift_2"); state != "disabled" {
		t.Fatalf("rejected batch state %s", state)
	}

	response = a.batch(proposer, promoBody("promo.cheap", 300, map[string]any{"count": 2, "cap": 0}))
	if response.Code != http.StatusCreated || response.Header().Get("Hd-Promo-State") != "active" || response.Header().Get("Hd-Approval-Id") != "" {
		t.Fatalf("batch %d %v", response.Code, response.Header())
	}
	status, body = a.call("POST", "/admin/promo-batches/promo.cheap/disable", second, map[string]any{"reason": "leaked"})
	if status != http.StatusOK || body["disabled"] != float64(2) {
		t.Fatalf("disable batch %d %v", status, body)
	}
	usage := a.promoCodes(second, "")["campaigns"].([]any)[0].(map[string]any)
	if usage["codes"] != float64(5) || usage["disabled"] != float64(5) || usage["active"] != float64(0) {
		t.Fatalf("usage %v", usage)
	}
}

func TestDisablingACodeIsImmediateAndAudited(t *testing.T) {
	a := newPromoAPI(t)
	liveOps, root := a.signIn("leyla"), a.signIn("root")
	a.call("POST", "/admin/promo-codes", liveOps, promoBody("promo.leak", 300, nil))

	status, body := a.call("POST", "/admin/promo-codes/promo.leak/disable", root, map[string]any{"reason": "leaked"})
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("POST", "/admin/promo-codes/promo.leak/disable", liveOps, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	a.now = start.Add(11 * time.Minute)
	status, body = a.call("POST", "/admin/promo-codes/promo.leak/disable", liveOps, map[string]any{"reason": "leaked"})
	expect(t, status, body, http.StatusForbidden, "step_up_required")
	a.call("POST", "/admin/session/step-up", liveOps, map[string]string{"id_token": "leyla", "nonce": "nonce"})
	for range 2 {
		status, body = a.call("POST", "/admin/promo-codes/promo.leak/disable", liveOps, map[string]any{"reason": "leaked", "ref": "OPS-9"})
		if status != http.StatusOK || body["id"] != "promo.leak" {
			t.Fatalf("disable %d %v", status, body)
		}
	}
	if state := a.promoState("promo.leak"); state != "disabled" {
		t.Fatalf("state %s", state)
	}
	status, body = a.call("POST", "/admin/promo-codes/promo.none/disable", liveOps, map[string]any{"reason": "leaked"})
	expect(t, status, body, http.StatusNotFound, "not_found")

	entries := a.entries(root, "object_type=promo_codes&action=disable_promo")
	if len(entries) != 1 || entries[0]["ref"] != "OPS-9" || entries[0]["after"].(map[string]any)["state"] != "disabled" {
		t.Fatalf("entries %v", entries)
	}
}

func TestTenThousandCodesComeOutInUnderTenSeconds(t *testing.T) {
	a := newPromoAPI(t)
	began := time.Now()
	response := a.batch(a.signIn("leyla"), promoBody("promo.mass", 300, map[string]any{"count": adminapi.MaxPromoBatchCount, "cap": 0}))
	took := time.Since(began)
	rows, err := csv.NewReader(response.Body).ReadAll()
	if response.Code != http.StatusCreated || err != nil || len(rows) != adminapi.MaxPromoBatchCount+1 {
		t.Fatalf("batch %d, %d rows, %v", response.Code, len(rows), err)
	}
	if took >= 10*time.Second {
		t.Fatalf("%d codes took %s", adminapi.MaxPromoBatchCount, took)
	}
	t.Logf("%d codes in %s", adminapi.MaxPromoBatchCount, took)
}
