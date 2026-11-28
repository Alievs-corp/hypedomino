package matchresult

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/matchreward"
	"github.com/Alievs-corp/hypedomino/server/internal/matchstore"
	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/wallet"
)

var ErrNotPlayed = errors.New("match.not_played")

const (
	PartCoins    = "coins"
	PartPassXp   = "pass_xp"
	PartMissions = "missions"
	PartRating   = "rating"
)

var partOfTopic = map[string]string{
	matchstore.TopicEconomy:  PartCoins,
	matchstore.TopicXp:       PartPassXp,
	matchstore.TopicMissions: PartMissions,
	matchstore.TopicRating:   PartRating,
}

type Coins struct {
	Match    int64 `json:"match"`
	FirstWin int64 `json:"first_win"`
}

type Mission struct {
	Mission string   `json:"mission"`
	Period  string   `json:"period"`
	Coins   int64    `json:"coins"`
	PassXP  int      `json:"pass_xp"`
	Medals  []string `json:"medals"`
}

type RatingChange struct {
	Placement  bool   `json:"placement"`
	LPDelta    int    `json:"lp_delta"`
	RankBefore string `json:"rank_before,omitempty"`
	LPBefore   int    `json:"lp_before"`
	RankAfter  string `json:"rank_after,omitempty"`
	LPAfter    int    `json:"lp_after"`
}

type Result struct {
	Match        string        `json:"match"`
	Kind         string        `json:"kind"`
	Result       string        `json:"result,omitempty"`
	Pending      []string      `json:"pending"`
	RetryAfterMs int64         `json:"retry_after_ms,omitempty"`
	Coins        Coins         `json:"coins"`
	PassXP       int64         `json:"pass_xp"`
	Missions     []Mission     `json:"missions"`
	Rating       *RatingChange `json:"rating"`
}

type Store struct {
	pool       *pgxpool.Pool
	retryAfter time.Duration
}

func NewStore(pool *pgxpool.Pool, retryAfter time.Duration) *Store {
	return &Store{pool: pool, retryAfter: retryAfter}
}

type seat struct {
	kind    string
	endedAt *time.Time
	result  *string
}

func (s *Store) Read(ctx context.Context, account, matchID string) (Result, error) {
	if _, err := uuid.Parse(matchID); err != nil {
		return Result{}, ErrNotPlayed
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)
	played, err := seatOf(ctx, tx, account, matchID)
	if err != nil {
		return Result{}, err
	}
	result := Result{Match: matchID, Kind: played.kind, Pending: []string{}, Missions: []Mission{}}
	if played.endedAt == nil {
		result.Pending = partsOf(expectedTopics(played.kind))
		result.RetryAfterMs = s.retryAfter.Milliseconds()
		return result, nil
	}
	if played.result != nil {
		result.Result = *played.result
	}
	if result.Pending, err = pendingParts(ctx, tx, matchID, played.kind); err != nil {
		return Result{}, err
	}
	if len(result.Pending) > 0 {
		result.RetryAfterMs = s.retryAfter.Milliseconds()
	}
	if result.Coins, err = matchCoins(ctx, tx, account, matchID, *played.endedAt); err != nil {
		return Result{}, err
	}
	if result.PassXP, err = matchPassXp(ctx, tx, account, matchID); err != nil {
		return Result{}, err
	}
	if result.Missions, err = completedMissions(ctx, tx, account, *played.endedAt); err != nil {
		return Result{}, err
	}
	if result.Rating, err = ratingChange(ctx, tx, account, matchID); err != nil {
		return Result{}, err
	}
	return result, nil
}

func seatOf(ctx context.Context, tx pgx.Tx, account, matchID string) (seat, error) {
	var played seat
	err := tx.QueryRow(ctx, `select m.kind, m.ended_at, p.result from match_participants p
		join matches m on m.id = p.match_id and m.started_at = p.started_at
		where p.match_id = $1 and p.account_id = $2`, matchID, account).Scan(&played.kind, &played.endedAt, &played.result)
	if errors.Is(err, pgx.ErrNoRows) {
		return seat{}, ErrNotPlayed
	}
	return played, err
}

