package dailylogin

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/reward"
)

func TestTheExampleIsTheDefaultCycle(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/economy/examples/daily_login.json")
	if err != nil {
		t.Fatal(err)
	}
	calendar, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for day := 1; day <= 7; day++ {
		total += calendar.Reward(day, 3).Coins
	}
	if total != 450 || len(calendar.Days) != len(Default().Days) {
		t.Fatalf("cycle pays %d over %d days", total, len(calendar.Days))
	}
}

func TestPublishRulesRefuseBrokenCalendars(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Calendar)
	}{
		{"wrong schema", func(c *Calendar) { c.Schema = 2 }},
		{"six days", func(c *Calendar) { c.Days = c.Days[:6] }},
		{"a day without a reward", func(c *Calendar) { c.Days[0].Coins = 0 }},
		{"coins and a bundle", func(c *Calendar) {
			c.Days[0].Bundle = &reward.Bundle{ID: "bundle.login", Revision: 1, Coins: 10}
		}},
		{"a bundle with a cosmetic", func(c *Calendar) {
			c.Days[6] = Day{Bundle: &reward.Bundle{ID: "bundle.login", Revision: 1, Cosmetics: []string{"emoji.laugh_spin"}}}
		}},
		{"a bundle with a medal", func(c *Calendar) {
			c.Days[6] = Day{Bundle: &reward.Bundle{ID: "bundle.login", Revision: 1, Medals: []string{"medal.wins_50"}}}
		}},
		{"a broken bundle", func(c *Calendar) { c.Days[6] = Day{Bundle: &reward.Bundle{ID: "login", Revision: 1, Coins: 10}} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			calendar := Default()
			c.change(calendar)
			if err := calendar.Validate(); !errors.Is(err, ErrInvalidCalendar) {
				t.Fatalf("Validate() = %v", err)
			}
		})
	}
}

func TestAThirtyDayCalendarWithABundlePublishes(t *testing.T) {
	calendar := &Calendar{Schema: Schema}
	for day := 1; day < 30; day++ {
		calendar.Days = append(calendar.Days, Day{Coins: 10})
	}
	calendar.Days = append(calendar.Days, Day{Bundle: &reward.Bundle{ID: "bundle.login_month", Revision: 1, Coins: 100, PassXP: 300}})
	if err := calendar.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := calendar.Reward(30, 5); got.ID != "bundle.login_month" || got.PassXP != 300 {
		t.Fatalf("day 30 = %+v", got)
	}
}

func TestTheCycleRestartsAfterItsLastDay(t *testing.T) {
	calendar := Default()
	cases := []struct {
		claims int64
		day    int
	}{{0, 1}, {1, 2}, {6, 7}, {7, 1}, {15, 2}}
	for _, c := range cases {
		if got := calendar.Next(c.claims); got != c.day {
			t.Errorf("after %d claims: day %d, want %d", c.claims, got, c.day)
		}
	}
}

func TestTheEconomyDayIsServerUtc(t *testing.T) {
	baku := time.FixedZone("AZT", 4*3600)
	cases := []struct {
		at   time.Time
		want string
	}{
		{time.Date(2026, 11, 3, 23, 59, 59, 0, time.UTC), "2026-11-03"},
		{time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC), "2026-11-04"},
		{time.Date(2026, 11, 4, 3, 30, 0, 0, baku), "2026-11-03"},
		{time.Date(2026, 11, 4, 4, 0, 0, 0, baku), "2026-11-04"},
	}
	for _, c := range cases {
		if got := EconomyDay(c.at); got != c.want {
			t.Errorf("%v: %s, want %s", c.at, got, c.want)
		}
	}
}
