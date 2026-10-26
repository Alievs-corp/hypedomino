package rules

import (
	"errors"
	"fmt"
)

var ErrTurnOver = errors.New("rules: the turn is over")

type Placement struct {
	Total int
	Score int
}

type Turn struct {
	board       *Board
	config      Config
	placements  int
	mayContinue bool
	score       int
}

func BeginTurn(board *Board, config Config) Turn {
	return Turn{board: board, config: config}
}

func (t *Turn) Placements() int {
	return t.placements
}

func (t *Turn) Score() int {
	return t.score
}

func (t *Turn) CanContinue() bool {
	return t.placements == 0 || t.mayContinue
}

func (t *Turn) Place(tile Tile, end End) (Placement, error) {
	if !t.CanContinue() {
		return Placement{}, ErrTurnOver
	}
	if t.placements > 0 && !tile.IsDouble() {
		return Placement{}, fmt.Errorf("%w: only a double may continue the turn", ErrIllegalPlacement)
	}
	trial := *t.board
	if err := trial.Place(tile, end); err != nil {
		return Placement{}, err
	}
	placement := t.outcome(&trial)
	if t.placements > 0 && placement.Score == 0 {
		return Placement{}, fmt.Errorf("%w: %s on %s does not score", ErrIllegalPlacement, tile, end)
	}
	*t.board = trial
	t.placements++
	t.score += placement.Score
	t.mayContinue = tile.IsDouble() && placement.Score > 0
	return placement, nil
}

func (t *Turn) outcome(board *Board) Placement {
	total := board.OpenEnds().Total()
	return Placement{Total: total, Score: PlacementScore(total, t.config)}
}

func (t *Turn) LegalMoves(hand Hand, moves []Move) []Move {
	if t.placements == 0 {
		return t.board.LegalMoves(hand, moves)
	}
	moves = moves[:0]
	if !t.mayContinue {
		return moves
	}
	for _, tile := range hand {
		if !tile.IsDouble() {
			continue
		}
		for end := EndLeft; end < endCount; end++ {
			if t.continuationScores(tile, end) {
				moves = append(moves, Move{Tile: tile, End: end})
			}
		}
	}
	return moves
}

func (t *Turn) continuationScores(tile Tile, end End) bool {
	if !t.board.CanPlace(tile, end) {
		return false
	}
	trial := *t.board
	if err := trial.Place(tile, end); err != nil {
		return false
	}
	return t.outcome(&trial).Score > 0
}
