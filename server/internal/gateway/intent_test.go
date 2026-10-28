package gateway

import (
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/replay"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestEncodeIntentRoundTripsThroughDecode(t *testing.T) {
	intents := []replay.Intent{
		{Kind: replay.KindOpen, Tile: rules.Tile{High: 6, Low: 6}},
		{Kind: replay.KindPlace, Tile: rules.Tile{High: 5, Low: 2}, End: rules.EndLeft},
		{Kind: replay.KindPlace, Tile: rules.Tile{High: 1, Low: 0}, End: rules.EndRight},
		{Kind: replay.KindEndTurn},
		{Kind: replay.KindDraw},
		{Kind: replay.KindPass},
	}
	for _, want := range intents {
		frame, ok := EncodeIntent(want)
		if !ok {
			t.Fatalf("%v not encoded", want)
		}
		got, ok := DecodeIntent(frame)
		if !ok || got != want {
			t.Fatalf("%v decoded as %v (%v)", want, got, ok)
		}
	}
	if _, ok := EncodeIntent(replay.Intent{Kind: replay.Kind(200)}); ok {
		t.Fatal("an unknown kind was encoded")
	}
}
