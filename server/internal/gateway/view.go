package gateway

import (
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func tileOf(tile *pb.Tile) (rules.Tile, bool) {
	if tile == nil || tile.GetLow() > rules.MaxPip || tile.GetHigh() > rules.MaxPip {
		return rules.Tile{}, false
	}
	parsed := rules.Tile{Low: uint8(tile.GetLow()), High: uint8(tile.GetHigh())}
	return parsed, parsed.Valid()
}

func endOf(end pb.BoardEnd) (rules.End, bool) {
	switch end {
	case pb.BoardEnd_BOARD_END_LEFT:
		return rules.EndLeft, true
	case pb.BoardEnd_BOARD_END_RIGHT:
		return rules.EndRight, true
	case pb.BoardEnd_BOARD_END_TOP:
		return rules.EndTop, true
	case pb.BoardEnd_BOARD_END_BOTTOM:
		return rules.EndBottom, true
	}
	return 0, false
}

func tileMessage(tile rules.Tile) *pb.Tile {
	return &pb.Tile{Low: uint32(tile.Low), High: uint32(tile.High)}
}

func endMessage(end rules.End) pb.BoardEnd {
	return pb.BoardEnd(end) + 1
}

func seatMessage(seat rules.Seat) pb.Seat {
	return pb.Seat(seat) + 1
}

func teamMessage(team rules.Team) pb.Team {
	return pb.Team(team) + 1
}

func controllerMessage(controller rules.Controller) pb.Controller {
	return pb.Controller(controller) + 1
}

func stateMessage(state room.State) pb.MatchState {
	return pb.MatchState(state) + 1
}

func pointsMessage(points rules.TeamPoints) *pb.TeamPoints {
	return &pb.TeamPoints{TeamA: uint32(points[rules.TeamA]), TeamB: uint32(points[rules.TeamB])}
}

func orientedMessages(tiles []rules.Oriented) []*pb.OrientedTile {
	if len(tiles) == 0 {
		return nil
	}
	out := make([]*pb.OrientedTile, len(tiles))
	for index, tile := range tiles {
		out[index] = &pb.OrientedTile{First: uint32(tile.First), Second: uint32(tile.Second)}
	}
	return out
}

func boardMessage(board room.BoardView) *pb.BoardView {
	out := &pb.BoardView{
		Line:   orientedMessages(board.Line),
		Top:    orientedMessages(board.Top),
		Bottom: orientedMessages(board.Bottom),
	}
	if board.HasSpinner {
		out.Spinner = tileMessage(board.Spinner)
	}
	for end := rules.EndLeft; end <= rules.EndBottom; end++ {
		if value, open := board.OpenEnds.Value(end); open {
			out.OpenEnds = append(out.OpenEnds, &pb.OpenEnd{End: endMessage(end), Value: uint32(value)})
		}
	}
	return out
}

func resultMessage(result rules.HandResult) *pb.HandResult {
	out := &pb.HandResult{
		Pips:       pointsMessage(result.Pips),
		Awarded:    pointsMessage(result.Awarded),
		CarryOver:  uint32(result.CarryOver),
		Scores:     pointsMessage(result.Scores),
		NextOpener: teamMessage(result.NextOpener),
	}
	if result.HasWinner {
		out.Winner = teamMessage(result.Winner)
	}
	return out
}

func ViewMessage(view room.View) *pb.MatchView {
	out := &pb.MatchView{
		State: stateMessage(view.State),
		Seat:  seatMessage(view.Seat),
		Mode:  pb.Mode(view.Mode),
	}
	for _, seat := range view.Mode.Seats() {
		out.Seats = append(out.Seats, &pb.SeatView{
			Seat:            seatMessage(seat),
			HandSize:        uint32(view.HandSizes[seat]),
			Controller:      controllerMessage(view.Controllers[seat]),
			LoadingProgress: uint32(view.Loading[seat]),
			Ready:           view.Ready[seat],
		})
	}
	for _, player := range view.Roster {
		out.Roster = append(out.Roster, &pb.RosterEntry{
			Seat:      seatMessage(player.Seat),
			Team:      teamMessage(player.Seat.Team()),
			Nickname:  player.Nickname,
			Character: player.Character,
			Frame:     player.Frame,
			Avatar:    player.Avatar,
			RankLabel: player.RankLabel,
			Emblems:   player.Emblems,
			Bot:       player.Bot,
		})
	}
	if view.State == room.StateLoading {
		out.LoadingDeadlineUnixMs = uint64(view.LoadingDeadline.UnixMilli())
		return out
	}
	if view.State == room.StateWaiting {
		return out
	}
	out.HandsPlayed = uint32(view.HandsPlayed)
	out.Scores = pointsMessage(view.Scores)
	out.CarryOver = uint32(view.CarryOver)
	for _, tile := range view.Hand {
		out.Hand = append(out.Hand, tileMessage(tile))
	}
	out.BoneyardSize = uint32(view.BoneyardSize)
	out.OpeningTeam = teamMessage(view.OpeningTeam)
	out.TurnDeadlineUnixMs = uint64(view.TurnDeadline.UnixMilli())
	if view.HasForced {
		out.ForcedOpening = &pb.Opening{Seat: seatMessage(view.ForcedOpening.Seat), Tile: tileMessage(view.ForcedOpening.Tile)}
	}
	if view.HasLastResult {
		out.LastResult = resultMessage(view.LastResult)
	}
	if view.HasWinner {
		out.Winner = teamMessage(view.Winner)
	}
	if view.State == room.StateFinished {
		return out
	}
	out.Current = seatMessage(view.Current)
	if view.State == room.StatePlaying {
		out.Board = boardMessage(view.Board)
		out.TurnPlacements = uint32(view.TurnPlacements)
		for _, move := range view.LegalMoves {
			out.LegalMoves = append(out.LegalMoves, &pb.LegalMove{Tile: tileMessage(move.Tile), End: endMessage(move.End)})
		}
	}
	return out
}
