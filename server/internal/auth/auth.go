package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
)

const MaxBodyBytes = 16 * 1024

type Verifier interface {
	IssueNonce(ctx context.Context) (string, error)
	Verify(ctx context.Context, provider idtoken.Provider, token, rawNonce string) (idtoken.Identity, error)
}

type Accounts interface {
	SignIn(ctx context.Context, identity idtoken.Identity) (account.Account, bool, error)
}

type Sessions interface {
	Create(account string) string
}

type Handler struct {
	verifier Verifier
	accounts Accounts
	sessions Sessions
	logger   *slog.Logger
	mux      *http.ServeMux
}

func New(verifier Verifier, accounts Accounts, sessions Sessions, logger *slog.Logger) *Handler {
	h := &Handler{verifier: verifier, accounts: accounts, sessions: sessions, logger: logger, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /auth/nonce", h.nonce)
	h.mux.HandleFunc("POST /auth/signin", h.signIn)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type nonceResponse struct {
	Nonce string `json:"nonce"`
}

type signInRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
	Nonce    string `json:"nonce"`
}

type signInResponse struct {
	SessionToken string `json:"session_token"`
	AccountID    string `json:"account_id"`
	Created      bool   `json:"created"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) nonce(w http.ResponseWriter, r *http.Request) {
	nonce, err := h.verifier.IssueNonce(r.Context())
	if err != nil {
		h.logger.Error("nonce issue failed", "error", err)
		reply(w, http.StatusServiceUnavailable, errorResponse{Error: "unavailable"})
		return
	}
	reply(w, http.StatusOK, nonceResponse{Nonce: nonce})
}

func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var request signInRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes)).Decode(&request); err != nil {
		reply(w, http.StatusBadRequest, errorResponse{Error: "malformed"})
		return
	}
	provider, ok := parseProvider(request.Provider)
	if !ok {
		reply(w, http.StatusBadRequest, errorResponse{Error: "unknown_provider"})
		return
	}
	identity, err := h.verifier.Verify(r.Context(), provider, request.IDToken, request.Nonce)
	if err != nil {
		h.refuse(w, err)
		return
	}
	a, created, err := h.accounts.SignIn(r.Context(), identity)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, signInResponse{SessionToken: h.sessions.Create(a.ID), AccountID: a.ID, Created: created})
}

func (h *Handler) refuse(w http.ResponseWriter, err error) {
	status, code := classify(err)
	if status >= http.StatusInternalServerError {
		h.logger.Error("sign-in failed", "error", err)
	} else {
		h.logger.Info("sign-in refused", "reason", code)
	}
	reply(w, status, errorResponse{Error: code})
}

func classify(err error) (int, string) {
	switch {
	case errors.Is(err, idtoken.ErrKeyFetch):
		return http.StatusServiceUnavailable, "provider_unavailable"
	case errors.Is(err, idtoken.ErrExpired), errors.Is(err, idtoken.ErrNotYetValid):
		return http.StatusUnauthorized, "token_expired"
	case errors.Is(err, idtoken.ErrNonce), errors.Is(err, idtoken.ErrNonceUsed):
		return http.StatusUnauthorized, "nonce_rejected"
	case errors.Is(err, idtoken.ErrMalformed), errors.Is(err, idtoken.ErrAlgorithm), errors.Is(err, idtoken.ErrUnknownKey),
		errors.Is(err, idtoken.ErrSignature), errors.Is(err, idtoken.ErrIssuer), errors.Is(err, idtoken.ErrAudience),
		errors.Is(err, idtoken.ErrSubject), errors.Is(err, idtoken.ErrUnknownProvider), errors.Is(err, account.ErrInvalidIdentity):
		return http.StatusUnauthorized, "token_rejected"
	case errors.Is(err, account.ErrBanned):
		return http.StatusForbidden, "banned"
	case errors.Is(err, account.ErrDeleted):
		return http.StatusGone, "deleted"
	}
	return http.StatusInternalServerError, "internal"
}

func parseProvider(name string) (idtoken.Provider, bool) {
	switch name {
	case idtoken.ProviderGoogle.String():
		return idtoken.ProviderGoogle, true
	case idtoken.ProviderApple.String():
		return idtoken.ProviderApple, true
	}
	return 0, false
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
