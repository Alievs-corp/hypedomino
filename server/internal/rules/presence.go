package rules

import (
	"fmt"
	"time"
)

type Controller uint8

const (
	ControllerPlayer Controller = iota
	ControllerBot
)

func (c Controller) String() string {
	if c == ControllerBot {
		return "bot"
	}
	return "player"
}

func (c Controller) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *Controller) UnmarshalText(text []byte) error {
	switch string(text) {
	case "player":
		*c = ControllerPlayer
	case "bot":
		*c = ControllerBot
	default:
		return fmt.Errorf("rules: invalid controller %q", text)
	}
	return nil
}

func TurnLimit(config Config) time.Duration {
	return time.Duration(config.TurnSeconds) * time.Second
}

func MovedBy(config Config, acted bool, actedAfter time.Duration) Controller {
	if acted && actedAfter < TurnLimit(config) {
		return ControllerPlayer
	}
	return ControllerBot
}

type seatPresence struct {
	consecutiveTimeouts int
	byBot               bool
}

type Presence struct {
	config Config
	seats  [seatCount]seatPresence
}

func NewPresence(config Config) Presence {
	return Presence{config: config}
}

func (p *Presence) Controller(seat Seat) Controller {
	if p.seats[seat].byBot {
		return ControllerBot
	}
	return ControllerPlayer
}

func (p *Presence) Timeout(seat Seat) Controller {
	s := &p.seats[seat]
	s.consecutiveTimeouts++
	if s.consecutiveTimeouts >= p.config.AfkConsecutiveTimeouts {
		s.byBot = true
	}
	return p.Controller(seat)
}

func (p *Presence) Move(seat Seat) {
	p.seats[seat].consecutiveTimeouts = 0
}

func (p *Presence) Reconnect(seat Seat) Controller {
	p.seats[seat] = seatPresence{}
	return p.Controller(seat)
}

func (p *Presence) Leave(seat Seat) {
	p.seats[seat].byBot = true
}
