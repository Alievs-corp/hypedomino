package gamestrings_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/db/dbtest"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
)

var now = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Main(m))
}

func newStore(t *testing.T) *gamestrings.Store {
	t.Helper()
	_, pool := dbtest.Fresh(t)
	return gamestrings.NewStore(pool)
}

func coins() gamestrings.Entry {
	return gamestrings.Entry{
		Key: "match.result.coins", Note: "Match end screen, coins is the reward.", MaxLength: 24,
		Placeholders: []gamestrings.Placeholder{{Name: "coins", Type: gamestrings.Integer}},
	}
}

func win() gamestrings.Entry {
	return gamestrings.Entry{Key: "match.result.win", Note: "Match end title.", MaxLength: 12}
}

func save(t *testing.T, store *gamestrings.Store, entries ...gamestrings.Entry) {
	t.Helper()
	for _, entry := range entries {
		if err := store.SaveEntry(context.Background(), entry, "content-1", now); err != nil {
			t.Fatalf("save %s: %v", entry.Key, err)
		}
	}
}

func set(t *testing.T, store *gamestrings.Store, key string, language gamestrings.Language, text string, status gamestrings.Status) {
	t.Helper()
	if err := store.SetValue(context.Background(), key, language, gamestrings.Value{Text: text, Status: status}, "content-1", now); err != nil {
		t.Fatalf("set %s %s: %v", key, language, err)
	}
}

func TestEntriesReadBackWithTheirValuesInKeyOrder(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	retired := gamestrings.Entry{Key: "lobby.banner.old", Note: "Retired banner.", MaxLength: 30, Deprecated: true}
	save(t, store, win(), coins(), retired)
	set(t, store, "match.result.coins", gamestrings.English, "{coins, plural, one {# coin} other {# coins}}", gamestrings.Approved)
	set(t, store, "match.result.coins", gamestrings.Russian, "{coins} монет", gamestrings.Draft)
	entries, err := store.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantCoins := coins()
	wantCoins.Values = map[gamestrings.Language]gamestrings.Value{
		gamestrings.English: {Text: "{coins, plural, one {# coin} other {# coins}}", Status: gamestrings.Approved},
		gamestrings.Russian: {Text: "{coins} монет", Status: gamestrings.Draft},
	}
	wantWin := win()
	wantWin.Placeholders = []gamestrings.Placeholder{}
	wantWin.Values = map[gamestrings.Language]gamestrings.Value{}
	retired.Placeholders = []gamestrings.Placeholder{}
	retired.Values = map[gamestrings.Language]gamestrings.Value{}
	if want := []gamestrings.Entry{retired, wantCoins, wantWin}; !reflect.DeepEqual(entries, want) {
		t.Fatalf("entries\n%+v\nwant\n%+v", entries, want)
	}
}

func TestSavingAnEntryAgainUpdatesItAndKeepsItsValues(t *testing.T) {
	store := newStore(t)
	save(t, store, win())
	set(t, store, "match.result.win", gamestrings.English, "Victory!", gamestrings.Approved)
	changed := win()
	changed.Note, changed.MaxLength, changed.Deprecated = "Match end title, above the score.", 16, true
	save(t, store, changed)
	entries, err := store.Entries(context.Background())
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries %+v %v", entries, err)
	}
	got := entries[0]
	if got.Note != changed.Note || got.MaxLength != 16 || !got.Deprecated || got.Values[gamestrings.English].Text != "Victory!" {
		t.Fatalf("entry %+v", got)
	}
}

func TestAnInvalidEntryIsNotSaved(t *testing.T) {
	store := newStore(t)
	bad := win()
	bad.MaxLength = 0
	if err := store.SaveEntry(context.Background(), bad, "content-1", now); !errors.Is(err, gamestrings.ErrInvalidEntry) {
		t.Fatalf("saved: %v", err)
	}
	if entries, err := store.Entries(context.Background()); err != nil || len(entries) != 0 {
		t.Fatalf("entries %+v %v", entries, err)
	}
}

