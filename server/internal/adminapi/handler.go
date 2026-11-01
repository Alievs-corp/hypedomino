package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/adminaccount"
	"github.com/Alievs-corp/hypedomino/server/internal/adminsession"
	"github.com/Alievs-corp/hypedomino/server/internal/approval"
	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/idtoken"
	"github.com/Alievs-corp/hypedomino/server/internal/playerrecord"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/shop"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

const MaxBodyBytes = 16 * 1024

type Config struct {
	Environment  string
	GoogleDomain string
	Session      adminsession.Config
	ApprovalTTL  time.Duration
	Shop         shop.Config

	GrantApprovalCoins int64
	TempBan            time.Duration
}

type Verifier interface {
	IssueNonce(ctx context.Context) (string, error)
	Verify(ctx context.Context, provider idtoken.Provider, token, rawNonce string) (idtoken.Identity, error)
}

type Handler struct {
	config    Config
	verifier  Verifier
	pool      *pgxpool.Pool
	accounts  *adminaccount.Store
	sessions  *adminsession.Store
	audit     *audit.Store
	strings   *gamestrings.Store
	approvals *approval.Store
	wallet    *wallet.Store
	now       func() time.Time
	logger    *slog.Logger
	mux       *http.ServeMux
}

func New(config Config, verifier Verifier, pool *pgxpool.Pool, now func() time.Time, logger *slog.Logger) *Handler {
	h := &Handler{
		config:    config,
		verifier:  verifier,
		pool:      pool,
		accounts:  adminaccount.NewStore(pool),
		sessions:  adminsession.NewStore(pool, config.Session),
		audit:     audit.NewStore(pool),
		strings:   gamestrings.NewStore(pool),
		approvals: approval.NewStore(pool, approval.Config{TTL: config.ApprovalTTL}),
		wallet:    wallet.NewStore(pool),
		now:       now,
		logger:    logger,
		mux:       http.NewServeMux(),
	}
	h.mux.HandleFunc("POST /admin/nonce", h.nonce)
	h.mux.HandleFunc("POST /admin/session", h.signIn)
	h.mux.HandleFunc("POST /admin/session/step-up", h.signedIn(h.stepUp))
	h.mux.HandleFunc("DELETE /admin/session", h.signedIn(h.signOut))
	h.mux.HandleFunc("GET /admin/me", h.signedIn(h.me))
	h.mux.HandleFunc("GET /admin/audit", h.allowed(rbac.AuditLog, rbac.View, h.searchAudit))
	h.mux.HandleFunc("GET /admin/accounts", h.allowed(rbac.AdminAccounts, rbac.View, h.listAccounts))
	h.mux.HandleFunc("POST /admin/accounts", h.allowed(rbac.AdminAccounts, rbac.Create, h.createAccount))
	h.mux.HandleFunc("PUT /admin/accounts/{id}/roles", h.allowed(rbac.AdminAccounts, rbac.Edit, h.setRoles))
	h.mux.HandleFunc("POST /admin/accounts/{id}/disable", h.allowed(rbac.AdminAccounts, rbac.Retire, h.disable))
	h.mux.HandleFunc("GET /admin/approvals", h.signedIn(h.pendingApprovals))
	h.mux.HandleFunc("POST /admin/approvals/{id}/approve", h.signedIn(h.approve))
	h.mux.HandleFunc("POST /admin/approvals/{id}/reject", h.signedIn(h.reject))
	h.mux.HandleFunc("GET /admin/config/rules", h.allowed(rbac.GameConfig, rbac.View, h.rules))
	h.mux.HandleFunc("PUT /admin/config/rules", h.allowed(rbac.GameConfig, rbac.Publish, h.publishRules))
	h.mux.HandleFunc("GET /admin/config/rules/history", h.allowed(rbac.GameConfig, rbac.View, h.rulesHistory))
	h.mux.HandleFunc("POST /admin/config/rules/rollback", h.allowed(rbac.GameConfig, rbac.Publish, h.rollbackRules))
	h.mux.HandleFunc("GET /admin/players", h.allowed(rbac.Players, rbac.View, h.searchPlayers))
	h.mux.HandleFunc("GET /admin/players/{id}", h.allowed(rbac.Players, rbac.View, h.playerRecord))
	h.mux.HandleFunc("POST /admin/players/{id}/coins", h.allowed(rbac.Players, rbac.ActOnPlayer, h.adjustCoins))
	h.mux.HandleFunc("POST /admin/players/{id}/temporary-ban", h.allowed(rbac.Players, rbac.ActOnPlayer, h.banTemporarily))
	h.mux.HandleFunc("DELETE /admin/players/{id}/temporary-ban", h.allowed(rbac.Players, rbac.ActOnPlayer, h.liftTemporaryBan))
	h.mux.HandleFunc("GET /admin/config/shop", h.allowed(rbac.ShopOffers, rbac.View, h.shopCatalog))
	h.mux.HandleFunc("PUT /admin/config/shop", h.allowed(rbac.ShopOffers, rbac.Publish, h.publishShop))
	h.mux.HandleFunc("GET /admin/config/shop/history", h.allowed(rbac.ShopOffers, rbac.View, h.shopHistory))
	h.mux.HandleFunc("POST /admin/config/shop/rollback", h.allowed(rbac.ShopOffers, rbac.Publish, h.rollbackShop))
	h.mux.HandleFunc("GET /admin/strings", h.allowed(rbac.GameStrings, rbac.View, h.listStrings))
	h.mux.HandleFunc("PUT /admin/strings/{key}", h.allowed(rbac.GameStrings, rbac.View, h.saveString))
	h.mux.HandleFunc("PUT /admin/strings/{key}/values/{language}", h.allowed(rbac.GameStrings, rbac.View, h.setValue))
	h.mux.HandleFunc("GET /admin/strings/export", h.allowed(rbac.GameStrings, rbac.View, h.exportStrings))
	h.mux.HandleFunc("POST /admin/strings/import", h.allowed(rbac.GameStrings, rbac.Edit, h.importStrings))
	h.mux.HandleFunc("GET /admin/strings/active", h.allowed(rbac.GameStrings, rbac.View, h.activeStrings))
	h.mux.HandleFunc("PUT /admin/strings/active", h.allowed(rbac.GameStrings, rbac.Publish, h.activateStrings))
	h.mux.HandleFunc("GET /admin/strings/bundles", h.allowed(rbac.GameStrings, rbac.View, h.listBundles))
	h.mux.HandleFunc("POST /admin/strings/bundles/{language}", h.allowed(rbac.GameStrings, rbac.Publish, h.publishBundle))
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type principal struct {
	account adminaccount.Account
	session adminsession.Session
	source  audit.Source
}

func (p principal) actor() audit.Actor {
	return audit.Actor{ID: p.account.ID, Roles: p.account.Roles}
}

func (h *Handler) audited(ctx context.Context, p principal, change func(tx pgx.Tx, now time.Time) ([]audit.Entry, error)) error {
	now := h.now()
	return pgx.BeginFunc(ctx, h.pool, func(tx pgx.Tx) error {
		entries, err := change(tx, now)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if _, err := h.audit.In(tx).Append(ctx, h.stamp(p, entry), now); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *Handler) stamp(p principal, entry audit.Entry) audit.Entry {
	entry.Actor, entry.Environment, entry.Source = p.actor(), h.config.Environment, p.source
	return entry
}

func (h *Handler) sensitive(p principal, note changeNote, now time.Time) error {
	if strings.TrimSpace(note.Reason) == "" {
		return errReasonRequired
	}
	if now.After(p.session.AuthenticatedAt.Add(h.config.Session.StepUp)) {
		return errStepUpRequired
	}
	return nil
}

type authedHandler func(w http.ResponseWriter, r *http.Request, p principal)

func (h *Handler) signedIn(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := h.authenticate(r)
		if err != nil {
			h.refuse(w, err)
			return
		}
		next(w, r, p)
	}
}

func (h *Handler) allowed(object rbac.Object, action rbac.Action, next authedHandler) http.HandlerFunc {
	return h.signedIn(func(w http.ResponseWriter, r *http.Request, p principal) {
		if err := h.allows(p, object, action); err != nil {
			h.refuse(w, err)
			return
		}
		next(w, r, p)
	})
}

func (h *Handler) authenticate(r *http.Request) (principal, error) {
	now := h.now()
	session, err := h.sessions.Authenticate(r.Context(), bearer(r), now)
	if err != nil {
		return principal{}, err
	}
	account, err := h.accounts.ByID(r.Context(), session.AccountID)
	if err != nil {
		return principal{}, err
	}
	if account.State != adminaccount.Active {
		if err := h.sessions.Revoke(r.Context(), session.ID, now); err != nil {
			return principal{}, err
		}
		return principal{}, adminsession.ErrRejected
	}
	return principal{account: account, session: session, source: source(r, session.ID)}, nil
}

func source(r *http.Request, sessionID string) audit.Source {
	return audit.Source{SessionID: sessionID, IP: clientIP(r), UserAgent: r.UserAgent()}
}

func clientIP(r *http.Request) netip.Addr {
	if forwarded := r.Header.Values("X-Forwarded-For"); len(forwarded) > 0 {
		hops := strings.Split(forwarded[len(forwarded)-1], ",")
		if addr, err := netip.ParseAddr(strings.TrimSpace(hops[len(hops)-1])); err == nil {
			return addr.Unmap()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}
	addr, _ := netip.ParseAddr(host)
	return addr.Unmap()
}

func bearer(r *http.Request) string {
	scheme, value, _ := strings.Cut(r.Header.Get("Authorization"), " ")
	if !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(value)
}

type errorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

var (
	errNotWorkspace     = errors.New("adminapi: identity is not a verified company account")
	errAccountRefused   = errors.New("adminapi: no active admin account for the identity")
	errDifferentAccount = errors.New("adminapi: step-up identity is another account")
	errGrantSeparately  = errors.New("adminapi: super_admin is granted to an existing account through an approval")
	errNoSuchVersion    = errors.New("adminapi: no such config version")
	errUnsupportedKind  = errors.New("adminapi: this kind of proposal cannot be carried out here yet")
	errOwnAccount       = errors.New("adminapi: an admin cannot change their own account")
	errMalformed        = errors.New("adminapi: malformed request")
	errForbidden        = errors.New("adminapi: the roles do not grant the action")
	errReasonRequired   = errors.New("adminapi: a reason is required")
	errStepUpRequired   = errors.New("adminapi: the action needs a recent sign-in")
)

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	return decodeUpTo(w, r, into, MaxBodyBytes)
}

func decodeUpTo(w http.ResponseWriter, r *http.Request, into any, limit int64) bool {
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit)).Decode(into)
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		reply(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "too_large"})
	case err != nil:
		reply(w, http.StatusBadRequest, errorResponse{Error: "malformed"})
	}
	return err == nil
}

