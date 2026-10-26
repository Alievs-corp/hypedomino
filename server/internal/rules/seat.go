package rules

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidMode = errors.New("rules: invalid mode")
	ErrInvalidSeat = errors.New("rules: invalid seat")
)

type Mode uint8

const (
	Mode1v1 Mode = iota + 1
	Mode2v2
)

func ParseMode(text string) (Mode, error) {
	switch text {
	case "1v1":
		return Mode1v1, nil
	case "2v2":
		return Mode2v2, nil
	}
	return 0, fmt.Errorf("%w: %q", ErrInvalidMode, text)
}

func (m Mode) String() string {
	switch m {
	case Mode1v1:
		return "1v1"
	case Mode2v2:
		return "2v2"
	}
	return "invalid"
}

func (m Mode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Mode) UnmarshalText(text []byte) error {
	parsed, err := ParseMode(string(text))
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

var (
	seats2v2 = [...]Seat{SeatNorth, SeatEast, SeatSouth, SeatWest}
	seats1v1 = [...]Seat{SeatX, SeatY}
)

func (m Mode) Seats() []Seat {
	switch m {
	case Mode1v1:
		return seats1v1[:]
	case Mode2v2:
		return seats2v2[:]
	}
	return nil
}

func (m Mode) Players() int {
	return len(m.Seats())
}

type Seat uint8

const (
	SeatNorth Seat = iota
	SeatEast
	SeatSouth
	SeatWest
	SeatX
	SeatY
	seatCount
)

var seatNames = [seatCount]string{"N", "E", "S", "W", "X", "Y"}

func ParseSeat(text string) (Seat, error) {
	for seat, name := range seatNames {
		if name == text {
			return Seat(seat), nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrInvalidSeat, text)
}

func (s Seat) String() string {
	if s >= seatCount {
		return "invalid"
	}
	return seatNames[s]
}

func (s Seat) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Seat) UnmarshalText(text []byte) error {
	parsed, err := ParseSeat(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}
