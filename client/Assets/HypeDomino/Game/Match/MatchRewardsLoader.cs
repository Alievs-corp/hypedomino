using System;
using HypeDomino.Net.Api;

namespace HypeDomino.Game.Match
{
    public enum MatchRewardsLoadOutcome
    {
        Loaded,
        Partial,
        NotPlayed,
        Offline,
        SignInRequired,
        RateLimited,
        Failed,
    }

    public readonly struct MatchRewardsLoadResult
    {
        public MatchRewardsLoadResult(MatchRewardsLoadOutcome outcome, MatchRewards rewards, bool final)
        {
            Outcome = outcome;
            Rewards = rewards;
            Final = final;
        }

        public MatchRewardsLoadOutcome Outcome { get; }
        public MatchRewards Rewards { get; }
        public bool Final { get; }

        public static MatchRewardsLoadOutcome RefusalOf(ApiReply reply) => reply.Outcome switch
        {
            ApiOutcome.Offline => MatchRewardsLoadOutcome.Offline,
            ApiOutcome.SignInRequired => MatchRewardsLoadOutcome.SignInRequired,
            ApiOutcome.RateLimited => MatchRewardsLoadOutcome.RateLimited,
            ApiOutcome.Rejected when reply.Error == MatchRewards.NotPlayed => MatchRewardsLoadOutcome.NotPlayed,
            _ => MatchRewardsLoadOutcome.Failed,
        };
    }

    public sealed class MatchRewardsLoader
    {
        public const int MaxReads = 5;

        readonly ApiClient api;
        readonly Action<ApiReply> answered;
        MatchResult result;
        Action<MatchRewardsLoadResult> done;
        string path;
        int reads;
        bool waiting;
        bool waitStarted;
        double waitSeconds;
        double readAgainAt;

        public MatchRewardsLoader(ApiClient api)
        {
            this.api = api;
            answered = Answered;
        }

        public bool Busy => done != null;

        public bool Load(MatchResult result, Action<MatchRewardsLoadResult> done)
        {
            if (Busy || result == null || done == null || string.IsNullOrEmpty(result.MatchId))
            {
                return false;
            }
            this.result = result;
            this.done = done;
            path = MatchRewards.PathOf(result.MatchId);
            reads = 0;
            ReadNow();
            return true;
        }

        public void Tick(double nowSeconds)
        {
            if (!waiting)
            {
                return;
            }
            if (!waitStarted)
            {
                waitStarted = true;
                readAgainAt = nowSeconds + waitSeconds;
                return;
            }
            if (nowSeconds < readAgainAt)
            {
                return;
            }
            waiting = false;
            ReadNow();
        }

        void ReadNow()
        {
            reads++;
            api.Get(path, answered);
        }

        void Answered(ApiReply reply)
        {
            var outcome = Read(reply, out var rewards);
            if (rewards != null && !result.ShowRewards(rewards))
            {
                outcome = MatchRewardsLoadOutcome.Failed;
                rewards = null;
            }
            var final = outcome != MatchRewardsLoadOutcome.Partial || reads >= MaxReads;
            var callback = done;
            if (final)
            {
                done = null;
                result = null;
            }
            else
            {
                waiting = true;
                waitStarted = false;
                waitSeconds = rewards.RetryAfterSeconds;
            }
            callback(new MatchRewardsLoadResult(outcome, rewards, final));
        }

        static MatchRewardsLoadOutcome Read(ApiReply reply, out MatchRewards rewards)
        {
            rewards = null;
            if (reply.Outcome != ApiOutcome.Ok)
            {
                return MatchRewardsLoadResult.RefusalOf(reply);
            }
            if (!MatchRewards.TryRead(reply.Body, out rewards))
            {
                return MatchRewardsLoadOutcome.Failed;
            }
            return rewards.Complete ? MatchRewardsLoadOutcome.Loaded : MatchRewardsLoadOutcome.Partial;
        }
    }
}
