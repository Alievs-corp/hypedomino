package objectstore_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Alievs-corp/hypedomino/server/internal/objectstore"
)

func TestASignedURLIsStableForAKnownRequest(t *testing.T) {
	store, err := objectstore.New(objectstore.Config{
		BaseURL:   "https://examplebucket.s3.amazonaws.com",
		Region:    "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		URLTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	signed := store.SignedGet("test.txt", time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC))
	const want = "X-Amz-Signature=aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404"
	if !strings.HasSuffix(signed, want) {
		t.Fatalf("signed %s", signed)
	}
}

func TestABadConfigIsRefused(t *testing.T) {
	good := objectstore.Config{BaseURL: "https://store.example/bucket", Region: "auto", AccessKey: "a", SecretKey: "s", URLTTL: time.Minute}
	cases := map[string]func(*objectstore.Config){
		"no host":   func(c *objectstore.Config) { c.BaseURL = "bucket" },
		"scheme":    func(c *objectstore.Config) { c.BaseURL = "ftp://store.example/bucket" },
		"no region": func(c *objectstore.Config) { c.Region = "" },
		"no access": func(c *objectstore.Config) { c.AccessKey = "" },
		"no secret": func(c *objectstore.Config) { c.SecretKey = "" },
		"no ttl":    func(c *objectstore.Config) { c.URLTTL = 0 },
		"long ttl":  func(c *objectstore.Config) { c.URLTTL = 8 * 24 * time.Hour },
	}
	for name, spoil := range cases {
		config := good
		spoil(&config)
		if _, err := objectstore.New(config); !errors.Is(err, objectstore.ErrConfig) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := objectstore.New(good); err != nil {
		t.Fatal(err)
	}
}

func TestAFilePutThroughASignedURLReadsBackFromARealStore(t *testing.T) {
	ctx := context.Background()
	const access, secret = "local-access", "local-secret-value"
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rustfs/rustfs:1.0.4",
			Env:          map[string]string{"RUSTFS_ACCESS_KEY": access, "RUSTFS_SECRET_KEY": secret},
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("9000/tcp"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { container.Terminate(context.Background()) })
	endpoint, err := container.PortEndpoint(ctx, "9000/tcp", "http")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	root, err := objectstore.New(objectstore.Config{BaseURL: endpoint, Region: "us-east-1", AccessKey: access, SecretKey: secret, URLTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	createBucketOnceReady(t, root, "assets")
	store, err := objectstore.New(objectstore.Config{BaseURL: endpoint + "/assets", Region: "us-east-1", AccessKey: access, SecretKey: secret, URLTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	const key = "cosmetic/avatar/fox/static/first upload.png"
	if _, err := store.Open(ctx, key, now); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("before upload %v", err)
	}
	content := []byte("the uploaded bytes")
	put(t, store.SignedPut(key, now), content)
	body, err := store.Open(ctx, key, now)
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	read, err := io.ReadAll(body)
	if err != nil || !bytes.Equal(read, content) {
		t.Fatalf("read %q %v", read, err)
	}

	otherObject := strings.Replace(store.SignedPut(key, now), "first%20upload", "other", 1)
	if status, detail := send(t, otherObject, []byte("forged")); status != http.StatusForbidden {
		t.Fatalf("a url signed for one object wrote another: %d %s", status, detail)
	}
}

func createBucketOnceReady(t *testing.T, root *objectstore.Store, bucket string) {
	t.Helper()
	deadline := time.Now().Add(time.Minute)
	for {
		status, detail := send(t, root.SignedPut(bucket, time.Now()), nil)
		if status == http.StatusOK {
			return
		}
		if status != http.StatusServiceUnavailable || time.Now().After(deadline) {
			t.Fatalf("create bucket %d %s", status, detail)
		}
		time.Sleep(time.Second)
	}
}

func put(t *testing.T, url string, content []byte) {
	t.Helper()
	if status, detail := send(t, url, content); status != http.StatusOK {
		t.Fatalf("put %d %s", status, detail)
	}
}

func send(t *testing.T, url string, content []byte) (int, []byte) {
	t.Helper()
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	detail, _ := io.ReadAll(response.Body)
	return response.StatusCode, detail
}
