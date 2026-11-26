package contentgate

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/liveops"
	"github.com/Alievs-corp/hypedomino/server/internal/medal"
	"github.com/Alievs-corp/hypedomino/server/internal/mission"
	"github.com/Alievs-corp/hypedomino/server/internal/pass"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

const (
	ResultPass      = "pass"
	ResultFail      = "fail"
	ResultUnchecked = "unchecked"
	VerdictGreen    = "green"
	VerdictRed      = "red"
	ThresholdPrefix = "ContentGateMin"
	MedalCategory   = "medal"
	LaunchTemplates = 3
)

var Domains = []string{cosmetics.Domain, medal.Domain, mission.Domain, pass.Domain, rating.Domain, shop.Domain, gamestrings.Domain, liveops.Domain}

type Failure struct {
	Object string `json:"object"`
	Rule   string `json:"rule"`
	Detail string `json:"detail,omitempty"`
}

type Gap struct {
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

type Check struct {
	ID        string    `json:"id"`
	Result    string    `json:"result"`
	Count     *int      `json:"count,omitempty"`
	Minimum   *int      `json:"minimum,omitempty"`
	Failures  []Failure `json:"failures"`
	Unchecked []Gap     `json:"unchecked"`
}

type Threshold struct {
	Key           string `json:"key"`
	Value         int    `json:"value"`
	Default       int    `json:"default"`
	ConfigVersion *int64 `json:"config_version"`
}

type Info struct {
	Ready        map[string]map[string]int `json:"ready"`
	LiveNotReady map[string]int            `json:"live_not_ready"`
}

type Result struct {
	Checks  []Check `json:"checks"`
	Info    Info    `json:"info"`
	Verdict string  `json:"verdict"`
}

type Inputs struct {
	LaunchAt          time.Time
	Now               time.Time
	Snapshots         map[string][]byte
	Shop              shop.Config
	Thresholds        []Threshold
	RankedSeasons     []rating.Season
	Templates         []promo.Template
	TemplatesReadable bool
	PassedStatics     map[string]bool
	Bundles           map[gamestrings.Language]map[string]bool
}

var (
	CosmeticGaps = []Gap{
		{Rule: "CR-4", Reason: "the admin API does not read the Addressables catalog active in production, so no address is resolved in it (V-17)"},
		{Rule: "CR-5", Reason: "only the static upload has server checks; animation and table layer files have no upload slot yet"},
		{Rule: "CR-2", Reason: "V-15 from a source to its record (every ref resolving to an existing record) has no validator yet; the record-to-item direction is checked"},
	}
	MedalGaps = []Gap{
		{Rule: "MR-3", Reason: "live events are not modelled yet, so an event mission's event is not checked for being published"},
		{Rule: "MR-4", Reason: "the admin API does not read the Addressables catalog active in production; only the static upload checks are read"},
		{Rule: "cultural_review", Reason: "the medal definition has no cultural review field yet (#141)"},
	}
	PassGaps = []Gap{
		{Rule: "PS-4", Reason: "store products carry no season binding or store status yet"},
		{Rule: "PS-5", Reason: "a pass season has no nameKey yet"},
	}
	PromoGaps = []Gap{
		{Rule: "PT-1", Reason: "a template records no launch role, so only that three templates exist is checked, not that they are the welcome, creator and compensation templates"},
	}
)

type category struct {
	check string
	rule  string
	key   string
	name  string
	tiers bool
}

var categories = []category{
	{check: "CG-2", rule: "MC-2", key: "ContentGateMinAvatarsAnimated", name: "avatar_animated", tiers: true},
	{check: "CG-3", rule: "MC-3", key: "ContentGateMinAvatarsStatic", name: "avatar_static", tiers: true},
	{check: "CG-4", rule: "MC-4", key: "ContentGateMinFrames", name: cosmetics.TypeFrame, tiers: true},
	{check: "CG-5", rule: "MC-5", key: "ContentGateMinTables", name: cosmetics.TypeTable, tiers: true},
	{check: "CG-6", rule: "MC-6", key: "ContentGateMinEmoji", name: cosmetics.TypeEmoji},
	{check: "CG-7", rule: "MC-7", key: "ContentGateMinMedals", name: MedalCategory},
}

func CategoryOf(item cosmetics.Item) string {
	if item.Type != cosmetics.TypeAvatar {
		return item.Type
	}
	if item.Animated {
		return "avatar_animated"
	}
	return "avatar_static"
}

type gate struct {
	in               Inputs
	rarities         cosmetics.Rarities
	tiers            []string
	catalog          *cosmetics.Catalog
	strictErrors     map[string]error
	sourceErrors     map[string]error
	itemFailures     map[string][]Failure
	catalogFailures  []Failure
	medals           *medal.Catalog
	medalFailures    map[string][]Failure
	medalCatalogFail []Failure
	missions         *mission.Catalog
	missionsFailure  *Failure
	passConfig       *pass.Snapshot
	passFailure      *Failure
	ranking          *rating.Snapshot
	rankingFailure   *Failure
}

func Run(in Inputs) Result {
	g := &gate{in: in, rarities: cosmetics.DefaultRarities(), strictErrors: map[string]error{}, sourceErrors: map[string]error{}, itemFailures: map[string][]Failure{}, medalFailures: map[string][]Failure{}}
	g.tiers = requiredTiers(g.rarities)
	g.load()
	g.judgeItems()
	g.judgeMedals()
	checks := make([]Check, 0, len(categories)+6)
	for _, c := range categories {
		checks = append(checks, g.countCheck(c))
	}
	checks = append(checks, g.readinessCheck(), g.missionCheck(), g.passCheck(), g.rankedCheck(), g.promoCheck(), g.thresholdCheck())
	verdict := VerdictGreen
	for _, check := range checks {
		if check.Result != ResultPass {
			verdict = VerdictRed
		}
	}
	return Result{Checks: checks, Info: g.info(), Verdict: verdict}
}

func requiredTiers(rarities cosmetics.Rarities) []string {
	tiers := make([]string, 0, len(rarities))
	for id := range rarities {
		if id != promo.MythicRarity {
			tiers = append(tiers, id)
		}
	}
	slices.SortFunc(tiers, func(a, b string) int { return rarities[a] - rarities[b] })
	return tiers
}

func (g *gate) load() {
	g.loadCosmetics()
	g.loadMedals()
	if raw := g.in.Snapshots[mission.Domain]; raw == nil {
		g.missionsFailure = &Failure{Object: mission.Domain, Rule: "MS-1", Detail: "the missions domain is not published"}
	} else if parsed, err := mission.Parse(raw); err != nil {
		g.missionsFailure = &Failure{Object: mission.Domain, Rule: "MS-1", Detail: err.Error()}
	} else {
		g.missions = parsed
	}
	if raw := g.in.Snapshots[pass.Domain]; raw == nil {
		g.passFailure = &Failure{Object: pass.Domain, Rule: "PS-1", Detail: "the pass domain is not published"}
	} else if parsed, err := pass.Parse(raw); err != nil {
		g.passFailure = &Failure{Object: pass.Domain, Rule: "PS-1", Detail: err.Error()}
	} else {
		g.passConfig = parsed
	}
	if raw := g.in.Snapshots[rating.Domain]; raw == nil {
		g.rankingFailure = &Failure{Object: rating.Domain, Rule: "RS-2", Detail: "the ranking domain is not published"}
	} else if parsed, err := rating.Parse(raw); err != nil {
		g.rankingFailure = &Failure{Object: rating.Domain, Rule: "RS-2", Detail: err.Error()}
	} else {
		g.ranking = parsed
	}
}

func (g *gate) loadCosmetics() {
	raw := g.in.Snapshots[cosmetics.Domain]
	if raw == nil {
		raw, _ = cosmetics.BuiltinManifest().Encode()
	}
	manifest, err := cosmetics.ParseManifest(raw)
	if err != nil {
		g.catalogFailures = append(g.catalogFailures, Failure{Object: cosmetics.Domain, Rule: "CR-2", Detail: err.Error()})
		g.catalog = &cosmetics.Catalog{}
		return
	}
	g.catalog, _ = cosmetics.Parse(raw)
	for _, item := range manifest.Items {
		var head cosmetics.Item
		if err := json.Unmarshal(item, &head); err != nil {
			continue
		}
		if err := cosmetics.ValidateManifest(cosmetics.Manifest{Items: []json.RawMessage{item}}, g.rarities); err != nil {
			g.strictErrors[head.ID] = err
		}
	}
}

func (g *gate) loadMedals() {
	raw := g.in.Snapshots[medal.Domain]
	if raw == nil {
		g.medals = medal.Default()
		g.medalCatalogFail = append(g.medalCatalogFail, Failure{Object: medal.Domain, Rule: "MR-1", Detail: "the medals domain is not published"})
		return
	}
	parsed, err := medal.Parse(raw)
	if err != nil {
		g.medals = medal.Default()
		g.medalCatalogFail = append(g.medalCatalogFail, Failure{Object: medal.Domain, Rule: "MR-2", Detail: err.Error()})
		return
	}
	g.medals = parsed
}

func (g *gate) judgeItems() {
	for _, item := range g.catalog.Items() {
		if item.Availability != cosmetics.Live {
			continue
		}
		fail := func(rule, detail string) {
			g.itemFailures[item.ID] = append(g.itemFailures[item.ID], Failure{Object: item.ID, Rule: rule, Detail: detail})
		}
		if err := g.strictErrors[item.ID]; err != nil {
			fail("CR-2", err.Error())
		}
		if g.rarities[item.Rarity] == 0 {
			fail("CR-3", fmt.Sprintf("rarityId %q is not in the rarity table", item.Rarity))
		}
		if !g.in.PassedStatics[item.ID] {
			fail("CR-5", "no static upload passed its checks")
		}
		if missing := g.missingLanguages(cosmetics.NameKeyOf(item.ID)); len(missing) > 0 {
			fail("CR-6", cosmetics.NameKeyOf(item.ID)+" missing in "+strings.Join(missing, ", "))
		}
	}
	g.catalogFailures = append(g.catalogFailures, g.grantListings()...)
}

func (g *gate) grantListings() []Failure {
	var failures []Failure
	itemLists := func(id string, source cosmetics.Source) bool {
		found, ok := g.catalog.Item(id)
		return ok && found.Grantable(source)
	}
	medalLists := func(id string, source cosmetics.Source) bool {
		found, ok := g.medals.Medal(id)
		return ok && found.Lists(source)
	}
	mythic := func(id string) bool {
		found, ok := g.catalog.Item(id)
		return ok && found.Rarity == promo.MythicRarity
	}
	record := func(domain string, err error) {
		if err != nil {
			g.sourceErrors[domain] = err
			failures = append(failures, Failure{Object: domain, Rule: "CR-2", Detail: err.Error()})
		}
	}
	if g.missions != nil {
		record(mission.Domain, g.missions.CheckSources(medalLists, itemLists))
	}
	if g.passConfig != nil {
		record(pass.Domain, g.passConfig.CheckSources(itemLists, mythic))
	}
	if g.ranking != nil {
		record(rating.Domain, g.ranking.CheckSources(medalLists, itemLists))
	}
	if raw := g.in.Snapshots[shop.Domain]; raw != nil {
		catalog, err := shop.Parse(raw, g.in.Shop)
		if err == nil {
			err = catalog.CheckSources(g.catalog)
		}
		record(shop.Domain, err)
	}
	return failures
}

func (g *gate) judgeMedals() {
	for _, m := range g.medals.Medals {
		if m.Availability != cosmetics.Live {
			continue
		}
		fail := func(rule, detail string) {
			g.medalFailures[m.ID] = append(g.medalFailures[m.ID], Failure{Object: m.ID, Rule: rule, Detail: detail})
		}
		if g.rarities[m.RarityID] == 0 || m.RarityID == promo.MythicRarity {
			fail("MR-2", fmt.Sprintf("rarityId %q is not a non-mythic tier of the rarity table", m.RarityID))
		}
		for _, source := range m.Sources {
			if detail := g.medalSourceProblem(source); detail != "" {
				fail("MR-3", detail)
			}
		}
		if !g.in.PassedStatics[m.ID] {
			fail("MR-4", "no static upload passed its checks")
		}
		for _, key := range []string{m.NameKey, m.DescriptionKey} {
			if missing := g.missingLanguages(key); len(missing) > 0 {
				fail("MR-5", key+" missing in "+strings.Join(missing, ", "))
			}
		}
	}
}

func (g *gate) medalSourceProblem(source cosmetics.Source) string {
	switch source.Type {
	case medal.SourceMission:
		if g.missions == nil {
			return "mission " + source.Ref + " cannot be resolved: the missions are not readable"
		}
		found, ok := g.missions.Mission(source.Ref)
		if !ok || found.Availability != mission.Live || (found.Type != mission.Achievement && found.Type != mission.Event) {
			return "mission " + source.Ref + " is not a live achievement or event mission"
		}
	case medal.SourceRankReward:
		if g.ranking == nil || !g.rankRewardExists(source.Ref) {
			return "rank reward " + source.Ref + " is not a rankRewards bundle of the published ranking config"
		}
	}
	return ""
}

func (g *gate) rankRewardExists(bundleID string) bool {
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		config, _ := g.ranking.Mode(mode)
		for _, bundle := range config.RankRewards {
			if bundle.ID == bundleID {
				return true
			}
		}
	}
	return false
}

