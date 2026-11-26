package contentgate_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/contentgate"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/medal"
	"github.com/Alievs-corp/hypedomino/server/internal/mission"
	"github.com/Alievs-corp/hypedomino/server/internal/pass"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/reward"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
)

var (
	launch      = time.Date(2027, 1, 15, 12, 0, 0, 0, time.UTC)
	seasonStart = launch.Add(-time.Hour)
	seasonEnd   = launch.AddDate(0, 1, 0)
	tiers       = []string{"common", "rare", "epic", "legendary"}
)

const rewardItem = "frame.reward"

type fixture struct {
	items            []map[string]any
	medals           *medal.Catalog
	missions         *mission.Catalog
	passConfig       *pass.Snapshot
	ranking          *rating.Snapshot
	seasons          []rating.Season
	templates        []promo.Template
	unreadablePromos bool
	thresholds       []contentgate.Threshold
	unpublished      []string
	withoutStatic    []string
	withoutName      map[gamestrings.Language][]string
	now              time.Time
}

func item(id string, rarity string, animated bool, sources ...cosmetics.Source) map[string]any {
	fields := map[string]any{
		"id": id, "type": cosmetics.TypeOf(id), "revision": 1, "nameKey": cosmetics.NameKeyOf(id), "rarityId": rarity,
		"animated": animated, "staticAddress": cosmetics.AddressOf(id, "static"), "availability": cosmetics.Live,
		"sources": sources, "artSource": "handmade", "culturalReview": "approved",
	}
	if animated {
		fields["animationAddress"] = cosmetics.AddressOf(id, "anim")
	}
	if cosmetics.TypeOf(id) == cosmetics.TypeEmoji {
		fields["group"] = "faces"
	}
	return fields
}

