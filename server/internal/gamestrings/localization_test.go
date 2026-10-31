package gamestrings

import (
	"errors"
	"reflect"
	"testing"
)

func TestALocalizationSnapshotKeepsTheDefaultChainsUnlessItSetsThem(t *testing.T) {
	localization, err := ParseLocalization([]byte(`{"bundles":{"en":3,"az":2},"fallbackChains":{"az":["az","en"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	wantChains := DefaultFallbackChains()
	wantChains[Azerbaijani] = []Language{Azerbaijani, English}
	want := Localization{Bundles: map[Language]int{English: 3, Azerbaijani: 2}, FallbackChains: wantChains}
	if !reflect.DeepEqual(localization, want) {
		t.Fatalf("localization %+v, want %+v", localization, want)
	}
	if localization, err := ParseLocalization([]byte(`{"bundles":{"en":1}}`)); err != nil || !reflect.DeepEqual(localization.FallbackChains, DefaultFallbackChains()) {
		t.Fatalf("defaults %+v %v", localization, err)
	}
}

func TestTheDefaultChainsEndInEnglish(t *testing.T) {
	for language, chain := range DefaultFallbackChains() {
		if err := checkChain(language, chain); err != nil {
			t.Error(err)
		}
	}
}

func TestALocalizationSnapshotIsRefusedWhenItBreaksTheDeliveryRules(t *testing.T) {
	tests := map[string]string{
		"not json":                `{"bundles":`,
		"unknown field":           `{"bundles":{"en":1},"chains":{}}`,
		"no en bundle":            `{"bundles":{"az":1}}`,
		"unknown bundle language": `{"bundles":{"en":1,"de":1}}`,
		"zero version":            `{"bundles":{"en":1,"ru":0}}`,
		"chain not ending in en":  `{"bundles":{"en":1},"fallbackChains":{"az":["az","tr"]}}`,
		"chain for another start": `{"bundles":{"en":1},"fallbackChains":{"az":["tr","en"]}}`,
		"empty chain":             `{"bundles":{"en":1},"fallbackChains":{"ru":[]}}`,
		"repeated step":           `{"bundles":{"en":1},"fallbackChains":{"az":["az","tr","tr","en"]}}`,
		"unknown step":            `{"bundles":{"en":1},"fallbackChains":{"az":["az","de","en"]}}`,
		"unknown chain language":  `{"bundles":{"en":1},"fallbackChains":{"de":["de","en"]}}`,
	}
	for name, raw := range tests {
		if _, err := ParseLocalization([]byte(raw)); !errors.Is(err, ErrInvalidLocalization) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
