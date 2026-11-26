package loadbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/Alievs-corp/hypedomino/server/internal/botclient"
	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

const TraceVersion = 1

var finishedState = pb.MatchState_MATCH_STATE_FINISHED.String()

var ErrBadTrace = errors.New("loadbot: bad trace")

type Trace struct {
	Version int         `json:"version"`
	Mode    string      `json:"mode"`
	Seed    uint64      `json:"seed"`
	Seats   []SeatTrace `json:"seats"`
}

type SeatTrace struct {
	Seat    string         `json:"seat"`
	Intents []TracedIntent `json:"intents"`
	Final   Checkpoint     `json:"final"`
}

type TracedIntent struct {
	ThinkMs int64  `json:"think_ms"`
	Frame   []byte `json:"frame"`
}

type Checkpoint struct {
	State       string `json:"state"`
	HandsPlayed uint32 `json:"hands_played"`
	ScoreA      uint32 `json:"score_a"`
	ScoreB      uint32 `json:"score_b"`
}

func checkpointOf(view *pb.MatchView) Checkpoint {
	return Checkpoint{State: view.GetState().String(), HandsPlayed: view.GetHandsPlayed(), ScoreA: view.GetScores().GetTeamA(), ScoreB: view.GetScores().GetTeamB()}
}

func (t *Trace) Intents() int {
	n := 0
	for _, seat := range t.Seats {
		n += len(seat.Intents)
	}
	return n
}

func (t *Trace) mode() (rules.Mode, error) {
	if t.Version != TraceVersion {
		return 0, fmt.Errorf("%w: version %d, want %d", ErrBadTrace, t.Version, TraceVersion)
	}
	mode, err := rules.ParseMode(t.Mode)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrBadTrace, err)
	}
	if len(t.Seats) != len(mode.Seats()) {
		return 0, fmt.Errorf("%w: %d seats for %s", ErrBadTrace, len(t.Seats), mode)
	}
	for index, seat := range mode.Seats() {
		if t.Seats[index].Seat != seat.String() {
			return 0, fmt.Errorf("%w: seat %d is %q, want %s", ErrBadTrace, index, t.Seats[index].Seat, seat)
		}
	}
	return mode, nil
}

