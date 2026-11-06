package adminapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/mission"
	"github.com/Alievs-corp/hypedomino/server/internal/pass"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
)

type cosmeticRulesView struct {
	Types        []string        `json:"types"`
	Rarities     []string        `json:"rarities"`
	Availability []string        `json:"availability"`
	SourceTypes  []string        `json:"source_types"`
	ArtSources   []string        `json:"art_sources"`
	Builtin      json.RawMessage `json:"builtin"`
}

type cosmeticCatalogView struct {
	catalogView
	Rules cosmeticRulesView `json:"rules"`
}

func (h *Handler) cosmeticCatalog(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, cosmetics.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	builtin, err := cosmetics.BuiltinManifest().Encode()
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, cosmeticCatalogView{
		catalogView: viewCatalog(version, raw),
		Rules: cosmeticRulesView{
			Types:        cosmetics.Types,
			Rarities:     rarityOrder(cosmetics.DefaultRarities()),
			Availability: []string{cosmetics.Live, cosmetics.Retired, cosmetics.Disabled},
			SourceTypes:  cosmetics.SourceTypes,
			ArtSources:   cosmetics.ArtSources,
			Builtin:      builtin,
		},
	})
}

func (h *Handler) cosmeticsDomain(rollback bool) configDomain {
	return configDomain{name: cosmetics.Domain, object: rbac.CosmeticItems,
		prepare: func(ctx context.Context, tx pgx.Tx, before, after []byte) ([]byte, error) {
			previous, err := publishedManifest(before)
			if err != nil {
				return nil, err
			}
			next, err := nextManifest(previous, after, rollback)
			if err != nil {
				return nil, err
			}
			raw, err := next.Encode()
			if err != nil {
				return nil, err
			}
			items, err := cosmetics.Parse(raw)
			if err != nil {
				return nil, err
			}
			if err := cosmetics.ValidateManifest(next, cosmetics.DefaultRarities()); err != nil {
				return nil, err
			}
			if err := cosmetics.CheckSuccessor(previous, next); err != nil {
				return nil, err
			}
			return raw, h.grantsStayListed(ctx, tx, items)
		},
		accept: func(_, after []byte) error {
			_, err := cosmetics.ParseManifest(after)
			return err
		},
	}
}

func publishedManifest(before []byte) (cosmetics.Manifest, error) {
	if before == nil {
		return cosmetics.BuiltinManifest(), nil
	}
	return cosmetics.ParseManifest(before)
}

func nextManifest(previous cosmetics.Manifest, after []byte, rollback bool) (cosmetics.Manifest, error) {
	if !rollback {
		return cosmetics.Stamp(previous, after)
	}
	target, err := cosmetics.ParseManifest(after)
	if err != nil {
		return cosmetics.Manifest{}, err
	}
	return cosmetics.Restore(previous, target)
}

func (h *Handler) grantsStayListed(ctx context.Context, tx pgx.Tx, items *cosmetics.Catalog) error {
	medals, err := publishedMedals(ctx, tx)
	if err != nil {
		return err
	}
	_, raw, err := latestSnapshot(ctx, tx, mission.Domain)
	if err != nil {
		return err
	}
	if raw != nil {
		missions, err := mission.Parse(raw)
		if err != nil {
			return err
		}
		if err := missionSourcesIn(items, missions, medals); err != nil {
			return err
		}
	}
	if _, raw, err = latestSnapshot(ctx, tx, pass.Domain); err != nil {
		return err
	}
	if raw != nil {
		config, err := pass.Parse(raw)
		if err != nil {
			return err
		}
		if err := trackSourcesIn(items, config); err != nil {
			return err
		}
	}
	ranked, err := publishedRating(ctx, tx)
	if err != nil {
		return err
	}
	if ranked != nil {
		if err := rankRewardSourcesIn(items, ranked, medals); err != nil {
			return err
		}
	}
	if _, raw, err = latestSnapshot(ctx, tx, shop.Domain); err != nil || raw == nil {
		return err
	}
	catalog, err := shop.Parse(raw, h.config.Shop)
	if err != nil {
		return err
	}
	return catalog.CheckChestPools(items)
}
