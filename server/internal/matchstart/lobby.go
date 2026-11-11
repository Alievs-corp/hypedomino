package matchstart

import (
	"context"

	"github.com/Alievs-corp/hypedomino/server/internal/bot"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/custom"
	"github.com/Alievs-corp/hypedomino/server/internal/matchmaking"
	"github.com/Alievs-corp/hypedomino/server/internal/room"
	"github.com/Alievs-corp/hypedomino/server/internal/rules"
)

var CharacterClips = []string{"idle", "win", "lose", "reaction"}

type Random interface {
	Uint64() uint64
}

func LobbyMember(freeze FreezeFunc, item func(id string) (cosmetics.Item, bool), muted MutedFunc) func(ctx context.Context, account string) (custom.Member, error) {
	return func(ctx context.Context, account string) (custom.Member, error) {
		loadout, err := freeze(ctx, account)
		if err != nil {
			return custom.Member{}, err
		}
		silenced, err := muted(ctx, account)
		if err != nil {
			return custom.Member{}, err
		}
		member := custom.Member{Account: account, Actions: loadout.Actions, Emoji: loadout.Emojis, Muted: silenced}
		if character, ok := item(loadout.Character); ok && character.Animated {
			member.Clips = CharacterClips
		}
		return member, nil
	}
}

func LobbyOpener(rooms Rooms, roster RosterFunc, config func() matchmaking.Config, clock room.Clock, random Random) custom.Opener {
	return func(options room.Options, seats []custom.SeatState) (*room.Room, error) {
		match := matchmaking.Match{Mode: options.Mode}
		for _, seat := range seats {
			match.Seats = append(match.Seats, matchmaking.Seat{Seat: seat.Seat, Account: seat.Account})
		}
		players, err := roster(context.Background(), match)
		if err != nil {
			return nil, err
		}
		options.Config, options.Bot, options.BotLevel = rules.DefaultConfig(), bot.DefaultConfig(), config().BotLevel(matchmaking.QueueCasual)
		options.Seed, options.Clock, options.Roster = random.Uint64(), clock, players
		return rooms.Open(options)
	}
}
