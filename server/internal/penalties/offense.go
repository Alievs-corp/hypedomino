package penalties

import (
	"github.com/Alievs-corp/hypedomino/server/internal/matchmaking"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

type Kind uint8

const (
	KindNone Kind = iota
	KindAfkReturned
	KindAfk
	KindLeaveReturned
	KindLeave
)

var kindNames = [...]string{"none", "afk_returned", "afk", "leave_returned", "leave"}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "invalid"
}

func (k Kind) Returned() bool {
	return k == KindAfkReturned || k == KindLeaveReturned
}

func (k Kind) Weight(c Config) int {
	switch k {
	case KindAfkReturned:
		return c.AfkReturnedWeight
	case KindAfk:
		return c.AfkWeight
	case KindLeaveReturned:
		return c.LeaveReturnedWeight
	case KindLeave:
		return c.LeaveWeight
	}
	return 0
}

func Detect(c Config, report rules.PresenceReport) Kind {
	kind := KindNone
	if report.MarkedAfk {
		kind = KindAfkReturned
		if report.ByBot {
			kind = KindAfk
		}
	}
	if !report.Left {
		return kind
	}
	leave := KindLeaveReturned
	if report.ByBot {
		leave = KindLeave
	}
	if leave.Weight(c) >= kind.Weight(c) {
		return leave
	}
	return kind
}

type SeatResult struct {
	CountsAsLoss         bool
	RatingPenalty        int
	RewardsKept          bool
	LossReductionPercent int
}

func Results(c Config, queue matchmaking.Queue, mode rules.Mode, winner rules.Team, hasWinner bool, offenses [6]Kind) [6]SeatResult {
	var results [6]SeatResult
	ranked := queue == matchmaking.QueueRanked
	for _, seat := range mode.Seats() {
		result := SeatResult{RewardsKept: true}
		if kind := offenses[seat]; kind != KindNone {
			result.CountsAsLoss = ranked
			result.RewardsKept = kind.Returned()
			if ranked {
				result.RatingPenalty = c.OffenseRatingPenalty
			}
		} else if partner, ok := seat.Partner(); ok && ranked && offenses[partner] != KindNone {
			if !hasWinner || winner != seat.Team() {
				result.LossReductionPercent = c.PartnerLossReductionPercent
			}
		}
		results[seat] = result
	}
	return results
}