func TestOnlyACheckedValueCanBeApproved(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, coins())
	value := func(text string, status gamestrings.Status) gamestrings.Value {
		return gamestrings.Value{Text: text, Status: status}
	}
	tests := []struct {
		name     string
		key      string
		language gamestrings.Language
		value    gamestrings.Value
		want     error
	}{
		{"unknown key", "match.result.lose", gamestrings.English, value("Defeat", gamestrings.Draft), gamestrings.ErrUnknownKey},
		{"unknown language", "match.result.coins", "de", value("{coins} Münzen", gamestrings.Draft), gamestrings.ErrInvalidValue},
		{"missing is not a value", "match.result.coins", gamestrings.English, value("", gamestrings.Missing), gamestrings.ErrInvalidValue},
		{"approved with a wrong placeholder", "match.result.coins", gamestrings.English, value("{coin} coins", gamestrings.Approved), gamestrings.ErrInvalidValue},
		{"approved ru without few and many", "match.result.coins", gamestrings.Russian, value("{coins, plural, one {# монета} other {# монеты}}", gamestrings.Approved), gamestrings.ErrInvalidValue},
	}
	for _, test := range tests {
		if err := store.SetValue(ctx, test.key, test.language, test.value, "content-1", now); !errors.Is(err, test.want) {
			t.Errorf("%s: %v", test.name, err)
		}
	}
	entries, err := store.Entries(ctx)
	if err != nil || len(entries[0].Values) != 0 {
		t.Fatalf("a rejected value was stored: %+v %v", entries, err)
	}
	set(t, store, "match.result.coins", gamestrings.English, "{coin} coins", gamestrings.Draft)
	if entries, _ := store.Entries(ctx); entries[0].Values[gamestrings.English].Status != gamestrings.Draft {
		t.Fatalf("a draft that fails its checks is kept for editing: %+v", entries)
	}
}

func TestPublishingStoresANumberedBundlePerLanguage(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, win(), coins())
	set(t, store, "match.result.win", gamestrings.English, "Victory!", gamestrings.Approved)
	set(t, store, "match.result.win", gamestrings.Azerbaijani, "Qələbə!", gamestrings.Approved)
	first, problems, err := store.Publish(ctx, gamestrings.English, "content-1", now)
	if err != nil || problems != nil || first.Version != 1 || first.By != "content-1" || !first.At.Equal(now) {
		t.Fatalf("first %+v %v %v", first, problems, err)
	}
	want, hash, _ := gamestrings.Bundle{Language: gamestrings.English, Strings: map[string]string{"match.result.win": "Victory!"}}.Encode()
	if string(first.Content) != string(want) || first.Hash != hash {
		t.Fatalf("content %s %s", first.Content, first.Hash)
	}
	again, _, err := store.Publish(ctx, gamestrings.English, "content-2", now.Add(time.Hour))
	if err != nil || !reflect.DeepEqual(again, first) {
		t.Fatalf("unchanged content published again as %+v %v", again, err)
	}
	set(t, store, "match.result.coins", gamestrings.English, "{coins, plural, one {# coin} other {# coins}}", gamestrings.Approved)
	second, _, err := store.Publish(ctx, gamestrings.English, "content-1", now.Add(time.Hour))
	if err != nil || second.Version != 2 || second.Hash == first.Hash {
		t.Fatalf("second %+v %v", second, err)
	}
	azerbaijani, _, err := store.Publish(ctx, gamestrings.Azerbaijani, "content-1", now)
	if err != nil || azerbaijani.Version != 1 {
		t.Fatalf("az %+v %v", azerbaijani, err)
	}
	earlier, err := store.Bundle(ctx, gamestrings.English, 1)
	if err != nil || !reflect.DeepEqual(earlier, first) {
		t.Fatalf("version 1 %+v %v", earlier, err)
	}
	if _, err := store.Bundle(ctx, gamestrings.English, 3); !errors.Is(err, gamestrings.ErrNoBundle) {
		t.Fatalf("version 3: %v", err)
	}
}

