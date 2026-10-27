package room

import (
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type BoardView struct {
	Line       []rules.Oriented
	Spinner    rules.Tile
	HasSpinner bool
	Top        []rules.Oriented
	Bottom     []rules.Oriented
	OpenEnds   rules.OpenEnds
}

type View struct {
	State          State
	Seat           rules.Seat
	Mode           rules.Mode
	Acked          uint64
	HandsPlayed    int
	Scores         rules.TeamPoints
	CarryOver      int
	Hand           []rules.Tile
	HandSizes      [6]int
	BoneyardSize   int
	Board          BoardView
	Current        rules.Seat
	OpeningTeam    rules.Team
	ForcedOpening  rules.Opening
	HasForced      bool
	TurnPlacements int
	TurnDeadline   time.Time
	Controllers    [6]rules.Controller
	LegalMoves     []rules.Move
	LastResult     rules.HandResult
	HasLastResult  bool
	Winner         rules.Team
	HasWinner      bool
}

func (r *Room) view(seat rules.Seat) View {
	v := View{State: r.state, Seat: seat, Mode: r.options.Mode, Acked: r.seats[seat].acked}
	for _, s := range r.options.Mode.Seats() {
		v.Controllers[s] = r.presence.Controller(s)
	}
	if r.recorder == nil {
		return v
	}
	m := r.match()
	v.HandsPlayed = m.HandsPlayed()
	v.Scores = m.Scores()
	v.CarryOver = m.CarryOver()
	v.Hand = append([]rules.Tile(nil), m.Hand(seat)...)
	for _, s := range r.options.Mode.Seats() {
		v.HandSizes[s] = len(m.Hand(s))
	}
	v.BoneyardSize = m.BoneyardSize()
	v.Current = m.Current()
	if acting, ok := r.actingSeat(); ok {
		v.Current = acting
	}
	v.OpeningTeam = m.OpeningTeam()
	v.ForcedOpening, v.HasForced = m.ForcedOpening()
	v.TurnDeadline = r.deadline
	v.LastResult, v.HasLastResult = m.LastResult()
	v.Winner, v.HasWinner = m.Winner()
	if m.Phase() == rules.PhasePlaying {
		board := m.Board()
		v.Board = BoardView{
			Line:     append([]rules.Oriented(nil), board.Line()...),
			Top:      append([]rules.Oriented(nil), board.Top()...),
			Bottom:   append([]rules.Oriented(nil), board.Bottom()...),
			OpenEnds: board.OpenEnds(),
		}
		v.Board.Spinner, v.Board.HasSpinner = board.Spinner()
		v.TurnPlacements = m.Turn().Placements()
		if m.Current() == seat {
			v.LegalMoves = append([]rules.Move(nil), m.LegalMoves(seat, r.moves[:0])...)
		}
	}
	return v
}
