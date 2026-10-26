package pb

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	hypedominov1 "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

var update = flag.Bool("update", false, "rewrite the binary golden frames from their json")

const (
	framesDir      = "../../../proto/testdata/frames"
	intentMaxBytes = 64
)

func TestGoldenFrames(t *testing.T) {
	jsonPaths, err := filepath.Glob(filepath.Join(framesDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(jsonPaths) == 0 {
		t.Fatalf("no golden frames in %s", framesDir)
	}
	covered := map[string]bool{}
	for _, jsonPath := range jsonPaths {
		name := strings.TrimSuffix(filepath.Base(jsonPath), ".json")
		t.Run(name, func(t *testing.T) {
			expected := readJSONFrame(t, jsonPath)
			binPath := strings.TrimSuffix(jsonPath, ".json") + ".bin"
			if *update {
				writeBinFrame(t, binPath, expected)
			}
			actual := readBinFrame(t, binPath)
			if !proto.Equal(expected, actual) {
				t.Fatalf("decoded frame differs\nwant: %v\ngot:  %v", expected, actual)
			}
			payload := payloadName(expected)
			covered[payload] = true
			if strings.HasSuffix(payload, "_intent") {
				assertIntentSize(t, binPath)
			}
		})
	}
	for _, payload := range payloadNames() {
		if !covered[payload] {
			t.Errorf("no golden frame for payload %q", payload)
		}
	}
}

func readJSONFrame(t *testing.T, path string) *hypedominov1.Envelope {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	frame := &hypedominov1.Envelope{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(data, frame); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if payloadName(frame) == "" {
		t.Fatalf("%s: frame has no payload", path)
	}
	return frame
}

func readBinFrame(t *testing.T, path string) *hypedominov1.Envelope {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update to write it)", err)
	}
	frame := &hypedominov1.Envelope{}
	if err := proto.Unmarshal(data, frame); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return frame
}

func writeBinFrame(t *testing.T, path string, frame *hypedominov1.Envelope) {
	t.Helper()
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertIntentSize(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() >= intentMaxBytes {
		t.Errorf("%s is %d bytes, intents must stay under %d", path, info.Size(), intentMaxBytes)
	}
}

func payloadName(frame *hypedominov1.Envelope) string {
	message := frame.ProtoReflect()
	field := message.WhichOneof(message.Descriptor().Oneofs().ByName("payload"))
	if field == nil {
		return ""
	}
	return string(field.Name())
}

func payloadNames() []string {
	fields := (&hypedominov1.Envelope{}).ProtoReflect().Descriptor().Oneofs().ByName("payload").Fields()
	names := make([]string, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		names = append(names, string(fields.Get(i).Name()))
	}
	return names
}
