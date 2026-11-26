package privacy_test

import (
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/privacy"
)

func TestParseReadsThePublishedPrivacyConfig(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		want  privacy.Config
		years int
	}{
		{"empty object keeps the defaults", `{}`, privacy.Config{}, 3},
		{"every key", `{"PrivacyPolicyVersion":"2026-11-01","TermsVersion":"v2.1_rc","ConsentRecordRetentionYears":5}`, privacy.Config{PrivacyPolicyVersion: "2026-11-01", TermsVersion: "v2.1_rc", ConsentRecordRetentionYears: 5}, 5},
		{"policy only", `{"PrivacyPolicyVersion":"7"}`, privacy.Config{PrivacyPolicyVersion: "7"}, 3},
		{"lowest retention", `{"ConsentRecordRetentionYears":1}`, privacy.Config{ConsentRecordRetentionYears: 1}, 1},
		{"highest retention", `{"ConsentRecordRetentionYears":10}`, privacy.Config{ConsentRecordRetentionYears: 10}, 10},
		{"longest version", `{"TermsVersion":"abcdefghijklmnopqrstuvwxyz012345"}`, privacy.Config{TermsVersion: "abcdefghijklmnopqrstuvwxyz012345"}, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := privacy.Parse([]byte(test.raw))
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want || got.RetentionYears() != test.years {
				t.Fatalf("got %+v retention %d, want %+v retention %d", got, got.RetentionYears(), test.want, test.years)
			}
		})
	}
}

func TestParseRefusesAnInvalidPrivacyConfig(t *testing.T) {
	for _, raw := range []string{
		``,
		`null`,
		`[]`,
		`{"PrivacyPolicyVersion":""}`,
		`{"PrivacyPolicyVersion":2026}`,
		`{"PrivacyPolicyVersion":"-2026"}`,
		`{"PrivacyPolicyVersion":"2026 11"}`,
		`{"TermsVersion":"abcdefghijklmnopqrstuvwxyz0123456"}`,
		`{"TermsVersion":null}`,
		`{"ConsentRecordRetentionYears":0}`,
		`{"ConsentRecordRetentionYears":11}`,
		`{"ConsentRecordRetentionYears":-3}`,
		`{"ConsentRecordRetentionYears":2.5}`,
		`{"ConsentRecordRetentionYears":"3"}`,
		`{"privacyPolicyVersion":"2026-11-01"}`,
		`{"MinAccountAgeByCountry":13}`,
	} {
		if _, err := privacy.Parse([]byte(raw)); !errors.Is(err, privacy.ErrInvalid) {
			t.Fatalf("%s: got %v, want ErrInvalid", raw, err)
		}
	}
}

func TestVersionsCarryThePublishedVersionsForTheHello(t *testing.T) {
	tests := []struct {
		config privacy.Config
		want   *pb.PrivacyVersions
	}{
		{privacy.Config{}, &pb.PrivacyVersions{}},
		{privacy.Config{PrivacyPolicyVersion: "2026-11-01", TermsVersion: "2026-10-01", ConsentRecordRetentionYears: 4}, &pb.PrivacyVersions{PrivacyPolicyVersion: "2026-11-01", TermsVersion: "2026-10-01"}},
	}
	for _, test := range tests {
		if got := test.config.Versions(); !proto.Equal(got, test.want) {
			t.Fatalf("%+v: got %v, want %v", test.config, got, test.want)
		}
	}
}