func complete() *fixture {
	byDefault := cosmetics.Source{Type: cosmetics.SourceDefault}
	event := cosmetics.Source{Type: "event", Ref: "launch_week"}
	f := &fixture{now: launch.Add(-48 * time.Hour), withoutName: map[gamestrings.Language][]string{}}
	for _, id := range []string{"avatar.starter", "frame.starter", "table.starter", "tile.starter", "emoji.starter"} {
		f.items = append(f.items, item(id, "common", false, byDefault))
	}
	for _, tier := range tiers {
		f.items = append(f.items,
			item("avatar.anim_"+tier, tier, true, event),
			item("avatar.still_"+tier, tier, false, event),
			item("frame."+tier, tier, false, event),
			item("table."+tier, tier, false, event),
		)
	}
	f.items = append(f.items, item(rewardItem, "rare", false,
		cosmetics.Source{Type: mission.SourceMission, Ref: "mission.first_win"},
		pass.TrackSource("season_1", pass.Free, 1),
		cosmetics.Source{Type: promo.SourceType, Ref: "campaign.launch"},
	))
	f.medals = medal.Default()
	f.medals.Medals = []medal.Medal{{ID: "medal.first", Revision: 1, NameKey: "medal.first.name", DescriptionKey: "medal.first.description", RarityID: "rare",
		StaticAddress: medal.StaticAddress("medal.first"), Sources: []cosmetics.Source{{Type: medal.SourceMission, Ref: "mission.first_win"}}, Availability: cosmetics.Live}}
	f.missions = mission.Default()
	for _, missionType := range []string{mission.Daily, mission.Weekly} {
		for slot := 1; slot <= len(f.missions.Slots(missionType)); slot++ {
			for _, variant := range []string{"a", "b"} {
				slug := missionType + "_" + string(rune('0'+slot)) + "_" + variant
				f.missions.Missions = append(f.missions.Missions, mission.Mission{ID: "mission." + slug, Revision: 1, Type: missionType, NameKey: "mission." + slug + ".name",
					Event: mission.MatchWon, Target: 1, Slot: slot, Weight: 1, Availability: mission.Live})
			}
		}
	}
	f.missions.Missions = append(f.missions.Missions, mission.Mission{ID: "mission.first_win", Revision: 1, Type: mission.Achievement, NameKey: "mission.first_win.name",
		Event: mission.MatchWon, Target: 1, Bundle: &reward.Bundle{ID: "bundle.first_win", Revision: 1, Coins: 10, Cosmetics: []string{rewardItem}, Medals: []string{"medal.first"}}, Availability: mission.Live})
	f.passConfig = pass.Default()
	f.passConfig.PremiumPlusLevels = 0
	f.passConfig.Seasons = []pass.Season{{ID: "season_1", Start: seasonStart, End: seasonEnd, Curve: pass.Curve{Levels: 1, XpBase: 100, XpStep: 10, XpMax: 200},
		Free:    map[int]reward.Bundle{1: {ID: "bundle.free_1", Revision: 1, Coins: 5, Cosmetics: []string{rewardItem}}},
		Premium: map[int]reward.Bundle{1: {ID: "bundle.premium_1", Revision: 1, Crystals: 5}}}}
	f.ranking = rating.Default()
	f.ranking.LeaderboardRegions = map[string][]string{"caucasus": {"AZ"}}
	f.ranking.Modes.OneVsOne.RankRewards = rankRewards("1v1")
	f.ranking.Modes.TwoVsTwo.RankRewards = rankRewards("2v2")
	f.seasons = []rating.Season{{ID: "season.launch", NameKey: "season.launch.name", Start: seasonStart, End: seasonEnd}}
	f.templates = []promo.Template{
		{ID: "promo.welcome", Campaign: "campaign.launch", Kind: "multi_use", Bundle: &reward.Bundle{ID: "bundle.welcome", Revision: 1, Coins: 50, Cosmetics: []string{rewardItem}}},
		{ID: "promo.creator", Campaign: "campaign.creator", Kind: "multi_use", Bundle: &reward.Bundle{ID: "bundle.creator", Revision: 1, Coins: 10}},
		{ID: "promo.compensation", Campaign: "campaign.compensation", Kind: "single_use"},
	}
	for _, key := range []string{"ContentGateMinAvatarsAnimated", "ContentGateMinAvatarsStatic", "ContentGateMinFrames", "ContentGateMinTables", "ContentGateMinEmoji", "ContentGateMinMedals"} {
		f.thresholds = append(f.thresholds, contentgate.Threshold{Key: key, Value: 1, Default: 1})
	}
	return f
}

func rankRewards(mode string) map[string]reward.Bundle {
	bundles := map[string]reward.Bundle{}
	for _, tier := range rating.Tiers() {
		bundles[tier] = reward.Bundle{ID: "bundle.rank_" + mode[:1] + "_" + strings.ToLower(tier), Revision: 1, Coins: 1}
	}
	return bundles
}

func (f *fixture) setItem(id, field string, value any) {
	for _, fields := range f.items {
		if fields["id"] == id {
			fields[field] = value
		}
	}
}

func (f *fixture) setThreshold(key string, value int) {
	for i := range f.thresholds {
		if f.thresholds[i].Key == key {
			f.thresholds[i].Value = value
		}
	}
}

