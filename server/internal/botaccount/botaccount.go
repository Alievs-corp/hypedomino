package botaccount

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/token"
)

const RefusalCode = "bot.casual_only"

var ErrCasualOnly = errors.New(RefusalCode)

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func Is(ctx context.Context, q Querier, account string) (bool, error) {
	var bot bool
	err := q.QueryRow(ctx, `select exists (select 1 from accounts where id = $1 and is_bot)`, account).Scan(&bot)
	return bot, err
}

func Check(q Querier) func(ctx context.Context, account string) (bool, error) {
	return func(ctx context.Context, account string) (bool, error) {
		return Is(ctx, q, account)
	}
}

func CasualOnly(verifier token.Verifier, isBot func(ctx context.Context, account string) (bool, error), logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := token.Authorize(verifier, r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		bot, err := isBot(r.Context(), claims.Account)
		if err != nil {
			logger.Error("bot account check failed", "error", err)
			refuse(w, http.StatusInternalServerError, "internal")
			return
		}
		if bot {
			refuse(w, http.StatusForbidden, RefusalCode)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func refuse(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code})
}
