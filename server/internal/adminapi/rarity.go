package adminapi

import (
	"encoding/json"
	"net/http"

	"github.com/Alievs-corp/hypedomino/server/internal/rarity"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type rarityCatalogView struct {
	catalogView
	Defaults json.RawMessage `json:"defaults"`
	Glow     []string        `json:"glow"`
}

func rarityDomain() configDomain {
	return configDomain{name: rarity.Domain, object: rbac.RarityTable, accept: func(_, after []byte) error {
		_, err := rarity.Parse(after)
		return err
	}}
}

func (h *Handler) rarityCatalog(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, rarity.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, rarityCatalogView{catalogView: viewCatalog(version, raw), Defaults: rarity.DefaultJSON(), Glow: rarity.GlowLevels()})
}
