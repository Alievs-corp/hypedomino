package rules

import (
	"errors"
	"fmt"
)

var ErrInvalidTeam = errors.New("rules: invalid team")

type Team uint8

const (
	TeamA Team = iota
	TeamB
	teamCount
)

var teamNames = [teamCount]string{"A", "B"}

func ParseTeam(text string) (Team, error) {
	for team, name := range teamNames {
		if name == text {
			return Team(team), nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrInvalidTeam, text)
}

func (t Team) String() string {
	if t >= teamCount {
		return "invalid"
	}
	return teamNames[t]
}

func (t Team) Other() Team {
	if t == TeamA {
		return TeamB
	}
	return TeamA
}

func (t Team) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Team) UnmarshalText(text []byte) error {
	parsed, err := ParseTeam(string(text))
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

var seatTeams = [seatCount]Team{TeamA, TeamB, TeamA, TeamB, TeamA, TeamB}

var seatNext = [seatCount]Seat{SeatEast, SeatSouth, SeatWest, SeatNorth, SeatY, SeatX}

func (s Seat) Team() Team {
	return seatTeams[s]
}

func (s Seat) Next() Seat {
	return seatNext[s]
}

func (s Seat) Partner() (Seat, bool) {
	switch s {
	case SeatNorth:
		return SeatSouth, true
	case SeatSouth:
		return SeatNorth, true
	case SeatEast:
		return SeatWest, true
	case SeatWest:
		return SeatEast, true
	}
	return s, false
}

func (m Mode) TeamSeats(team Team, seats []Seat) []Seat {
	seats = seats[:0]
	for _, seat := range m.Seats() {
		if seat.Team() == team {
			seats = append(seats, seat)
		}
	}
	return seats
}
