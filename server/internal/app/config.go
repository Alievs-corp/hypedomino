package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/cluster"
	"github.com/Alievs-corp/hypedomino/server/internal/gateway"
	"github.com/Alievs-corp/hypedomino/server/internal/session"
)

type Config struct {
	ListenAddr      string
	LogLevel        string
	Origins         []string
	ShutdownTimeout time.Duration
	DrainTimeout    time.Duration
	Gateway         gateway.Config
	Session         session.Config
	Redis           string
	Cluster         cluster.Config
	Relay           cluster.RelayConfig
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:      ":8080",
		LogLevel:        "info",
		ShutdownTimeout: 10 * time.Second,
		DrainTimeout:    15 * time.Minute,
		Gateway:         gateway.DefaultConfig(),
		Session:         session.DefaultConfig(),
		Cluster:         defaultCluster(),
		Relay:           cluster.DefaultRelayConfig(),
	}
}

func defaultCluster() cluster.Config {
	config := cluster.DefaultConfig()
	config.NodeID, _ = os.Hostname()
	return config
}

type Lookup func(name string) (string, bool)

type envReader struct {
	lookup Lookup
	errs   []error
}

func (r *envReader) text(name string, value *string) {
	if raw, ok := r.lookup(name); ok {
		*value = raw
	}
}

func (r *envReader) list(name string, value *[]string) {
	raw, ok := r.lookup(name)
	if !ok {
		return
	}
	*value = nil
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			*value = append(*value, item)
		}
	}
}

func (r *envReader) integer(name string, value *int) {
	raw, ok := r.lookup(name)
	if !ok {
		return
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a non-negative integer", name, raw))
		return
	}
	*value = parsed
}

func (r *envReader) fraction(name string, value *float64) {
	raw, ok := r.lookup(name)
	if !ok {
		return
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || parsed < 0 || parsed > 1 {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a fraction between 0 and 1", name, raw))
		return
	}
	*value = parsed
}

func (r *envReader) duration(name string, value *time.Duration) {
	raw, ok := r.lookup(name)
	if !ok {
		return
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed < 0 {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a non-negative duration", name, raw))
		return
	}
	*value = parsed
}

func ConfigFromEnv(lookup Lookup) (Config, error) {
	config := DefaultConfig()
	r := &envReader{lookup: lookup}
	r.text("HD_LISTEN_ADDR", &config.ListenAddr)
	r.text("HD_LOG_LEVEL", &config.LogLevel)
	r.list("HD_ORIGINS", &config.Origins)
	r.duration("HD_SHUTDOWN_TIMEOUT", &config.ShutdownTimeout)
	r.duration("HD_DRAIN_TIMEOUT", &config.DrainTimeout)
	r.text("HD_MIN_APP_VERSION", &config.Gateway.MinimumAppVersion)
	r.text("HD_CATALOG_VERSION", &config.Gateway.CatalogVersion)
	r.integer("HD_MAX_FRAME_BYTES", &config.Gateway.MaxFrameBytes)
	r.integer("HD_WRITE_QUEUE", &config.Gateway.WriteQueue)
	r.duration("HD_HELLO_TIMEOUT", &config.Gateway.HelloTimeout)
	r.duration("HD_WRITE_TIMEOUT", &config.Gateway.WriteTimeout)
	r.duration("HD_CLOSE_TIMEOUT", &config.Gateway.CloseTimeout)
	r.integer("HD_COMPRESSION_THRESHOLD", &config.Gateway.CompressionThreshold)
	r.integer("HD_INTENTS_PER_SECOND", &config.Gateway.IntentsPerSecond)
	r.integer("HD_INTENT_BURST", &config.Gateway.IntentBurst)
	r.duration("HD_PING_INTERVAL", &config.Gateway.PingInterval)
	r.duration("HD_PONG_TIMEOUT", &config.Gateway.PongTimeout)
	r.duration("HD_RECONNECT_GRACE", &config.Session.ReconnectGrace)
	r.text("HD_REDIS_URL", &config.Redis)
	r.text("HD_NODE_ID", &config.Cluster.NodeID)
	r.text("HD_NODE_ADDR", &config.Cluster.Addr)
	r.integer("HD_NODE_CAPACITY", &config.Cluster.Capacity)
	r.duration("HD_NODE_HEARTBEAT", &config.Cluster.Heartbeat)
	r.duration("HD_NODE_TTL", &config.Cluster.NodeTTL)
	r.duration("HD_ROUTE_TTL", &config.Cluster.RouteTTL)
	r.fraction("HD_SCALE_OUT_AT", &config.Cluster.ScaleOutAt)
	r.fraction("HD_SCALE_IN_AT", &config.Cluster.ScaleInAt)
	r.duration("HD_RELAY_DIAL_TIMEOUT", &config.Relay.DialTimeout)
	if config.Redis != "" {
		if err := config.Cluster.Validate(); err != nil {
			r.errs = append(r.errs, err)
		}
	}
	if _, ok := parseLevel(config.LogLevel); !ok {
		r.errs = append(r.errs, fmt.Errorf("HD_LOG_LEVEL: %q is not debug, info, warn or error", config.LogLevel))
	}
	return config, errors.Join(r.errs...)
}
