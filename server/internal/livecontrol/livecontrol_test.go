package livecontrol_test

import (
	"errors"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/livecontrol"
)

func TestParseChecksTheControls(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		valid bool
	}{
		{"version only", `{"minimum_app_version":"1.4.0"}`, true},
		{"enforced allowlist", `{"minimum_app_version":"1.4.0","build_allowlist_enforced":true,"build_allowlist":["abc"]}`, true},
		{"listed but not enforced", `{"minimum_app_version":"1.4.0","build_allowlist":["abc"]}`, true},
		{"both kill switches", `{"minimum_app_version":"1.4.0","purchases_disabled":true,"ranked_disabled":true}`, true},
		{"no version", `{}`, false},
		{"bad version", `{"minimum_app_version":"latest"}`, false},
		{"enforced and empty", `{"minimum_app_version":"1.4.0","build_allowlist_enforced":true}`, false},
		{"empty hash", `{"minimum_app_version":"1.4.0","build_allowlist":[""]}`, false},
		{"unknown field", `{"minimum_app_version":"1.4.0","purchases_off":true}`, false},
		{"not json", `{`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := livecontrol.Parse([]byte(c.raw))
			if c.valid && err != nil {
				t.Fatalf("refused: %v", err)
			}
			if !c.valid && !errors.Is(err, livecontrol.ErrInvalid) {
				t.Fatalf("got %v, want ErrInvalid", err)
			}
		})
	}
}

func TestAdmissionCarriesTheVersionAndTheAllowlist(t *testing.T) {
	controls, err := livecontrol.Parse([]byte(`{"minimum_app_version":"1.4.0","build_allowlist_enforced":true,"build_allowlist":["abc"]}`))
	if err != nil {
		t.Fatal(err)
	}
	admission := controls.Admission()
	if !admission.Admits("1.4.0", "abc") || admission.Admits("1.3.9", "abc") || admission.Admits("1.4.0", "leaked") {
		t.Fatalf("admission %+v", admission)
	}
}
