package profile

import (
	"errors"
	"strings"
	"testing"
)

func TestDisplayNormalisesAndStripsInvisibleCharacters(t *testing.T) {
	cases := map[string]string{
		"  Ayşe  Han ":         "Ayşe Han",
		"Zero​Width":           "ZeroWidth",
		"‮Reversed":            "Reversed",
		"ﬁne":                  "fine",
		"Ａｙｌａ":                 "Ayla",
		"tab\tand\nnewline":    "tab and newline",
		"Café":                "Café",
		"​​":                   "",
		"İlkin":                "İlkin",
		"À́̂bc":               "À́̂bc",
		strings.Repeat(" ", 3): "",
	}
	for raw, want := range cases {
		if got := Display(raw); got != want {
			t.Fatalf("Display(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestKeyFoldsCaseTurkishLettersAndConfusables(t *testing.T) {
	cases := map[string]string{
		"Ayşe":      "ayse",
		"AYSE":      "ayse",
		"İlkin":     "ilkin",
		"Ilkin":     "ilkin",
		"ılkın":     "ilkin",
		"i̇lk":      "ilk",
		"Сatherine": "catherine",
		"рrо":       "pro",
		"Günəş":     "gunas",
		"çağ_1":     "cag_1",
		"L33T":      "l33t",
	}
	for display, want := range cases {
		if got := Key(display); got != want {
			t.Fatalf("Key(%q) = %q, want %q", display, got, want)
		}
	}
	if Key("Ayşe") != Key("AYSE") || Key("Ayşe") != Key("ayse") {
		t.Fatal("Ayşe and Ayse must collide")
	}
}

func TestValidateAppliesTheFormatRules(t *testing.T) {
	config := DefaultConfig()
	cases := map[string]error{
		"Ayla":                  nil,
		"ayla_07":               nil,
		"Ayşe Han":              nil,
		"Günəş-1":               nil,
		"Иван":                  nil,
		"Иван Smith":            nil,
		"a.b":                   nil,
		"a__b":                  nil,
		"Café":                  nil,
		"ab":                    ErrInvalidLength,
		"":                      ErrInvalidLength,
		strings.Repeat("a", 17): ErrInvalidLength,
		strings.Repeat("a", 16): nil,
		"Ayla😀":                 ErrInvalidCharacters,
		"Ayla!":                 ErrInvalidCharacters,
		"Ayla@home":             ErrInvalidCharacters,
		"中文名字":                  ErrInvalidCharacters,
		"à́b":                  ErrInvalidCharacters,
		"́abc":                  ErrInvalidCharacters,
		"_Ayla":                 ErrInvalidSeparators,
		"Ayla.":                 ErrInvalidSeparators,
		"a___b":                 ErrInvalidSeparators,
		"a...b":                 ErrInvalidSeparators,
		"Иvan":                  ErrMixedScripts,
		"Ivаn":                  ErrMixedScripts,
		"12345":                 ErrDigitsOnly,
		"12_34":                 ErrDigitsOnly,
		"١٢٣٤":                  ErrInvalidCharacters,
	}
	for display, want := range cases {
		if got := config.Validate(display); !errors.Is(got, want) {
			t.Fatalf("Validate(%q) = %v, want %v", display, got, want)
		}
	}
	config.NicknameAllowMixedScript = true
	if err := config.Validate("Иvan"); err != nil {
		t.Fatalf("mixed scripts allowed by config: %v", err)
	}
	config.NicknameMinLength, config.NicknameMaxLength = 2, 4
	if config.Validate("ab") != nil || !errors.Is(config.Validate("abcde"), ErrInvalidLength) {
		t.Fatal("configured lengths were not applied")
	}
}
