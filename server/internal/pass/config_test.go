package pass

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/reward"
)

func example(t *testing.T) *Snapshot {
	t.Helper()
	raw, err := os.ReadFile("../../../docs/pass/examples/pass.json")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestTheExampleSeasonFollowsTheSampleOfTheSpec(t *testing.T) {
	season, ok := example(t).Season("season_1")
	if !ok {
		t.Fatal("season_1 missing")
	}
	if season.Curve != defaultCurve {
		t.Fatalf("curve = %+v", season.Curve)
	}
	if len(season.Premium) != 100 || len(season.Free) != 60 {
		t.Fatalf("premium %d, free %d levels", len(season.Premium), len(season.Free))
	}
	if got := season.Premium[100].Cosmetics; len(got) != 1 || got[0] != "character.s1_sea_queen" {
		t.Fatalf("premium 100 = %v", got)
	}
}

func TestPublishRulesRefuseBrokenSeasons(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Snapshot)
	}{
		{"free track over budget", func(s *Snapshot) { s.FreeTrackCoins = 1499 }},
		{"premium track over budget", func(s *Snapshot) { s.PremiumTrackCoins = 2999 }},
		{"premium level without bundle", func(s *Snapshot) { delete(s.Seasons[0].Premium, 42) }},
		{"level outside the table", func(s *Snapshot) { s.Seasons[0].Free[101] = coinBundle("bundle.extra", 1) }},
		{"track bundle with a medal", func(s *Snapshot) {
			b := s.Seasons[0].Free[1]
			b.Medals = []string{"medal.wins_50"}
			s.Seasons[0].Free[1] = b
		}},
		{"free track bundle with crystals", func(s *Snapshot) {
			b := s.Seasons[0].Free[1]
			b.Crystals = 10
			s.Seasons[0].Free[1] = b
		}},
		{"track bundle with pass xp", func(s *Snapshot) {
			b := s.Seasons[0].Premium[2]
			b.PassXP = 100
			s.Seasons[0].Premium[2] = b
		}},
		{"levels not above premium plus levels", func(s *Snapshot) { s.PremiumPlusLevels = 100 }},
		{"curve value below one", func(s *Snapshot) { s.Seasons[0].Curve.XpStep = 0 }},
		{"xp max below xp base", func(s *Snapshot) { s.Seasons[0].Curve.XpMax = 299 }},
		{"bad season id", func(s *Snapshot) { s.Seasons[0].ID = "spring" }},
		{"end before start", func(s *Snapshot) { s.Seasons[0].End = s.Seasons[0].Start }},
		{"overlapping seasons", func(s *Snapshot) {
			next := s.Seasons[0]
			next.ID = "season_2"
			next.Start = s.Seasons[0].End.Add(-time.Hour)
			next.End = next.Start.Add(90 * 24 * time.Hour)
			s.Seasons = append(s.Seasons, next)
		}},
		{"season listed twice", func(s *Snapshot) {
			again := s.Seasons[0]
			again.Start, again.End = again.End, again.End.Add(time.Hour)
			s.Seasons = append(s.Seasons, again)
		}},
		{"unknown schema", func(s *Snapshot) { s.Schema = 2 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			snapshot := example(t)
			c.change(snapshot)
			if err := snapshot.Validate(); !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("Validate() = %v", err)
			}
		})
	}
}

func TestAGapBetweenSeasonsIsAllowed(t *testing.T) {
	snapshot := example(t)
	next := snapshot.Seasons[0]
	next.ID = "season_2"
	next.Start = snapshot.Seasons[0].End.Add(24 * time.Hour)
	next.End = next.Start.Add(90 * 24 * time.Hour)
	snapshot.Seasons = append(snapshot.Seasons, next)
	if err := snapshot.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, ok := snapshot.Active(snapshot.Seasons[0].End.Add(time.Hour)); ok {
		t.Fatal("a season is active inside the gap")
	}
	if active, ok := snapshot.Active(next.Start); !ok || active.ID != "season_2" {
		t.Fatalf("Active(start of season_2) = %s, %v", active.ID, ok)
	}
}

func TestARunningSeasonKeepsItsLevelsCurveAndBundles(t *testing.T) {
	running := example(t).Seasons[0].Start.Add(time.Hour)
	cases := []struct {
		name   string
		change func(*Snapshot)
		frozen bool
	}{
		{"curve changed", func(s *Snapshot) { s.Seasons[0].Curve.XpMax = 900 }, true},
		{"bundle moved", func(s *Snapshot) { s.Seasons[0].Free[2] = coinBundle("bundle.moved", 30) }, true},
		{"season removed", func(s *Snapshot) { s.Seasons = nil }, true},
		{"window extended", func(s *Snapshot) { s.Seasons[0].End = s.Seasons[0].End.Add(24 * time.Hour) }, false},
		{"next season added", func(s *Snapshot) {
			next := s.Seasons[0]
			next.ID = "season_2"
			next.Start = s.Seasons[0].End
			next.End = next.Start.Add(90 * 24 * time.Hour)
			s.Seasons = append(s.Seasons, next)
		}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			next := example(t)
			c.change(next)
			err := CheckFrozen(example(t), next, running)
			if errors.Is(err, ErrFrozen) != c.frozen {
				t.Fatalf("CheckFrozen() = %v", err)
			}
		})
	}
}

func TestASeasonThatHasNotStartedMayChange(t *testing.T) {
	published := example(t)
	next := example(t)
	next.Seasons[0].Curve.XpMax = 900
	if err := CheckFrozen(published, next, published.Seasons[0].Start.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func coinBundle(id string, coins int64) reward.Bundle {
	return reward.Bundle{ID: id, Revision: 1, Coins: coins}
}
