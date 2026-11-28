using System;
using System.Collections.Generic;
using HypeDomino.Game.Ranking;
using HypeDomino.Net;

namespace HypeDomino.Game.Match
{
    [Flags]
    public enum MatchRewardParts
    {
        None = 0,
        Coins = 1,
        PassXp = 2,
        Missions = 4,
        Rating = 8,
    }

    public readonly struct MatchMissionReward
    {
        readonly string[] medals;

        public MatchMissionReward(string mission, string period, long coins, long passXp, string[] medals)
        {
            Mission = mission;
            Period = period;
            Coins = coins;
            PassXp = passXp;
            this.medals = medals;
        }

        public string Mission { get; }
        public string Period { get; }
        public long Coins { get; }
        public long PassXp { get; }
        public IReadOnlyList<string> Medals => medals;
    }

    public readonly struct MatchRatingChange
    {
        public MatchRatingChange(bool placement, long lpDelta, RankBadge before, long lpBefore, RankBadge after, long lpAfter)
        {
            Placement = placement;
            LpDelta = lpDelta;
            Before = before;
            LpBefore = lpBefore;
            After = after;
            LpAfter = lpAfter;
        }

        public bool Placement { get; }
        public long LpDelta { get; }
        public RankBadge Before { get; }
        public long LpBefore { get; }
        public RankBadge After { get; }
        public long LpAfter { get; }
    }

    public sealed class MatchRewards
    {
        public const string NotPlayed = "match.not_played";

        const string PathPrefix = "/matches/";
        const string PathSuffix = "/result";
        const double MillisecondsPerSecond = 1000;

        readonly MatchMissionReward[] missions;

        MatchRewards(string matchId, MatchRewardParts pending, bool complete, double retryAfterSeconds, long matchCoins, long firstWinCoins, long passXp, MatchMissionReward[] missions, bool hasRating, MatchRatingChange rating)
        {
            MatchId = matchId;
            Pending = pending;
            Complete = complete;
            RetryAfterSeconds = retryAfterSeconds;
            MatchCoins = matchCoins;
            FirstWinCoins = firstWinCoins;
            PassXp = passXp;
            this.missions = missions;
            HasRating = hasRating;
            Rating = rating;
        }

        public string MatchId { get; }
        public MatchRewardParts Pending { get; }
        public bool Complete { get; }
        public double RetryAfterSeconds { get; }
        public long MatchCoins { get; }
        public long FirstWinCoins { get; }
        public long PassXp { get; }
        public IReadOnlyList<MatchMissionReward> Missions => missions;
        public bool HasRating { get; }
        public MatchRatingChange Rating { get; }

        public bool IsPending(MatchRewardParts part) => (Pending & part) != 0;

        public static string PathOf(string matchId) => PathPrefix + Uri.EscapeDataString(matchId) + PathSuffix;

        public static bool TryRead(string json, out MatchRewards rewards)
        {
            rewards = null;
            if (!JsonMap.TryParse(json, out var reply) || !reply.TryString("match", out var matchId) || string.IsNullOrEmpty(matchId))
            {
                return false;
            }
            if (!reply.TryStrings("pending", out var pendingNames) || !TryRetry(reply, pendingNames.Length > 0, out var retryAfterSeconds))
            {
                return false;
            }
            if (!reply.TryMap("coins", out var coins) || !TryCount(coins, "match", out var matchCoins) || !TryCount(coins, "first_win", out var firstWinCoins) || !TryCount(reply, "pass_xp", out var passXp))
            {
                return false;
            }
            if (!TryMissions(reply, out var missions) || !TryRating(reply, out var hasRating, out var rating))
            {
                return false;
            }
            rewards = new MatchRewards(matchId, PartsOf(pendingNames), pendingNames.Length == 0, retryAfterSeconds, matchCoins, firstWinCoins, passXp, missions, hasRating, rating);
            return true;
        }

        static bool TryRetry(JsonMap reply, bool pending, out double seconds)
        {
            seconds = 0;
            if (!pending)
            {
                return true;
            }
            if (!reply.TryInteger("retry_after_ms", out var milliseconds) || milliseconds < 1)
            {
                return false;
            }
            seconds = milliseconds / MillisecondsPerSecond;
            return true;
        }

        static MatchRewardParts PartsOf(string[] names)
        {
            var parts = MatchRewardParts.None;
            foreach (var name in names)
            {
                parts |= name switch
                {
                    "coins" => MatchRewardParts.Coins,
                    "pass_xp" => MatchRewardParts.PassXp,
                    "missions" => MatchRewardParts.Missions,
                    "rating" => MatchRewardParts.Rating,
                    _ => MatchRewardParts.None,
                };
            }
            return parts;
        }

        static bool TryCount(JsonMap map, string key, out long count) => map.TryInteger(key, out count) && count >= 0;

        static bool TryMissions(JsonMap reply, out MatchMissionReward[] missions)
        {
            missions = null;
            if (!reply.TryMaps("missions", out var maps))
            {
                return false;
            }
            var read = new MatchMissionReward[maps.Length];
            for (var i = 0; i < maps.Length; i++)
            {
                var map = maps[i];
                if (!map.TryString("mission", out var mission) || string.IsNullOrEmpty(mission) || !map.TryString("period", out var period) || string.IsNullOrEmpty(period))
                {
                    return false;
                }
                if (!TryCount(map, "coins", out var coins) || !TryCount(map, "pass_xp", out var passXp) || !map.TryStrings("medals", out var medals))
                {
                    return false;
                }
                read[i] = new MatchMissionReward(mission, period, coins, passXp, medals);
            }
            missions = read;
            return true;
        }

        static bool TryRating(JsonMap reply, out bool hasRating, out MatchRatingChange rating)
        {
            hasRating = false;
            rating = default;
            if (reply.IsAbsent("rating"))
            {
                return true;
            }
            if (!reply.TryMap("rating", out var map) || !map.TryBool("placement", out var placement) || !map.TryInteger("lp_delta", out var lpDelta))
            {
                return false;
            }
            if (placement)
            {
                hasRating = true;
                rating = new MatchRatingChange(true, lpDelta, default, 0, default, 0);
                return true;
            }
            if (!TryRank(map, "rank_before", out var before) || !TryCount(map, "lp_before", out var lpBefore) || !TryRank(map, "rank_after", out var after) || !TryCount(map, "lp_after", out var lpAfter))
            {
                return false;
            }
            hasRating = true;
            rating = new MatchRatingChange(false, lpDelta, before, lpBefore, after, lpAfter);
            return true;
        }

        static bool TryRank(JsonMap map, string key, out RankBadge badge)
        {
            badge = default;
            return map.TryString(key, out var label) && RankBadge.TryRead(label, out badge) && badge.Placed;
        }
    }
}
