package privacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

const (
	Domain                             = "privacy"
	PrivacyPolicyVersion               = "PrivacyPolicyVersion"
	TermsVersion                       = "TermsVersion"
	ConsentRecordRetentionYears        = "ConsentRecordRetentionYears"
	PrivacyPolicyURL                   = "PrivacyPolicyUrl"
	TermsOfUseURL                      = "TermsOfUseUrl"
	MaxURLLength                       = 512
	DefaultConsentRecordRetentionYears = 3
	MaxConsentRecordRetentionYears     = 10
)

var ErrInvalid = errors.New("privacy: invalid privacy snapshot")

var versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)

type Config struct {
	PrivacyPolicyVersion        string
	TermsVersion                string
	ConsentRecordRetentionYears int
	PrivacyPolicyURL            string
	TermsOfUseURL               string
}

func Parse(raw []byte) (Config, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return Config{}, fmt.Errorf("%w: not a JSON object: %v", ErrInvalid, err)
	}
	var config Config
	for name, field := range fields {
		if err := config.set(name, field); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}

func (c *Config) set(name string, field json.RawMessage) error {
	switch name {
	case PrivacyPolicyVersion:
		return version(name, field, &c.PrivacyPolicyVersion)
	case TermsVersion:
		return version(name, field, &c.TermsVersion)
	case ConsentRecordRetentionYears:
		years, err := strconv.Atoi(string(field))
		if err != nil || years < 1 || years > MaxConsentRecordRetentionYears {
			return fmt.Errorf("%w: %s must be a whole number from 1 to %d, got %s", ErrInvalid, name, MaxConsentRecordRetentionYears, field)
		}
		c.ConsentRecordRetentionYears = years
		return nil
	case PrivacyPolicyURL:
		return link(name, field, &c.PrivacyPolicyURL)
	case TermsOfUseURL:
		return link(name, field, &c.TermsOfUseURL)
	}
	return fmt.Errorf("%w: unknown key %s", ErrInvalid, name)
}

func version(name string, field json.RawMessage, target *string) error {
	var text string
	if err := json.Unmarshal(field, &text); err != nil || !versionPattern.MatchString(text) {
		return fmt.Errorf("%w: %s must be 1 to 32 letters, digits, dots, dashes or underscores starting with a letter or digit, got %s", ErrInvalid, name, field)
	}
	*target = text
	return nil
}

func link(name string, field json.RawMessage, target *string) error {
	var text string
	if err := json.Unmarshal(field, &text); err != nil || !validLink(text) {
		return fmt.Errorf("%w: %s must be an https URL with a host, at most %d characters, got %s", ErrInvalid, name, MaxURLLength, field)
	}
	*target = text
	return nil
}

func validLink(text string) bool {
	if text == "" || len(text) > MaxURLLength || strings.IndexFunc(text, blank) >= 0 {
		return false
	}
	parsed, err := url.Parse(text)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func blank(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }

func (c Config) RetentionYears() int {
	if c.ConsentRecordRetentionYears == 0 {
		return DefaultConsentRecordRetentionYears
	}
	return c.ConsentRecordRetentionYears
}

func (c Config) Versions() *pb.PrivacyVersions {
	return &pb.PrivacyVersions{
		PrivacyPolicyVersion: c.PrivacyPolicyVersion,
		TermsVersion:         c.TermsVersion,
		PrivacyPolicyUrl:     c.PrivacyPolicyURL,
		TermsOfUseUrl:        c.TermsOfUseURL,
	}
}