func (g *gate) missingLanguages(key string) []string {
	var missing []string
	for _, language := range gamestrings.Languages {
		if !g.in.Bundles[language][key] {
			missing = append(missing, string(language))
		}
	}
	return missing
}

func (g *gate) cosmeticProblem(id string) string {
	item, ok := g.catalog.Item(id)
	switch {
	case !ok:
		return "cosmetic " + id + " is not in the manifest"
	case item.Availability != cosmetics.Live:
		return "cosmetic " + id + " is " + item.Availability
	case len(g.itemFailures[id]) > 0:
		return "cosmetic " + id + " fails " + ruleList(g.itemFailures[id])
	}
	return ""
}

func (g *gate) medalProblem(id string) string {
	found, ok := g.medals.Medal(id)
	switch {
	case !ok:
		return "medal " + id + " is not in the medals"
	case found.Availability != cosmetics.Live:
		return "medal " + id + " is " + found.Availability
	case len(g.medalFailures[id]) > 0:
		return "medal " + id + " fails " + ruleList(g.medalFailures[id])
	}
	return ""
}

func ruleList(failures []Failure) string {
	var ids []string
	for _, failure := range failures {
		if !slices.Contains(ids, failure.Rule) {
			ids = append(ids, failure.Rule)
		}
	}
	return strings.Join(ids, ", ")
}

