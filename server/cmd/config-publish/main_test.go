package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func TestPublishValidatesAndStoresTheSnapshot(t *testing.T) {
	url, pool := dbtest.Fresh(t)
	ctx := context.Background()
	var out bytes.Buffer
	if err := run(ctx, url, shop.Domain, "../../../docs/economy/examples/shop.json", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "shop version ") {
		t.Fatalf("output %q", out.String())
	}
	var count int
	if err := pool.QueryRow(ctx, `select count(*) from config_snapshots where domain = 'shop'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("stored snapshots %d %v", count, err)
	}
	broken := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(broken, []byte(`{"schema":1,"offers":[{"id":"x"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(ctx, url, shop.Domain, broken, &out); !errors.Is(err, shop.ErrInvalid) {
		t.Fatalf("invalid catalog: %v", err)
	}
	if err := run(ctx, url, "weather", "../../../docs/economy/examples/shop.json", &out); !errors.Is(err, errDomain) {
		t.Fatalf("unknown domain: %v", err)
	}
	if err := run(ctx, url, shop.Domain, filepath.Join(t.TempDir(), "missing.json"), &out); err == nil {
		t.Fatal("missing file published")
	}
	if err := pool.QueryRow(ctx, `select count(*) from config_snapshots`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("snapshots after refusals %d %v", count, err)
	}
}

func TestRulesSnapshotsAreCheckedBeforeTheyArePublished(t *testing.T) {
	url, pool := dbtest.Fresh(t)
	ctx := context.Background()
	write := func(content string) string {
		path := filepath.Join(t.TempDir(), "rules.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	var out bytes.Buffer
	if err := run(ctx, url, rulesconfig.Domain, write(`{"TurnSeconds": 61}`), &out); !errors.Is(err, rulesconfig.ErrInvalid) {
		t.Fatalf("out of bounds turn time: %v", err)
	}
	if err := run(ctx, url, rulesconfig.Domain, write(`{"TurnSeconds": 25}`), &out); err != nil {
		t.Fatal(err)
	}
	watcher := snapshot.New(pool, rulesconfig.Domain, rulesconfig.Parse, slog.New(slog.DiscardHandler))
	if _, err := watcher.Load(ctx); err != nil {
		t.Fatal(err)
	}
	tuning, _, _ := watcher.Current()
	if got := tuning.Apply(rules.DefaultConfig()).TurnSeconds; got != 25 {
		t.Fatalf("published turn time %d", got)
	}
}

func TestACosmeticsManifestMustFollowThePublishedOne(t *testing.T) {
	url, _ := dbtest.Fresh(t)
	ctx := context.Background()
	raw, err := os.ReadFile("../../../docs/cosmetics/examples/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	first, err := cosmetics.ParseManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	write := func(m cosmetics.Manifest) string {
		t.Helper()
		encoded, err := m.Encode()
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "manifest.json")
		if err := os.WriteFile(path, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	var out bytes.Buffer
	if err := run(ctx, url, cosmetics.Domain, write(first), &out); err != nil {
		t.Fatalf("first manifest: %v", err)
	}
	if err := run(ctx, url, cosmetics.Domain, write(first), &out); !errors.Is(err, cosmetics.ErrSuccessor) {
		t.Fatalf("the same manifestVersion again: %v", err)
	}
	dropped := first
	dropped.ManifestVersion++
	dropped.Items = first.Items[:len(first.Items)-1]
	if err := run(ctx, url, cosmetics.Domain, write(dropped), &out); !errors.Is(err, cosmetics.ErrSuccessor) {
		t.Fatalf("a manifest without an earlier item: %v", err)
	}
	next := first
	next.ManifestVersion++
	if err := run(ctx, url, cosmetics.Domain, write(next), &out); err != nil {
		t.Fatalf("the next manifest: %v", err)
	}
	if err := run(ctx, url, cosmetics.Domain, write(cosmetics.Manifest{SchemaVersion: 1, ManifestVersion: 9}), &out); !errors.Is(err, cosmetics.ErrInvalid) {
		t.Fatalf("an invalid manifest: %v", err)
	}
}

func TestALocalizationSnapshotNeedsPublishedBundlesAndRollsBackByVersion(t *testing.T) {
	url, pool := dbtest.Fresh(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	write := func(content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "localization.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	var out bytes.Buffer
	active := write(`{"bundles":{"en":1,"ru":2}}`)
	if err := run(ctx, url, gamestrings.Domain, active, &out); !errors.Is(err, gamestrings.ErrNoBundle) {
		t.Fatalf("unpublished bundles: %v", err)
	}
	if err := run(ctx, url, gamestrings.Domain, write(`{"bundles":{"ru":1}}`), &out); !errors.Is(err, gamestrings.ErrInvalidLocalization) {
		t.Fatalf("no en bundle: %v", err)
	}
	store := gamestrings.NewStore(pool)
	if err := store.SaveEntry(ctx, gamestrings.Entry{Key: "match.result.win", Note: "Match end title.", MaxLength: 12}, "content-1", now); err != nil {
		t.Fatal(err)
	}
	for _, value := range []struct {
		language gamestrings.Language
		text     string
	}{{gamestrings.English, "Victory"}, {gamestrings.Russian, "Победа"}, {gamestrings.Russian, "Победа!"}} {
		if err := store.SetValue(ctx, "match.result.win", value.language, gamestrings.Value{Text: value.text, Status: gamestrings.Approved}, "content-1", now); err != nil {
			t.Fatal(err)
		}
		if _, problems, err := store.Publish(ctx, value.language, "content-1", now); err != nil {
			t.Fatalf("publish %s: %v %v", value.language, problems, err)
		}
	}
	if err := run(ctx, url, gamestrings.Domain, active, &out); err != nil {
		t.Fatal(err)
	}
	if err := run(ctx, url, gamestrings.Domain, write(`{"bundles":{"en":1,"ru":1}}`), &out); err != nil {
		t.Fatalf("rollback to ru 1: %v", err)
	}
	watcher := snapshot.New(pool, gamestrings.Domain, gamestrings.ParseLocalization, slog.New(slog.DiscardHandler))
	if _, err := watcher.Load(ctx); err != nil {
		t.Fatal(err)
	}
	localization, version, _ := watcher.Current()
	if localization.Bundles[gamestrings.Russian] != 1 || version != 2 {
		t.Fatalf("active %+v at version %d", localization, version)
	}
}
