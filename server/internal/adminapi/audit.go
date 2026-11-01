package adminapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type auditEntryView struct {
	ID            int64           `json:"id"`
	At            time.Time       `json:"at"`
	ActorID       string          `json:"actor_id"`
	ActorRoles    []rbac.Role     `json:"actor_roles"`
	Environment   string          `json:"environment"`
	Action        string          `json:"action"`
	ObjectType    rbac.Object     `json:"object_type"`
	ObjectID      string          `json:"object_id"`
	Before        json.RawMessage `json:"before,omitempty"`
	After         json.RawMessage `json:"after,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	ApprovalID    int64           `json:"approval_id,omitempty"`
	Ref           string          `json:"ref,omitempty"`
	ChangeSet     string          `json:"change_set,omitempty"`
	ConfigVersion int64           `json:"config_version,omitempty"`
	SessionID     string          `json:"session_id"`
	SourceIP      string          `json:"source_ip,omitempty"`
	UserAgent     string          `json:"user_agent,omitempty"`
}

type auditResponse struct {
	Entries []auditEntryView `json:"entries"`
}

func viewEntry(r audit.Recorded) auditEntryView {
	view := auditEntryView{
		ID: r.ID, At: r.At, ActorID: r.Actor.ID, ActorRoles: r.Actor.Roles, Environment: r.Environment, Action: r.Action,
		ObjectType: r.Object.Type, ObjectID: r.Object.ID, Before: r.Before, After: r.After, Reason: r.Reason,
		ApprovalID: r.ApprovalID, Ref: r.Ref, ChangeSet: r.ChangeSet, ConfigVersion: r.ConfigVersion,
		SessionID: r.Source.SessionID, UserAgent: r.Source.UserAgent,
	}
	if r.Source.IP.IsValid() {
		view.SourceIP = r.Source.IP.String()
	}
	return view
}

func auditFilter(query url.Values) (audit.Filter, error) {
	filter := audit.Filter{
		ActorID:        query.Get("actor"),
		Object:         audit.Object{Type: rbac.Object(query.Get("object_type")), ID: query.Get("object_id")},
		Action:         query.Get("action"),
		ReasonContains: query.Get("reason"),
	}
	var err error
	if filter.From, err = optionalInstant(query.Get("from")); err != nil {
		return audit.Filter{}, err
	}
	if filter.To, err = optionalInstant(query.Get("to")); err != nil {
		return audit.Filter{}, err
	}
	if filter.BeforeID, err = optionalNumber(query.Get("before_id")); err != nil {
		return audit.Filter{}, err
	}
	limit, err := optionalNumber(query.Get("limit"))
	filter.Limit = int(limit)
	return filter, err
}

func optionalInstant(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	instant, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, errMalformed
	}
	return instant, nil
}

func optionalNumber(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	number, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || number <= 0 {
		return 0, errMalformed
	}
	return number, nil
}

func (h *Handler) searchAudit(w http.ResponseWriter, r *http.Request, p principal) {
	filter, err := auditFilter(r.URL.Query())
	if err != nil {
		h.refuse(w, err)
		return
	}
	entries, err := h.audit.Search(r.Context(), p.account.Roles, filter)
	if err != nil {
		h.refuse(w, err)
		return
	}
	views := make([]auditEntryView, 0, len(entries))
	for _, entry := range entries {
		views = append(views, viewEntry(entry))
	}
	reply(w, http.StatusOK, auditResponse{Entries: views})
}
