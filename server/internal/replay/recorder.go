package replay

import (
	"fmt"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type Table interface {
	Open(seat rules.Seat, tile rules.Tile) (int, error)
	Place(seat rules.Seat, tile rules.Tile, end rules.End) (rules.Placement, error)
	EndTurn(seat rules.Seat) error
	Draw(seat rules.Seat) (rules.Tile, error)
	Pass(seat rules.Seat) error
}

type Recorder struct {
	*rules.Match
	record Record
}

func NewRecorder(mode rules.Mode, config rules.Config, seed uint64) *Recorder {
	return &Recorder{
		Match:  rules.NewMatch(mode, config, rules.NewSeededRandom(seed)),
		record: Record{Mode: mode, Seed: seed, Config: config},
	}
}

func (r *Recorder) Record() Record {
	record := r.record
	record.Intents = append([]Intent(nil), r.record.Intents...)
	record.Result, record.HasResult = resultOf(r.Match)
	return record
}

func resultOf(m *rules.Match) (Result, bool) {
	if m.Phase() != rules.PhaseFinished {
		return Result{}, false
	}
	winner, _ := m.Winner()
	return Result{Scores: m.Scores(), HandsPlayed: m.HandsPlayed(), Winner: winner, HasWinner: true}, true
}

func (r *Recorder) Open(seat rules.Seat, tile rules.Tile) (int, error) {
	score, err := r.Match.Open(seat, tile)
	r.note(err, Intent{Kind: KindOpen, Seat: seat, Tile: tile})
	return score, err
}

func (r *Recorder) Place(seat rules.Seat, tile rules.Tile, end rules.End) (rules.Placement, error) {
	placement, err := r.Match.Place(seat, tile, end)
	r.note(err, Intent{Kind: KindPlace, Seat: seat, Tile: tile, End: end})
	return placement, err
}

func (r *Recorder) EndTurn(seat rules.Seat) error {
	err := r.Match.EndTurn(seat)
	r.note(err, Intent{Kind: KindEndTurn, Seat: seat})
	return err
}

func (r *Recorder) Draw(seat rules.Seat) (rules.Tile, error) {
	tile, err := r.Match.Draw(seat)
	r.note(err, Intent{Kind: KindDraw, Seat: seat})
	return tile, err
}

func (r *Recorder) Pass(seat rules.Seat) error {
	err := r.Match.Pass(seat)
	r.note(err, Intent{Kind: KindPass, Seat: seat})
	return err
}

func (r *Recorder) note(err error, intent Intent) {
	if err == nil {
		r.record.Intents = append(r.record.Intents, intent)
	}
}

func Apply(table Table, intent Intent) error {
	switch intent.Kind {
	case KindOpen:
		_, err := table.Open(intent.Seat, intent.Tile)
		return err
	case KindPlace:
		_, err := table.Place(intent.Seat, intent.Tile, intent.End)
		return err
	case KindEndTurn:
		return table.EndTurn(intent.Seat)
	case KindDraw:
		_, err := table.Draw(intent.Seat)
		return err
	case KindPass:
		return table.Pass(intent.Seat)
	}
	return fmt.Errorf("%w: intent kind %d", ErrBadRecord, intent.Kind)
}

func Replay(record Record) (*rules.Match, error) {
	m := rules.NewMatch(record.Mode, record.Config, rules.NewSeededRandom(record.Seed))
	for index, intent := range record.Intents {
		if err := Apply(m, intent); err != nil {
			return m, fmt.Errorf("%w: intent %d: %v", ErrIntentRefused, index, err)
		}
	}
	return m, nil
}

func Verify(record Record) error {
	m, err := Replay(record)
	if err != nil {
		return err
	}
	result, finished := resultOf(m)
	if finished != record.HasResult || result != record.Result {
		return fmt.Errorf("%w: replayed %+v finished=%v, recorded %+v finished=%v", ErrResultDiffers, result, finished, record.Result, record.HasResult)
	}
	return nil
}
