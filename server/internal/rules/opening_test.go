package rules

import "testing"

func TestFirstOpenerPrefersTheOpeningTileThenTheLowestDouble(t *testing.T) {
	config := DefaultConfig()
	cases := []struct {
		name   string
		mode   Mode
		hands  map[Seat]Hand
		opener Seat
		tile   string
		ok     bool
	}{
		{"2v2 holder of 2-3", Mode2v2, map[Seat]Hand{SeatNorth: tiles("1-1"), SeatEast: tiles("0-0"), SeatSouth: tiles("2-3"), SeatWest: tiles("6-6")}, SeatSouth, "2-3", true},
		{"1v1 2-3 beats a double", Mode1v1, map[Seat]Hand{SeatX: tiles("1-1"), SeatY: tiles("2-3")}, SeatY, "2-3", true},
		{"1v1 lowest double", Mode1v1, map[Seat]Hand{SeatX: tiles("4-4", "0-0"), SeatY: tiles("3-3")}, SeatY, "3-3", true},
		{"1v1 0-0 is last", Mode1v1, map[Seat]Hand{SeatX: tiles("0-0"), SeatY: tiles("6-6")}, SeatY, "6-6", true},
		{"1v1 only 0-0", Mode1v1, map[Seat]Hand{SeatX: tiles("0-1"), SeatY: tiles("0-0")}, SeatY, "0-0", true},
		{"1v1 redeal", Mode1v1, map[Seat]Hand{SeatX: tiles("0-1", "4-5"), SeatY: tiles("1-2")}, SeatX, "", false},
		{"2v2 never falls back to doubles", Mode2v2, map[Seat]Hand{SeatNorth: tiles("1-1"), SeatEast: nil, SeatSouth: nil, SeatWest: nil}, SeatNorth, "", false},
	}
	for _, c := range cases {
		opening, ok := FirstOpener(c.mode, config, c.hands)
		if ok != c.ok {
			t.Errorf("%s: ok = %v", c.name, ok)
			continue
		}
		if ok && (opening.Seat != c.opener || opening.Tile.String() != c.tile) {
			t.Errorf("%s: opening = %s %s, want %s %s", c.name, opening.Seat, opening.Tile, c.opener, c.tile)
		}
	}
}

func TestFirstOpenerUsesTheConfiguredTile(t *testing.T) {
	config := DefaultConfig()
	config.FirstHandOpeningTile = MustTile(6, 4)
	hands := map[Seat]Hand{SeatX: tiles("2-3"), SeatY: tiles("4-6")}
	opening, ok := FirstOpener(Mode1v1, config, hands)
	if !ok || opening.Seat != SeatY || opening.Tile != MustTile(4, 6) {
		t.Fatalf("opening = %+v %v", opening, ok)
	}
}
