package matchresult

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

const Route = "GET /matches/{id}/result"

type Reader interface {
	Read(ctx context.Context, account, matchID string) (Result, error)
}

type Handler struct {
	verifier token.Verifier
	reader   Reader
	logger   *slog.Logger
	mux      *http.ServeMux
}

func NewHandler(verifier token.Verifier, reader Reader, logger *slog.Logger) *Handler {
	h := &Handler{verifier: verifier, reader: reader, logger: logger, mux: http.NewServeMux()}
	h.mux.HandleFunc(Route, h.result)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) result(w http.ResponseWriter, r *http.Request) {
	claims, err := token.Authorize(h.verifier, r)
	switch {
	case errors.Is(err, token.ErrExpired):
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_expired"})
		return
	case err != nil:
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_rejected"})
		return
	}
	result, err := h.reader.Read(r.Context(), claims.Account, r.PathValue("id"))
	switch {
	case errors.Is(err, ErrNotPlayed):
		reply(w, http.StatusNotFound, errorResponse{Error: ErrNotPlayed.Error()})
	case err != nil:
		h.logger.Error("match result request failed", "error", err)
		reply(w, http.StatusInternalServerError, errorResponse{Error: "internal"})
	default:
		reply(w, http.StatusOK, result)
	}
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
