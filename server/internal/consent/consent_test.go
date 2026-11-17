package consent_test

import (
	"fmt"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/consent"
)

func TestChangedListsOnlyPurposesThatDiffer(t *testing.T) {
	for _, tc := range []struct {
		from, to consent.Choices
		want     string
	}{
		{consent.Choices{}, consent.Choices{}, "[]"},
		{consent.Choices{}, consent.Choices{Analytics: true}, "[analytics]"},
		{consent.Choices{Analytics: true}, consent.Choices{Diagnostics: true}, "[analytics diagnostics]"},
		{consent.Choices{Analytics: true, Diagnostics: true}, consent.Choices{Analytics: true, Diagnostics: true}, "[]"},
	} {
		if got := fmt.Sprint(tc.from.Changed(tc.to)); got != tc.want {
			t.Fatalf("%+v to %+v changed %s, want %s", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestOfReadsEachPurpose(t *testing.T) {
	choices := consent.Choices{Diagnostics: true}
	if choices.Of(consent.Analytics) || !choices.Of(consent.Diagnostics) || choices.Of("promotional_push") {
		t.Fatalf("%+v read wrongly", choices)
	}
}