func newCheck(id string) *Check {
	return &Check{ID: id, Failures: []Failure{}, Unchecked: []Gap{}}
}

func (c *Check) fail(object, rule, detail string) {
	c.Failures = append(c.Failures, Failure{Object: object, Rule: rule, Detail: detail})
}

func (c *Check) leaves(gaps ...Gap) {
	for _, gap := range gaps {
		if !slices.Contains(c.Unchecked, gap) {
			c.Unchecked = append(c.Unchecked, gap)
		}
	}
}

func (c *Check) granting(cosmeticIDs, medalIDs []string) {
	if len(cosmeticIDs) > 0 {
		c.leaves(CosmeticGaps...)
	}
	if len(medalIDs) > 0 {
		c.leaves(MedalGaps...)
	}
}

func (c *Check) settle() Check {
	switch {
	case len(c.Failures) > 0:
		c.Result = ResultFail
	case len(c.Unchecked) > 0:
		c.Result = ResultUnchecked
	default:
		c.Result = ResultPass
	}
	return *c
}

func (g *gate) threshold(key string) (Threshold, bool) {
	for _, t := range g.in.Thresholds {
		if t.Key == key {
			return t, true
		}
	}
	return Threshold{}, false
}

func (g *gate) readyByCategory() map[string][]string {
	ready := map[string][]string{}
	for _, item := range g.catalog.Items() {
		if item.Availability == cosmetics.Live && len(g.itemFailures[item.ID]) == 0 {
			ready[CategoryOf(item)] = append(ready[CategoryOf(item)], item.Rarity)
		}
	}
	for _, m := range g.medals.Medals {
		if m.Availability == cosmetics.Live && len(g.medalFailures[m.ID]) == 0 {
			ready[MedalCategory] = append(ready[MedalCategory], m.RarityID)
		}
	}
	return ready
}

