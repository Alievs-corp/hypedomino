package app

import (
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
)

func openTimed(t *testing.T, rooms *Rooms) (*room.Room, time.Time) {
	t.Helper()
	opened := time.Now()
	r, err := rooms.Open(room.Options{Mode: rules.Mode1v1, Config: rules.DefaultConfig(), Bot: bot.DefaultConfig(), BotLevel: bot.Hard, Seed: 9, Clock: realClock{}})
	if err != nil {
		t.Fatal(err)
	}
	return r, opened
}

func loadingTime(r *room.Room, opened time.Time) time.Duration {
	reply := make(chan room.View, 1)
	r.Snapshot(rules.SeatX, reply)
	return (<-reply).LoadingDeadline.Sub(opened).Round(time.Second)
}

func TestNewRoomsTakeTheTunedRulesAndOpenRoomsKeepTheirs(t *testing.T) {
	rooms := NewRooms()
	defer rooms.Stop()
	before, beforeOpened := openTimed(t, rooms)
	tuning, err := rulesconfig.Parse([]byte(`{"MatchLoadingSeconds": 40}`))
	if err != nil {
		t.Fatal(err)
	}
	rooms.Tune(tuning.Apply)
	after, afterOpened := openTimed(t, rooms)
	if got := loadingTime(after, afterOpened); got != 40*time.Second {
		t.Fatalf("a room opened after the change loads for %v", got)
	}
	if got := loadingTime(before, beforeOpened); got != 10*time.Second {
		t.Fatalf("a room opened before the change loads for %v", got)
	}
}
