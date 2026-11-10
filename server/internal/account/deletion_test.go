package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

func signedUp(t *testing.T, store *account.Store, pool *pgxpool.Pool, subject, nickname string) account.Account {
	t.Helper()
	ctx := context.Background()
	a, _, err := store.SignIn(ctx, idtoken.Identity{Provider: idtoken.ProviderGoogle, Subject: subject, Email: subject + "@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `insert into profiles (account_id, nickname, nickname_key) values ($1, $2, lower($2))`, a.ID, nickname); err != nil {
		t.Fatal(err)
	}
	return a
}

func count(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestDeletionWaitsForTheGraceAndCanBeCancelled(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := account.New(pool)
	ctx := context.Background()
	ayla := signedUp(t, store, pool, "g-ayla", "Ayla")
	identity := idtoken.Identity{Provider: idtoken.ProviderGoogle, Subject: "g-ayla"}
	requestedAt := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx, `insert into push_devices (token, account_id, platform, registered_at) values ('push-1', $1, 'android', now())`, ayla.ID); err != nil {
		t.Fatal(err)
	}
	if at, err := store.RequestDeletion(ctx, ayla.ID, requestedAt); err != nil || !at.Equal(requestedAt) {
		t.Fatalf("request: %v %v", at, err)
	}
	if n := count(t, pool, `select count(*) from push_devices where account_id = $1`, ayla.ID); n != 0 {
		t.Fatalf("push devices after the request: %d", n)
	}
	pending, err := store.Lookup(ctx, identity)
	if !errors.Is(err, account.ErrDeletionPending) || pending.ID != ayla.ID || pending.DeletionAt == nil || !pending.DeletionAt.Equal(requestedAt) {
		t.Fatalf("lookup while pending: %+v %v", pending, err)
	}
	if deleted, err := store.DeleteDue(ctx, requestedAt.Add(-time.Second), requestedAt); err != nil || deleted != 0 {
		t.Fatalf("delete before the grace: %d %v", deleted, err)
	}
	if err := store.CancelDeletion(ctx, ayla.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.CancelDeletion(ctx, ayla.ID); !errors.Is(err, account.ErrNoDeletionPending) {
		t.Fatalf("second cancel: %v", err)
	}
	if restored, _, err := store.SignIn(ctx, identity); err != nil || restored.ID != ayla.ID || restored.DeletionAt != nil {
		t.Fatalf("sign-in after cancel: %+v %v", restored, err)
	}
}

func TestBannedAccountsCannotRequestDeletion(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := account.New(pool)
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	temporary := signedUp(t, store, pool, "g-temp", "Temp")
	if _, err := pool.Exec(ctx, `update accounts set banned_until = $2 where id = $1`, temporary.ID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	permanent := signedUp(t, store, pool, "g-perm", "Perm")
	if _, err := pool.Exec(ctx, `update accounts set state = $2 where id = $1`, permanent.ID, account.StateBanned); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{temporary.ID, permanent.ID} {
		if _, err := store.RequestDeletion(ctx, id, now); !errors.Is(err, account.ErrBanned) {
			t.Fatalf("%s: %v", id, err)
		}
	}
}

func TestDeletionErasesPersonalDataAndKeepsSharedHistoryUnderATombstone(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := account.New(pool)
	ctx := context.Background()
	ayla := signedUp(t, store, pool, "g-ayla", "Ayla")
	ben := signedUp(t, store, pool, "g-ben", "Ben")
	for _, statement := range []string{
		`insert into friendships (account_id, friend_id) values ($1, $2), ($2, $1)`,
		`insert into nickname_history (account_id, nickname, nickname_key) values ($1, 'Old', 'old')`,
		`insert into purchases (platform, transaction_id, account_id, offer_id, product_id, environment, state, purchased_at) values ('google', 'tx-1', $1, 'offer', 'product', 'production', 'granted', now())`,
	} {
		if _, err := pool.Exec(ctx, statement, ayla.ID, ben.ID); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
	requestedAt := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	if _, err := store.RequestDeletion(ctx, ayla.ID, requestedAt); err != nil {
		t.Fatal(err)
	}
	now := requestedAt.Add(30 * 24 * time.Hour)
	if deleted, err := store.DeleteDue(ctx, now.Add(-30*24*time.Hour), now); err != nil || deleted != 1 {
		t.Fatalf("delete due: %d %v", deleted, err)
	}
	if deleted, err := store.DeleteDue(ctx, now, now); err != nil || deleted != 0 {
		t.Fatalf("second run: %d %v", deleted, err)
	}
	checks := map[string]int{
		`select count(*) from accounts where id in ($1, $2) and state = 'active'`:                                                                                   1,
		`select count(*) from auth_identities where subject = 'g-ayla' or email like 'g-ayla%' or account_id = $1 or $2 = ''`:                                       0,
		`select count(*) from profiles where nickname = 'Ayla' or account_id = $1 or $2 = ''`:                                                                       0,
		`select count(*) from nickname_history where nickname = 'Old' or account_id = $1 or $2 = ''`:                                                                0,
		`select count(*) from friendships where account_id in ($1, $2)`:                                                                                             0,
		`select count(*) from purchases p join accounts a on a.id = p.account_id where a.state = 'deleted' and a.deleted_at is not null and a.id not in ($1, $2)`:   1,
		`select count(*) from nickname_history h join accounts a on a.id = h.account_id where h.nickname = 'Ayla' and a.state = 'deleted' and a.id not in ($1, $2)`: 1,
	}
	for query, want := range checks {
		if got := count(t, pool, query, ayla.ID, ben.ID); got != want {
			t.Fatalf("%s: %d, want %d", query, got, want)
		}
	}
	fresh, created, err := store.SignIn(ctx, idtoken.Identity{Provider: idtoken.ProviderGoogle, Subject: "g-ayla"})
	if err != nil || !created || fresh.ID == ayla.ID {
		t.Fatalf("sign-in after deletion: %+v %v %v", fresh, created, err)
	}
}
