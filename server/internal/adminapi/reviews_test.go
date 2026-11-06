package adminapi_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/risk"
)

func (a *api) signal(account, kind string, weight int, at time.Time) {
	a.t.Helper()
	signal := risk.Signal{Account: account, Kind: kind, Weight: weight, Dedupe: fmt.Sprint(at.UnixNano()), Evidence: map[string]any{"count": 1}, At: at}
	if _, err := risk.NewStore(a.pool).Record(context.Background(), signal); err != nil {
		a.t.Fatal(err)
	}
}

func (a *api) shadowbanned(account string) bool {
	a.t.Helper()
	var banned bool
	if err := a.pool.QueryRow(context.Background(), `select shadowbanned_at is not null from accounts where id = $1`, account).Scan(&banned); err != nil {
		a.t.Fatal(err)
	}
	return banned
}

func reviewID(t *testing.T, body map[string]any, index int) string {
	t.Helper()
	reviews := body["reviews"].([]any)
	return fmt.Sprintf("%.0f", reviews[index].(map[string]any)["id"].(float64))
}

func TestSupportReadsAndClearsTheReviewQueue(t *testing.T) {
	a := newAPI(t)
	a.now = time.Now().UTC()
	a.create("orxan@example.com", rbac.Support)
	support, content, root := a.signIn("orxan"), a.signIn("ayla"), a.signIn("root")
	id := a.player("g-1", "kamal@mail.example", "Kamal")
	a.signal(id, risk.KindFakeReceipt, 20, a.now)
	wave := risk.NewWave(a.pool, risk.DefaultConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	a.signal(id, risk.KindReplay, 40, a.now.Add(time.Second))
	if _, err := wave.RunDue(context.Background(), a.now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if !a.shadowbanned(id) {
		t.Fatal("the wave did not shadowban")
	}

	status, body := a.call("GET", "/admin/reviews", content, nil)
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("GET", "/admin/reviews", support, nil)
	reviews := body["reviews"].([]any)
	if status != http.StatusOK || len(reviews) != 1 {
		t.Fatalf("reviews %d %v", status, body)
	}
	review := reviews[0].(map[string]any)
	if review["account_id"] != id || review["nickname"] != "Kamal" || review["cause"] != risk.CauseHardSignal || len(review["signals"].([]any)) != 2 {
		t.Fatalf("review %v", review)
	}
	path := "/admin/reviews/" + reviewID(t, body, 0) + "/clear"
	a.now = a.now.Add(2 * time.Minute)

	status, body = a.call("POST", path, support, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", path, support, map[string]any{"reason": "bad network", "ref": "OPS-2"})
	if status != http.StatusOK || body["shadowbanned"] != false {
		t.Fatalf("clear %d %v", status, body)
	}
	if a.shadowbanned(id) {
		t.Fatal("clearing kept the shadowban")
	}
	status, body = a.call("POST", path, support, map[string]any{"reason": "again"})
	expect(t, status, body, http.StatusConflict, "decided")
	status, body = a.call("POST", "/admin/reviews/999999/clear", support, map[string]any{"reason": "none"})
	expect(t, status, body, http.StatusNotFound, "not_found")

	entries := a.entries(root, "object_type=queues&action="+adminapi.ClearReviewAction)
	if len(entries) != 1 || entries[0]["reason"] != "bad network" || entries[0]["ref"] != "OPS-2" {
		t.Fatalf("entries %v", entries)
	}
	if _, err := wave.RunDue(context.Background(), a.now.Add(risk.DefaultConfig().BanWaveInterval+time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, body := a.call("GET", "/admin/reviews", support, nil); len(body["reviews"].([]any)) != 0 || a.shadowbanned(id) {
		t.Fatalf("cleared signals acted again: %v", body)
	}
}

func TestConfirmingAReviewKeepsTheShadowbanAndWaitsForNewSignals(t *testing.T) {
	a := newAPI(t)
	a.now = time.Now().UTC()
	a.create("orxan@example.com", rbac.Support)
	support := a.signIn("orxan")
	id := a.player("g-1", "kamal@mail.example", "Kamal")
	config := risk.DefaultConfig()
	wave := risk.NewWave(a.pool, config, slog.New(slog.NewTextHandler(io.Discard, nil)))
	a.signal(id, risk.KindReplay, config.ShadowbanScore, a.now)
	if _, err := wave.RunDue(context.Background(), a.now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	_, body := a.call("GET", "/admin/reviews", support, nil)
	status, body := a.call("POST", "/admin/reviews/"+reviewID(t, body, 0)+"/confirm", support, map[string]any{"reason": "cheat proven"})
	if status != http.StatusOK || body["shadowbanned"] != true {
		t.Fatalf("confirm %d %v", status, body)
	}
	decided := time.Now().UTC()
	if _, err := wave.RunDue(context.Background(), decided.Add(config.BanWaveInterval)); err != nil {
		t.Fatal(err)
	}
	if _, body := a.call("GET", "/admin/reviews", support, nil); len(body["reviews"].([]any)) != 0 {
		t.Fatalf("a handled signal reopened the review: %v", body)
	}
	a.signal(id, risk.KindReplay, config.ReviewScore, decided.Add(time.Hour))
	if _, err := wave.RunDue(context.Background(), decided.Add(2*config.BanWaveInterval)); err != nil {
		t.Fatal(err)
	}
	if _, body := a.call("GET", "/admin/reviews", support, nil); len(body["reviews"].([]any)) != 1 {
		t.Fatalf("a new signal opened no review: %v", body)
	}
}
