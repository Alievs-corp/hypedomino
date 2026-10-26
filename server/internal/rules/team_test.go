package rules

import (
	"errors"
	"testing"
)

func TestSeatTeamsAlternateAndTurnsPassClockwise(t *testing.T) {
	cases := []struct {
		seat    Seat
		team    Team
		next    Seat
		partner Seat
		paired  bool
	}{
		{SeatNorth, TeamA, SeatEast, SeatSouth, true},
		{SeatEast, TeamB, SeatSouth, SeatWest, true},
		{SeatSouth, TeamA, SeatWest, SeatNorth, true},
		{SeatWest, TeamB, SeatNorth, SeatEast, true},
		{SeatX, TeamA, SeatY, SeatX, false},
		{SeatY, TeamB, SeatX, SeatY, false},
	}
	for _, c := range cases {
		if c.seat.Team() != c.team || c.seat.Next() != c.next {
			t.Errorf("%s: team %s next %s", c.seat, c.seat.Team(), c.seat.Next())
		}
		partner, paired := c.seat.Partner()
		if partner != c.partner || paired != c.paired {
			t.Errorf("%s: partner %s %v", c.seat, partner, paired)
		}
		if c.seat.Next().Team() == c.seat.Team() {
			t.Errorf("%s passes the turn to its own team", c.seat)
		}
	}
	for _, mode := range []Mode{Mode1v1, Mode2v2} {
		seat := mode.Seats()[0]
		for i := 0; i < mode.Players(); i++ {
			seat = seat.Next()
		}
		if seat != mode.Seats()[0] {
			t.Errorf("%s: %d turns do not return to the first seat", mode, mode.Players())
		}
	}
}

func TestTeamText(t *testing.T) {
	for _, text := range []string{"A", "B"} {
		team, err := ParseTeam(text)
		if err != nil || team.String() != text {
			t.Fatalf("ParseTeam(%q) = %v, %v", text, team, err)
		}
	}
	if _, err := ParseTeam("C"); !errors.Is(err, ErrInvalidTeam) {
		t.Errorf("ParseTeam(C) = %v", err)
	}
	if TeamA.Other() != TeamB || TeamB.Other() != TeamA || Team(5).String() != "invalid" {
		t.Error("Other or String is wrong")
	}
}

func TestTeamSeats(t *testing.T) {
	var buffer [4]Seat
	seats := Mode2v2.TeamSeats(TeamA, buffer[:0])
	if len(seats) != 2 || seats[0] != SeatNorth || seats[1] != SeatSouth {
		t.Fatalf("2v2 team A seats %v", seats)
	}
	seats = Mode2v2.TeamSeats(TeamB, buffer[:0])
	if len(seats) != 2 || seats[0] != SeatEast || seats[1] != SeatWest {
		t.Fatalf("2v2 team B seats %v", seats)
	}
	seats = Mode1v1.TeamSeats(TeamB, buffer[:0])
	if len(seats) != 1 || seats[0] != SeatY {
		t.Fatalf("1v1 team B seats %v", seats)
	}
	allocations := testing.AllocsPerRun(100, func() {
		Mode2v2.TeamSeats(TeamA, buffer[:0])
		SeatNorth.Next()
		SeatNorth.Team()
	})
	if allocations != 0 {
		t.Fatalf("seat lookups allocate %.0f times", allocations)
	}
}