func TestAnInvalidApprovedValueBlocksPublishing(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, win(), coins())
	set(t, store, "match.result.win", gamestrings.English, "Victory!", gamestrings.Approved)
	set(t, store, "match.result.coins", gamestrings.English, "{coins, plural, one {# coin} other {# coins}}", gamestrings.Approved)
	renamed := coins()
	renamed.Placeholders = []gamestrings.Placeholder{{Name: "amount", Type: gamestrings.Integer}}
	save(t, store, renamed)
	_, problems, err := store.Publish(ctx, gamestrings.English, "content-1", now)
	if !errors.Is(err, gamestrings.ErrBlocked) || len(problems) != 1 || problems[0].Key != "match.result.coins" {
		t.Fatalf("published over a stale placeholder: %v %v", problems, err)
	}
	if _, err := store.Bundle(ctx, gamestrings.English, 1); !errors.Is(err, gamestrings.ErrNoBundle) {
		t.Fatalf("a blocked bundle was stored: %v", err)
	}
}

func TestPublishedBundlesCannotBeChanged(t *testing.T) {
	_, pool := dbtest.Fresh(t)
	store := gamestrings.NewStore(pool)
	ctx := context.Background()
	save(t, store, win())
	set(t, store, "match.result.win", gamestrings.English, "Victory!", gamestrings.Approved)
	if _, _, err := store.Publish(ctx, gamestrings.English, "content-1", now); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`update game_string_bundles set content = 'x'`,
		`delete from game_string_bundles`,
		`truncate game_string_bundles`,
	} {
		if _, err := pool.Exec(ctx, statement); err == nil {
			t.Errorf("%s succeeded", statement)
		}
	}
}

func TestAnImportStoresChangedTextsAsDraftsForReview(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, win(), coins())
	set(t, store, "match.result.win", gamestrings.Russian, "Победа", gamestrings.Approved)
	set(t, store, "match.result.coins", gamestrings.Russian, "{coins} монета", gamestrings.Approved)
	file := "key,ru\nmatch.result.win,Победа\nmatch.result.coins,\"{coins, plural, one {# монета} few {# монеты} many {# монет} other {# монеты}}\"\n"
	changes, problems, err := store.Import(ctx, gamestrings.Russian, gamestrings.CSV, []byte(file), "translator-ru", now)
	if err != nil || problems != nil || len(changes) != 1 || changes[0].Key != "match.result.coins" {
		t.Fatalf("changes %+v problems %+v err %v", changes, problems, err)
	}
	entries, err := store.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]gamestrings.Value{
		"match.result.coins": {Text: "{coins, plural, one {# монета} few {# монеты} many {# монет} other {# монеты}}", Status: gamestrings.Draft},
		"match.result.win":   {Text: "Победа", Status: gamestrings.Approved},
	}
	for _, entry := range entries {
		if got := entry.Values[gamestrings.Russian]; got != want[entry.Key] {
			t.Errorf("%s: %+v, want %+v", entry.Key, got, want[entry.Key])
		}
	}
}

func TestABlockedImportStoresNothing(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, win(), coins())
	file := `{"language":"ru","rows":[{"key":"match.result.win","text":"Победа"},{"key":"match.result.coins","text":"{coin} монет"}]}`
	changes, problems, err := store.Import(ctx, gamestrings.Russian, gamestrings.JSON, []byte(file), "translator-ru", now)
	if !errors.Is(err, gamestrings.ErrImportBlocked) || changes != nil || len(problems) != 1 || problems[0].Key != "match.result.coins" {
		t.Fatalf("changes %+v problems %+v err %v", changes, problems, err)
	}
	entries, err := store.Entries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if len(entry.Values) != 0 {
			t.Fatalf("%s was stored: %+v", entry.Key, entry.Values)
		}
	}
}

func TestAnExportImportsBackWithoutChanges(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	save(t, store, win(), coins())
	set(t, store, "match.result.win", gamestrings.Turkish, "Zafer", gamestrings.Approved)
	for _, format := range []gamestrings.Format{gamestrings.CSV, gamestrings.JSON} {
		data, err := store.Export(ctx, gamestrings.Turkish, format)
		if err != nil {
			t.Fatal(err)
		}
		changes, problems, err := store.Import(ctx, gamestrings.Turkish, format, data, "translator-tr", now)
		if err != nil || problems != nil || changes != nil {
			t.Errorf("%s: changes %+v problems %+v err %v", format, changes, problems, err)
		}
	}
}
