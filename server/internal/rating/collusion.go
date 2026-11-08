package rating

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/risk"
)

func (s *Store) SetCollusionWeight(weight int) {
	s.collusionWeight = weight
}

func collusive(kind string) bool {
	return kind == SignalWinTrade || kind == SignalFeeding || kind == SignalPartnerFit
}

func (s *Store) reportCollusion(ctx context.Context, tx pgx.Tx, side Signal, at time.Time) error {
	if s.collusionWeight <= 0 || !collusive(side.Kind) {
		return nil
	}
	_, err := risk.RecordIn(ctx, tx, risk.Signal{
		Account: side.Account, Kind: risk.KindCollusion, Weight: s.collusionWeight,
		Dedupe: side.Kind + ":" + side.Counterpart + ":" + side.Mode + ":" + at.UTC().Format(time.DateOnly),
		Evidence: map[string]any{
			"signal": side.Kind, "counterpart": side.Counterpart, "mode": side.Mode,
			"matches": side.Matches, "wins": side.Wins, "window_hours": side.WindowHours,
		},
		At: at,
	})
	return err
}
