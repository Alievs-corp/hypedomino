package rules

type Config struct {
	HandSize                int
	TargetScore             int
	ScoringMultiple         int
	HandPointsLockThreshold int
	FirstHandOpeningTile    Tile
	ScoringOpeningTile      Tile
	ScoringOpeningPoints    int
}

func DefaultConfig() Config {
	return Config{
		HandSize:                7,
		TargetScore:             365,
		ScoringMultiple:         5,
		HandPointsLockThreshold: 300,
		FirstHandOpeningTile:    Tile{Low: 2, High: 3},
		ScoringOpeningTile:      Tile{Low: 5, High: 5},
		ScoringOpeningPoints:    10,
	}
}
