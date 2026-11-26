package loadbot

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

func TestARecordedMatchReplaysAsTheSameMatch(t *testing.T) {
	short, _ := startServer(t, 30)
	long, _ := startServer(t, 365)
	cases := []struct {
		name     string
		target   string
		mode     rules.Mode
		length   time.Duration
		thinkMin time.Duration
		thinkMax time.Duration
		finished bool
	}{
		{name: "whole 2v2 match", target: short, mode: rules.Mode2v2, finished: true},
		{name: "whole 1v1 match", target: short, mode: rules.Mode1v1, finished: true},
		{name: "2v2 match cut by the length", target: long, mode: rules.Mode2v2, length: 300 * time.Millisecond, thinkMin: 2 * time.Millisecond, thinkMax: 6 * time.Millisecond},
	}
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			trace, err := Record(ctx, RecordConfig{Target: tc.target, AppVersion: "1.0.0", Mode: tc.mode, Seed: 7, Room: uint64(100 + index), ThinkMin: tc.thinkMin, ThinkMax: tc.thinkMax, ThinkFollow: time.Millisecond, Length: tc.length})
			if err != nil {
				t.Fatal(err)
			}
			if trace.Intents() == 0 {
				t.Fatal("no intent recorded")
			}
			for _, seat := range trace.Seats {
				if got := seat.Final.State == finishedState; got != tc.finished {
					t.Fatalf("seat %s final %+v, finished want %v", seat.Seat, seat.Final, tc.finished)
				}
			}
			path := filepath.Join(t.TempDir(), "match.json")
			if err := trace.Write(path); err != nil {
				t.Fatal(err)
			}
			read, err := ReadTrace(path)
			if err != nil {
				t.Fatal(err)
			}
			report, err := Replay(ctx, read, ReplayConfig{Target: tc.target, AppVersion: "1.0.0", Room: uint64(1000 + 10*index), Runs: 2, Parallel: 2, Budget: time.Second, RunTimeout: 30 * time.Second})
			if err != nil {
				t.Fatal(err)
			}
			t.Log("\n" + report.Text())
			if !report.Pass || len(report.Runs) != 4 {
				t.Fatalf("replay failed: %v", report.Failures)
			}
			for _, run := range report.Runs {
				if run.Moves != uint64(trace.Intents()) || run.P99 <= 0 || run.Max <= 0 {
					t.Fatalf("run %d: %d moves of %d intents, p99 %s max %s", run.Index, run.Moves, trace.Intents(), run.P99, run.Max)
				}
			}
			if report.Moves != uint64(4*trace.Intents()) {
				t.Fatalf("%d moves in all runs, want %d", report.Moves, 4*trace.Intents())
			}
		})
	}
}

func TestAReplayReportsADivergingMatchAndABadTrace(t *testing.T) {
	target, _ := startServer(t, 30)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	trace, err := Record(ctx, RecordConfig{Target: target, AppVersion: "1.0.0", Mode: rules.Mode2v2, Seed: 11, Room: 200, ThinkFollow: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		change  func(*Trace)
		refused error
		failure string
	}{
		{name: "another deal", change: func(t *Trace) { t.Seed = 12 }, failure: "diverged"},
		{name: "a missing intent", change: func(t *Trace) { t.Seats[1].Intents = t.Seats[1].Intents[1:] }, failure: "diverged"},
		{name: "an unknown version", change: func(t *Trace) { t.Version = 2 }, refused: ErrBadTrace},
		{name: "seats of another mode", change: func(t *Trace) { t.Mode = rules.Mode1v1.String() }, refused: ErrBadTrace},
	}
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := *trace
			changed.Seats = append([]SeatTrace(nil), trace.Seats...)
			tc.change(&changed)
			report, err := Replay(ctx, &changed, ReplayConfig{Target: target, AppVersion: "1.0.0", Room: uint64(2000 + index), Budget: time.Second, RunTimeout: 5 * time.Second})
			if tc.refused != nil {
				if !errors.Is(err, tc.refused) {
					t.Fatalf("replay: %v, want %v", err, tc.refused)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if report.Pass || !strings.Contains(strings.Join(report.Failures, ";"), tc.failure) {
				t.Fatalf("pass %v failures %v, want %q", report.Pass, report.Failures, tc.failure)
			}
		})
	}
}

func TestReplayJudgeAppliesTheBudgetToEveryRun(t *testing.T) {
	cases := []struct {
		name     string
		runs     []ReplayRun
		pass     bool
		worstP99 time.Duration
	}{
		{name: "every run under the budget", runs: []ReplayRun{{P99: 20 * time.Millisecond}, {P99: 40 * time.Millisecond}}, pass: true, worstP99: 40 * time.Millisecond},
		{name: "one run at the budget", runs: []ReplayRun{{P99: 20 * time.Millisecond}, {Index: 1, P99: 100 * time.Millisecond}}, worstP99: 100 * time.Millisecond},
		{name: "a diverged run", runs: []ReplayRun{{P99: time.Millisecond, Failure: "seat N diverged"}}, worstP99: time.Millisecond},
		{name: "no run"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := &ReplayReport{Budget: 100 * time.Millisecond, Runs: tc.runs}
			rep.judge()
			if rep.Pass != tc.pass || rep.WorstP99 != tc.worstP99 {
				t.Fatalf("pass %v worst %s failures %v", rep.Pass, rep.WorstP99, rep.Failures)
			}
			if !tc.pass && !strings.Contains(rep.Text(), "FAIL: ") {
				t.Fatal(rep.Text())
			}
		})
	}
}
