package rules

import (
	"errors"
	"fmt"
)

var ErrTileNotHeld = errors.New("rules: tile not in hand")

type Move struct {
	Tile Tile
	End  End
}

func (b *Board) LegalMoves(hand Hand, moves []Move) []Move {
	moves = moves[:0]
	for _, tile := range hand {
		for end := EndLeft; end < endCount; end++ {
			if b.CanPlace(tile, end) {
				moves = append(moves, Move{Tile: tile, End: end})
			}
		}
	}
	return moves
}

func (b *Board) HasLegalMove(hand Hand) bool {
	for _, tile := range hand {
		for end := EndLeft; end < endCount; end++ {
			if b.CanPlace(tile, end) {
				return true
			}
		}
	}
	return false
}

func (b *Board) Validate(hand Hand, tile Tile, end End) error {
	if !hand.Contains(tile) {
		return fmt.Errorf("%w: %s", ErrTileNotHeld, tile)
	}
	if !b.CanPlace(tile, end) {
		return fmt.Errorf("%w: %s on %s", ErrIllegalPlacement, tile, end)
	}
	return nil
}

type TurnOutcome uint8

const (
	OutcomePlace TurnOutcome = iota
	OutcomePass
)

func (o TurnOutcome) String() string {
	if o == OutcomePass {
		return "pass"
	}
	return "place"
}

func (o TurnOutcome) MarshalText() ([]byte, error) {
	return []byte(o.String()), nil
}

func (o *TurnOutcome) UnmarshalText(text []byte) error {
	switch string(text) {
	case "place":
		*o = OutcomePlace
	case "pass":
		*o = OutcomePass
	default:
		return fmt.Errorf("rules: invalid turn outcome %q", text)
	}
	return nil
}

func DrawUntilPlayable(mode Mode, board *Board, hand *Hand, boneyard *Boneyard, drawn []Tile) ([]Tile, TurnOutcome) {
	drawn = drawn[:0]
	if board.HasLegalMove(*hand) {
		return drawn, OutcomePlace
	}
	if mode != Mode1v1 {
		return drawn, OutcomePass
	}
	for len(*boneyard) > 1 {
		tile, _ := boneyard.Draw()
		hand.Add(tile)
		drawn = append(drawn, tile)
		if board.CanPlaceAnywhere(tile) {
			return drawn, OutcomePlace
		}
	}
	return drawn, OutcomePass
}

func (b *Board) CanPlaceAnywhere(tile Tile) bool {
	for end := EndLeft; end < endCount; end++ {
		if b.CanPlace(tile, end) {
			return true
		}
	}
	return false
}
