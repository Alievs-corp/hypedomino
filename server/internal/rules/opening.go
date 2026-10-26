package rules

type Opening struct {
	Seat Seat
	Tile Tile
}

var firstHandDoubleOrder = [...]Tile{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {0, 0}}

func FirstOpener(mode Mode, config Config, hands map[Seat]Hand) (Opening, bool) {
	if opening, ok := holderOf(mode, hands, config.FirstHandOpeningTile); ok {
		return opening, true
	}
	if mode != Mode1v1 {
		return Opening{}, false
	}
	for _, double := range firstHandDoubleOrder {
		if opening, ok := holderOf(mode, hands, double); ok {
			return opening, true
		}
	}
	return Opening{}, false
}

func holderOf(mode Mode, hands map[Seat]Hand, tile Tile) (Opening, bool) {
	for _, seat := range mode.Seats() {
		if hands[seat].Contains(tile) {
			return Opening{Seat: seat, Tile: tile}, true
		}
	}
	return Opening{}, false
}
