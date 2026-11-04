package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/medal"
	"github.com/Alievs-corp/hypedomino/server/internal/mission"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var (
	errMissionRemoved = errors.New("adminapi: a published mission stays in the catalog; retire it instead")
	errMedalRemoved   = errors.New("adminapi: a published medal stays in the catalog; retire or disable it instead")
)

type missionRulesView struct {
	Types  []string `json:"types"`
	Events []string `json:"events"`
	Queues []string `json:"queues"`
	Modes  []string `json:"modes"`
}

type missionCatalogView struct {
	catalogView
	Rules missionRulesView `json:"rules"`
}

type medalRulesView struct {
	Rarities     []string `json:"rarities"`
	SourceTypes  []string `json:"source_types"`
	Availability []string `json:"availability"`
}

type medalCatalogView struct {
	catalogView
	Rules medalRulesView `json:"rules"`
}

func (h *Handler) missionCatalog(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, mission.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	modes := []string{rules.Mode1v1.String(), rules.Mode2v2.String()}
	reply(w, http.StatusOK, missionCatalogView{
		catalogView: viewCatalog(version, raw),
		Rules:       missionRulesView{Types: mission.Types, Events: mission.ProgressEvents, Queues: mission.Queues, Modes: modes},
	})
}

func (h *Handler) medalCatalog(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, medal.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, medalCatalogView{
		catalogView: viewCatalog(version, raw),
		Rules: medalRulesView{
			Rarities:     rarityOrder(cosmetics.DefaultRarities()),
			SourceTypes:  []string{medal.SourceMission, medal.SourceRankReward},
			Availability: []string{cosmetics.Live, cosmetics.Retired, cosmetics.Disabled},
		},
	})
}

func rarityOrder(rarities cosmetics.Rarities) []string {
	ids := make([]string, 0, len(rarities))
	for id := range rarities {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b string) int { return rarities[a] - rarities[b] })
	return ids
}

func (h *Handler) missionDomain(rollback bool) configDomain {
	return configDomain{name: mission.Domain, object: rbac.Missions,
		prepare: func(ctx context.Context, tx pgx.Tx, before, after []byte) ([]byte, error) {
			stamped, err := stampMissions(before, after, rollback)
			if err != nil {
				return nil, err
			}
			catalog, err := mission.Parse(stamped)
			if err != nil {
				return nil, err
			}
			medals, err := publishedMedals(ctx, tx)
			if err != nil {
				return nil, err
			}
			return stamped, missionSourcesListed(ctx, tx, catalog, medals)
		},
		accept: func(_, after []byte) error {
			_, err := mission.Parse(after)
			return err
		},
	}
}

func (h *Handler) medalDomain(rollback bool) configDomain {
	return configDomain{name: medal.Domain, object: rbac.Medals,
		prepare: func(ctx context.Context, tx pgx.Tx, before, after []byte) ([]byte, error) {
			stamped, err := stampMedals(before, after, rollback)
			if err != nil {
				return nil, err
			}
			medals, err := medal.Parse(stamped)
			if err != nil {
				return nil, err
			}
			ranked, err := publishedRating(ctx, tx)
			if err != nil {
				return nil, err
			}
			if ranked != nil {
				if err := rankRewardSourcesListed(ctx, tx, ranked, medals); err != nil {
					return nil, err
				}
			}
			_, raw, err := latestSnapshot(ctx, tx, mission.Domain)
			if err != nil || raw == nil {
				return stamped, err
			}
			missions, err := mission.Parse(raw)
			if err != nil {
				return nil, err
			}
			return stamped, missionSourcesListed(ctx, tx, missions, medals)
		},
		accept: func(_, after []byte) error {
			_, err := medal.Parse(after)
			return err
		},
	}
}

func publishedMedals(ctx context.Context, tx pgx.Tx) (*medal.Catalog, error) {
	_, raw, err := latestSnapshot(ctx, tx, medal.Domain)
	if err != nil || raw == nil {
		return medal.Default(), err
	}
	return medal.Parse(raw)
}

func missionSourcesListed(ctx context.Context, tx pgx.Tx, missions *mission.Catalog, medals *medal.Catalog) error {
	items, err := cosmetics.Published(ctx, tx)
	if err != nil {
		return err
	}
	medalLists := func(id string, source cosmetics.Source) bool {
		found, ok := medals.Medal(id)
		return ok && found.Lists(source)
	}
	itemLists := func(id string, source cosmetics.Source) bool {
		found, ok := items.Item(id)
		return ok && found.Grantable(source)
	}
	return missions.CheckSources(medalLists, itemLists)
}

func stampMissions(before, after []byte, rollback bool) ([]byte, error) {
	next, current := mission.Default(), mission.Default()
	if err := json.Unmarshal(after, next); err != nil {
		return nil, fmt.Errorf("%w: %v", mission.ErrInvalidCatalog, err)
	}
	if before != nil {
		if err := json.Unmarshal(before, current); err != nil {
			return nil, err
		}
	}
	previous := make(map[string]mission.Mission, len(current.Missions))
	for _, old := range current.Missions {
		previous[old.ID] = old
	}
	listed := make(map[string]bool, len(next.Missions))
	for i := range next.Missions {
		m := &next.Missions[i]
		listed[m.ID] = true
		if slug, ok := strings.CutPrefix(m.ID, "mission."); ok {
			m.NameKey = "mission." + slug + ".name"
		}
		old, existed := previous[m.ID]
		m.Revision = old.Revision
		if !existed || !sameMission(old, *m) {
			m.Revision++
		}
	}
	for _, old := range current.Missions {
		if listed[old.ID] {
			continue
		}
		if !rollback {
			return nil, fmt.Errorf("%w: %s", errMissionRemoved, old.ID)
		}
		if old.Availability != mission.Retired {
			old.Availability = mission.Retired
			old.Revision++
		}
		next.Missions = append(next.Missions, old)
	}
	return json.Marshal(next)
}

func sameMission(a, b mission.Mission) bool {
	a.Revision, b.Revision = 0, 0
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}

func stampMedals(before, after []byte, rollback bool) ([]byte, error) {
	next, current := medal.Default(), medal.Default()
	if err := json.Unmarshal(after, next); err != nil {
		return nil, fmt.Errorf("%w: %v", medal.ErrInvalidCatalog, err)
	}
	if before != nil {
		if err := json.Unmarshal(before, current); err != nil {
			return nil, err
		}
	}
	previous := make(map[string]medal.Medal, len(current.Medals))
	for _, old := range current.Medals {
		previous[old.ID] = old
	}
	listed := make(map[string]bool, len(next.Medals))
	for i := range next.Medals {
		m := &next.Medals[i]
		listed[m.ID] = true
		if slug, ok := strings.CutPrefix(m.ID, "medal."); ok {
			m.NameKey, m.DescriptionKey, m.StaticAddress = "medal."+slug+".name", "medal."+slug+".description", "medal/"+slug+"/static"
		}
		old, existed := previous[m.ID]
		m.Revision = old.Revision
		if !existed || !sameMedal(old, *m) {
			m.Revision++
		}
	}
	for _, old := range current.Medals {
		if listed[old.ID] {
			continue
		}
		if !rollback {
			return nil, fmt.Errorf("%w: %s", errMedalRemoved, old.ID)
		}
		if old.Availability == cosmetics.Live {
			old.Availability = cosmetics.Retired
			old.Revision++
		}
		next.Medals = append(next.Medals, old)
	}
	return json.Marshal(next)
}

func sameMedal(a, b medal.Medal) bool {
	a.Revision, b.Revision = 0, 0
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}