func encode(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func (f *fixture) inputs(t *testing.T) contentgate.Inputs {
	t.Helper()
	snapshots := map[string][]byte{
		cosmetics.Domain: encode(t, map[string]any{"schemaVersion": 1, "manifestVersion": 1, "items": f.items}),
		medal.Domain:     encode(t, f.medals),
		mission.Domain:   encode(t, f.missions),
		pass.Domain:      encode(t, f.passConfig),
		rating.Domain:    encode(t, f.ranking),
	}
	for _, domain := range f.unpublished {
		delete(snapshots, domain)
	}
	statics := map[string]bool{}
	var keys []string
	for _, fields := range f.items {
		id := fields["id"].(string)
		statics[id] = !slices.Contains(f.withoutStatic, id)
		keys = append(keys, cosmetics.NameKeyOf(id))
	}
	for _, m := range f.medals.Medals {
		statics[m.ID] = !slices.Contains(f.withoutStatic, m.ID)
		keys = append(keys, m.NameKey, m.DescriptionKey)
	}
	for _, m := range f.missions.Missions {
		keys = append(keys, m.NameKey)
	}
	for _, season := range f.seasons {
		keys = append(keys, season.NameKey)
	}
	bundles := map[gamestrings.Language]map[string]bool{}
	for _, language := range gamestrings.Languages {
		bundles[language] = map[string]bool{}
		for _, key := range keys {
			bundles[language][key] = !slices.Contains(f.withoutName[language], key)
		}
	}
	return contentgate.Inputs{LaunchAt: launch, Now: f.now, Snapshots: snapshots, Shop: shop.DefaultConfig(), Thresholds: f.thresholds, RankedSeasons: f.seasons,
		Templates: f.templates, TemplatesReadable: !f.unreadablePromos, PassedStatics: statics, Bundles: bundles}
}

func checkOf(t *testing.T, result contentgate.Result, id string) contentgate.Check {
	t.Helper()
	for _, check := range result.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("no check %s in %v", id, result.Checks)
	return contentgate.Check{}
}

func failsWith(check contentgate.Check, rule, detail string) bool {
	return slices.ContainsFunc(check.Failures, func(f contentgate.Failure) bool { return f.Rule == rule && strings.Contains(f.Detail, detail) })
}

func TestACompleteCatalogueFailsNoCheckAndLeavesOnlyTheUncheckableRules(t *testing.T) {
	result := contentgate.Run(complete().inputs(t))
	want := []string{"CG-2", "CG-3", "CG-4", "CG-5", "CG-6", "CG-7", "CG-8", "CG-9", "CG-10", "CG-11", "CG-12", "CG-13"}
	var ids []string
	for _, check := range result.Checks {
		ids = append(ids, check.ID)
		if len(check.Failures) != 0 {
			t.Errorf("%s fails: %v", check.ID, check.Failures)
		}
	}
	if !slices.Equal(ids, want) {
		t.Fatalf("checks %v", ids)
	}
	if checkOf(t, result, "CG-13").Result != contentgate.ResultPass {
		t.Fatalf("thresholds %v", checkOf(t, result, "CG-13"))
	}
	for _, id := range []string{"CG-2", "CG-8", "CG-10", "CG-12"} {
		if check := checkOf(t, result, id); check.Result != contentgate.ResultUnchecked || len(check.Unchecked) == 0 {
			t.Fatalf("%s should stay unchecked: %v", id, check)
		}
	}
	if result.Verdict != contentgate.VerdictRed {
		t.Fatalf("an unchecked rule never gives green: %s", result.Verdict)
	}
	if count := checkOf(t, result, "CG-2").Count; count == nil || *count != 4 {
		t.Fatalf("animated avatars counted %v", count)
	}
	if result.Info.Ready["avatar_static"]["common"] != 2 || result.Info.Ready[contentgate.MedalCategory]["rare"] != 1 || len(result.Info.LiveNotReady) != 0 {
		t.Fatalf("info %v", result.Info)
	}
}

func TestEachCheckFailsOnOneMissingCondition(t *testing.T) {
	cases := []struct {
		name   string
		change func(f *fixture)
		check  string
		rule   string
		detail string
	}{
		{"CG-2 below its minimum", func(f *fixture) { f.setThreshold("ContentGateMinAvatarsAnimated", 5) }, "CG-2", "MC-2", "4 ready"},
		{"CG-2 without a legendary animated avatar", func(f *fixture) { f.setItem("avatar.anim_legendary", "availability", cosmetics.Retired) }, "CG-2", "MC-2", "no ready legendary"},
		{"CG-3 static avatar without a passed upload", func(f *fixture) { f.withoutStatic = []string{"avatar.still_epic"} }, "CG-3", "MC-3", "no ready epic"},
		{"CG-4 frame without a Turkish name", func(f *fixture) {
			f.withoutName[gamestrings.Turkish] = []string{cosmetics.NameKeyOf("frame.rare")}
			f.setItem(rewardItem, "rarityId", "common")
		}, "CG-4", "MC-4", "no ready rare"},
		{"CG-5 below its minimum", func(f *fixture) { f.setThreshold("ContentGateMinTables", 6) }, "CG-5", "MC-5", "5 ready"},
		{"CG-6 below its minimum", func(f *fixture) { f.setThreshold("ContentGateMinEmoji", 2) }, "CG-6", "MC-6", "1 ready"},
		{"CG-7 without published medals", func(f *fixture) { f.unpublished = []string{medal.Domain} }, "CG-7", "MR-1", "not published"},
		{"CG-7 below its minimum", func(f *fixture) { f.setThreshold("ContentGateMinMedals", 2) }, "CG-7", "MC-7", "1 ready"},
		{"CG-8 item failing its manifest rules", func(f *fixture) { f.setItem("frame.epic", "culturalReview", "pending") }, "CG-8", "CR-2", "culturalReview"},
		{"CG-8 item without a passed upload", func(f *fixture) { f.withoutStatic = []string{"table.rare"} }, "CG-8", "CR-5", "static upload"},
		{"CG-8 item without a Russian name", func(f *fixture) {
			f.withoutName[gamestrings.Russian] = []string{cosmetics.NameKeyOf("emoji.starter")}
		}, "CG-8", "CR-6", "missing in ru"},
		{"CG-8 grant the item does not list", func(f *fixture) {
			f.setItem(rewardItem, "sources", []cosmetics.Source{{Type: "event", Ref: "launch_week"}})
		}, "CG-8", "CR-2", "do not list"},
		{"CG-8 mythic medal", func(f *fixture) { f.medals.Medals[0].RarityID = promo.MythicRarity }, "CG-8", "MR-2", "mythic"},
		{"CG-8 medal from a retired mission", func(f *fixture) { f.missions.Missions[len(f.missions.Missions)-1].Availability = mission.Retired }, "CG-8", "MR-3", "mission.first_win"},
		{"CG-8 medal without a passed upload", func(f *fixture) { f.withoutStatic = []string{"medal.first"} }, "CG-8", "MR-4", "static upload"},
		{"CG-8 medal without an Azerbaijani description", func(f *fixture) {
			f.withoutName[gamestrings.Azerbaijani] = []string{"medal.first.description"}
		}, "CG-8", "MR-5", "missing in az"},
		{"CG-9 without published missions", func(f *fixture) { f.unpublished = []string{mission.Domain} }, "CG-9", "MS-1", "not published"},
		{"CG-9 slot with one live mission", func(f *fixture) { f.missions.Missions[0].Availability = mission.Retired }, "CG-9", "MS-2", "1 live missions, 2 needed"},
		{"CG-9 slot short of its rerolls", func(f *fixture) { f.missions.DailyRerolls = 2 }, "CG-9", "MS-2", "2 live missions, 3 needed"},
		{"CG-9 mission without a Turkish name", func(f *fixture) {
			f.withoutName[gamestrings.Turkish] = []string{"mission.weekly_2_b.name"}
		}, "CG-9", "MS-3", "mission.weekly_2_b.name"},
		{"CG-9 achievement granting an item that is not ready", func(f *fixture) { f.withoutStatic = []string{rewardItem} }, "CG-9", "MS-4", rewardItem},
		{"CG-10 without a season at launch", func(f *fixture) {
			f.passConfig.Seasons[0].Start, f.passConfig.Seasons[0].End = seasonEnd, seasonEnd.AddDate(0, 1, 0)
		}, "CG-10", "CG-10", "no pass season"},
		{"CG-10 without published pass seasons", func(f *fixture) { f.unpublished = []string{pass.Domain} }, "CG-10", "PS-1", "not published"},
		{"CG-10 track item that does not list the track", func(f *fixture) {
			f.setItem(rewardItem, "sources", []cosmetics.Source{{Type: mission.SourceMission, Ref: "mission.first_win"}, {Type: promo.SourceType, Ref: "campaign.launch"}})
		}, "CG-10", "PS-1", "do not list"},
		{"CG-10 window apart from the ranked season", func(f *fixture) { f.passConfig.Seasons[0].End = seasonEnd.Add(time.Hour) }, "CG-10", "PS-2", "season.launch"},
		{"CG-10 track item that is not ready", func(f *fixture) { f.withoutStatic = []string{rewardItem} }, "CG-10", "PS-3", rewardItem},
		{"CG-11 without a ranked season at launch", func(f *fixture) { f.seasons = nil }, "CG-11", "CG-11", "no ranked season"},
		{"CG-11 season closed", func(f *fixture) {
			closed := launch
			f.seasons[0].Closed1v1, f.seasons[0].Closed2v2 = &closed, &closed
		}, "CG-11", "RS-1", "closed"},
		{"CG-11 without a published ranking config", func(f *fixture) { f.unpublished = []string{rating.Domain} }, "CG-11", "RS-2", "not published"},
		{"CG-11 tier without a rank reward", func(f *fixture) { delete(f.ranking.Modes.TwoVsTwo.RankRewards, rating.TierMaster) }, "CG-11", "RS-3", "no rankRewards"},
		{"CG-11 rank reward granting an item that is not ready", func(f *fixture) {
			bundle := f.ranking.Modes.OneVsOne.RankRewards["Gold"]
			bundle.Cosmetics = []string{"frame.missing"}
			f.ranking.Modes.OneVsOne.RankRewards["Gold"] = bundle
		}, "CG-11", "RS-3", "frame.missing"},
		{"CG-11 without leaderboard regions", func(f *fixture) { f.ranking.LeaderboardRegions = nil }, "CG-11", "RS-4", "leaderboardRegions"},
		{"CG-11 season without a Russian name", func(f *fixture) {
			f.withoutName[gamestrings.Russian] = []string{"season.launch.name"}
		}, "CG-11", "RS-5", "missing in ru"},
		{"CG-12 templates unreadable", func(f *fixture) { f.unreadablePromos = true }, "CG-12", "PT-1", "HD_PROMO_KEY"},
		{"CG-12 two templates", func(f *fixture) { f.templates = f.templates[:2] }, "CG-12", "PT-1", "2 templates"},
		{"CG-12 template bundle with crystals", func(f *fixture) { f.templates[1].Bundle.Crystals = 5 }, "CG-12", "PT-2", "crystals"},
		{"CG-12 template item without the promo source", func(f *fixture) { f.templates[0].Campaign = "campaign.other" }, "CG-12", "PT-3", "promo source campaign.other"},
		{"CG-13 threshold below its default", func(f *fixture) {
			f.thresholds[5].Default = 40
			f.thresholds[5].Value = 39
		}, "CG-13", "CG-13", "below the repo default 40"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := complete()
			c.change(f)
			result := contentgate.Run(f.inputs(t))
			check := checkOf(t, result, c.check)
			if check.Result != contentgate.ResultFail || !failsWith(check, c.rule, c.detail) {
				t.Fatalf("%s %s: %v", check.ID, check.Result, check.Failures)
			}
			if result.Verdict != contentgate.VerdictRed {
				t.Fatalf("verdict %s", result.Verdict)
			}
		})
	}
}

func TestAnUnpublishedManifestFallsBackToTheBuiltInOneWhoseItemsLackUploads(t *testing.T) {
	f := complete()
	f.unpublished = []string{cosmetics.Domain}
	inputs := f.inputs(t)
	inputs.PassedStatics = map[string]bool{}
	result := contentgate.Run(inputs)
	if check := checkOf(t, result, "CG-8"); !failsWith(check, "CR-5", "") || result.Info.LiveNotReady[cosmetics.TypeTile] != 1 {
		t.Fatalf("builtin manifest %v %v", check.Failures, result.Info)
	}
}
