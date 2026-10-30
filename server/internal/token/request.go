package token

import (
	"net/http"
	"strings"
)

type Verifier interface {
	Verify(access string) (Claims, error)
}

func Authorize(verifier Verifier, r *http.Request) (Claims, error) {
	scheme, value, _ := strings.Cut(r.Header.Get("Authorization"), " ")
	if !strings.EqualFold(scheme, "Bearer") {
		value = ""
	}
	return verifier.Verify(strings.TrimSpace(value))
}
