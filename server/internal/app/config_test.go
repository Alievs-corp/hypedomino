package app

import (
	"strings"
	"testing"
	"time"
)

func lookup(values map[string]string) Lookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func TestConfigDefaultsWhenTheEnvironmentIsEmpty(t *testing.T) {
	config, err := ConfigFromEnv(lookup(nil))
	if err != nil {
		t.Fatal(err)
	}
	if config.ListenAddr != ":8080" || config.LogLevel != "info" || config.DrainTimeout != 15*time.Minute || config.Gateway.MaxFrameBytes != 1024 || config.Origins != nil {
		t.Fatalf("defaults %+v", config)
	}
}

func TestConfigReadsEveryVariable(t *testing.T) {
	config, err := ConfigFromEnv(lookup(map[string]string{
		"HD_LISTEN_ADDR":           "127.0.0.1:9000",
		"HD_LOG_LEVEL":             "DEBUG",
		"HD_ORIGINS":               "play.example.com, admin.example.com,",
		"HD_SHUTDOWN_TIMEOUT":      "3s",
		"HD_DRAIN_TIMEOUT":         "2m",
		"HD_MIN_APP_VERSION":       "1.4.0",
		"HD_CATALOG_VERSION":       "2026.11.01-2",
		"HD_MAX_FRAME_BYTES":       "2048",
		"HD_WRITE_QUEUE":           "64",
		"HD_HELLO_TIMEOUT":         "1s",
		"HD_WRITE_TIMEOUT":         "2s",
		"HD_CLOSE_TIMEOUT":         "500ms",
		"HD_COMPRESSION_THRESHOLD": "512",
		"HD_INTENTS_PER_SECOND":    "20",
		"HD_INTENT_BURST":          "40",
		"HD_PING_INTERVAL":         "7s",
		"HD_PONG_TIMEOUT":          "3s",
		"HD_RECONNECT_GRACE":       "45s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	g := config.Gateway
	if config.ListenAddr != "127.0.0.1:9000" || config.LogLevel != "DEBUG" || len(config.Origins) != 2 || config.Origins[1] != "admin.example.com" {
		t.Fatalf("app config %+v", config)
	}
	if config.ShutdownTimeout != 3*time.Second || config.DrainTimeout != 2*time.Minute {
		t.Fatalf("timeouts %+v", config)
	}
	if config.Session.ReconnectGrace != 45*time.Second || g.PingInterval != 7*time.Second || g.PongTimeout != 3*time.Second {
		t.Fatalf("durations %+v %+v", config.Session, g)
	}
	if g.MinimumAppVersion != "1.4.0" || g.CatalogVersion != "2026.11.01-2" || g.MaxFrameBytes != 2048 || g.WriteQueue != 64 || g.HelloTimeout != time.Second || g.WriteTimeout != 2*time.Second || g.CloseTimeout != 500*time.Millisecond || g.CompressionThreshold != 512 || g.IntentsPerSecond != 20 || g.IntentBurst != 40 {
		t.Fatalf("gateway config %+v", g)
	}
}

func TestConfigReportsEveryBadValueAtOnce(t *testing.T) {
	_, err := ConfigFromEnv(lookup(map[string]string{
		"HD_LOG_LEVEL":       "loud",
		"HD_DRAIN_TIMEOUT":   "soon",
		"HD_WRITE_QUEUE":     "-1",
		"HD_INTENT_BURST":    "many",
		"HD_HELLO_TIMEOUT":   "-5s",
		"HD_MAX_FRAME_BYTES": "1024",
	}))
	if err == nil {
		t.Fatal("bad values were accepted")
	}
	for _, name := range []string{"HD_LOG_LEVEL", "HD_DRAIN_TIMEOUT", "HD_WRITE_QUEUE", "HD_INTENT_BURST", "HD_HELLO_TIMEOUT"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not name %s: %v", name, err)
		}
	}
	if strings.Contains(err.Error(), "HD_MAX_FRAME_BYTES") {
		t.Errorf("a valid value was reported: %v", err)
	}
}
