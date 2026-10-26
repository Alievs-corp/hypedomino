package rules

type Config struct {
	HandSize                int
	ScoringMultiple         int
	HandPointsLockThreshold int
	ScoringOpeningTile      Tile
	ScoringOpeningPoints    int
}

func DefaultConfig() Config {
	return Config{
		HandSize:                7,
		ScoringMultiple:         5,
		HandPointsLockThreshold: 300,
		ScoringOpeningTile:      Tile{Low: 5, High: 5},
		ScoringOpeningPoints:    10,
	}
}
