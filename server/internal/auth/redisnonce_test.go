package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Alievs-corp/hypedomino/server/internal/auth"
	"github.com/Alievs-corp/hypedomino/server/internal/cluster/redistest"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

func TestMain(m *testing.M) {
	os.Exit(redistest.Main(m))
}

func TestRedisNoncesAreSingleUseAndExpire(t *testing.T) {
	options, err := redis.ParseURL(redistest.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(options)
	t.Cleanup(func() { client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	config := idtoken.DefaultConfig()
	config.NonceTTL = 2 * time.Second
	var nonces idtoken.NonceStore = auth.NewRedisNonces(client, "hd:test:"+t.Name()+":")
	verifier := idtoken.New(idtoken.Options{Config: config, Nonces: nonces})
	nonce, err := verifier.IssueNonce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := nonces.Consume(ctx, "never-issued", time.Now()); ok || err != nil {
		t.Fatalf("unknown nonce consumed: %v %v", ok, err)
	}
	if ok, err := nonces.Consume(ctx, nonce, time.Now()); !ok || err != nil {
		t.Fatalf("first consume: %v %v", ok, err)
	}
	if ok, err := nonces.Consume(ctx, nonce, time.Now()); ok || err != nil {
		t.Fatalf("second consume: %v %v", ok, err)
	}
	expiring, err := verifier.IssueNonce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	if ok, err := nonces.Consume(ctx, expiring, time.Now()); ok || err != nil {
		t.Fatalf("expired nonce consumed: %v %v", ok, err)
	}
}
