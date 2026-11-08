package adminapi

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/report"
)

const (
	ReportTargetLimit   = 100
	UpholdReportsAction = "uphold_reports"
	RejectReportsAction = "reject_reports"
)

type reportsResponse struct {
	Targets []report.Target `json:"targets"`
}

type reportsDecision struct {
	AccountID string `json:"account_id"`
	State     string `json:"state"`
	Closed    int64  `json:"closed"`
}

func (h *Handler) listReports(w http.ResponseWriter, r *http.Request, p principal) {
	targets, err := report.Open(r.Context(), h.pool, ReportTargetLimit)
	if err != nil {
		h.refuse(w, err)
		return
	}
	if targets == nil {
		targets = []report.Target{}
	}
	reply(w, http.StatusOK, reportsResponse{Targets: targets})
}

func (h *Handler) decideReports(state, action string) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, p principal) {
		var note changeNote
		if !decode(w, r, &note) {
			return
		}
		decision := reportsDecision{AccountID: r.PathValue("id"), State: state}
		err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
			if err := h.sensitive(p, note, now); err != nil {
				return nil, err
			}
			if err := playerrecord.Lock(r.Context(), tx, decision.AccountID); err != nil {
				return nil, err
			}
			closed, err := report.Decide(r.Context(), tx, decision.AccountID, state, p.account.ID, now)
			if err != nil {
				return nil, err
			}
			decision.Closed = closed
			return []audit.Entry{{Action: action, Object: audit.Object{Type: rbac.Players, ID: decision.AccountID}, After: marshal(decision), Reason: note.Reason, Ref: note.Ref}}, nil
		})
		if err != nil {
			h.refuse(w, err)
			return
		}
		reply(w, http.StatusOK, decision)
	}
}
