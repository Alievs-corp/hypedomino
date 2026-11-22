package liveops

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func documentedRows(t *testing.T, path, header string) map[string][]string {
	t.Helper()
	doc, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, table, found := strings.Cut(string(doc), header)
	if !found {
		t.Fatalf("%s has no config table", path)
	}
	rows := map[string][]string{}
	for _, line := range strings.Split(table, "\n") {
		if !strings.HasPrefix(line, "| ") {
			break
		}
		cells := strings.Split(strings.Trim(line, "| "), " | ")
		rows[strings.Trim(cells[0], "`")] = cells[1:]
	}
	return rows
}

func keysWithPrefix(prefix string) []Key {
	var matching []Key
	for _, key := range Keys() {
		if strings.HasPrefix(key.Name, prefix) {
			matching = append(matching, key)
		}
	}
	return matching
}

func TestTheKeysMatchTheConfigTableOfTheContentGateSpec(t *testing.T) {
	documented := documentedRows(t, "../../../docs/launch-content-gate.md", "| Key | Default | Section |\n|---|---|---|\n")
	gateKeys := keysWithPrefix("ContentGateMin")
	if len(documented) != len(gateKeys) {
		t.Fatalf("launch-content-gate.md lists %d keys, code has %d", len(documented), len(gateKeys))
	}
	for _, key := range gateKeys {
		if got := documented[key.Name][0]; got != strconv.Itoa(key.Default) {
			t.Errorf("%s: spec default %q, code %d", key.Name, got, key.Default)
		}
		if key.Min != 0 || key.Max != nil || key.Default < key.Min {
			t.Errorf("%s: min %d, default %d, spec says integers >= 0", key.Name, key.Min, key.Default)
		}
	}
}

func TestTheAnalyticsKeysMatchTheConfigTableOfTheAnalyticsSpec(t *testing.T) {
	documented := documentedRows(t, "../../../docs/analytics.md", "| Key | Default | Min | Max | Meaning |\n|---|---|---|---|---|\n")
	analyticsKeys := keysWithPrefix("Analytics")
	if len(documented) != 3 || len(analyticsKeys) != 3 {
		t.Fatalf("analytics.md lists %d keys, code has %d", len(documented), len(analyticsKeys))
	}
	for _, key := range analyticsKeys {
		row := documented[key.Name]
		if len(row) < 3 || key.Max == nil || row[0] != strconv.Itoa(key.Default) || row[1] != strconv.Itoa(key.Min) || row[2] != strconv.Itoa(*key.Max) {
			t.Fatalf("%s: spec %v, code default %d min %d max %v", key.Name, row, key.Default, key.Min, key.Max)
		}
		if key.Default < key.Min || key.Default > *key.Max {
			t.Errorf("%s: default %d outside its bounds", key.Name, key.Default)
		}
	}
	if batch := keysWithPrefix(AnalyticsBatchSize)[0]; *batch.Max != AnalyticsMaxBatchSize {
		t.Errorf("batch size max %d, body limit allows %d", *batch.Max, AnalyticsMaxBatchSize)
	}
}

func TestTheAnalyticsValuesFallBackToTheirDefaults(t *testing.T) {
	tests := []struct {
		name   string
		values Values
		want   Analytics
	}{
		{"none published", nil, Analytics{BatchSize: 50, QueueLimit: 500, FlushSeconds: 60}},
		{"one published", Values{AnalyticsFlushSeconds: 120}, Analytics{BatchSize: 50, QueueLimit: 500, FlushSeconds: 120}},
		{"every key", Values{AnalyticsBatchSize: 400, AnalyticsQueueLimit: 400, AnalyticsFlushSeconds: 10}, Analytics{BatchSize: 400, QueueLimit: 400, FlushSeconds: 10}},
		{"content gate only", Values{"ContentGateMinFrames": 25}, Analytics{BatchSize: 50, QueueLimit: 500, FlushSeconds: 60}},
	}
	for _, test := range tests {
		if got := test.values.Analytics(); got != test.want {
			t.Errorf("%s: %+v, want %+v", test.name, got, test.want)
		}
		config := test.values.AnalyticsConfig()
		if config.GetBatchSize() != uint32(test.want.BatchSize) || config.GetQueueLimit() != uint32(test.want.QueueLimit) || config.GetFlushSeconds() != uint32(test.want.FlushSeconds) {
			t.Errorf("%s: hello config %v, want %+v", test.name, config, test.want)
		}
	}
}