func (h *Handler) refuse(w http.ResponseWriter, err error) {
	status, code := classify(err)
	if status >= http.StatusInternalServerError {
		h.logger.Error("admin request failed", "error", err)
	} else {
		h.logger.Info("admin request refused", "reason", code)
	}
	response := errorResponse{Error: code}
	if status == http.StatusBadRequest {
		response.Detail = err.Error()
	}
	reply(w, status, response)
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
		errors.Is(err, idtoken.ErrSubject), errors.Is(err, idtoken.ErrUnknownProvider):
		return http.StatusUnauthorized, "token_rejected"
	case errors.Is(err, adminsession.ErrRejected):
		return http.StatusUnauthorized, "session_rejected"
	case errors.Is(err, errNotWorkspace):
		return http.StatusForbidden, "not_company_account"
	case errors.Is(err, errAccountRefused):
		return http.StatusForbidden, "account_refused"
	case errors.Is(err, errDifferentAccount):
		return http.StatusForbidden, "different_account"
	case errors.Is(err, adminaccount.ErrNotFound), errors.Is(err, playerrecord.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, adminaccount.ErrTaken):
		return http.StatusConflict, "email_taken"
	case errors.Is(err, errGrantSeparately):
		return http.StatusConflict, "grant_by_approval"
	case errors.Is(err, errUnsupportedKind):
		return http.StatusConflict, "not_supported"
	case errors.Is(err, approval.ErrSelfDecision):
		return http.StatusForbidden, "self_decision"
	case errors.Is(err, approval.ErrNotAllowed):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, approval.ErrDecided):
		return http.StatusConflict, "decided"
	case errors.Is(err, approval.ErrExpired):
		return http.StatusConflict, "expired"
	case errors.Is(err, approval.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, approval.ErrInvalid):
		return http.StatusBadRequest, "invalid"
	case errors.Is(err, errOwnAccount):
		return http.StatusConflict, "own_account"
	case errors.Is(err, adminaccount.ErrInvalid), errors.Is(err, audit.ErrInvalid), errors.Is(err, errMalformed),
		errors.Is(err, playerrecord.ErrEmptyQuery):
		return http.StatusBadRequest, "invalid"
	case errors.Is(err, errBadAdjustment), errors.Is(err, wallet.ErrInvalid), errors.Is(err, wallet.ErrWrongSign):
		return http.StatusBadRequest, "invalid"
	case errors.Is(err, wallet.ErrInsufficient):
		return http.StatusConflict, "insufficient"
	case errors.Is(err, errBanNotConfigured):
		return http.StatusConflict, "not_configured"
	case errors.Is(err, errNotBanned):
		return http.StatusConflict, "not_banned"
	case errors.Is(err, errReasonRequired):
		return http.StatusBadRequest, "reason_required"
	case errors.Is(err, errStepUpRequired):
		return http.StatusForbidden, "step_up_required"
	case errors.Is(err, audit.ErrForbidden), errors.Is(err, errForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, gamestrings.ErrInvalidLocalization), errors.Is(err, rulesconfig.ErrInvalid),
		errors.Is(err, shop.ErrInvalid), errors.Is(err, errOfferRemoved):
		return http.StatusBadRequest, "invalid"
	case errors.Is(err, errNoSuchVersion):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, gamestrings.ErrUnknownKey), errors.Is(err, gamestrings.ErrNoBundle):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, gamestrings.ErrInvalidEntry), errors.Is(err, gamestrings.ErrInvalidValue),
		errors.Is(err, gamestrings.ErrInvalidImport):
		return http.StatusBadRequest, "invalid"
	}
	return http.StatusInternalServerError, "internal"
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
