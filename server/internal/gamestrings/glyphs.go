package gamestrings

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

var ErrUndrawable = errors.New("gamestrings: characters the client fonts do not draw")

//go:embed font-characters.json
var fontCharactersJSON []byte

var drawable = mustDrawable(fontCharactersJSON)

func mustDrawable(data []byte) map[rune]bool {
	var set struct {
		Characters string `json:"characters"`
	}
	if err := json.Unmarshal(data, &set); err != nil {
		panic(err)
	}
	if set.Characters == "" {
		panic("gamestrings: font-characters.json lists no characters")
	}
	drawn := map[rune]bool{}
	for _, r := range set.Characters {
		drawn[r] = true
	}
	return drawn
}

func UndrawableCharacters(text string) []string {
	var found []string
	for _, r := range text {
		character := string(r)
		if !drawable[r] && !slices.Contains(found, character) {
			found = append(found, character)
		}
	}
	return found
}

func (b Bundle) UndrawableProblems() []Problem {
	var problems []Problem
	for _, key := range slices.Sorted(maps.Keys(b.Strings)) {
		characters := UndrawableCharacters(b.Strings[key])
		if characters == nil {
			continue
		}
		problems = append(problems, Problem{Key: key, Characters: characters, Err: fmt.Errorf("%w: %s", ErrUndrawable, codePointList(characters))})
	}
	return problems
}

func codePointList(characters []string) string {
	labels := make([]string, 0, len(characters))
	for _, character := range characters {
		labels = append(labels, fmt.Sprintf("%q U+%04X", character, []rune(character)[0]))
	}
	return strings.Join(labels, ", ")
}
