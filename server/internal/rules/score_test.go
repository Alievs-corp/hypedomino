package rules

import "testing"

func TestPlacementScore(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		total int
		score int
		level int
	}{
		{0, 0, 0},
		{-5, 0, 0},
		{3, 0, 0},
		{5, 5, 1},
		{7, 0, 0},
		{10, 10, 2},
		{11, 0, 0},
		{15, 15, 3},
		{20, 20, 4},
		{25, 25, 5},
		{45, 45, 9},
	}
	for _, c := range cases {
		score := PlacementScore(c.total, config)
		if score != c.score {
			t.Errorf("PlacementScore(%d) = %d, want %d", c.total, score, c.score)
		}
		if level := ScoreLevel(score, config); level != c.level {
			t.Errorf("ScoreLevel(%d) = %d, want %d", score, level, c.level)
		}
	}
}

func TestPlacementScoreIgnoresTeamScore(t *testing.T) {
	config := DefaultConfig()
	board := NewBoard(MustTile(2, 3))
	mustPlace(t, &board, "3-3", EndRight)
	mustPlace(t, &board, "2-4", EndLeft)
	total := board.OpenEnds().Total()
	if total != 10 || PlacementScore(total, config) != 10 {
		t.Fatalf("total %d score %d", total, PlacementScore(total, config))
	}
}

func TestOpeningScore(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		name      string
		firstHand bool
		tile      Tile
		teamScore int
		score     int
	}{
		{"5-5 in a later hand", false, MustTile(5, 5), 140, 10},
		{"5-5 at 299", false, MustTile(5, 5), 299, 10},
		{"5-5 at exactly 300", false, MustTile(5, 5), 300, 0},
		{"5-5 above 300", false, MustTile(5, 5), 345, 0},
		{"5-5 in the first hand", true, MustTile(5, 5), 0, 0},
		{"6-4 adds to 10 but is not 5-5", false, MustTile(4, 6), 140, 0},
		{"1-4 adds to 5", false, MustTile(1, 4), 50, 0},
		{"2-3 first hand", true, MustTile(2, 3), 0, 0},
	}
	for _, c := range cases {
		if score := OpeningScore(config, c.firstHand, c.tile, c.teamScore); score != c.score {
			t.Errorf("%s: score %d, want %d", c.name, score, c.score)
		}
	}
}

func TestScoreFunctionsDoNotAllocate(t *testing.T) {
	config := DefaultConfig()
	allocations := testing.AllocsPerRun(100, func() {
		PlacementScore(10, config)
		OpeningScore(config, false, MustTile(5, 5), 140)
		ScoreLevel(10, config)
	})
	if allocations != 0 {
		t.Fatalf("scoring allocates %.0f times", allocations)
	}
}
