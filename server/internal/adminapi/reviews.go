package adminapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/risk"
)

const (
	ReviewListLimit     = 100
	ReviewSignalLimit   = 50
	ClearReviewAction   = "clear_review"
	ConfirmReviewAction = "confirm_review"
)

type reviewsResponse struct {
	Reviews []risk.Review `json:"reviews"`
}

type reviewState struct {
	AccountID    string `json:"account_id"`
	Shadowbanned bool   `json:"shadowbanned"`
}

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request, p principal) {
	reviews, err := risk.OpenReviews(r.Context(), h.pool, ReviewListLimit, ReviewSignalLimit)
	if err != nil {
		h.refuse(w, err)
		return
	}
	if reviews == nil {
		reviews = []risk.Review{}
	}
	reply(w, http.StatusOK, reviewsResponse{Reviews: reviews})
}

func (h *Handler) decideReview(decision, action string) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, p principal) {
		var note changeNote
		if !decode(w, r, &note) {
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			h.refuse(w, risk.ErrNoReview)
			return
		}
		var state reviewState
		err = h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
			if err := h.sensitive(p, note, now); err != nil {
				return nil, err
			}
			decided, err := risk.DecideReview(r.Context(), tx, id, decision, p.account.ID, now)
			if err != nil {
				return nil, err
			}
			state = reviewState{AccountID: decided.Account, Shadowbanned: decided.StillShadowbanned}
			return []audit.Entry{{
				Action: action, Object: audit.Object{Type: rbac.Queues, ID: r.PathValue("id")},
				Before: marshal(reviewState{AccountID: decided.Account, Shadowbanned: decided.WasShadowbanned}), After: marshal(state),
				Reason: note.Reason, Ref: note.Ref,
			}}, nil
		})
		if err != nil {
			h.refuse(w, err)
			return
		}
		reply(w, http.StatusOK, state)
	}
}