func ReadTrace(path string) (*Trace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t := &Trace{}
	if err := json.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadTrace, err)
	}
	if _, err := t.mode(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Trace) Write(path string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

type RecordConfig struct {
	Target      string
	AppVersion  string
	Mode        rules.Mode
	Seed        uint64
	Room        uint64
	ThinkMin    time.Duration
	ThinkMax    time.Duration
	ThinkFollow time.Duration
	Length      time.Duration
}

type ReplayConfig struct {
	Target     string
	AppVersion string
	Room       uint64
	Runs       int
	Parallel   int
	Budget     time.Duration
	RunTimeout time.Duration
}

type tracedSeat struct {
	client     *botclient.Client
	frames     chan readResult
	pending    uint64
	sentAt     time.Time
	turn       turnKey
	checkpoint Checkpoint
}

func dialTraced(ctx context.Context, target string, appVersion string, token string) (*tracedSeat, error) {
	c, err := botclient.Dial(ctx, target, botclient.Options{Token: token, AppVersion: appVersion})
	if err != nil {
		return nil, err
	}
	s := &tracedSeat{client: c, frames: make(chan readResult, 16)}
	go func() {
		defer close(s.frames)
		for {
			frame, err := c.Receive()
			s.frames <- readResult{frame, err}
			if err != nil {
				return
			}
		}
	}()
	return s, nil
}

func (s *tracedSeat) view(frame *pb.Envelope) (*pb.MatchView, error) {
	if rejected := frame.GetError(); rejected != nil {
		return nil, fmt.Errorf("intent %d rejected: %s", rejected.GetRejectedSequence(), rejected.GetCode())
	}
	s.client.Observe(frame)
	return frame.GetMatchView(), nil
}

func (s *tracedSeat) settle(view *pb.MatchView, acked uint64) (time.Duration, bool) {
	if s.pending == 0 || acked < s.pending {
		return 0, false
	}
	s.pending = 0
	s.checkpoint = checkpointOf(view)
	return time.Since(s.sentAt), true
}

func (s *tracedSeat) send(frame *pb.Envelope) error {
	s.sentAt = time.Now()
	if err := s.client.Send(frame); err != nil {
		return err
	}
	s.pending = s.client.Sequence
	return nil
}

func wait(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func dropAll(seats []*tracedSeat) {
	for _, s := range seats {
		if s != nil {
			s.client.Drop()
		}
	}
}

func Record(ctx context.Context, config RecordConfig) (*Trace, error) {
	if config.Target == "" || config.Mode.Players() == 0 {
		return nil, errors.New("loadbot: record needs a target and a mode")
	}
	ctx, abort := context.WithCancel(ctx)
	defer abort()
	stop, cancel := ctx, context.CancelFunc(func() {})
	if config.Length > 0 {
		stop, cancel = context.WithTimeout(ctx, config.Length)
	}
	defer cancel()
	trace := &Trace{Version: TraceVersion, Mode: config.Mode.String(), Seed: config.Seed}
	seats := make([]*tracedSeat, len(config.Mode.Seats()))
	defer dropAll(seats)
	for index, seat := range config.Mode.Seats() {
		s, err := dialTraced(ctx, config.Target, config.AppVersion, SeededToken(config.Room, config.Mode, seat, config.Seed))
		if err != nil {
			return nil, err
		}
		seats[index] = s
		trace.Seats = append(trace.Seats, SeatTrace{Seat: seat.String()})
	}
	errs := make([]error, len(seats))
	var group sync.WaitGroup
	for index := range seats {
		group.Add(1)
		go func() {
			defer group.Done()
			random := rand.New(rand.NewPCG(config.Seed, uint64(index)))
			errs[index] = recordSeat(ctx, stop, config, seats[index], &trace.Seats[index], random)
			if errs[index] != nil {
				abort()
			}
		}()
	}
	group.Wait()
	return trace, errors.Join(errs...)
}

func recordSeat(ctx context.Context, stop context.Context, config RecordConfig, s *tracedSeat, out *SeatTrace, random *rand.Rand) error {
	stopped := stop.Done()
	for {
		select {
		case <-stopped:
			if s.pending == 0 {
				return nil
			}
			stopped = nil
			continue
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-s.frames:
			if !ok || res.err != nil {
				return fmt.Errorf("seat %s: connection lost: %v", out.Seat, res.err)
			}
			view, err := s.view(res.frame)
			if err != nil {
				return fmt.Errorf("seat %s: %w", out.Seat, err)
			}
			if view == nil {
				continue
			}
			s.settle(view, res.frame.GetAckedSequence())
			out.Final = s.checkpoint
			if botclient.Finished(view) {
				out.Final = checkpointOf(view)
				return nil
			}
			if stop.Err() != nil && s.pending == 0 {
				return nil
			}
			if botclient.NeedsReady(view) {
				if err := s.client.Ready(); err != nil {
					return fmt.Errorf("seat %s: ready: %w", out.Seat, err)
				}
				continue
			}
			if botclient.Idle(view) || s.pending != 0 || stop.Err() != nil {
				continue
			}
			key := keyOf(view)
			think := thinkTime(config, random, key != s.turn)
			if !wait(stop, think) {
				return nil
			}
			s.turn = key
			intent := botclient.NextIntent(view)
			frame, err := proto.Marshal(intent)
			if err != nil {
				return err
			}
			out.Intents = append(out.Intents, TracedIntent{ThinkMs: think.Milliseconds(), Frame: frame})
			if err := s.send(intent); err != nil {
				return fmt.Errorf("seat %s: send: %w", out.Seat, err)
			}
		}
	}
}

func thinkTime(config RecordConfig, random *rand.Rand, firstOfTurn bool) time.Duration {
	if !firstOfTurn {
		return config.ThinkFollow.Truncate(time.Millisecond)
	}
	think := config.ThinkMin
	if span := config.ThinkMax - config.ThinkMin; span > 0 {
		think += time.Duration(random.Int64N(int64(span)))
	}
	return think.Truncate(time.Millisecond)
}

type ReplayRun struct {
	Index   int           `json:"index"`
	Room    uint64        `json:"room"`
	Moves   uint64        `json:"moves"`
	P50     time.Duration `json:"p50_ns"`
	P95     time.Duration `json:"p95_ns"`
	P99     time.Duration `json:"p99_ns"`
	Max     time.Duration `json:"max_ns"`
	Elapsed time.Duration `json:"elapsed_ns"`
	Failure string        `json:"failure,omitempty"`
}

type ReplayReport struct {
	Mode     string        `json:"mode"`
	Seed     uint64        `json:"seed"`
	Intents  int           `json:"intents"`
	Budget   time.Duration `json:"budget_ns"`
	Runs     []ReplayRun   `json:"runs"`
	Moves    uint64        `json:"moves"`
	P50      time.Duration `json:"p50_ns"`
	P99      time.Duration `json:"p99_ns"`
	Max      time.Duration `json:"max_ns"`
	WorstP99 time.Duration `json:"worst_run_p99_ns"`
	Pass     bool          `json:"pass"`
	Failures []string      `json:"failures,omitempty"`
}

func Replay(ctx context.Context, trace *Trace, config ReplayConfig) (*ReplayReport, error) {
	mode, err := trace.mode()
	if err != nil {
		return nil, err
	}
	if config.Target == "" {
		return nil, errors.New("loadbot: replay needs a target")
	}
	runs, parallel := max(config.Runs, 1), max(config.Parallel, 1)
	report := &ReplayReport{Mode: trace.Mode, Seed: trace.Seed, Intents: trace.Intents(), Budget: config.Budget}
	overall := &Histogram{}
	for index := 0; index < runs; index++ {
		batch := make([]ReplayRun, parallel)
		var group sync.WaitGroup
		for slot := range batch {
			group.Add(1)
			go func() {
				defer group.Done()
				room := config.Room + uint64(index*parallel+slot)
				batch[slot] = replayOnce(ctx, trace, mode, config, room, overall)
				batch[slot].Index = index*parallel + slot
			}()
		}
		group.Wait()
		report.Runs = append(report.Runs, batch...)
	}
	report.Moves, report.P50, report.P99, report.Max = overall.Count(), overall.Percentile(50), overall.Percentile(99), overall.Max()
	report.judge()
	return report, nil
}

func (rep *ReplayReport) judge() {
	rep.Failures = nil
	for _, run := range rep.Runs {
		rep.WorstP99 = max(rep.WorstP99, run.P99)
		switch {
		case run.Failure != "":
			rep.Failures = append(rep.Failures, fmt.Sprintf("run %d: %s", run.Index, run.Failure))
		case rep.Budget > 0 && run.P99 >= rep.Budget:
			rep.Failures = append(rep.Failures, fmt.Sprintf("run %d: p99 %s is not under the budget %s", run.Index, run.P99, rep.Budget))
		}
	}
	if len(rep.Runs) == 0 {
		rep.Failures = append(rep.Failures, "no run")
	}
	rep.Pass = len(rep.Failures) == 0
}

func (rep *ReplayReport) Text() string {
	text := fmt.Sprintf("replay %s seed %d, %d intents per run, %d runs\n", rep.Mode, rep.Seed, rep.Intents, len(rep.Runs))
	text += "run  room  moves  p50  p95  p99  max  elapsed\n"
	for _, run := range rep.Runs {
		text += fmt.Sprintf("%3d %d %d %s %s %s %s %s %s\n", run.Index, run.Room, run.Moves, run.P50, run.P95, run.P99, run.Max, run.Elapsed.Round(time.Millisecond), run.Failure)
	}
	text += fmt.Sprintf("all runs: moves %d p50 %s p99 %s max %s, worst run p99 %s, budget %s\n", rep.Moves, rep.P50, rep.P99, rep.Max, rep.WorstP99, rep.Budget)
	if rep.Pass {
		return text + "PASS\n"
	}
	for _, failure := range rep.Failures {
		text += "FAIL: " + failure + "\n"
	}
	return text
}

func replayOnce(ctx context.Context, trace *Trace, mode rules.Mode, config ReplayConfig, room uint64, overall *Histogram) ReplayRun {
	run := ReplayRun{Room: room}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if config.RunTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.RunTimeout)
		defer cancel()
	}
	started := time.Now()
	latencies := &Histogram{}
	seats := make([]*tracedSeat, len(trace.Seats))
	defer dropAll(seats)
	for index, seat := range mode.Seats() {
		s, err := dialTraced(ctx, config.Target, config.AppVersion, SeededToken(room, mode, seat, trace.Seed))
		if err != nil {
			run.Failure = fmt.Sprintf("seat %s: dial: %v", seat, err)
			return run
		}
		seats[index] = s
	}
	errs := make([]error, len(seats))
	var group sync.WaitGroup
	for index := range seats {
		group.Add(1)
		go func() {
			defer group.Done()
			errs[index] = replaySeat(ctx, seats[index], &trace.Seats[index], func(d time.Duration) {
				latencies.Add(d)
				overall.Add(d)
			})
			if errs[index] != nil {
				cancel()
			}
		}()
	}
	group.Wait()
	run.Elapsed = time.Since(started)
	run.Moves, run.P50, run.P95, run.P99, run.Max = latencies.Count(), latencies.Percentile(50), latencies.Percentile(95), latencies.Percentile(99), latencies.Max()
	if err := errors.Join(errs...); err != nil {
		run.Failure = err.Error()
	}
	return run
}

