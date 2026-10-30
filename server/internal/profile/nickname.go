package profile

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	ErrInvalidLength     = errors.New("nickname.invalid_length")
	ErrInvalidCharacters = errors.New("nickname.invalid_characters")
	ErrInvalidSeparators = errors.New("nickname.invalid_separators")
	ErrMixedScripts      = errors.New("nickname.mixed_scripts")
	ErrDigitsOnly        = errors.New("nickname.digits_only")
	ErrTaken             = errors.New("nickname.taken")
	ErrCooldown          = errors.New("nickname.cooldown")
)

type Config struct {
	NicknameMinLength        int
	NicknameMaxLength        int
	NicknameAllowMixedScript bool
	NicknameRenameCooldown   time.Duration
	NicknameRelease          time.Duration
}

func DefaultConfig() Config {
	return Config{
		NicknameMinLength:      3,
		NicknameMaxLength:      16,
		NicknameRenameCooldown: 30 * 24 * time.Hour,
		NicknameRelease:        30 * 24 * time.Hour,
	}
}

const separators = "_.- "

var confusables = map[rune]rune{
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x', 'і': 'i',
	'ә': 'a', 'ə': 'a', 'ı': 'i', 'ö': 'o', 'ü': 'u', 'ş': 's', 'ç': 'c', 'ğ': 'g',
}

func Display(raw string) string {
	var b strings.Builder
	space := false
	for _, r := range norm.NFKC.String(raw) {
		switch {
		case unicode.Is(unicode.Cf, r):
			continue
		case unicode.IsSpace(r):
			space = true
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	return b.String()
}

func Key(display string) string {
	folded := strings.ReplaceAll(strings.ToLower(strings.ReplaceAll(display, "İ", "i")), "i̇", "i")
	var b strings.Builder
	for _, r := range folded {
		if canonical, ok := confusables[r]; ok {
			r = canonical
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (c Config) Validate(display string) error {
	if display == "" {
		return ErrInvalidLength
	}
	length, digits, letters := 0, 0, 0
	marks := 0
	run, last := 0, rune(0)
	hasLatin, hasCyrillic := false, false
	for _, r := range display {
		switch {
		case unicode.Is(unicode.Mn, r):
			marks++
			if marks > 1 || length == 0 {
				return ErrInvalidCharacters
			}
			continue
		case r >= '0' && r <= '9':
			digits++
		case unicode.Is(unicode.Latin, r) && unicode.IsLetter(r):
			hasLatin = true
			letters++
		case unicode.Is(unicode.Cyrillic, r) && unicode.IsLetter(r):
			hasCyrillic = true
			letters++
		case strings.ContainsRune(separators, r):
			if length == 0 {
				return ErrInvalidSeparators
			}
			if !c.NicknameAllowMixedScript && hasLatin && hasCyrillic {
				return ErrMixedScripts
			}
			hasLatin, hasCyrillic = false, false
		default:
			return ErrInvalidCharacters
		}
		marks = 0
		if r == last && strings.ContainsRune(separators, r) {
			run++
			if run > 2 {
				return ErrInvalidSeparators
			}
		} else {
			run = 1
		}
		last = r
		length++
	}
	switch {
	case strings.ContainsRune(separators, last):
		return ErrInvalidSeparators
	case !c.NicknameAllowMixedScript && hasLatin && hasCyrillic:
		return ErrMixedScripts
	case length < c.NicknameMinLength || length > c.NicknameMaxLength:
		return ErrInvalidLength
	case letters == 0 && digits > 0:
		return ErrDigitsOnly
	}
	return nil
}
