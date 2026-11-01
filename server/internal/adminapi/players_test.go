package adminapi_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
	"github.com/Alievs-corp/hypedomino/server/internal/profile"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

func (a *api) player(subject, email string, nicknames ...string) string {
	a.t.Helper()
	ctx := context.Background()
	created, _, err := account.New(a.pool).SignIn(ctx, idtoken.Identity{Provider: idtoken.ProviderGoogle, Subject: subject, Email: email, EmailVerified: true})
	if err != nil {
		a.t.Fatal(err)
	}
	profiles := profile.NewStore(a.pool, profile.DefaultConfig())
	for i, nickname := range nicknames {
		if _, err := profiles.SetNickname(ctx, created.ID, nickname, start.Add(time.Duration(i)*31*24*time.Hour)); err != nil {
			a.t.Fatal(err)
		}
	}
	return created.ID
}

func TestSupportFindsAPlayerAndEveryRecordViewIsAudited(t *testing.T) {
	a := newAPI(t)
	a.create("orxan@example.com", rbac.Support)
	support, root, content := a.signIn("orxan"), a.signIn("root"), a.signIn("ayla")
	id := a.player("g-1", "Kamal@Mail.example", "OldName", "Kamal_Star")
	a.player("g-2", "other@mail.example", "Kamala")
	a.player("g-3", "third@mail.example", "Kamxal")
	if _, err := wallet.NewStore(a.pool).Apply(context.Background(), wallet.Entry{
		Account: id, Amount: 250, Reason: wallet.MatchReward, Source: wallet.Source{Type: "match", Ref: "m-1"}, Key: "match:m-1", ConfigVersion: 1,
	}, start); err != nil {
		t.Fatal(err)
	}

	for query, want := range map[string]int{"kamal": 2, "KAMAL_": 1, "kam%": 0, "kam": 3, "kamal@mail.example": 1, id: 1, "nobody": 0, "oldname": 0} {
		status, body := a.call("GET", "/admin/players?q="+url.QueryEscape(query), support, nil)
		if status != http.StatusOK || len(body["players"].([]any)) != want {
			t.Fatalf("search %q: %d %v", query, status, body)
		}
	}
	status, body := a.call("GET", "/admin/players?q=%20", support, nil)
	expect(t, status, body, http.StatusBadRequest, "invalid")
	if entries := a.entries(root, "object_type=players"); len(entries) != 0 {
		t.Fatalf("a search wrote entries: %v", entries)
	}

	status, body = a.call("GET", "/admin/players/"+id, support, nil)
	if status != http.StatusOK {
		t.Fatalf("record %d %v", status, body)
	}
	player := body["account"].(map[string]any)
	history := body["nickname_history"].([]any)
	ledger := body["ledger"].([]any)
	if player["nickname"] != "Kamal_Star" || player["emails"].([]any)[0] != "Kamal@Mail.example" || player["identities"] != nil ||
		len(history) != 1 || history[0].(map[string]any)["nickname"] != "OldName" ||
		body["balance"] != float64(250) || len(ledger) != 1 || ledger[0].(map[string]any)["reason"] != "match_reward" {
		t.Fatalf("record %v", body)
	}
	status, body = a.call("GET", "/admin/players/"+id, root, nil)
	if identities := body["account"].(map[string]any)["identities"].([]any); status != http.StatusOK || len(identities) != 1 ||
		identities[0].(map[string]any)["subject"] != "g-1" {
		t.Fatalf("super_admin record %d %v", status, body)
	}
	for _, missing := range []string{"00000000-0000-0000-0000-000000000000", "kamal"} {
		status, body = a.call("GET", "/admin/players/"+missing, support, nil)
		expect(t, status, body, http.StatusNotFound, "not_found")
	}
	for _, path := range []string{"/admin/players?q=kamal", "/admin/players/" + id} {
		status, body = a.call("GET", path, content, nil)
		expect(t, status, body, http.StatusForbidden, "forbidden")
	}

	entries := a.entries(root, "object_type=players&object_id="+id)
	if len(entries) != 2 || entries[0]["action"] != "view" || entries[1]["actor_id"] == entries[0]["actor_id"] {
		t.Fatalf("view entries %v", entries)
	}
}
