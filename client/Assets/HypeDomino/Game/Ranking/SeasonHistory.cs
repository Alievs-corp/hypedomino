using System;
using HypeDomino.Net;
using UnityEngine;

namespace HypeDomino.Game.Ranking
{
    [Serializable]
    public sealed class SeasonRewardReply
    {
        public bool granted;
        public long coins;
        public string bundle_id;
        public string[] emblems;
    }

    [Serializable]
    public sealed class PastModeReply
    {
        public string mode;
        public string rank;
        public string tier;
        public int lp;
        public int position;
        public SeasonRewardReply reward;
    }

    [Serializable]
    public sealed class PastSeasonReply
    {
        public string id;
        public string name_key;
        public string starts_at;
        public string ends_at;
        public PastModeReply[] modes;
    }

    [Serializable]
    public sealed class SeasonHistoryReply
    {
        public PastSeasonReply[] seasons;

        public static SeasonHistoryReply FromJson(string json) => JsonUtility.FromJson<SeasonHistoryReply>(json);
    }

    public readonly struct PastSeasonMode
    {
        public PastSeasonMode(string mode, RankBadge badge, int lp, int position, bool rewarded, long rewardCoins)
        {
            Mode = mode;
            Badge = badge;
            LP = lp;
            Position = position;
            Rewarded = rewarded;
            RewardCoins = rewardCoins;
        }

        public string Mode { get; }

        public RankBadge Badge { get; }

        public int LP { get; }

        public int Position { get; }

        public bool Rewarded { get; }

        public long RewardCoins { get; }
    }

    public sealed class PastSeason
    {
        readonly PastSeasonMode[] modes;

        internal PastSeason(PastSeasonReply season, PastSeasonMode[] modes)
        {
            Id = season.id;
            NameKey = season.name_key ?? string.Empty;
            this.modes = modes;
        }

        public string Id { get; }

        public string NameKey { get; }

        public int ModeCount => modes.Length;

        public PastSeasonMode ModeAt(int index) => modes[index];

        public bool TryFind(string mode, out PastSeasonMode found)
        {
            for (var i = 0; i < modes.Length; i++)
            {
                if (string.Equals(modes[i].Mode, mode, StringComparison.Ordinal))
                {
                    found = modes[i];
                    return true;
                }
            }
            found = default;
            return false;
        }
    }

    public sealed class SeasonHistory
    {
        public const string Path = "/ranked/seasons/history";

        readonly PastSeason[] seasons;

        SeasonHistory(PastSeason[] seasons) => this.seasons = seasons;

        public int Count => seasons.Length;

        public PastSeason SeasonAt(int index) => seasons[index];

        public static bool TryRead(string json, out SeasonHistory history)
        {
            history = null;
            if (!TryParse(json, out var reply))
            {
                return false;
            }
            var replies = reply.seasons ?? Array.Empty<PastSeasonReply>();
            var read = new PastSeason[replies.Length];
            for (var i = 0; i < replies.Length; i++)
            {
                if (!TrySeason(replies[i], out read[i]))
                {
                    return false;
                }
            }
            history = new SeasonHistory(read);
            return true;
        }

        static bool TryParse(string json, out SeasonHistoryReply reply)
        {
            reply = null;
            if (string.IsNullOrEmpty(json))
            {
                return false;
            }
            try
            {
                reply = SeasonHistoryReply.FromJson(json);
            }
            catch (ArgumentException)
            {
                return false;
            }
            return reply != null;
        }

        static bool TrySeason(PastSeasonReply reply, out PastSeason season)
        {
            season = null;
            if (reply == null || string.IsNullOrEmpty(reply.id) || reply.modes == null || reply.modes.Length == 0)
            {
                return false;
            }
            if (!ServerClock.TryReadUnixMs(reply.starts_at, out var startsUnixMs) || !ServerClock.TryReadUnixMs(reply.ends_at, out var endsUnixMs) || endsUnixMs <= startsUnixMs)
            {
                return false;
            }
            var modes = new PastSeasonMode[reply.modes.Length];
            for (var i = 0; i < modes.Length; i++)
            {
                if (!TryMode(reply.modes[i], out modes[i]))
                {
                    return false;
                }
                for (var j = 0; j < i; j++)
                {
                    if (string.Equals(modes[j].Mode, modes[i].Mode, StringComparison.Ordinal))
                    {
                        return false;
                    }
                }
            }
            season = new PastSeason(reply, modes);
            return true;
        }

        static bool TryMode(PastModeReply reply, out PastSeasonMode mode)
        {
            mode = default;
            if (reply == null || string.IsNullOrEmpty(reply.mode) || reply.lp < 0 || reply.position < 1 || !TryBadge(reply.rank, reply.tier, out var badge))
            {
                return false;
            }
            var reward = reply.reward ?? new SeasonRewardReply();
            if (reward.coins < 0 || (!reward.granted && reward.coins != 0))
            {
                return false;
            }
            mode = new PastSeasonMode(reply.mode, badge, reply.lp, reply.position, reward.granted, reward.coins);
            return true;
        }

        static bool TryBadge(string rank, string tier, out RankBadge badge)
        {
            if (!RankBadge.TryRead(rank, out badge) || !badge.Placed)
            {
                return false;
            }
            return !badge.Apex || (RankBadge.TryRead(tier, out badge) && badge.Apex);
        }
    }
}
