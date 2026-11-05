package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/approval"
	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/livecontrol"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

const (
	ProposeLiveControlsAction = "propose_live_controls"
	SwitchLiveControlsAction  = "switch_live_controls"
)

var errStaleLiveControls = errors.New("adminapi: live controls changed after this proposal")

type liveControlsRequest struct {
	Controls json.RawMessage `json:"controls"`
	changeNote
}

type liveControlsChange struct {
	BaseVersion int64           `json:"base_version"`
	Controls    json.RawMessage `json:"controls"`
}

type liveControlsView struct {
	ConfigVersion int64           `json:"config_version"`
	Controls      json.RawMessage `json:"controls"`
	ApprovalID    int64           `json:"approval_id,omitempty"`
}

func (h *Handler) liveControls(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, livecontrol.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, liveControlsView{ConfigVersion: version, Controls: viewCatalog(version, raw).Catalog})
}

func (h *Handler) switchLiveControls(w http.ResponseWriter, r *http.Request, p principal) {
	var request liveControlsRequest
	if !decode(w, r, &request) {
		return
	}
	var view liveControlsView
	err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		if err := h.sensitive(p, request.changeNote, now); err != nil {
			return nil, err
		}
		if _, err := livecontrol.Parse(request.Controls); err != nil {
			return nil, err
		}
		if err := lockDomain(r.Context(), tx, livecontrol.Domain); err != nil {
			return nil, err
		}
		version, before, err := latestSnapshot(r.Context(), tx, livecontrol.Domain)
		if err != nil {
			return nil, err
		}
		if h.config.Environment != audit.Production {
			view.Controls = request.Controls
			if view.ConfigVersion, err = snapshot.Publish(r.Context(), tx, livecontrol.Domain, request.Controls); err != nil {
				return nil, err
			}
			return []audit.Entry{liveControlsEntry(SwitchLiveControlsAction, before, request.Controls, view.ConfigVersion, request.changeNote)}, nil
		}
		view.ConfigVersion, view.Controls = version, viewCatalog(version, before).Catalog
		change := marshal(liveControlsChange{BaseVersion: version, Controls: request.Controls})
		proposal, err := h.approvals.In(tx).Propose(r.Context(), approval.LiveControl, livecontrol.Domain, change, request.Reason, p.actor(), now)
		if err != nil {
			return nil, err
		}
		view.ApprovalID = proposal.ID
		return []audit.Entry{liveControlsEntry(ProposeLiveControlsAction, before, request.Controls, 0, request.changeNote)}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	status := http.StatusOK
	if view.ApprovalID != 0 {
		status = http.StatusAccepted
	}
	reply(w, status, view)
}

func liveControlsEntry(action string, before, after []byte, version int64, note changeNote) audit.Entry {
	return audit.Entry{
		Action: action, Object: audit.Object{Type: rbac.LiveControls, ID: livecontrol.Domain}, Before: before,
		After: json.RawMessage(after), ConfigVersion: version, Reason: note.Reason, Ref: note.Ref,
	}
}

func (h *Handler) applyLiveControls(ctx context.Context, tx pgx.Tx, p principal, proposal approval.Proposal, note changeNote, now time.Time) error {
	var change liveControlsChange
	if err := json.Unmarshal(proposal.Change, &change); err != nil {
		return err
	}
	if err := lockDomain(ctx, tx, livecontrol.Domain); err != nil {
		return err
	}
	current, before, err := latestSnapshot(ctx, tx, livecontrol.Domain)
	if err != nil {
		return err
	}
	if current != change.BaseVersion {
		return fmt.Errorf("%w: proposed on version %d, now %d", errStaleLiveControls, change.BaseVersion, current)
	}
	version, err := snapshot.Publish(ctx, tx, livecontrol.Domain, change.Controls)
	if err != nil {
		return err
	}
	decision := audit.Entry{Action: "approve", Object: proposal.Object, After: proposal.Change, Reason: note.Reason, Ref: note.Ref}
	recorded, err := h.audit.In(tx).Append(ctx, h.stamp(p, decision), now)
	if err != nil {
		return err
	}
	applied := liveControlsEntry(SwitchLiveControlsAction, before, change.Controls, version, changeNote{Reason: proposal.Reason})
	applied.ApprovalID = recorded.ID
	_, err = h.audit.In(tx).Append(ctx, h.stamp(p, applied), now)
	return err
}