func (g *gate) countCheck(c category) Check {
	check := newCheck(c.check)
	if c.name == MedalCategory {
		check.leaves(MedalGaps...)
		for _, failure := range g.medalCatalogFail {
			check.fail(failure.Object, failure.Rule, failure.Detail)
		}
	} else {
		check.leaves(CosmeticGaps...)
	}
	rarities := g.readyByCategory()[c.name]
	count := len(rarities)
	check.Count = &count
	minimum, ok := g.threshold(c.key)
	if !ok {
		check.fail(c.key, c.rule, "the threshold is missing")
		return check.settle()
	}
	check.Minimum = &minimum.Value
	if count < minimum.Value {
		check.fail(c.name, c.rule, fmt.Sprintf("%d ready, %s is %d", count, c.key, minimum.Value))
	}
	if c.tiers {
		for _, tier := range g.tiers {
			if !slices.Contains(rarities, tier) {
				check.fail(c.name, c.rule, "no ready "+tier+" item")
			}
		}
	}
	return check.settle()
}

func (g *gate) readinessCheck() Check {
	check := newCheck("CG-8")
	check.leaves(CosmeticGaps...)
	check.leaves(MedalGaps...)
	for _, failure := range g.catalogFailures {
		check.fail(failure.Object, failure.Rule, failure.Detail)
	}
	for _, failure := range g.medalCatalogFail {
		check.fail(failure.Object, failure.Rule, failure.Detail)
	}
	for _, item := range g.catalog.Items() {
		for _, failure := range g.itemFailures[item.ID] {
			check.fail(failure.Object, failure.Rule, failure.Detail)
		}
	}
	for _, m := range g.medals.Medals {
		for _, failure := range g.medalFailures[m.ID] {
			check.fail(failure.Object, failure.Rule, failure.Detail)
		}
	}
	return check.settle()
}

