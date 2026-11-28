package bottoken

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	secret = bytes.Repeat([]byte{7}, MinSecretBytes)
	now    = time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
)

func signed(t *testing.T, key []byte, head header, body claims) string {
	t.Helper()
	h, err := json.Marshal(head)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	input := segment.EncodeToString(h) + "." + segment.EncodeToString(b)
	return input + "." + segment.EncodeToString(mac(key, input))
}

func valid() claims {
	return claims{Issuer: Issuer, Audience: Audience, Bot: "staging-bot-1", IssuedAt: now.Unix(), ExpiresAt: now.Add(Lifetime).Unix()}
}

func TestMintedTokenVerifiesAsItsBotForOneHour(t *testing.T) {
	token, err := Mint(secret, "staging-bot-1", now)
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []time.Time{now, now.Add(Lifetime - time.Second), now.Add(-ClockSkew)} {
		if bot, err := Verify(secret, token, at); err != nil || bot != "staging-bot-1" {
			t.Fatalf("at %v: %q, %v", at, bot, err)
		}
	}
	if _, err := Verify(secret, token, now.Add(Lifetime+ClockSkew)); !errors.Is(err, ErrExpired) {
		t.Fatalf("after its hour: %v", err)
	}
}

func TestVerifyRefusesEveryTamperedOrForeignToken(t *testing.T) {
	other := bytes.Repeat([]byte{9}, MinSecretBytes)
	hs256 := header{Algorithm: algorithm, Type: "JWT"}
	minted, err := Mint(secret, "staging-bot-1", now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(minted, ".")
	forged, _ := json.Marshal(claims{Issuer: Issuer, Audience: Audience, Bot: "staging-bot-2", IssuedAt: now.Unix(), ExpiresAt: now.Add(Lifetime).Unix()})
	with := func(change func(*claims)) claims {
		c := valid()
		change(&c)
		return c
	}
	cases := []struct {
		name  string
		token string
		want  error
	}{
		{"other secret", signed(t, other, hs256, valid()), ErrSignature},
		{"swapped body", parts[0] + "." + segment.EncodeToString(forged) + "." + parts[2], ErrSignature},
		{"alg none", signed(t, secret, header{Algorithm: "none", Type: "JWT"}, valid()), ErrMalformed},
		{"two parts", parts[0] + "." + parts[1], ErrMalformed},
		{"oversized", strings.Repeat("a", maxTokenSize+1), ErrMalformed},
		{"access token issuer", signed(t, secret, hs256, with(func(c *claims) { c.Issuer = "hypedomino" })), ErrIssuer},
		{"other audience", signed(t, secret, hs256, with(func(c *claims) { c.Audience = "hypedomino" })), ErrAudience},
		{"empty bot", signed(t, secret, hs256, with(func(c *claims) { c.Bot = "" })), ErrBotName},
		{"bot name with a path", signed(t, secret, hs256, with(func(c *claims) { c.Bot = "Bot/../1" })), ErrBotName},
		{"longer than an hour", signed(t, secret, hs256, with(func(c *claims) { c.ExpiresAt = now.Add(Lifetime + time.Second).Unix() })), ErrLifetime},
		{"expiry before issue", signed(t, secret, hs256, with(func(c *claims) { c.ExpiresAt = c.IssuedAt })), ErrLifetime},
		{"issued in the future", signed(t, secret, hs256, with(func(c *claims) {
			c.IssuedAt = now.Add(2 * ClockSkew).Unix()
			c.ExpiresAt = now.Add(2*ClockSkew + Lifetime).Unix()
		})), ErrNotYet},
		{"expired", signed(t, secret, hs256, with(func(c *claims) {
			c.IssuedAt = now.Add(-2 * Lifetime).Unix()
			c.ExpiresAt = now.Add(-Lifetime).Unix()
		})), ErrExpired},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if bot, err := Verify(secret, c.token, now); !errors.Is(err, c.want) {
				t.Fatalf("got %q, %v; want %v", bot, err, c.want)
			}
		})
	}
}

func TestSecretsShorterThan32BytesAreRefused(t *testing.T) {
	if _, err := ParseSecret(base64.StdEncoding.EncodeToString(make([]byte, MinSecretBytes-1))); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("short secret: %v", err)
	}
	if _, err := ParseSecret("not base64!"); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("not base64: %v", err)
	}
	parsed, err := ParseSecret(base64.StdEncoding.EncodeToString(secret))
	if err != nil || !bytes.Equal(parsed, secret) {
		t.Fatalf("secret %x, %v", parsed, err)
	}
	if _, err := Mint(secret[:MinSecretBytes-1], "staging-bot-1", now); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("mint with a short secret: %v", err)
	}
	if _, err := Verify(nil, "a.b.c", now); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("verify without a secret: %v", err)
	}
}

func TestMintRefusesABadBotName(t *testing.T) {
	if _, err := Mint(secret, "Staging Bot", now); !errors.Is(err, ErrBotName) {
		t.Fatalf("bad name: %v", err)
	}
}
