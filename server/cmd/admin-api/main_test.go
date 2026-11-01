package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/adminsession"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
)

func env(values map[string]string) lookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func TestConfigReadsTheAdminEnvironment(t *testing.T) {
	got, err := configFromEnv(env(map[string]string{
		"HD_DATABASE_URL":           "postgres://admin",
		"HD_ADMIN_ENVIRONMENT":      "staging",
		"HD_ADMIN_GOOGLE_DOMAIN":    "example.com",
		"HD_ADMIN_GOOGLE_AUDIENCES": " web.apps.example , ,other ",
		"HD_ADMIN_SESSION_IDLE":     "20m",
		"HD_ADMIN_STEP_UP":          "5m",
		"HD_ADMIN_APPROVAL_TTL":     "48h",
		"HD_SHOP_DAILY_SLOTS":       "6",
		"HD_SHOP_RESET_HOUR_UTC":    "4",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := config{
		listenAddr: ":8090", logLevel: "info", database: "postgres://admin", audiences: []string{"web.apps.example", "other"},
		api: adminapi.Config{
			Environment: "staging", GoogleDomain: "example.com",
			Session:     adminsession.Config{Idle: 20 * time.Minute, Max: 12 * time.Hour, StepUp: 5 * time.Minute},
			ApprovalTTL: 48 * time.Hour,
			Shop:        shop.Config{DailySlots: 6, WeeklySlots: 2, ResetHourUTC: 4},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config %+v, want %+v", got, want)
	}
}

func TestConfigNamesEveryMissingOrInvalidValue(t *testing.T) {
	_, err := configFromEnv(env(map[string]string{"HD_ADMIN_ENVIRONMENT": "prod", "HD_ADMIN_SESSION_MAX": "0s", "HD_SHOP_WEEKLY_SLOTS": "two", "HD_SHOP_RESET_HOUR_UTC": "24"}))
	if err == nil {
		t.Fatal("empty environment accepted")
	}
	for _, name := range []string{"HD_DATABASE_URL", "HD_ADMIN_ENVIRONMENT", "HD_ADMIN_GOOGLE_DOMAIN", "HD_ADMIN_GOOGLE_AUDIENCES", "HD_ADMIN_SESSION_MAX", "HD_SHOP_WEEKLY_SLOTS", "HD_SHOP_RESET_HOUR_UTC"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %s", err, name)
		}
	}
}
