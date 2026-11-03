package token

import "net/http"

type Verifier interface {
	Verify(access string) (Claims, error)
}

func Authorize(verifier Verifier, r *http.Request) (Claims, error) {
	return verifier.Verify(bearer(r))
}
