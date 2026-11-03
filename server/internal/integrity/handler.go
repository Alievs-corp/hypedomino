package integrity

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

const MaxBodyBytes = 1024

type Handler struct {
	verifier token.Verifier
	nonces   *Nonces
	logger   *slog.Logger
	mux      *http.ServeMux
}

func NewHandler(verifier token.Verifier, nonces *Nonces, logger *slog.Logger) *Handler {
	h := &Handler{verifier: verifier, nonces: nonces, logger: logger, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /integrity/nonce", h.nonce)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type nonceRequest struct {
	Purpose string `json:"purpose"`
}

type nonceResponse struct {
	Nonce     string `json:"nonce"`
	ExpiresIn int    `json:"expires_in"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) nonce(w http.ResponseWriter, r *http.Request) {
	claims, err := token.Authorize(h.verifier, r)
	switch {
	case errors.Is(err, token.ErrExpired):
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_expired"})
		return
	case err != nil:
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_rejected"})
		return
	}
	var request nonceRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes)).Decode(&request) != nil {
		reply(w, http.StatusBadRequest, errorResponse{Error: "bad_request"})
		return
	}
	purpose, ok := parsePurpose(request.Purpose)
	if !ok {
		reply(w, http.StatusBadRequest, errorResponse{Error: "unknown_purpose"})
		return
	}
	nonce, err := h.nonces.Issue(r.Context(), Binding{Account: claims.Account, Session: claims.Session, Purpose: purpose})
	if err != nil {
		h.logger.Error("integrity nonce not issued", "account", claims.Account, "error", err)
		reply(w, http.StatusServiceUnavailable, errorResponse{Error: "unavailable"})
		return
	}
	reply(w, http.StatusOK, nonceResponse{Nonce: nonce.Value, ExpiresIn: int(h.nonces.config.TTL.Seconds())})
}

func parsePurpose(name string) (Purpose, bool) {
	for purpose := PurposeSessionStart; purpose.valid(); purpose++ {
		if purpose.String() == name {
			return purpose, true
		}
	}
	return 0, false
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
