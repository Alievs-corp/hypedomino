package adminapi

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/pass"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type playersView struct {
	Players []playerrecord.Summary `json:"players"`
}

func (h *Handler) searchPlayers(w http.ResponseWriter, r *http.Request, p principal) {
	query := r.URL.Query()
	var found []playerrecord.Summary
	var err error
	switch {
	case !query.Has("subject"):
		found, err = playerrecord.Search(r.Context(), h.pool, query.Get("q"))
	case !slices.Contains(p.account.Roles, rbac.SuperAdmin):
		err = errForbidden
	default:
		found, err = playerrecord.SearchIdentity(r.Context(), h.pool, query.Get("provider"), query.Get("subject"))
	}
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
		if err = passStanding(r.Context(), tx, &record, now); err != nil {
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

func passStanding(ctx context.Context, tx pgx.Tx, record *playerrecord.Record, now time.Time) error {
	if record.Profile == nil {
		return nil
	}
	_, raw, err := latestSnapshot(ctx, tx, pass.Domain)
	if err != nil {
		return err
	}
	config := pass.Default()
	if raw != nil {
		if config, err = pass.Parse(raw); err != nil {
			return err
		}
	}
	record.Profile.Level, record.Profile.XP, err = config.Standing(ctx, tx, record.Account.ID, now)
	return err
}

func (h *Handler) matchReplay(w http.ResponseWriter, r *http.Request, p principal) {
	var match playerrecord.MatchReplay
	err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		var err error
		if match, err = playerrecord.LoadReplay(r.Context(), tx, r.PathValue("id")); err != nil {
			return nil, err
		}
		return []audit.Entry{{Action: ViewReplayAction, Object: audit.Object{Type: rbac.Players, ID: match.MatchID}}}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, match)
}
