package adminapi_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/objectstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type memoryStore struct {
	files   map[string][]byte
	broken  bool
	offline bool
}

func (m *memoryStore) SignedPut(key string, _ time.Time) string {
	return "https://store.example/put/" + key
}

func (m *memoryStore) SignedGet(key string, _ time.Time) string {
	return "https://store.example/get/" + key
}

func (m *memoryStore) Open(_ context.Context, key string, _ time.Time) (io.ReadCloser, error) {
	if m.offline {
		return nil, objectstore.ErrUnavailable
	}
	file, ok := m.files[key]
	if !ok {
		return nil, objectstore.ErrNotFound
	}
	if m.broken {
		return io.NopCloser(io.MultiReader(bytes.NewReader(file[:8]), iotest.ErrReader(errors.New("connection reset")))), nil
	}
	return io.NopCloser(bytes.NewReader(file)), nil
}

func (m *memoryStore) upload(uploadURL string, file []byte) {
	m.files[strings.TrimPrefix(uploadURL, "https://store.example/put/")] = file
}

func newAssetAPI(t *testing.T) (*api, *memoryStore) {
	t.Helper()
	store := &memoryStore{files: map[string][]byte{}}
	return newAPIWith(t, func(c *adminapi.Config) { c.Assets = store }), store
}

func encodedPNG(t *testing.T, picture image.Image) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := png.Encode(&out, picture); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func startUpload(t *testing.T, a *api, token string) (string, string) {
	t.Helper()
	status, body := a.call("POST", "/admin/assets/uploads", token, map[string]any{"item_id": "avatar.fox", "slot": "static"})
	expect(t, status, body, http.StatusOK, "")
	return body["asset"].(map[string]any)["id"].(string), body["upload_url"].(string)
}

func TestAStaticFileIsUploadedMeasuredAndPassed(t *testing.T) {
	a, store := newAssetAPI(t)
	a.create("leyla@example.com", rbac.LiveOps)
	content, liveOps := a.signIn("ayla"), a.signIn("leyla")

	status, body := a.call("POST", "/admin/assets/uploads", liveOps, map[string]any{"item_id": "avatar.fox", "slot": "static"})
	expect(t, status, body, http.StatusForbidden, "forbidden")

	id, uploadURL := startUpload(t, a, content)
	if !strings.HasPrefix(uploadURL, "https://store.example/put/cosmetic/avatar/fox/static/"+id) {
		t.Fatalf("url %s", uploadURL)
	}

	status, body = a.call("POST", "/admin/assets/"+id+"/check", content, nil)
	expect(t, status, body, http.StatusConflict, "asset_file_missing")

	file := encodedPNG(t, image.NewNRGBA(image.Rect(0, 0, 64, 48)))
	store.upload(uploadURL, file)
	status, body = a.call("POST", "/admin/assets/"+id+"/check", content, nil)
	expect(t, status, body, http.StatusOK, "")
	asset := body["asset"].(map[string]any)
	sum := sha256.Sum256(file)
	if asset["status"] != "passed" || asset["format"] != "png" || asset["width"] != float64(64) || asset["height"] != float64(48) || asset["frames"] != float64(1) ||
		asset["bytes"] != float64(len(file)) || asset["content_hash"] != hex.EncodeToString(sum[:]) || asset["address"] != "cosmetic/avatar/fox/static" {
		t.Fatalf("asset %v", asset)
	}

	status, body = a.call("POST", "/admin/assets/"+id+"/check", content, nil)
	expect(t, status, body, http.StatusConflict, "asset_already_checked")

	status, body = a.call("GET", "/admin/assets", liveOps, nil)
	expect(t, status, body, http.StatusOK, "")
	listed := body["assets"].([]any)[0].(map[string]any)
	if body["store_configured"] != true || !strings.HasPrefix(listed["view_url"].(string), "https://store.example/get/") {
		t.Fatalf("list %v", body)
	}
}

func TestEveryAssetStepIsAudited(t *testing.T) {
	a, store := newAssetAPI(t)
	token := a.signIn("ayla")
	id, uploadURL := startUpload(t, a, token)
	store.upload(uploadURL, encodedPNG(t, image.NewNRGBA(image.Rect(0, 0, 8, 8))))
	status, body := a.call("POST", "/admin/assets/"+id+"/check", token, nil)
	expect(t, status, body, http.StatusOK, "")

	status, body = a.call("GET", "/admin/audit?object_type=assets", a.signIn("root"), nil)
	expect(t, status, body, http.StatusOK, "")
	var actions []string
	for _, entry := range body["entries"].([]any) {
		actions = append(actions, entry.(map[string]any)["action"].(string))
	}
	if strings.Join(actions, ",") != "check,create" {
		t.Fatalf("audit %v", actions)
	}
}

