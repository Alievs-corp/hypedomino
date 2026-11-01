package adminapi

import (
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type playersView struct {
	Players []playerrecord.Summary `json:"players"`
}

func (h *Handler) searchPlayers(w http.ResponseWriter, r *http.Request, p principal) {
	found, err := playerrecord.Search(r.Context(), h.pool, r.URL.Query().Get("q"))
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, playersView{Players: found})
}

func (h *Handler) playerRecord(w http.ResponseWriter, r *http.Request, p principal) {
	var record playerrecord.Record
	err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		var err error
		if record, err = playerrecord.Load(r.Context(), tx, r.PathValue("id"), slices.Contains(p.account.Roles, rbac.SuperAdmin)); err != nil {
			return nil, err
		}
		return []audit.Entry{{Action: "view", Object: audit.Object{Type: rbac.Players, ID: record.Account.ID}}}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, record)
}
