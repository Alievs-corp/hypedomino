package consent

import "errors"

type Purpose string

const (
	Analytics   Purpose = "analytics"
	Diagnostics Purpose = "diagnostics"
)

var (
	ErrMalformed = errors.New("consent.malformed")
	ErrContext   = errors.New("consent.context_required")
)

type Choices struct {
	Analytics   bool `json:"analytics"`
	Diagnostics bool `json:"diagnostics"`
}

type Change struct {
	Choices
	PolicyVersion string `json:"policyVersion"`
	AppVersion    string `json:"appVersion"`
}

func (c Choices) Of(purpose Purpose) bool {
	switch purpose {
	case Analytics:
		return c.Analytics
	case Diagnostics:
		return c.Diagnostics
	}
	return false
}

func (c *Choices) set(purpose Purpose, granted bool) {
	switch purpose {
	case Analytics:
		c.Analytics = granted
	case Diagnostics:
		c.Diagnostics = granted
	}
}

func (c Choices) Changed(next Choices) []Purpose {
	var changed []Purpose
	for _, purpose := range []Purpose{Analytics, Diagnostics} {
		if c.Of(purpose) != next.Of(purpose) {
			changed = append(changed, purpose)
		}
	}
	return changed
}
