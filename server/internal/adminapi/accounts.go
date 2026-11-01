package adminapi

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/adminaccount"
	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

type changeNote struct {
	Reason string `json:"reason"`
	Ref    string `json:"ref"`
}

type createAccountRequest struct {
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name"`
	Roles       []rbac.Role `json:"roles"`
	changeNote
}

type setRolesRequest struct {
	Roles []rbac.Role `json:"roles"`
	changeNote
}

type accountsResponse struct {
	Accounts []accountView `json:"accounts"`
}

type auditedAccount struct {
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name"`
	Roles       []rbac.Role `json:"roles"`
	State       string      `json:"state"`
}

func snapshotOf(a adminaccount.Account) json.RawMessage {
	raw, _ := json.Marshal(auditedAccount{Email: a.Email, DisplayName: a.DisplayName, Roles: a.Roles, State: string(a.State)})
	return raw
}

func accountEntry(action string, note changeNote, before, after adminaccount.Account) audit.Entry {
	entry := audit.Entry{Action: action, Object: audit.Object{Type: rbac.AdminAccounts, ID: after.ID}, Reason: note.Reason, Ref: note.Ref}
	if before.ID != "" {
		entry.Before = snapshotOf(before)
	}
	entry.After = snapshotOf(after)
	return entry
}

type accountChange func(accounts *adminaccount.Store, now time.Time) (adminaccount.Account, audit.Entry, error)

func (h *Handler) changeAccount(ctx context.Context, p principal, change accountChange) (adminaccount.Account, error) {
	var changed adminaccount.Account
	err := h.audited(ctx, p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		account, entry, err := change(h.accounts.In(tx), now)
		changed = account
		return []audit.Entry{entry}, err
	})
	return changed, err
}

func (h *Handler) replyAccount(w http.ResponseWriter, status int, account adminaccount.Account, err error) {
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, status, viewAccount(account))
}

func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request, p principal) {
	accounts, err := h.accounts.List(r.Context())
	if err != nil {
		h.refuse(w, err)
		return
	}
	views := make([]accountView, 0, len(accounts))
	for _, account := range accounts {
		views = append(views, viewAccount(account))
	}
	reply(w, http.StatusOK, accountsResponse{Accounts: views})
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request, p principal) {
	var request createAccountRequest
	if !decode(w, r, &request) {
		return
	}
	if slices.Contains(request.Roles, rbac.SuperAdmin) {
		h.refuse(w, errApprovalRequired)
		return
	}
	account, err := h.changeAccount(r.Context(), p, func(accounts *adminaccount.Store, now time.Time) (adminaccount.Account, audit.Entry, error) {
		created, err := accounts.Create(r.Context(), request.Email, request.DisplayName, request.Roles, now)
		return created, accountEntry("create", request.changeNote, adminaccount.Account{}, created), err
	})
	h.replyAccount(w, http.StatusCreated, account, err)
}

func (h *Handler) setRoles(w http.ResponseWriter, r *http.Request, p principal) {
	var request setRolesRequest
	if !decode(w, r, &request) {
		return
	}
	id := r.PathValue("id")
	if id == p.account.ID {
		h.refuse(w, errOwnAccount)
		return
	}
	account, err := h.changeAccount(r.Context(), p, func(accounts *adminaccount.Store, now time.Time) (adminaccount.Account, audit.Entry, error) {
		before, err := accounts.Lock(r.Context(), id)
		if err != nil {
			return adminaccount.Account{}, audit.Entry{}, err
		}
		if slices.Contains(request.Roles, rbac.SuperAdmin) && !slices.Contains(before.Roles, rbac.SuperAdmin) {
			return adminaccount.Account{}, audit.Entry{}, errApprovalRequired
		}
		after, err := accounts.SetRoles(r.Context(), id, request.Roles)
		return after, accountEntry("edit", request.changeNote, before, after), err
	})
	h.replyAccount(w, http.StatusOK, account, err)
}

func (h *Handler) disable(w http.ResponseWriter, r *http.Request, p principal) {
	var request changeNote
	if !decode(w, r, &request) {
		return
	}
	id := r.PathValue("id")
	if id == p.account.ID {
		h.refuse(w, errOwnAccount)
		return
	}
	account, err := h.changeAccount(r.Context(), p, func(accounts *adminaccount.Store, now time.Time) (adminaccount.Account, audit.Entry, error) {
		before, err := accounts.Lock(r.Context(), id)
		if err != nil {
			return adminaccount.Account{}, audit.Entry{}, err
		}
		after, err := accounts.Disable(r.Context(), id, now)
		return after, accountEntry("retire", request, before, after), err
	})
	h.replyAccount(w, http.StatusOK, account, err)
}
