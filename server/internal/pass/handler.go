package pass

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/reward"
	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

type Enterer interface {
	Enter(ctx context.Context, account string, now time.Time) (Season, Progress, int64, error)
}

type Handler struct {
	verifier token.Verifier
	store    Enterer
	logger   *slog.Logger
	now      func() time.Time
}

func NewHandler(verifier token.Verifier, store Enterer, logger *slog.Logger, now func() time.Time) *Handler {
	return &Handler{verifier: verifier, store: store, logger: logger, now: now}
}

type errorResponse struct {
	Error string `json:"error"`
}

type View struct {
	Season        string                           `json:"season"`
	StartsAt      time.Time                        `json:"starts_at"`
	EndsAt        time.Time                        `json:"ends_at"`
	ConfigVersion int64                            `json:"config_version"`
	Levels        int                              `json:"levels"`
	XP            int64                            `json:"xp"`
	Level         int                              `json:"level"`
	LevelXP       int64                            `json:"level_xp"`
	NextLevelXP   *int64                           `json:"next_level_xp"`
	Premium       bool                             `json:"premium"`
	Tracks        map[string]map[int]reward.Bundle `json:"tracks"`
}

func NewView(season Season, progress Progress, version int64) View {
	level := season.Curve.LevelFor(progress.XP)
	view := View{
		Season:        season.ID,
		StartsAt:      season.Start,
		EndsAt:        season.End,
		ConfigVersion: version,
		Levels:        season.Curve.Levels,
		XP:            progress.XP,
		Level:         level,
		LevelXP:       season.Curve.XpToReach(level),
		Premium:       progress.Premium,
		Tracks:        map[string]map[int]reward.Bundle{Free: season.Free, Premium: season.Premium},
	}
	if level < season.Curve.Levels {
		next := season.Curve.XpToReach(level + 1)
		view.NextLevelXP = &next
	}
	return view
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		reply(w, http.StatusMethodNotAllowed, errorResponse{Error: "method_not_allowed"})
		return
	}
	claims, err := token.Authorize(h.verifier, r)
	switch {
	case errors.Is(err, token.ErrExpired):
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_expired"})
		return
	case err != nil:
		reply(w, http.StatusUnauthorized, errorResponse{Error: "access_rejected"})
		return
	}
	season, progress, version, err := h.store.Enter(r.Context(), claims.Account, h.now())
	switch {
	case errors.Is(err, ErrNoSeason):
		reply(w, http.StatusNotFound, errorResponse{Error: "pass.no_season"})
	case err != nil:
		h.logger.Error("pass view failed", "error", err)
		reply(w, http.StatusInternalServerError, errorResponse{Error: "internal"})
	default:
		reply(w, http.StatusOK, NewView(season, progress, version))
	}
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
