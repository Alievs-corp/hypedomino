package rules

func PlacementScore(total int, config Config) int {
	if total <= 0 || total%config.ScoringMultiple != 0 {
		return 0
	}
	return total
}

func OpeningScore(config Config, firstHand bool, tile Tile, teamScoreAtHandStart int) int {
	if firstHand || tile != config.ScoringOpeningTile || teamScoreAtHandStart >= config.HandPointsLockThreshold {
		return 0
	}
	return config.ScoringOpeningPoints
}

func ScoreLevel(score int, config Config) int {
	if score <= 0 {
		return 0
	}
	return score / config.ScoringMultiple
}