func TestSnapshotsAreCheckedAgainstTheKeys(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Values
	}{
		{"every key", `{"ContentGateMinAvatarsAnimated": 31, "ContentGateMinAvatarsStatic": 61, "ContentGateMinFrames": 21, "ContentGateMinTables": 16, "ContentGateMinEmoji": 41, "ContentGateMinMedals": 41}`, Values{"ContentGateMinAvatarsAnimated": 31, "ContentGateMinAvatarsStatic": 61, "ContentGateMinFrames": 21, "ContentGateMinTables": 16, "ContentGateMinEmoji": 41, "ContentGateMinMedals": 41}},
		{"one key", `{"ContentGateMinFrames": 25}`, Values{"ContentGateMinFrames": 25}},
		{"zero", `{"ContentGateMinEmoji": 0}`, Values{"ContentGateMinEmoji": 0}},
		{"below the default", `{"ContentGateMinMedals": 10}`, Values{"ContentGateMinMedals": 10}},
		{"large", `{"ContentGateMinTables": 1000}`, Values{"ContentGateMinTables": 1000}},
		{"no keys", `{}`, Values{}},
		{"analytics keys", `{"AnalyticsBatchSize": 100, "AnalyticsQueueLimit": 1000, "AnalyticsFlushSeconds": 30}`, Values{"AnalyticsBatchSize": 100, "AnalyticsQueueLimit": 1000, "AnalyticsFlushSeconds": 30}},
		{"analytics bounds", `{"AnalyticsBatchSize": 400, "AnalyticsQueueLimit": 5000, "AnalyticsFlushSeconds": 3600}`, Values{"AnalyticsBatchSize": 400, "AnalyticsQueueLimit": 5000, "AnalyticsFlushSeconds": 3600}},
		{"queue limit equal to the batch", `{"AnalyticsBatchSize": 1, "AnalyticsQueueLimit": 1, "AnalyticsFlushSeconds": 10}`, Values{"AnalyticsBatchSize": 1, "AnalyticsQueueLimit": 1, "AnalyticsFlushSeconds": 10}},
	}
	for _, test := range tests {
		got, err := Parse([]byte(test.raw))
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: %v %v", test.name, got, err)
		}
	}
	for name, raw := range map[string]string{
		"unknown key":                   `{"ContentGateMinChests": 5}`,
		"rules key":                     `{"TurnSeconds": 20}`,
		"negative":                      `{"ContentGateMinFrames": -1}`,
		"fraction":                      `{"ContentGateMinFrames": 20.5}`,
		"string":                        `{"ContentGateMinFrames": "20"}`,
		"null value":                    `{"ContentGateMinFrames": null}`,
		"boolean":                       `{"ContentGateMinFrames": true}`,
		"array":                         `[20]`,
		"null":                          `null`,
		"not JSON":                      `ContentGateMinFrames=20`,
		"batch of zero":                 `{"AnalyticsBatchSize": 0}`,
		"batch above the body limit":    `{"AnalyticsBatchSize": 401, "AnalyticsQueueLimit": 1000}`,
		"queue above its maximum":       `{"AnalyticsQueueLimit": 5001}`,
		"flush below its minimum":       `{"AnalyticsFlushSeconds": 9}`,
		"flush above its maximum":       `{"AnalyticsFlushSeconds": 3601}`,
		"queue below the batch":         `{"AnalyticsBatchSize": 100, "AnalyticsQueueLimit": 99}`,
		"queue below the default batch": `{"AnalyticsQueueLimit": 49}`,
		"fractional flush":              `{"AnalyticsFlushSeconds": 60.5}`,
	} {
		if _, err := Parse([]byte(raw)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s accepted: %v", name, err)
		}
	}
}
