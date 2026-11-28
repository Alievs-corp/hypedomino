package gamestrings

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestTheFontsDrawEveryLanguageAlphabetButNotOtherLetters(t *testing.T) {
	cases := map[string][]string{
		"Victory!": nil,
		"Qələbə! Ğğ İı Şş Çç Öö Üü":              nil,
		"Kazandın — Âlâ Îmân Ûmît":               nil,
		"Победа! Ёлка №1 «ура» …":                nil,
		"{coins, plural, one {# ₼} other {# ₼}}": nil,
		"Café crème":      {"é", "è"},
		"Straße ß":        {"ß"},
		"Київ":            {"ї"},
		"cafe\u0301":      {"\u0301"},
		"two\nlines":      {"\n"},
		"zero\u200bwidth": {"\u200b"},
		"win 🎉🎉":          {"🎉"},
	}
	for text, want := range cases {
		if got := UndrawableCharacters(text); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: %q, want %q", text, got, want)
		}
	}
}

func TestEveryApprovedBuiltInDefaultStaysDrawable(t *testing.T) {
	for _, language := range Languages {
		bundle, problems := BuildBundle(catalog(), language)
		if problems != nil {
			t.Fatalf("%s: %v", language, problems)
		}
		if undrawable := bundle.UndrawableProblems(); undrawable != nil {
			t.Errorf("%s: %+v", language, undrawable)
		}
	}
}

func TestUndrawableProblemsNameTheKeyAndTheCharacters(t *testing.T) {
	bundle := Bundle{Language: Turkish, Strings: map[string]string{
		"match.result.win":  "Kazandın!",
		"match.result.lose": "Kaybettín ñ í",
		"lobby.title":       "Café",
	}}
	problems := bundle.UndrawableProblems()
	if len(problems) != 2 || problems[0].Key != "lobby.title" || problems[1].Key != "match.result.lose" {
		t.Fatalf("problems %+v", problems)
	}
	if !reflect.DeepEqual(problems[1].Characters, []string{"í", "ñ"}) || !errors.Is(problems[1].Err, ErrUndrawable) {
		t.Fatalf("lose %+v", problems[1])
	}
	if !strings.Contains(problems[1].Err.Error(), `"í" U+00ED, "ñ" U+00F1`) {
		t.Fatalf("message %v", problems[1].Err)
	}
}