func (g *gate) missionCheck() Check {
	check := newCheck("CG-9")
	if g.missions == nil {
		check.fail(g.missionsFailure.Object, g.missionsFailure.Rule, g.missionsFailure.Detail)
		return check.settle()
	}
	for _, missionType := range []string{mission.Daily, mission.Weekly} {
		rerolls := g.missions.DailyRerolls
		if missionType == mission.Weekly {
			rerolls = g.missions.WeeklyRerolls
		}
		needed := max(2, rerolls+1)
		for slot := 1; slot <= len(g.missions.Slots(missionType)); slot++ {
			live := 0
			for _, m := range g.missions.Missions {
				if m.Type == missionType && m.Slot == slot && m.Availability == mission.Live {
					live++
				}
			}
			if live < needed {
				check.fail(fmt.Sprintf("%s slot %d", missionType, slot), "MS-2", fmt.Sprintf("%d live missions, %d needed", live, needed))
			}
		}
	}
	for _, m := range g.missions.Missions {
		if m.Availability != mission.Live {
			continue
		}
		if missing := g.missingLanguages(m.NameKey); len(missing) > 0 {
			check.fail(m.ID, "MS-3", m.NameKey+" missing in "+strings.Join(missing, ", "))
		}
		if m.Type != mission.Achievement && m.Type != mission.Event {
			continue
		}
		if m.Bundle == nil {
			check.fail(m.ID, "MS-4", "no bundle")
			continue
		}
		g.bundleReady(check, m.ID, "MS-4", m.Bundle.Cosmetics, m.Bundle.Medals)
	}
	return check.settle()
}

func (g *gate) bundleReady(check *Check, object, rule string, cosmeticIDs, medalIDs []string) {
	check.granting(cosmeticIDs, medalIDs)
	for _, id := range cosmeticIDs {
		if problem := g.cosmeticProblem(id); problem != "" {
			check.fail(object, rule, problem)
		}
	}
	for _, id := range medalIDs {
		if problem := g.medalProblem(id); problem != "" {
			check.fail(object, rule, problem)
		}
	}
}

func (g *gate) rankedSeason() (rating.Season, bool) {
	for _, season := range g.in.RankedSeasons {
		if !g.in.LaunchAt.Before(season.Start) && g.in.LaunchAt.Before(season.End) {
			return season, true
		}
	}
	return rating.Season{}, false
}

func (g *gate) passCheck() Check {
	check := newCheck("CG-10")
	check.leaves(PassGaps...)
	if g.passConfig == nil {
		check.fail(g.passFailure.Object, g.passFailure.Rule, g.passFailure.Detail)
		return check.settle()
	}
	if err := g.sourceErrors[pass.Domain]; err != nil {
		check.fail(pass.Domain, "PS-1", err.Error())
	}
	season, ok := g.passConfig.Active(g.in.LaunchAt)
	if !ok {
		check.fail(pass.Domain, "CG-10", "no pass season contains launchAt")
		return check.settle()
	}
	if ranked, found := g.rankedSeason(); !found {
		check.fail(season.ID, "PS-2", "no ranked season contains launchAt")
	} else if !ranked.Start.Equal(season.Start) || !ranked.End.Equal(season.End) {
		check.fail(season.ID, "PS-2", fmt.Sprintf("window %s to %s, ranked season %s runs %s to %s", stamp(season.Start), stamp(season.End), ranked.ID, stamp(ranked.Start), stamp(ranked.End)))
	}
	for level := 1; level <= season.Curve.Levels; level++ {
		if _, ok := season.Premium[level]; !ok {
			check.fail(season.ID, "PS-3", fmt.Sprintf("premium level %d has no bundle", level))
		}
	}
	for _, track := range []string{pass.Free, pass.Premium} {
		for level := 1; level <= season.Curve.Levels; level++ {
			bundle, ok := season.Track(track)[level]
			if ok {
				g.bundleReady(check, pass.TrackSource(season.ID, track, level).Ref, "PS-3", bundle.Cosmetics, nil)
			}
		}
	}
	return check.settle()
}

func stamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}

func (g *gate) rankedCheck() Check {
	check := newCheck("CG-11")
	season, ok := g.rankedSeason()
	if !ok {
		check.fail(rating.Domain, "CG-11", "no ranked season contains launchAt")
		return check.settle()
	}
	if state := season.State(g.in.Now); state != rating.StateLive && state != rating.StateScheduled {
		check.fail(season.ID, "RS-1", "state "+state)
	}
	if missing := g.missingLanguages(season.NameKey); len(missing) > 0 {
		check.fail(season.ID, "RS-5", season.NameKey+" missing in "+strings.Join(missing, ", "))
	}
	if g.ranking == nil {
		check.fail(g.rankingFailure.Object, g.rankingFailure.Rule, g.rankingFailure.Detail)
		return check.settle()
	}
	if err := g.sourceErrors[rating.Domain]; err != nil {
		check.fail(rating.Domain, "RS-2", err.Error())
	}
	if len(g.ranking.LeaderboardRegions) == 0 {
		check.fail(rating.Domain, "RS-4", "leaderboardRegions is empty")
	}
	for _, mode := range []rules.Mode{rules.Mode1v1, rules.Mode2v2} {
		config, _ := g.ranking.Mode(mode)
		for _, tier := range rating.Tiers() {
			bundle, ok := config.RankRewards[tier]
			object := mode.String() + " " + tier
			if !ok {
				check.fail(object, "RS-3", "no rankRewards bundle")
				continue
			}
			g.bundleReady(check, object, "RS-3", bundle.Cosmetics, bundle.Medals)
		}
	}
	return check.settle()
}

func (g *gate) promoCheck() Check {
	check := newCheck("CG-12")
	check.leaves(PromoGaps...)
	if !g.in.TemplatesReadable {
		check.fail("promo_templates", "PT-1", "the promo templates cannot be read without HD_PROMO_KEY")
		return check.settle()
	}
	if len(g.in.Templates) < LaunchTemplates {
		check.fail("promo_templates", "PT-1", fmt.Sprintf("%d templates, the launch needs %d", len(g.in.Templates), LaunchTemplates))
	}
	for _, t := range g.in.Templates {
		if strings.TrimSpace(t.Campaign) == "" {
			check.fail(t.ID, "PT-2", "no campaignId")
		}
		if t.Bundle == nil {
			continue
		}
		if err := t.Bundle.ValidateFor(wallet.Promo); err != nil {
			check.fail(t.ID, "PT-2", err.Error())
		}
		check.granting(t.Bundle.Cosmetics, nil)
		source := cosmetics.Source{Type: promo.SourceType, Ref: t.Campaign}
		for _, id := range t.Bundle.Cosmetics {
			if problem := g.cosmeticProblem(id); problem != "" {
				check.fail(t.ID, "PT-3", problem)
			}
			if item, ok := g.catalog.Item(id); ok && !item.Grantable(source) {
				check.fail(t.ID, "PT-3", id+" does not list the promo source "+t.Campaign)
			}
		}
	}
	return check.settle()
}

func (g *gate) thresholdCheck() Check {
	check := newCheck("CG-13")
	for _, c := range categories {
		t, ok := g.threshold(c.key)
		switch {
		case !ok:
			check.fail(c.key, "CG-13", "the threshold is missing")
		case t.Value < t.Default:
			check.fail(c.key, "CG-13", fmt.Sprintf("%d is below the repo default %d", t.Value, t.Default))
		}
	}
	return check.settle()
}

func (g *gate) info() Info {
	info := Info{Ready: map[string]map[string]int{}, LiveNotReady: map[string]int{}}
	for name, rarities := range g.readyByCategory() {
		info.Ready[name] = map[string]int{}
		for _, rarity := range rarities {
			info.Ready[name][rarity]++
		}
	}
	for _, item := range g.catalog.Items() {
		if item.Availability == cosmetics.Live && len(g.itemFailures[item.ID]) > 0 {
			info.LiveNotReady[CategoryOf(item)]++
		}
	}
	for _, m := range g.medals.Medals {
		if m.Availability == cosmetics.Live && len(g.medalFailures[m.ID]) > 0 {
			info.LiveNotReady[MedalCategory]++
		}
	}
	return info
}
