package rating

import "context"

const SeasonHistoryLimit = 20

type PastReward struct {
	Granted  bool     `json:"granted"`
	Coins    int64    `json:"coins"`
	BundleID string   `json:"bundle_id,omitempty"`
	Emblems  []string `json:"emblems"`
}

type PastMode struct {
	Mode     string     `json:"mode"`
	Rank     string     `json:"rank"`
	Tier     string     `json:"tier"`
	LP       int        `json:"lp"`
	Position int        `json:"position"`
	Reward   PastReward `json:"reward"`
}

type PastSeason struct {
	SeasonClock
	Modes []PastMode `json:"modes"`
}

type SeasonHistory struct {
	Seasons []PastSeason `json:"seasons"`
}

func (s *Store) SeasonHistory(ctx context.Context, account string) (SeasonHistory, error) {
	rows, err := s.pool.Query(ctx, `with played as (
		select s.id, s.name_key, s.starts_at, s.ends_at from ranked_seasons s
		where s.id in (select season_id from season_results where account_id = $1)
		order by s.ends_at desc limit $2
	)
	select p.id, p.name_key, p.starts_at, p.ends_at, r.mode, r.rank_index, r.apex_tier, r.lp, r.position,
		w.account_id is not null, coalesce(w.coins, 0), coalesce(w.bundle_id, ''), coalesce(w.emblems, '{}')
	from played p
	join season_results r on r.season_id = p.id and r.account_id = $1
	left join season_rewards w on w.season_id = r.season_id and w.mode = r.mode and w.account_id = r.account_id
	order by p.ends_at desc, r.mode`, account, SeasonHistoryLimit)
	if err != nil {
		return SeasonHistory{}, err
	}
	defer rows.Close()
	history := SeasonHistory{Seasons: []PastSeason{}}
	for rows.Next() {
		var season SeasonClock
		var mode PastMode
		var rank Rank
		var apex *string
		if err := rows.Scan(&season.ID, &season.NameKey, &season.StartsAt, &season.EndsAt, &mode.Mode, &rank.Index, &apex, &rank.LP, &mode.Position,
			&mode.Reward.Granted, &mode.Reward.Coins, &mode.Reward.BundleID, &mode.Reward.Emblems); err != nil {
			return SeasonHistory{}, err
		}
		history.Seasons = appendPast(history.Seasons, season, finalMode(mode, rank, apex))
	}
	return history, rows.Err()
}

func finalMode(mode PastMode, rank Rank, apex *string) PastMode {
	mode.Tier, mode.LP, mode.Rank = tierOf(rank, apex), rank.LP, rank.String()
	if rank.Apex() {
		mode.Rank = mode.Tier
	}
	if mode.Reward.Emblems == nil {
		mode.Reward.Emblems = []string{}
	}
	return mode
}

func appendPast(seasons []PastSeason, season SeasonClock, mode PastMode) []PastSeason {
	if last := len(seasons) - 1; last >= 0 && seasons[last].ID == season.ID {
		seasons[last].Modes = append(seasons[last].Modes, mode)
		return seasons
	}
	season.StartsAt, season.EndsAt = season.StartsAt.UTC(), season.EndsAt.UTC()
	return append(seasons, PastSeason{SeasonClock: season, Modes: []PastMode{mode}})
}