func replaySeat(ctx context.Context, s *tracedSeat, recorded *SeatTrace, observe func(time.Duration)) error {
	next := 0
	for {
		if next == len(recorded.Intents) && s.pending == 0 && (recorded.Final.State != finishedState || s.checkpoint.State == finishedState) {
			if s.checkpoint != recorded.Final {
				return fmt.Errorf("seat %s diverged: %+v, recorded %+v", recorded.Seat, s.checkpoint, recorded.Final)
			}
			return nil
		}
		var res readResult
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok = <-s.frames:
		}
		if !ok || res.err != nil {
			return fmt.Errorf("seat %s: connection lost after %d of %d intents: %v", recorded.Seat, next, len(recorded.Intents), res.err)
		}
		view, err := s.view(res.frame)
		if err != nil {
			return fmt.Errorf("seat %s diverged: %w", recorded.Seat, err)
		}
		if view == nil {
			continue
		}
		if latency, settled := s.settle(view, res.frame.GetAckedSequence()); settled {
			observe(latency)
		}
		if botclient.Finished(view) {
			s.checkpoint = checkpointOf(view)
			if next < len(recorded.Intents) {
				return fmt.Errorf("seat %s diverged: match finished after %d of %d intents", recorded.Seat, next, len(recorded.Intents))
			}
			continue
		}
		if botclient.NeedsReady(view) {
			if err := s.client.Ready(); err != nil {
				return fmt.Errorf("seat %s: ready: %w", recorded.Seat, err)
			}
			continue
		}
		if botclient.Idle(view) || s.pending != 0 || next == len(recorded.Intents) {
			continue
		}
		intent := recorded.Intents[next]
		frame := &pb.Envelope{}
		if err := proto.Unmarshal(intent.Frame, frame); err != nil {
			return fmt.Errorf("%w: seat %s intent %d: %v", ErrBadTrace, recorded.Seat, next, err)
		}
		if !wait(ctx, time.Duration(intent.ThinkMs)*time.Millisecond) {
			return ctx.Err()
		}
		if err := s.send(frame); err != nil {
			return fmt.Errorf("seat %s: send: %w", recorded.Seat, err)
		}
		next++
	}
}
