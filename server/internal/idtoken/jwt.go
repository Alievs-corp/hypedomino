package idtoken

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

const algorithmRS256 = "RS256"

var segmentEncoding = base64.RawURLEncoding.Strict()

type header struct {
	Algorithm string          `json:"alg"`
	KeyID     string          `json:"kid"`
	Critical  json.RawMessage `json:"crit"`
}

type parsedToken struct {
	header       header
	payload      []byte
	signature    []byte
	signingInput string
}

func parse(token string) (parsedToken, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return parsedToken{}, ErrMalformed
	}
	decoded, err := parseHeader(segments[0])
	if err != nil {
		return parsedToken{}, err
	}
	payload, err := segmentEncoding.DecodeString(segments[1])
	if err != nil {
		return parsedToken{}, ErrMalformed
	}
	signature, err := segmentEncoding.DecodeString(segments[2])
	if err != nil || len(signature) == 0 {
		return parsedToken{}, ErrMalformed
	}
	return parsedToken{
		header:       decoded,
		payload:      payload,
		signature:    signature,
		signingInput: segments[0] + "." + segments[1],
	}, nil
}

func parseHeader(segment string) (header, error) {
	headerBytes, err := segmentEncoding.DecodeString(segment)
	if err != nil {
		return header{}, ErrMalformed
	}
	var decoded header
	if err := json.Unmarshal(headerBytes, &decoded); err != nil {
		return header{}, ErrMalformed
	}
	if decoded.Critical != nil {
		return header{}, ErrMalformed
	}
	if decoded.Algorithm != algorithmRS256 {
		return header{}, ErrAlgorithm
	}
	if decoded.KeyID == "" {
		return header{}, ErrUnknownKey
	}
	return decoded, nil
}

func (t parsedToken) verifySignature(key *rsa.PublicKey) error {
	digest := sha256.Sum256([]byte(t.signingInput))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], t.signature) != nil {
		return ErrSignature
	}
	return nil
}