func expectedTopics(kind string) []string {
	var topics []string
	for _, message := range matchstore.MatchMessages(kind, "", nil) {
		if _, ok := partOfTopic[message.Topic]; ok {
			topics = append(topics, message.Topic)
		}
	}
	return topics
}

func partsOf(topics []string) []string {
	parts := make([]string, 0, len(topics))
	for _, topic := range topics {
		parts = append(parts, partOfTopic[topic])
	}
	return parts
}

func pendingParts(ctx context.Context, tx pgx.Tx, matchID, kind string) ([]string, error) {
	expected := expectedTopics(kind)
	rows, err := tx.Query(ctx, `select topic from outbox where key = $1 and topic = any($2) and published_at is null`, "match:"+matchID, expected)
	if err != nil {
		return nil, err
	}
	unpublished, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	pending := slices.DeleteFunc(expected, func(topic string) bool { return !slices.Contains(unpublished, topic) })
	return partsOf(pending), nil
}

func matchCoins(ctx context.Context, tx pgx.Tx, account, matchID string, endedAt time.Time) (Coins, error) {
	keys := []string{"match:" + matchID, "first_win:" + matchreward.EconomyDay(endedAt).Format(time.DateOnly)}
	rows, err := tx.Query(ctx, `select reason, amount from ledger_entries
		where account_id = $1 and currency = $2 and idempotency_key = any($3) and source_type = $4 and source_ref = $5`,
		account, string(wallet.Coin), keys, matchreward.SourceMatch, matchID)
	if err != nil {
		return Coins{}, err
	}
	defer rows.Close()
	var coins Coins
	for rows.Next() {
		var reason string
		var amount int64
		if err := rows.Scan(&reason, &amount); err != nil {
			return Coins{}, err
		}
		switch wallet.Reason(reason) {
		case wallet.MatchReward:
			coins.Match += amount
		case wallet.FirstWin:
			coins.FirstWin += amount
		}
	}
	return coins, rows.Err()
}

func matchPassXp(ctx context.Context, tx pgx.Tx, account, matchID string) (int64, error) {
	var xp int64
	err := tx.QueryRow(ctx, `select coalesce(sum(amount), 0) from pass_xp_events where account_id = $1 and idempotency_key = $2`, account, "match:"+matchID).Scan(&xp)
	return xp, err
}

type missionGrant struct {
	Coins  int64    `json:"coins"`
	PassXP int      `json:"pass_xp"`
	Medals []string `json:"medals"`
}

func completedMissions(ctx context.Context, tx pgx.Tx, account string, endedAt time.Time) ([]Mission, error) {
	rows, err := tx.Query(ctx, `select a.mission_id, a.period, g.report from mission_assignments a
		left join reward_grants g on g.account_id = a.account_id and g.idempotency_key = 'mission:' || a.mission_id || ':' || a.period and g.revoked_at is null
		where a.account_id = $1 and a.completed_at = $2 order by a.mission_id, a.period`, account, endedAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	missions := []Mission{}
	for rows.Next() {
		var mission Mission
		var report []byte
		if err := rows.Scan(&mission.Mission, &mission.Period, &report); err != nil {
			return nil, err
		}
		var grant missionGrant
		if report != nil {
			if err := json.Unmarshal(report, &grant); err != nil {
				return nil, err
			}
		}
		mission.Coins, mission.PassXP, mission.Medals = grant.Coins, grant.PassXP, grant.Medals
		if mission.Medals == nil {
			mission.Medals = []string{}
		}
		missions = append(missions, mission)
	}
	return missions, rows.Err()
}

func ratingChange(ctx context.Context, tx pgx.Tx, account, matchID string) (*RatingChange, error) {
	var change RatingChange
	var before, after int
	err := tx.QueryRow(ctx, `select placement, lp_delta, rank_before, lp_before, rank_after, lp_after from rating_changes where match_id = $1 and account_id = $2`,
		matchID, account).Scan(&change.Placement, &change.LPDelta, &before, &change.LPBefore, &after, &change.LPAfter)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if change.Placement {
		return &RatingChange{Placement: true}, nil
	}
	change.RankBefore, change.RankAfter = rating.Rank{Index: before}.String(), rating.Rank{Index: after}.String()
	return &change, nil
}
