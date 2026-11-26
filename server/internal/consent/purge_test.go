package consent_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/consent"
)

func TestRetentionCountsFromTheEndOfTheDeletionYear(t *testing.T) {
	tests := []struct {
		deleted time.Time
		now     time.Time
		years   int
		purged  bool
	}{
		{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC), time.Date(2029, 12, 31, 23, 59, 59, 0, time.UTC), 3, false},
		{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC), time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), 3, true},
		{time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), 3, true},
		{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC), 3, false},
		{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC), 3, true},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC), 1, false},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC), 1, true},
		{time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2029, 12, 31, 23, 0, 0, 0, time.FixedZone("UTC-3", -3*3600)), 3, true},
	}
	for _, test := range tests {
		if got := test.deleted.Before(consent.RetentionCutoff(test.now, test.years)); got != test.purged {
			t.Fatalf("deleted %v, now %v, %d years: purged %t, want %t", test.deleted, test.now, test.years, got, test.purged)
		}
	}
}

func TestPurgeWaitsForThePrivacyConfig(t *testing.T) {
	purge := consent.NewPurge(nil, func() (int, bool) { return 3, false }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	purged, err := purge.Once(context.Background(), now)
	if purged != 0 || err != nil {
		t.Fatalf("purged %d %v before the privacy config was loaded", purged, err)
	}
	if got := exposed(t, purge); got != `hd_consent_purge_runs_total{result="waiting"} 1` {
		t.Fatalf("metrics %s", got)
	}
}

func TestPurgeDeletesOnlyExpiredRecordsOfDeletedAccounts(t *testing.T) {
	w := newWorld(t)
	runAt := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	expired := w.tombstone(time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), 501)
	kept := w.tombstone(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), 2)
	active := w.player()
	w.insertRecords(active, 3)
	purge := consent.NewPurge(w.store, func() (int, bool) { return 3, true }, slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, want := range []int64{501, 0} {
		purged, err := purge.Once(context.Background(), runAt)
		if err != nil || purged != want {
			t.Fatalf("purged %d %v, want %d", purged, err, want)
		}
	}
	for account, want := range map[string]int{expired: 0, kept: 2, active: 3} {
		if got := w.count(account); got != want {
			t.Fatalf("%s keeps %d records, want %d", account, got, want)
		}
	}
	if got := exposed(t, purge); got != "hd_consent_purge_runs_total{result=\"ok\"} 2\nhd_consent_records_purged_total 501" {
		t.Fatalf("metrics %s", got)
	}
}

func TestPurgeFollowsTheConfiguredRetention(t *testing.T) {
	w := newWorld(t)
	tombstone := w.tombstone(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 1)
	years := 5
	purge := consent.NewPurge(w.store, func() (int, bool) { return years, true }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	runAt := time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC)

	if purged, err := purge.Once(context.Background(), runAt); err != nil || purged != 0 {
		t.Fatalf("five years: purged %d %v", purged, err)
	}
	years = 3
	if purged, err := purge.Once(context.Background(), runAt); err != nil || purged != 1 || w.count(tombstone) != 0 {
		t.Fatalf("three years: purged %d %v", purged, err)
	}
}

func (w *world) tombstone(deletedAt time.Time, records int) string {
	w.t.Helper()
	var id string
	if err := w.pool.QueryRow(context.Background(), `insert into accounts (state, created_at, deleted_at) values ('deleted', $1, $1) returning id::text`, deletedAt).Scan(&id); err != nil {
		w.t.Fatal(err)
	}
	w.insertRecords(id, records)
	return id
}

func (w *world) insertRecords(account string, records int) {
	w.t.Helper()
	if _, err := w.pool.Exec(context.Background(), `insert into consent_records (account_id, purpose, granted, policy_version, source, app_version, recorded_at)
		select $1::uuid, 'analytics', true, '2026-11', 'settings', '1.4.0', $2::timestamptz from generate_series(1, $3::int)`, account, now, records); err != nil {
		w.t.Fatal(err)
	}
}

func (w *world) count(account string) int {
	w.t.Helper()
	var count int
	if err := w.pool.QueryRow(context.Background(), `select count(*) from consent_records where account_id = $1`, account).Scan(&count); err != nil {
		w.t.Fatal(err)
	}
	return count
}

func exposed(t *testing.T, purge *consent.Purge) string {
	t.Helper()
	var out strings.Builder
	purge.Runs().WriteTo(&out)
	purge.Purged().WriteTo(&out)
	var series []string
	for _, line := range strings.Split(out.String(), "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			series = append(series, line)
		}
	}
	return strings.Join(series, "\n")
}
