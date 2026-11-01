package adminapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

const (
	TemporaryBanAction     = "temporary_ban"
	LiftTemporaryBanAction = "lift_temporary_ban"
)

var (
	errBanNotConfigured = errors.New("adminapi: HD_TEMP_BAN_HOURS is not set")
	errNotBanned        = errors.New("adminapi: the account has no running temporary ban")
)

type banState struct {
	BannedUntil *time.Time `json:"banned_until"`
}

type bannedState struct {
	BannedUntil     *time.Time `json:"banned_until"`
	SessionsRevoked int        `json:"sessions_revoked"`
}

type banView struct {
	AccountID       string     `json:"account_id"`
	BannedUntil     *time.Time `json:"banned_until"`
	SessionsRevoked int        `json:"sessions_revoked"`
}

func (h *Handler) changeBan(w http.ResponseWriter, r *http.Request, p principal, change func(tx pgx.Tx, accountID string, current *time.Time, now time.Time) (banView, audit.Entry, error)) {
	var note changeNote
	if !decode(w, r, &note) {
		return
	}
	accountID := r.PathValue("id")
	var view banView
	err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		if err := h.sensitive(p, note, now); err != nil {
			return nil, err
		}
		if err := playerrecord.Lock(r.Context(), tx, accountID); err != nil {
			return nil, err
		}
		var current *time.Time
		if err := tx.QueryRow(r.Context(), `select banned_until from accounts where id = $1`, accountID).Scan(&current); err != nil {
			return nil, err
		}
		if current != nil {
			utc := current.UTC()
			current = &utc
		}
		var entry audit.Entry
		var err error
		if view, entry, err = change(tx, accountID, current, now); err != nil {
			return nil, err
		}
		entry.Object, entry.Reason, entry.Ref = audit.Object{Type: rbac.Players, ID: accountID}, note.Reason, note.Ref
		return []audit.Entry{entry}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, view)
}

func (h *Handler) banTemporarily(w http.ResponseWriter, r *http.Request, p principal) {
	if h.config.TempBan <= 0 {
		h.refuse(w, errBanNotConfigured)
		return
	}
	h.changeBan(w, r, p, func(tx pgx.Tx, accountID string, current *time.Time, now time.Time) (banView, audit.Entry, error) {
		until := now.Add(h.config.TempBan).UTC().Truncate(time.Microsecond)
		if _, err := tx.Exec(r.Context(), `update accounts set banned_until = $2 where id = $1`, accountID, until); err != nil {
			return banView{}, audit.Entry{}, err
		}
		revoked, err := token.RevokeAccountIn(r.Context(), tx, accountID, token.ReasonBanned, now)
		if err != nil {
			return banView{}, audit.Entry{}, err
		}
		entry := audit.Entry{Action: TemporaryBanAction, Before: marshal(banState{BannedUntil: current}), After: marshal(bannedState{BannedUntil: &until, SessionsRevoked: revoked})}
		return banView{AccountID: accountID, BannedUntil: &until, SessionsRevoked: revoked}, entry, nil
	})
}

func (h *Handler) liftTemporaryBan(w http.ResponseWriter, r *http.Request, p principal) {
	h.changeBan(w, r, p, func(tx pgx.Tx, accountID string, current *time.Time, now time.Time) (banView, audit.Entry, error) {
		if current == nil || !now.Before(*current) {
			return banView{}, audit.Entry{}, errNotBanned
		}
		if _, err := tx.Exec(r.Context(), `update accounts set banned_until = null where id = $1`, accountID); err != nil {
			return banView{}, audit.Entry{}, err
		}
		entry := audit.Entry{Action: LiftTemporaryBanAction, Before: marshal(banState{BannedUntil: current}), After: marshal(banState{})}
		return banView{AccountID: accountID}, entry, nil
	})
}
