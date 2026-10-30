package wallet_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

var start = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func account(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `insert into accounts default values returning id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

var match = wallet.Source{Type: "match", Ref: "m1"}

func TestEntriesAreIdempotentPerAccountAndKey(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := wallet.NewStore(pool)
	ctx := context.Background()
	ayla, ben := account(t, pool), account(t, pool)
	if balance, err := store.Balance(ctx, ayla); err != nil || balance != 0 {
		t.Fatalf("balance before any entry %d %v", balance, err)
	}
	first, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: 120, Reason: wallet.MatchReward, Source: match, Key: "match:m1", ConfigVersion: 3}, start)
	if err != nil || first.Repeated || first.BalanceAfter != 120 || first.ID == 0 || !first.At.Equal(start) || first.ConfigVersion != 3 {
		t.Fatalf("first entry %+v %v", first, err)
	}
	again, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: 999, Reason: wallet.MatchReward, Source: match, Key: "match:m1", ConfigVersion: 4}, start.Add(time.Hour))
	if err != nil || !again.Repeated || again.ID != first.ID || again.Amount != 120 || again.BalanceAfter != 120 {
		t.Fatalf("repeated entry %+v %v", again, err)
	}
	if balance, err := store.Balance(ctx, ayla); err != nil || balance != 120 {
		t.Fatalf("balance after the repeat %d %v", balance, err)
	}
	other, err := store.Apply(ctx, wallet.Entry{Account: ben, Amount: 50, Reason: wallet.MatchReward, Source: match, Key: "match:m1"}, start)
	if err != nil || other.Repeated || other.BalanceAfter != 50 {
		t.Fatalf("same key on another account %+v %v", other, err)
	}
	history, err := store.History(ctx, ayla, 10)
	if err != nil || len(history) != 1 || history[0].ID != first.ID {
		t.Fatalf("history %+v %v", history, err)
	}
}

func TestEntriesAreValidatedAndSpendsNeverGoNegative(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := wallet.NewStore(pool)
	ctx := context.Background()
	ayla := account(t, pool)
	for name, entry := range map[string]struct {
		entry wallet.Entry
		err   error
	}{
		"no key":         {wallet.Entry{Account: ayla, Amount: 10, Reason: wallet.MatchReward, Source: match}, wallet.ErrInvalid},
		"no source":      {wallet.Entry{Account: ayla, Amount: 10, Reason: wallet.MatchReward, Key: "k"}, wallet.ErrInvalid},
		"zero":           {wallet.Entry{Account: ayla, Amount: 0, Reason: wallet.MatchReward, Source: match, Key: "k"}, wallet.ErrInvalid},
		"reason":         {wallet.Entry{Account: ayla, Amount: 10, Reason: "gift", Source: match, Key: "k"}, wallet.ErrUnknownReason},
		"credit sign":    {wallet.Entry{Account: ayla, Amount: -10, Reason: wallet.MatchReward, Source: match, Key: "k"}, wallet.ErrWrongSign},
		"debit sign":     {wallet.Entry{Account: ayla, Amount: 10, Reason: wallet.Rename, Source: match, Key: "k"}, wallet.ErrWrongSign},
		"overspend":      {wallet.Entry{Account: ayla, Amount: -1, Reason: wallet.Rename, Source: match, Key: "spend:1"}, wallet.ErrInsufficient},
		"admin negative": {wallet.Entry{Account: ayla, Amount: -5, Reason: wallet.AdminAdjust, Source: wallet.Source{Type: "admin", Ref: "a1"}, Key: "admin:a1"}, wallet.ErrInsufficient},
	} {
		if _, err := store.Apply(ctx, entry.entry, start); !errors.Is(err, entry.err) {
			t.Fatalf("%s: %v, want %v", name, err, entry.err)
		}
	}
	if history, err := store.History(ctx, ayla, 10); err != nil || len(history) != 0 {
		t.Fatalf("refused entries were written: %+v %v", history, err)
	}
	if _, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: 1000, Reason: wallet.Promo, Source: wallet.Source{Type: "promo", Ref: "c1"}, Key: "promo:c1"}, start); err != nil {
		t.Fatal(err)
	}
	spent, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: -1000, Reason: wallet.Rename, Source: wallet.Source{Type: "rename", Ref: "r1"}, Key: "spend:r1"}, start)
	if err != nil || spent.BalanceAfter != 0 {
		t.Fatalf("spend to zero %+v %v", spent, err)
	}
	if adjusted, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: 7, Reason: wallet.AdminAdjust, Source: wallet.Source{Type: "admin", Ref: "a2"}, Key: "admin:a2"}, start); err != nil || adjusted.BalanceAfter != 7 {
		t.Fatalf("positive adjust %+v %v", adjusted, err)
	}
	if adjusted, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: -7, Reason: wallet.AdminAdjust, Source: wallet.Source{Type: "admin", Ref: "a3"}, Key: "admin:a3"}, start); err != nil || adjusted.BalanceAfter != 0 {
		t.Fatalf("negative adjust %+v %v", adjusted, err)
	}
}

func TestParallelSpendsNeverOverdraw(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := wallet.NewStore(pool)
	ctx := context.Background()
	ayla := account(t, pool)
	if _, err := store.Apply(ctx, wallet.Entry{Account: ayla, Amount: 500, Reason: wallet.Purchase, Source: wallet.Source{Type: "store", Ref: "t1"}, Key: "purchase:google:t1"}, start); err != nil {
		t.Fatal(err)
	}
	const attempts = 12
	var wg sync.WaitGroup
	errs := make([]error, attempts)
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = store.Apply(ctx, wallet.Entry{Account: ayla, Amount: -100, Reason: wallet.ShopPurchase, Source: wallet.Source{Type: "offer", Ref: "o1"}, Key: fmt.Sprintf("spend:%d", i)}, start)
		}()
	}
	wg.Wait()
	won := 0
	for _, err := range errs {
		switch {
		case err == nil:
			won++
		case !errors.Is(err, wallet.ErrInsufficient):
			t.Fatalf("unexpected error %v", err)
		}
	}
	balance, err := store.Balance(ctx, ayla)
	if err != nil || won != 5 || balance != 0 {
		t.Fatalf("%d spends won, balance %d %v", won, balance, err)
	}
	history, err := store.History(ctx, ayla, 100)
	if err != nil || len(history) != 6 || history[0].BalanceAfter != 0 || history[5].BalanceAfter != 500 {
		t.Fatalf("history %+v %v", history, err)
	}
}

func TestAPurchaseTransactionIsGrantedToOneAccountOnly(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := wallet.NewStore(pool)
	ctx := context.Background()
	ayla, ben := account(t, pool), account(t, pool)
	entry := wallet.Entry{Account: ayla, Amount: 500, Reason: wallet.Purchase, Source: wallet.Source{Type: "store", Ref: "t1"}, Key: "purchase:google:t1"}
	if _, err := store.Apply(ctx, entry, start); err != nil {
		t.Fatal(err)
	}
	entry.Account = ben
	if _, err := store.Apply(ctx, entry, start); !errors.Is(err, wallet.ErrPurchaseReuse) {
		t.Fatalf("purchase reused on another account: %v", err)
	}
	if balance, err := store.Balance(ctx, ben); err != nil || balance != 0 {
		t.Fatalf("ben's balance %d %v", balance, err)
	}
}

func TestAuditFindsTamperedBalancesAndReconcileRestoresThem(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := wallet.NewStore(pool)
	ctx := context.Background()
	ayla, ben := account(t, pool), account(t, pool)
	for _, id := range []string{ayla, ben} {
		if _, err := store.Apply(ctx, wallet.Entry{Account: id, Amount: 300, Reason: wallet.DailyLogin, Source: wallet.Source{Type: "daily", Ref: "d1"}, Key: "daily_login:1"}, start); err != nil {
			t.Fatal(err)
		}
	}
	if mismatches, err := store.Audit(ctx); err != nil || len(mismatches) != 0 {
		t.Fatalf("audit of consistent wallets %+v %v", mismatches, err)
	}
	if _, err := pool.Exec(ctx, `update wallets set balance = 9999 where account_id = $1`, ben); err != nil {
		t.Fatal(err)
	}
	mismatches, err := store.Audit(ctx)
	if err != nil || len(mismatches) != 1 || mismatches[0] != (wallet.Mismatch{Account: ben, Stored: 9999, Computed: 300}) {
		t.Fatalf("audit %+v %v", mismatches, err)
	}
	if m, fixed, err := store.Reconcile(ctx, ayla, start); err != nil || fixed || m.Stored != 300 {
		t.Fatalf("reconcile of a consistent wallet %+v %v %v", m, fixed, err)
	}
	if m, fixed, err := store.Reconcile(ctx, ben, start); err != nil || !fixed || m.Stored != 9999 || m.Computed != 300 {
		t.Fatalf("reconcile %+v %v %v", m, fixed, err)
	}
	if balance, err := store.Balance(ctx, ben); err != nil || balance != 300 {
		t.Fatalf("balance after reconcile %d %v", balance, err)
	}
}
