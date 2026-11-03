package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/adminapi"
	"github.com/Alievs-corp/hypedomino/server/internal/adminsession"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/ratelimit"
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
		"HD_DATABASE_URL":                 "postgres://admin",
		"HD_ADMIN_ENVIRONMENT":            "staging",
		"HD_ADMIN_GOOGLE_DOMAIN":          "example.com",
		"HD_ADMIN_GOOGLE_AUDIENCES":       " web.apps.example , ,other ",
		"HD_ADMIN_SESSION_IDLE":           "20m",
		"HD_ADMIN_STEP_UP":                "5m",
		"HD_ADMIN_APPROVAL_TTL":           "48h",
		"HD_SHOP_DAILY_SLOTS":             "6",
		"HD_SHOP_RESET_HOUR_UTC":          "4",
		"HD_ADMIN_GRANT_APPROVAL_COINS":   "5000",
		"HD_TEMP_BAN_HOURS":               "72",
		"HD_READ_TIMEOUT":                 "30s",
		"HD_ADMIN_RATE_ACCOUNT_BURST":     "80",
		"HD_ADMIN_RATE_SIGNIN_PER_MINUTE": "4",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := config{
		listenAddr: ":8090", logLevel: "info", database: "postgres://admin", audiences: []string{"web.apps.example", "other"}, readLimit: 30 * time.Second, idleLimit: time.Minute,
		api: adminapi.Config{
			Environment: "staging", GoogleDomain: "example.com",
			Session:            adminsession.Config{Idle: 20 * time.Minute, Max: 12 * time.Hour, StepUp: 5 * time.Minute},
			ApprovalTTL:        48 * time.Hour,
			Shop:               shop.Config{DailySlots: 6, WeeklySlots: 2, ResetHourUTC: 4},
			GrantApprovalCoins: 5000,
			TempBan:            72 * time.Hour,
			Limits:             adminapi.Limits{Account: ratelimit.Rule{PerMinute: 120, Burst: 80}, SignIn: ratelimit.Rule{PerMinute: 4, Burst: 10}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config %+v, want %+v", got, want)
	}
}

func TestConfigNamesEveryMissingOrInvalidValue(t *testing.T) {
	_, err := configFromEnv(env(map[string]string{"HD_ADMIN_ENVIRONMENT": "prod", "HD_ADMIN_SESSION_MAX": "0s", "HD_SHOP_WEEKLY_SLOTS": "two", "HD_SHOP_RESET_HOUR_UTC": "24", "HD_ADMIN_GRANT_APPROVAL_COINS": "0", "HD_TEMP_BAN_HOURS": "-1", "HD_ADMIN_RATE_SIGNIN_BURST": "0"}))
	if err == nil {
		t.Fatal("empty environment accepted")
	}
	for _, name := range []string{"HD_DATABASE_URL", "HD_ADMIN_ENVIRONMENT", "HD_ADMIN_GOOGLE_DOMAIN", "HD_ADMIN_GOOGLE_AUDIENCES", "HD_ADMIN_SESSION_MAX", "HD_SHOP_WEEKLY_SLOTS", "HD_SHOP_RESET_HOUR_UTC", "HD_ADMIN_GRANT_APPROVAL_COINS", "HD_TEMP_BAN_HOURS", "HD_ADMIN_RATE_*"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %s", err, name)
		}
	}
}

func TestConfigTakesPromoCodesOnlyWithTheirKey(t *testing.T) {
	base := map[string]string{"HD_DATABASE_URL": "postgres://admin", "HD_ADMIN_ENVIRONMENT": "staging", "HD_ADMIN_GOOGLE_DOMAIN": "example.com", "HD_ADMIN_GOOGLE_AUDIENCES": "web"}
	with := func(extra map[string]string) map[string]string {
		merged := map[string]string{}
		for key, value := range base {
			merged[key] = value
		}
		for key, value := range extra {
			merged[key] = value
		}
		return merged
	}
	got, err := configFromEnv(env(with(map[string]string{"HD_PROMO_KEY": strings.Repeat("k", 32), "HD_PROMO_COINS_APPROVAL_THRESHOLD": "500"})))
	if err != nil || got.api.Promo.CoinsApprovalThreshold != 500 || got.api.Promo.CodeLength != promo.DefaultConfig().CodeLength || got.api.Cosmetics == nil {
		t.Fatalf("config %+v, %v", got.api, err)
	}
	_, err = configFromEnv(env(with(map[string]string{"HD_PROMO_KEY": "short", "HD_PROMO_CODE_LENGTH": "3", "HD_COSMETICS_MANIFEST": "missing.json"})))
	for _, name := range []string{"HD_PROMO_KEY", "HD_PROMO_*", "HD_COSMETICS_MANIFEST"} {
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("error %v does not name %s", err, name)
		}
	}
}
