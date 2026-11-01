package adminapi_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func (a *api) playerSignIn(subject string) error {
	a.t.Helper()
	_, _, err := account.New(a.pool).SignIn(context.Background(), idtoken.Identity{Provider: idtoken.ProviderGoogle, Subject: subject, EmailVerified: true})
	return err
}

func (a *api) liveSessions(accountID string) int {
	a.t.Helper()
	var count int
	if err := a.pool.QueryRow(context.Background(), `select count(*) from auth_sessions where account_id = $1 and revoked_at is null`, accountID).Scan(&count); err != nil {
		a.t.Fatal(err)
	}
	return count
}

func TestATemporaryBanBlocksSignInRevokesSessionsAndCanBeLifted(t *testing.T) {
	a := newAPI(t)
	a.now = time.Now().UTC()
	a.create("orxan@example.com", rbac.Support)
	support, content, root := a.signIn("orxan"), a.signIn("ayla"), a.signIn("root")
	id := a.player("g-1", "kamal@mail.example", "Kamal")
	if _, err := a.pool.Exec(context.Background(), `insert into auth_sessions (account_id, device_id, refresh_hash, expires_at) values ($1, 'd-1', '\x01', $2), ($1, 'd-2', '\x02', $2)`,
		id, a.now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	path := "/admin/players/" + id + "/temporary-ban"

	status, body := a.call("DELETE", path, support, map[string]any{"reason": "appeal"})
	expect(t, status, body, http.StatusConflict, "not_banned")
	status, body = a.call("POST", path, support, map[string]any{})
	expect(t, status, body, http.StatusBadRequest, "reason_required")
	status, body = a.call("POST", path, content, map[string]any{"reason": "cheating"})
	expect(t, status, body, http.StatusForbidden, "forbidden")

	status, body = a.call("POST", path, support, map[string]any{"reason": "speed hack", "ref": "REV-12"})
	until, _ := time.Parse(time.RFC3339Nano, body["banned_until"].(string))
	if status != http.StatusOK || body["sessions_revoked"] != float64(2) || !until.Equal(a.now.Add(72*time.Hour).Truncate(time.Microsecond)) {
		t.Fatalf("ban %d %v", status, body)
	}
	if err := a.playerSignIn("g-1"); !errors.Is(err, account.ErrBanned) {
		t.Fatalf("sign-in while banned: %v", err)
	}
	if got := a.liveSessions(id); got != 0 {
		t.Fatalf("live sessions %d", got)
	}
	status, body = a.call("GET", "/admin/players/"+id, support, nil)
	if status != http.StatusOK || body["account"].(map[string]any)["banned_until"] == nil {
		t.Fatalf("record %d %v", status, body)
	}

	status, body = a.call("DELETE", path, support, map[string]any{"reason": "appeal upheld"})
	if status != http.StatusOK || body["banned_until"] != nil {
		t.Fatalf("lift %d %v", status, body)
	}
	if err := a.playerSignIn("g-1"); err != nil {
		t.Fatalf("sign-in after lift: %v", err)
	}

	entries := a.entries(root, "object_type=players&object_id="+id)
	if len(entries) != 3 || entries[0]["action"] != "lift_temporary_ban" || entries[0]["reason"] != "appeal upheld" ||
		entries[0]["before"].(map[string]any)["banned_until"] != until.Format(time.RFC3339Nano) || entries[1]["action"] != "view" ||
		entries[2]["action"] != "temporary_ban" || entries[2]["ref"] != "REV-12" || entries[2]["after"].(map[string]any)["sessions_revoked"] != float64(2) {
		t.Fatalf("entries %v", entries)
	}
}

func TestATemporaryBanNeedsItsConfiguredLength(t *testing.T) {
	a := newAPIWith(t, func(config *adminapi.Config) { config.TempBan = 0 })
	id := a.player("g-1", "kamal@mail.example", "Kamal")
	status, body := a.call("POST", "/admin/players/"+id+"/temporary-ban", a.signIn("root"), map[string]any{"reason": "speed hack"})
	expect(t, status, body, http.StatusConflict, "not_configured")
	if err := a.playerSignIn("g-1"); err != nil {
		t.Fatalf("sign-in: %v", err)
	}
}

func TestOnlySuperAdminSearchesByAuthIdentity(t *testing.T) {
	a := newAPI(t)
	a.create("orxan@example.com", rbac.Support)
	id := a.player("g-1", "kamal@mail.example", "Kamal")
	status, body := a.call("GET", "/admin/players?provider=google&subject=g-1", a.signIn("orxan"), nil)
	expect(t, status, body, http.StatusForbidden, "forbidden")
	status, body = a.call("GET", "/admin/players?provider=google&subject=g-1", a.signIn("root"), nil)
	if found := body["players"].([]any); status != http.StatusOK || len(found) != 1 || found[0].(map[string]any)["id"] != id {
		t.Fatalf("identity search %d %v", status, body)
	}
	status, body = a.call("GET", "/admin/players?provider=apple&subject=g-1", a.signIn("root"), nil)
	if status != http.StatusOK || len(body["players"].([]any)) != 0 {
		t.Fatalf("other provider %d %v", status, body)
	}
}
