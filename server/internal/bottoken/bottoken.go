package bottoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	Issuer         = "hypedomino-bot-minter"
	Audience       = "hypedomino-bot-signin"
	Lifetime       = time.Hour
	ClockSkew      = time.Minute
	MinSecretBytes = 32
	algorithm      = "HS256"
	maxTokenSize   = 2048
)

var (
	ErrWeakSecret = errors.New("bottoken: secret is not base64 of at least 32 bytes")
	ErrMalformed  = errors.New("bottoken: malformed bot token")
	ErrSignature  = errors.New("bottoken: bad signature")
	ErrIssuer     = errors.New("bottoken: issuer not trusted")
	ErrAudience   = errors.New("bottoken: audience is not bot sign-in")
	ErrBotName    = errors.New("bottoken: bot name is not 1 to 64 lowercase letters, digits, '-' or '_'")
	ErrExpired    = errors.New("bottoken: bot token expired")
	ErrNotYet     = errors.New("bottoken: bot token issued in the future")
	ErrLifetime   = errors.New("bottoken: bot token lives longer than one hour")
)

var (
	segment     = base64.RawURLEncoding.Strict()
	namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
)

type claims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Bot       string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

func ParseSecret(encoded string) ([]byte, error) {
	secret, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(secret) < MinSecretBytes {
		return nil, ErrWeakSecret
	}
	return secret, nil
}

func Mint(secret []byte, bot string, now time.Time) (string, error) {
	if len(secret) < MinSecretBytes {
		return "", ErrWeakSecret
	}
	if !namePattern.MatchString(bot) {
		return "", ErrBotName
	}
	head, err := json.Marshal(header{Algorithm: algorithm, Type: "JWT"})
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(claims{Issuer: Issuer, Audience: Audience, Bot: bot, IssuedAt: now.Unix(), ExpiresAt: now.Add(Lifetime).Unix()})
	if err != nil {
		return "", err
	}
	input := segment.EncodeToString(head) + "." + segment.EncodeToString(body)
	return input + "." + segment.EncodeToString(mac(secret, input)), nil
}

func Verify(secret []byte, token string, now time.Time) (string, error) {
	c, err := signedClaims(secret, token)
	if err != nil {
		return "", err
	}
	issued, expires := time.Unix(c.IssuedAt, 0), time.Unix(c.ExpiresAt, 0)
	switch {
	case c.Issuer != Issuer:
		return "", ErrIssuer
	case c.Audience != Audience:
		return "", ErrAudience
	case !namePattern.MatchString(c.Bot):
		return "", ErrBotName
	case !expires.After(issued) || expires.Sub(issued) > Lifetime:
		return "", ErrLifetime
	case issued.After(now.Add(ClockSkew)):
		return "", ErrNotYet
	case !now.Add(-ClockSkew).Before(expires):
		return "", ErrExpired
	}
	return c.Bot, nil
}

func signedClaims(secret []byte, token string) (claims, error) {
	if len(secret) < MinSecretBytes {
		return claims{}, ErrWeakSecret
	}
	if len(token) > maxTokenSize {
		return claims{}, ErrMalformed
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims{}, ErrMalformed
	}
	headBytes, err := segment.DecodeString(parts[0])
	if err != nil {
		return claims{}, ErrMalformed
	}
	var head header
	if json.Unmarshal(headBytes, &head) != nil || head.Algorithm != algorithm {
		return claims{}, ErrMalformed
	}
	signature, err := segment.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, mac(secret, parts[0]+"."+parts[1])) {
		return claims{}, ErrSignature
	}
	body, err := segment.DecodeString(parts[1])
	if err != nil {
		return claims{}, ErrMalformed
	}
	var c claims
	if err := json.Unmarshal(body, &c); err != nil {
		return claims{}, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	return c, nil
}

func mac(secret []byte, input string) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(input))
	return h.Sum(nil)
}
