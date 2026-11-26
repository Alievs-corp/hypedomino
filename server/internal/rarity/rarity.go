package rarity

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

const (
	Domain  = "rarity"
	MaxPips = 6
)

var ErrInvalid = errors.New("rarity: invalid rarity table")

var (
	tierID = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	hex    = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

type Tier struct {
	Ordinal int    `json:"ordinal"`
	NameKey string `json:"nameKey"`
	Fill    string `json:"fill"`
	On      string `json:"on"`
	Text    string `json:"text"`
	Pips    int    `json:"pips"`
	Glow    string `json:"glow"`
	Flow    bool   `json:"flow"`
}

type Table struct {
	DiamondMinSize int             `json:"diamondMinSize"`
	FlowPeriodMs   int             `json:"flowPeriodMs"`
	Tiers          map[string]Tier `json:"tiers"`
}

type Glow struct {
	Alpha float64 `json:"alpha"`
	Scale float64 `json:"scale"`
}

type tokens struct {
	Glow                   map[string]Glow `json:"glow"`
	Grounds                []string        `json:"grounds"`
	TextMinimumContrast    float64         `json:"textMinimumContrast"`
	NonTextMinimumContrast float64         `json:"nonTextMinimumContrast"`
	Table                  json.RawMessage `json:"table"`
}

//go:embed defaults.json
var defaultsJSON []byte

var builtin = loadTokens()

func loadTokens() tokens {
	var loaded tokens
	if err := json.Unmarshal(defaultsJSON, &loaded); err != nil {
		panic(fmt.Errorf("rarity: built-in defaults: %w", err))
	}
	return loaded
}

func Default() Table {
	table, err := Parse(builtin.Table)
	if err != nil {
		panic(err)
	}
	return table
}

func DefaultJSON() json.RawMessage {
	return builtin.Table
}

func GlowLevels() []string {
	levels := make([]string, 0, len(builtin.Glow))
	for name := range builtin.Glow {
		levels = append(levels, name)
	}
	sort.Slice(levels, func(a, b int) bool {
		return builtin.Glow[levels[a]].Alpha < builtin.Glow[levels[b]].Alpha || builtin.Glow[levels[a]].Alpha == builtin.Glow[levels[b]].Alpha && levels[a] < levels[b]
	})
	return levels
}

func Parse(raw []byte) (Table, error) {
	var table Table
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&table); err != nil {
		return Table{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := table.validate(); err != nil {
		return Table{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return table, nil
}

func (t Table) validate() error {
	if t.DiamondMinSize < 1 {
		return errors.New("diamondMinSize must be a positive integer")
	}
	if t.FlowPeriodMs < 1 {
		return errors.New("flowPeriodMs must be a positive integer")
	}
	if len(t.Tiers) == 0 {
		return errors.New("at least one tier is required")
	}
	for _, id := range t.ids() {
		if err := checkTier(id, t.Tiers[id]); err != nil {
			return err
		}
	}
	if err := t.checkOrder(); err != nil {
		return err
	}
	return t.keepsBuiltinTiers()
}

func checkTier(id string, tier Tier) error {
	switch {
	case !tierID.MatchString(id):
		return fmt.Errorf("tier id %q must match %s", id, tierID)
	case tier.NameKey != "rarity."+id+".name":
		return fmt.Errorf("%s nameKey must be rarity.%s.name", id, id)
	case !hex.MatchString(tier.Fill) || !hex.MatchString(tier.On) || !hex.MatchString(tier.Text):
		return fmt.Errorf("%s fill, on and text must be #RRGGBB", id)
	case tier.Pips < 1 || tier.Pips > MaxPips:
		return fmt.Errorf("%s pips must be 1 to %d", id, MaxPips)
	}
	if _, known := builtin.Glow[tier.Glow]; !known {
		return fmt.Errorf("%s glow %q is not a glow level", id, tier.Glow)
	}
	return checkContrast(id, tier)
}

func checkContrast(id string, tier Tier) error {
	for _, ground := range builtin.Grounds {
		if ratio := contrast(tier.Fill, ground); ratio < builtin.NonTextMinimumContrast {
			return fmt.Errorf("%s fill on %s is %.2f, needs %v", id, ground, ratio, builtin.NonTextMinimumContrast)
		}
		if ratio := contrast(tier.Text, ground); ratio < builtin.TextMinimumContrast {
			return fmt.Errorf("%s text on %s is %.2f, needs %v", id, ground, ratio, builtin.TextMinimumContrast)
		}
	}
	if ratio := contrast(tier.On, tier.Fill); ratio < builtin.TextMinimumContrast {
		return fmt.Errorf("%s on on fill is %.2f, needs %v", id, ratio, builtin.TextMinimumContrast)
	}
	return nil
}

func (t Table) checkOrder() error {
	ids := t.ids()
	pips := make(map[int]string, len(ids))
	for at, id := range ids {
		tier := t.Tiers[id]
		if tier.Ordinal != at+1 {
			return fmt.Errorf("ordinals must be 1 to %d without gaps or repeats", len(ids))
		}
		if other, taken := pips[tier.Pips]; taken {
			return fmt.Errorf("%s and %s share %d pips", other, id, tier.Pips)
		}
		pips[tier.Pips] = id
		if at > 0 && builtin.Glow[tier.Glow].Alpha < builtin.Glow[t.Tiers[ids[at-1]].Glow].Alpha {
			return fmt.Errorf("%s glow is weaker than the glow of a lower tier", id)
		}
	}
	return nil
}

func (t Table) keepsBuiltinTiers() error {
	var defaults struct {
		Tiers map[string]json.RawMessage `json:"tiers"`
	}
	if err := json.Unmarshal(builtin.Table, &defaults); err != nil {
		return err
	}
	for id := range defaults.Tiers {
		if _, kept := t.Tiers[id]; !kept {
			return fmt.Errorf("tier %s is referenced by items and cannot be removed", id)
		}
	}
	return nil
}

func (t Table) ids() []string {
	ids := make([]string, 0, len(t.Tiers))
	for id := range t.Tiers {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(a, b int) bool {
		return t.Tiers[ids[a]].Ordinal < t.Tiers[ids[b]].Ordinal || t.Tiers[ids[a]].Ordinal == t.Tiers[ids[b]].Ordinal && ids[a] < ids[b]
	})
	return ids
}

func (t Table) Hello() *pb.RarityTableConfig {
	ids := t.ids()
	tiers := make([]*pb.RarityTierConfig, 0, len(ids))
	for _, id := range ids {
		tier := t.Tiers[id]
		glow := builtin.Glow[tier.Glow]
		tiers = append(tiers, &pb.RarityTierConfig{
			Id: id, Ordinal: uint32(tier.Ordinal), NameKey: tier.NameKey,
			Fill: tier.Fill, On: tier.On, Text: tier.Text, Pips: uint32(tier.Pips),
			GlowAlpha: float32(glow.Alpha), GlowScale: float32(glow.Scale), Flow: tier.Flow,
		})
	}
	return &pb.RarityTableConfig{DiamondMinSize: uint32(t.DiamondMinSize), FlowPeriodMs: uint32(t.FlowPeriodMs), Tiers: tiers}
}

func contrast(first, second string) float64 {
	light, dark := luminance(first), luminance(second)
	if dark > light {
		light, dark = dark, light
	}
	return math.Floor((light+0.05)/(dark+0.05)*100) / 100
}

func luminance(color string) float64 {
	return 0.2126*channel(color[1:3]) + 0.7152*channel(color[3:5]) + 0.0722*channel(color[5:7])
}

func channel(pair string) float64 {
	value, _ := strconv.ParseUint(pair, 16, 8)
	linear := float64(value) / 255
	if linear <= 0.04045 {
		return linear / 12.92
	}
	return math.Pow((linear+0.055)/1.055, 2.4)
}
