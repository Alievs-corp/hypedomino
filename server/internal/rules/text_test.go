package rules

import (
	"encoding/json"
	"testing"
)

func TestTextTypesRoundTripThroughJSON(t *testing.T) {
	type record struct {
		Mode    Mode
		Seat    Seat
		Team    Team
		End     End
		Tile    Tile
		Placed  Oriented
		Outcome TurnOutcome
		Status  HandStatus
		Match   MatchOutcome
	}
	want := record{
		Mode:    Mode2v2,
		Seat:    SeatWest,
		Team:    TeamB,
		End:     EndBottom,
		Tile:    MustTile(2, 5),
		Placed:  Oriented{First: 5, Second: 2},
		Outcome: OutcomePass,
		Status:  HandBlocked,
		Match:   MatchWon,
	}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	const expected = `{"Mode":"2v2","Seat":"W","Team":"B","End":"bottom","Tile":"2-5","Placed":"5-2","Outcome":"pass","Status":"block","Match":"won"}`
	if string(data) != expected {
		t.Fatalf("marshal = %s", data)
	}
	var got record
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}

func TestTextTypesRejectInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		into any
	}{
		{"mode", new(Mode)},
		{"seat", new(Seat)},
		{"team", new(Team)},
		{"end", new(End)},
		{"tile", new(Tile)},
		{"oriented", new(Oriented)},
		{"outcome", new(TurnOutcome)},
		{"status", new(HandStatus)},
		{"match", new(MatchOutcome)},
	}
	for _, c := range cases {
		if err := json.Unmarshal([]byte(`"nonsense"`), c.into); err == nil {
			t.Errorf("%s accepted nonsense", c.name)
		}
	}
	if Mode(9).String() != "invalid" || Seat(9).String() != "invalid" || End(9).String() != "invalid" {
		t.Error("out of range values have names")
	}
	if _, err := ParseOriented("7-1"); err == nil {
		t.Error("ParseOriented accepted a pip above the maximum")
	}
	if _, open := (OpenEnds{}).Value(End(9)); open {
		t.Error("an out of range end is open")
	}
	var empty Board
	for end := EndLeft; end <= End(9); end++ {
		if _, open := empty.EndPip(end); open {
			t.Errorf("empty board has an open %s end", end)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("MustTile accepted an invalid tile")
		}
	}()
	MustTile(7, 0)
}
