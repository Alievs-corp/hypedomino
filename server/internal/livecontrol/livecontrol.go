package livecontrol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
)

const Domain = "live_controls"

var ErrInvalid = errors.New("livecontrol: invalid controls")

type Controls struct {
	MinimumAppVersion      string   `json:"minimum_app_version"`
	BuildAllowlistEnforced bool     `json:"build_allowlist_enforced"`
	BuildAllowlist         []string `json:"build_allowlist"`
	PurchasesDisabled      bool     `json:"purchases_disabled"`
	RankedDisabled         bool     `json:"ranked_disabled"`
}

func Parse(raw []byte) (Controls, error) {
	var controls Controls
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&controls); err != nil {
		return Controls{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return controls, controls.validate()
}

func (c Controls) validate() error {
	if !gateway.VersionAtLeast(c.MinimumAppVersion, "0.0.0") {
		return fmt.Errorf("%w: minimum_app_version %q is not major.minor.patch", ErrInvalid, c.MinimumAppVersion)
	}
	if c.BuildAllowlistEnforced && len(c.BuildAllowlist) == 0 {
		return fmt.Errorf("%w: an enforced build allowlist cannot be empty", ErrInvalid)
	}
	if slices.Contains(c.BuildAllowlist, "") {
		return fmt.Errorf("%w: empty build hash", ErrInvalid)
	}
	return nil
}

func (c Controls) Admission() gateway.Admission {
	return gateway.Admission{MinimumAppVersion: c.MinimumAppVersion, EnforceBuilds: c.BuildAllowlistEnforced, Builds: c.BuildAllowlist}
}
