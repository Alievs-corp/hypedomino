package rules

import "testing"

func TestHandPointsRoundsUpToTheScoringMultiple(t *testing.T) {
	cases := []struct {
		multiple, pips, want int
	}{
		{5, 0, 0}, {5, 1, 5}, {5, 4, 5}, {5, 5, 5}, {5, 6, 10}, {5, 9, 10}, {5, 10, 10},
		{5, 21, 25}, {5, 22, 25}, {5, 25, 25}, {5, 26, 30}, {5, 78, 80}, {5, 168, 170},
		{10, 0, 0}, {10, 1, 10}, {10, 10, 10}, {10, 11, 20},
		{1, 7, 7},
	}
	for _, c := range cases {
		config := DefaultConfig()
		config.ScoringMultiple = c.multiple
		if got := HandPoints(c.pips, config); got != c.want {
			t.Errorf("HandPoints(%d, multiple %d) = %d, want %d", c.pips, c.multiple, got, c.want)
		}
	}
}

func FuzzHandPoints(f *testing.F) {
	f.Add(22, 5)
	f.Add(0, 5)
	f.Add(168, 10)
	f.Fuzz(func(t *testing.T, pips, multiple int) {
		if pips < 0 || pips > FullSetPips() || multiple < 1 || multiple > 50 {
			t.Skip()
		}
		config := DefaultConfig()
		config.ScoringMultiple = multiple
		points := HandPoints(pips, config)
		if (pips == 0) != (points == 0) {
			t.Fatalf("HandPoints(%d) = %d: zero only for zero", pips, points)
		}
		if points < pips || points-pips >= multiple || points%multiple != 0 {
			t.Fatalf("HandPoints(%d, multiple %d) = %d", pips, multiple, points)
		}
	})
}

func FullSetPips() int {
	set := FullSet()
	return Hand(set[:]).Pips()
}

func TestHandStatusText(t *testing.T) {
	for _, status := range []HandStatus{HandContinues, HandDomino, HandBlocked} {
		var parsed HandStatus
		if err := parsed.UnmarshalText([]byte(status.String())); err != nil || parsed != status {
			t.Errorf("%s does not round-trip: %v %v", status, parsed, err)
		}
	}
	var parsed HandStatus
	if err := parsed.UnmarshalText([]byte("draw")); err == nil {
		t.Error("draw parsed as a hand status")
	}
	if HandStatus(9).String() != "invalid" {
		t.Error("out of range status has a name")
	}
}

func TestHandStatusOneVoneBoneyardKeepsHandOpen(t *testing.T) {
	board := lineBoard(t, "6-3", "3-2")
	hands := map[Seat]Hand{SeatX: tiles("1-1", "4-5"), SeatY: tiles("0-0", "1-4")}
	cases := []struct {
		boneyard Boneyard
		want     HandStatus
	}{
		{Boneyard(tiles("0-1", "2-5", "4-4")), HandContinues},
		{Boneyard(tiles("0-1", "2-5")), HandContinues},
		{Boneyard(tiles("2-5")), HandBlocked},
		{Boneyard(nil), HandBlocked},
	}
	for _, c := range cases {
		if got := Mode1v1.HandStatus(&board, hands, c.boneyard); got != c.want {
			t.Errorf("boneyard %v: %s, want %s", c.boneyard, got, c.want)
		}
	}
	hands[SeatY] = nil
	if got := Mode1v1.HandStatus(&board, hands, Boneyard(tiles("0-1", "2-5"))); got != HandDomino {
		t.Errorf("empty hand with a full boneyard: %s", got)
	}
}

func TestDominoResultCountsOnlyTheOpponents(t *testing.T) {
	config := DefaultConfig()
	hands := map[Seat]Hand{
		SeatNorth: tiles("6-6", "5-6", "5-5"),
		SeatEast:  tiles("0-1"),
		SeatSouth: nil,
		SeatWest:  tiles("0-2"),
	}
	scores := HandScores{AtHandStart: TeamPoints{40, 75}, Current: TeamPoints{55, 75}, CarryOver: 20}
	got := DominoResult(Mode2v2, config, SeatSouth, hands, scores)
	want := HandResult{
		Pips:       TeamPoints{33, 3},
		Winner:     TeamA,
		HasWinner:  true,
		Awarded:    TeamPoints{25, 0},
		CarryOver:  0,
		Scores:     TeamPoints{80, 75},
		NextOpener: TeamA,
	}
	if got != want {
		t.Fatalf("DominoResult = %+v, want %+v", got, want)
	}
}

func TestDominoResultLockedTeamDiscardsHandPointsAndCarryOver(t *testing.T) {
	config := DefaultConfig()
	hands := map[Seat]Hand{SeatX: nil, SeatY: tiles("4-5")}
	scores := HandScores{AtHandStart: TeamPoints{300, 100}, Current: TeamPoints{310, 100}, CarryOver: 35}
	got := DominoResult(Mode1v1, config, SeatX, hands, scores)
	if got.Awarded != (TeamPoints{}) || got.Scores != scores.Current || got.CarryOver != 0 {
		t.Fatalf("locked winner still scored: %+v", got)
	}
	if !got.HasWinner || got.Winner != TeamA || got.NextOpener != TeamA || got.Pips != (TeamPoints{0, 9}) {
		t.Fatalf("locked winner lost the hand: %+v", got)
	}
	scores.AtHandStart[TeamA] = 299
	if got := DominoResult(Mode1v1, config, SeatX, hands, scores); got.Awarded[TeamA] != 45 {
		t.Fatalf("299 at hand start awarded %d, want 45", got.Awarded[TeamA])
	}
}

func TestHandEndDoesNotAllocate(t *testing.T) {
	board := lineBoard(t, "6-3", "3-2")
	hands := map[Seat]Hand{
		SeatNorth: tiles("1-1"),
		SeatEast:  tiles("4-5"),
		SeatSouth: tiles("0-0"),
		SeatWest:  tiles("1-4"),
	}
	config := DefaultConfig()
	allocations := testing.AllocsPerRun(100, func() {
		Mode2v2.HandStatus(&board, hands, nil)
		DominoResult(Mode2v2, config, SeatNorth, hands, HandScores{})
	})
	if allocations != 0 {
		t.Fatalf("hand end allocates %.0f times", allocations)
	}
}