func TestAFileThatFailsItsChecksIsRejectedAndNotViewable(t *testing.T) {
	cases := map[string]struct {
		file   func(*testing.T) []byte
		failed string
	}{
		"not a png": {func(*testing.T) []byte { return []byte("not an image at all") }, "format"},
		"no alpha":  {func(t *testing.T) []byte { return encodedPNG(t, image.NewGray(image.Rect(0, 0, 8, 8))) }, "alpha"},
		"empty":     {func(*testing.T) []byte { return []byte{} }, "format"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			a, store := newAssetAPI(t)
			token := a.signIn("ayla")
			id, uploadURL := startUpload(t, a, token)
			store.upload(uploadURL, c.file(t))
			status, body := a.call("POST", "/admin/assets/"+id+"/check", token, nil)
			expect(t, status, body, http.StatusOK, "")
			asset := body["asset"].(map[string]any)
			checks := asset["checks"].(map[string]any)
			if asset["status"] != "rejected" || checks[c.failed] != false || asset["view_url"] != nil {
				t.Fatalf("asset %v", asset)
			}
			if _, measured := checks["alpha"]; c.failed == "format" && measured {
				t.Fatalf("a file that is not a png reported on alpha: %v", checks)
			}
		})
	}
}

func TestABrokenStoreLeavesTheAssetWaiting(t *testing.T) {
	a, store := newAssetAPI(t)
	token := a.signIn("ayla")
	id, uploadURL := startUpload(t, a, token)
	store.upload(uploadURL, encodedPNG(t, image.NewNRGBA(image.Rect(0, 0, 8, 8))))

	store.broken = true
	status, body := a.call("POST", "/admin/assets/"+id+"/check", token, nil)
	expect(t, status, body, http.StatusServiceUnavailable, "asset_store_unavailable")
	store.broken, store.offline = false, true
	status, body = a.call("POST", "/admin/assets/"+id+"/check", token, nil)
	expect(t, status, body, http.StatusServiceUnavailable, "asset_store_unavailable")

	store.offline = false
	status, body = a.call("POST", "/admin/assets/"+id+"/check", token, nil)
	expect(t, status, body, http.StatusOK, "")
	if body["asset"].(map[string]any)["status"] != "passed" {
		t.Fatalf("after recovery %v", body)
	}
}

func TestAnUploadIsRefusedForABadItemOrASlotWithoutAFormat(t *testing.T) {
	a, _ := newAssetAPI(t)
	token := a.signIn("ayla")
	status, body := a.call("POST", "/admin/assets/uploads", token, map[string]any{"item_id": "../avatar.fox", "slot": "static"})
	expect(t, status, body, http.StatusBadRequest, "invalid")
	for _, slot := range []string{"anim", "clip/idle", "layer/base", ""} {
		status, body = a.call("POST", "/admin/assets/uploads", token, map[string]any{"item_id": "avatar.fox", "slot": slot})
		expect(t, status, body, http.StatusBadRequest, "asset_slot_unsupported")
	}
	status, body = a.call("POST", "/admin/assets/not-an-id/check", token, nil)
	expect(t, status, body, http.StatusNotFound, "not_found")
	status, body = a.call("POST", "/admin/assets/5f2c9d3e-6d4a-4f4e-9d11-1a2b3c4d5e6f/check", token, nil)
	expect(t, status, body, http.StatusNotFound, "not_found")
}

func TestWithoutAStoreTheListWorksAndUploadsAreRefused(t *testing.T) {
	a := newAPI(t)
	token := a.signIn("ayla")
	status, body := a.call("GET", "/admin/assets", token, nil)
	expect(t, status, body, http.StatusOK, "")
	if body["store_configured"] != false || len(body["assets"].([]any)) != 0 {
		t.Fatalf("list %v", body)
	}
	status, body = a.call("POST", "/admin/assets/uploads", token, map[string]any{"item_id": "avatar.fox", "slot": "static"})
	expect(t, status, body, http.StatusServiceUnavailable, "asset_store_off")
}
