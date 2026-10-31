package gamestrings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const Domain = "localization"

var ErrInvalidLocalization = errors.New("gamestrings: invalid localization snapshot")

type Localization struct {
	Bundles        map[Language]int        `json:"bundles"`
	FallbackChains map[Language][]Language `json:"fallbackChains"`
}

func DefaultFallbackChains() map[Language][]Language {
	return map[Language][]Language{
		Azerbaijani: {Azerbaijani, Turkish, English},
		English:     {English},
		Turkish:     {Turkish, English},
		Russian:     {Russian, English},
	}
}

func ParseLocalization(raw []byte) (Localization, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var localization Localization
	if err := decoder.Decode(&localization); err != nil {
		return Localization{}, fmt.Errorf("%w: %w", ErrInvalidLocalization, err)
	}
	if err := checkBundles(localization.Bundles); err != nil {
		return Localization{}, err
	}
	chains := DefaultFallbackChains()
	for language, chain := range localization.FallbackChains {
		if err := checkChain(language, chain); err != nil {
			return Localization{}, err
		}
		chains[language] = chain
	}
	localization.FallbackChains = chains
	return localization, nil
}

func checkBundles(bundles map[Language]int) error {
	if _, has := bundles[English]; !has {
		return fmt.Errorf("%w: no en bundle", ErrInvalidLocalization)
	}
	for language, version := range bundles {
		if !slices.Contains(Languages, language) {
			return fmt.Errorf("%w: unknown language %q", ErrInvalidLocalization, language)
		}
		if version <= 0 {
			return fmt.Errorf("%w: %s bundle version %d", ErrInvalidLocalization, language, version)
		}
	}
	return nil
}

func checkChain(language Language, chain []Language) error {
	switch {
	case !slices.Contains(Languages, language):
		return fmt.Errorf("%w: chain for unknown language %q", ErrInvalidLocalization, language)
	case len(chain) == 0 || chain[0] != language:
		return fmt.Errorf("%w: %s chain must start with %s", ErrInvalidLocalization, language, language)
	case chain[len(chain)-1] != English:
		return fmt.Errorf("%w: %s chain must end with en", ErrInvalidLocalization, language)
	}
	for i, step := range chain {
		if !slices.Contains(Languages, step) || slices.Contains(chain[:i], step) {
			return fmt.Errorf("%w: %s chain %v", ErrInvalidLocalization, language, chain)
		}
	}
	return nil
}

func (s *Store) CheckActive(ctx context.Context, localization Localization) error {
	for language, version := range localization.Bundles {
		if _, err := s.Bundle(ctx, language, version); err != nil {
			return fmt.Errorf("%w: %s version %d: %w", ErrInvalidLocalization, language, version, err)
		}
	}
	english, err := s.Bundle(ctx, English, localization.Bundles[English])
	if err != nil {
		return err
	}
	var bundle Bundle
	if err := json.Unmarshal(english.Content, &bundle); err != nil {
		return err
	}
	entries, err := s.Entries(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, has := bundle.Strings[entry.Key]; !entry.Deprecated && !has {
			return fmt.Errorf("%w: en version %d lacks %s", ErrInvalidLocalization, english.Version, entry.Key)
		}
	}
	return nil
}
