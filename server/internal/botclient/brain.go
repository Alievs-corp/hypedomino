package botclient

import (
	"errors"
	"fmt"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var ErrBoardMismatch = errors.New("botclient: rebuilt board differs from the view")

type Brain struct {
	player *bot.Player
	config rules.Config
	table  table
}

func NewBrain(level bot.Level, botConfig bot.Config, config rules.Config, random rules.Random) *Brain {
	return &Brain{player: bot.New(level, botConfig, random), config: config}
}

func (b *Brain) Decide(view *pb.MatchView) (*pb.Envelope, error) {
	if view.GetState() == pb.MatchState_MATCH_STATE_PLAYING && len(view.GetLegalMoves()) == 0 {
		return fallback(view), nil
	}
	if err := b.table.load(b.config, view); err != nil {
		return nil, err
	}
	if _, err := b.player.Act(&b.table, b.table.seat); err != nil {
		return nil, err
	}
	if b.table.chosen == nil {
		return nil, fmt.Errorf("botclient: bot chose nothing for %s", b.table.seat)
	}
	return b.table.chosen, nil
}

func fallback(view *pb.MatchView) *pb.Envelope {
	switch {
	case view.GetTurnPlacements() > 0:
		return &pb.Envelope{Payload: &pb.Envelope_EndTurnIntent{EndTurnIntent: &pb.EndTurnIntent{}}}
	case view.GetMode() == pb.Mode_MODE_1V1 && view.GetBoneyardSize() > 1:
		return &pb.Envelope{Payload: &pb.Envelope_DrawIntent{DrawIntent: &pb.DrawIntent{}}}
	}
	return &pb.Envelope{Payload: &pb.Envelope_PassIntent{PassIntent: &pb.PassIntent{}}}
}

type table struct {
	config  rules.Config
	view    *pb.MatchView
	seat    rules.Seat
	board   rules.Board
	hand    rules.Hand
	moves   []rules.Move
	turn    rules.Turn
	opening rules.Opening
	forced  bool
	chosen  *pb.Envelope
}

func (t *table) load(config rules.Config, view *pb.MatchView) error {
	t.config, t.view, t.seat, t.chosen = config, view, seatOf(view.GetSeat()), nil
	t.hand = t.hand[:0]
	for _, tile := range view.GetHand() {
		t.hand = append(t.hand, tileOf(tile))
	}
	t.moves = t.moves[:0]
	for _, move := range view.GetLegalMoves() {
		t.moves = append(t.moves, rules.Move{Tile: tileOf(move.GetTile()), End: endOf(move.GetEnd())})
	}
	t.forced = view.GetForcedOpening() != nil
	if t.forced {
		t.opening = rules.Opening{Seat: seatOf(view.GetForcedOpening().GetSeat()), Tile: tileOf(view.GetForcedOpening().GetTile())}
	}
	if len(view.GetBoard().GetLine()) == 0 {
		return nil
	}
	board, err := RebuildBoard(view.GetBoard())
	t.board = board
	return err
}

func (t *table) Mode() rules.Mode {
	if t.view.GetMode() == pb.Mode_MODE_2V2 {
		return rules.Mode2v2
	}
	return rules.Mode1v1
}

func (t *table) Config() rules.Config { return t.config }

func (t *table) Phase() rules.Phase {
	switch t.view.GetState() {
	case pb.MatchState_MATCH_STATE_OPENING:
		return rules.PhaseOpening
	case pb.MatchState_MATCH_STATE_PLAYING:
		return rules.PhasePlaying
	}
	return rules.PhaseFinished
}

func (t *table) Scores() rules.TeamPoints {
	scores := t.view.GetScores()
	return rules.TeamPoints{int(scores.GetTeamA()), int(scores.GetTeamB())}
}

func (t *table) HandsPlayed() int           { return int(t.view.GetHandsPlayed()) }
func (t *table) Current() rules.Seat        { return seatOf(t.view.GetCurrent()) }
func (t *table) OpeningTeam() rules.Team    { return rules.Team(t.view.GetOpeningTeam() - 1) }
func (t *table) Board() *rules.Board        { return &t.board }
func (t *table) Turn() *rules.Turn          { return &t.turn }
func (t *table) Hand(rules.Seat) rules.Hand { return t.hand }

func (t *table) ForcedOpening() (rules.Opening, bool) {
	return t.opening, t.forced
}

func (t *table) LegalMoves(seat rules.Seat, moves []rules.Move) []rules.Move {
	return append(moves, t.moves...)
}

func (t *table) Open(seat rules.Seat, tile rules.Tile) (int, error) {
	t.chosen = place(tileMessage(tile), pb.BoardEnd_BOARD_END_UNSPECIFIED)
	return 0, nil
}

func (t *table) Place(seat rules.Seat, tile rules.Tile, end rules.End) (rules.Placement, error) {
	t.chosen = place(tileMessage(tile), pb.BoardEnd(end)+1)
	return rules.Placement{}, nil
}

func (t *table) EndTurn(rules.Seat) error {
	t.chosen = &pb.Envelope{Payload: &pb.Envelope_EndTurnIntent{EndTurnIntent: &pb.EndTurnIntent{}}}
	return nil
}

func (t *table) Draw(rules.Seat) (rules.Tile, error) {
	t.chosen = &pb.Envelope{Payload: &pb.Envelope_DrawIntent{DrawIntent: &pb.DrawIntent{}}}
	return rules.Tile{}, nil
}

func (t *table) Pass(rules.Seat) error {
	t.chosen = &pb.Envelope{Payload: &pb.Envelope_PassIntent{PassIntent: &pb.PassIntent{}}}
	return nil
}

func RebuildBoard(view *pb.BoardView) (rules.Board, error) {
	line := view.GetLine()
	for start := range line {
		if spinner := view.GetSpinner(); spinner != nil && tileOf(spinner) != orientedOf(line[start]).Tile() {
			continue
		}
		if board, ok := rebuildFrom(view, start); ok {
			return board, nil
		}
	}
	return rules.Board{}, ErrBoardMismatch
}

func rebuildFrom(view *pb.BoardView, start int) (rules.Board, bool) {
	line := view.GetLine()
	board := rules.NewBoard(orientedOf(line[start]).Tile())
	for left, right := start-1, start+1; left >= 0 || right < len(line); left, right = left-1, right+1 {
		if left >= 0 && board.Place(orientedOf(line[left]).Tile(), rules.EndLeft) != nil {
			return board, false
		}
		if right < len(line) && board.Place(orientedOf(line[right]).Tile(), rules.EndRight) != nil {
			return board, false
		}
	}
	for _, arm := range []struct {
		tiles []*pb.OrientedTile
		end   rules.End
	}{{view.GetTop(), rules.EndTop}, {view.GetBottom(), rules.EndBottom}} {
		for _, tile := range arm.tiles {
			if board.Place(orientedOf(tile).Tile(), arm.end) != nil {
				return board, false
			}
		}
	}
	return board, sameTiles(board.Line(), line) && sameTiles(board.Top(), view.GetTop()) && sameTiles(board.Bottom(), view.GetBottom())
}

func sameTiles(tiles []rules.Oriented, messages []*pb.OrientedTile) bool {
	if len(tiles) != len(messages) {
		return false
	}
	for index, tile := range tiles {
		if tile != orientedOf(messages[index]) {
			return false
		}
	}
	return true
}

func seatOf(seat pb.Seat) rules.Seat {
	return rules.Seat(seat - 1)
}

func endOf(end pb.BoardEnd) rules.End {
	return rules.End(end - 1)
}

func tileOf(tile *pb.Tile) rules.Tile {
	return rules.Tile{Low: uint8(tile.GetLow()), High: uint8(tile.GetHigh())}
}

func orientedOf(tile *pb.OrientedTile) rules.Oriented {
	return rules.Oriented{First: uint8(tile.GetFirst()), Second: uint8(tile.GetSecond())}
}

func tileMessage(tile rules.Tile) *pb.Tile {
	return &pb.Tile{Low: uint32(tile.Low), High: uint32(tile.High)}
}
