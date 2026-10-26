package rules

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestFullSetHoldsEveryTileOnce(t *testing.T) {
	set := FullSet()
	seen := map[Tile]bool{}
	for index, tile := range set {
		if !tile.Valid() {
			t.Fatalf("tile %d is invalid: %v", index, tile)
		}
		if seen[tile] {
			t.Fatalf("tile %s appears twice", tile)
		}
		seen[tile] = true
		if tileIndex(tile) != index {
			t.Fatalf("tileIndex(%s) = %d, want %d", tile, tileIndex(tile), index)
		}
	}
	if len(seen) != SetSize {
		t.Fatalf("set holds %d tiles, want %d", len(seen), SetSize)
	}
}

func TestTileParseAndText(t *testing.T) {
	cases := []struct {
		text   string
		tile   Tile
		pips   int
		double bool
	}{
		{"0-0", Tile{0, 0}, 0, true},
		{"2-3", Tile{2, 3}, 5, false},
		{"3-2", Tile{2, 3}, 5, false},
		{"6-6", Tile{6, 6}, 12, true},
	}
	for _, c := range cases {
		tile, err := ParseTile(c.text)
		if err != nil {
			t.Fatalf("ParseTile(%q): %v", c.text, err)
		}
		if tile != c.tile || tile.Pips() != c.pips || tile.IsDouble() != c.double {
			t.Fatalf("ParseTile(%q) = %v pips %d double %v", c.text, tile, tile.Pips(), tile.IsDouble())
		}
		var decoded Tile
		if err := json.Unmarshal([]byte(strconv.Quote(c.text)), &decoded); err != nil || decoded != c.tile {
			t.Fatalf("json %q = %v, %v", c.text, decoded, err)
		}
		if encoded, _ := json.Marshal(c.tile); string(encoded) != strconv.Quote(c.tile.String()) {
			t.Fatalf("json of %v = %s", c.tile, encoded)
		}
	}
	for _, text := range []string{"", "7-0", "0-7", "1", "1-2-3", "a-b", "-1-2"} {
		if _, err := ParseTile(text); err == nil {
			t.Errorf("ParseTile(%q) accepted", text)
		}
	}
	if _, err := NewTile(7, 0); err == nil {
		t.Error("NewTile(7, 0) accepted")
	}
}

func TestTileHas(t *testing.T) {
	tile := MustTile(2, 5)
	if !tile.Has(2) || !tile.Has(5) || tile.Has(3) {
		t.Fatalf("Has on %s is wrong", tile)
	}
}

func TestHandOperations(t *testing.T) {
	hand := Hand{MustTile(0, 1), MustTile(2, 3), MustTile(6, 6)}
	if hand.Pips() != 1+5+12 {
		t.Fatalf("Pips = %d", hand.Pips())
	}
	if !hand.Contains(MustTile(2, 3)) || hand.Contains(MustTile(4, 4)) {
		t.Fatal("Contains is wrong")
	}
	if hand.Remove(MustTile(4, 4)) {
		t.Fatal("removed a tile that is not held")
	}
	if !hand.Remove(MustTile(2, 3)) || len(hand) != 2 || hand.Contains(MustTile(2, 3)) {
		t.Fatalf("Remove left %v", hand)
	}
	hand.Add(MustTile(4, 4))
	if len(hand) != 3 || !hand.Contains(MustTile(4, 4)) {
		t.Fatalf("Add left %v", hand)
	}
}

func TestBoneyardDrawsInOrder(t *testing.T) {
	boneyard := Boneyard{MustTile(1, 4), MustTile(5, 0)}
	first, ok := boneyard.Draw()
	if !ok || first != MustTile(1, 4) || len(boneyard) != 1 {
		t.Fatalf("first draw = %v %v, left %v", first, ok, boneyard)
	}
	second, ok := boneyard.Draw()
	if !ok || second != MustTile(0, 5) {
		t.Fatalf("second draw = %v %v", second, ok)
	}
	if _, ok := boneyard.Draw(); ok {
		t.Fatal("drew from an empty boneyard")
	}
}

func TestModeAndSeatText(t *testing.T) {
	for _, text := range []string{"1v1", "2v2"} {
		mode, err := ParseMode(text)
		if err != nil || mode.String() != text {
			t.Fatalf("ParseMode(%q) = %v, %v", text, mode, err)
		}
	}
	if _, err := ParseMode("3v3"); err == nil {
		t.Error("ParseMode(3v3) accepted")
	}
	if Mode1v1.Players() != 2 || Mode2v2.Players() != 4 || Mode(0).Players() != 0 {
		t.Fatal("Players is wrong")
	}
	for _, text := range []string{"N", "E", "S", "W", "X", "Y"} {
		seat, err := ParseSeat(text)
		if err != nil || seat.String() != text {
			t.Fatalf("ParseSeat(%q) = %v, %v", text, seat, err)
		}
	}
	if _, err := ParseSeat("Z"); err == nil {
		t.Error("ParseSeat(Z) accepted")
	}
	var hands map[Seat]Hand
	if err := json.Unmarshal([]byte("{\"N\":[\"0-0\"],\"E\":[\"6-6\"]}"), &hands); err != nil {
		t.Fatal(err)
	}
	if hands[SeatNorth][0] != MustTile(0, 0) || hands[SeatEast][0] != MustTile(6, 6) {
		t.Fatalf("decoded %v", hands)
	}
}
