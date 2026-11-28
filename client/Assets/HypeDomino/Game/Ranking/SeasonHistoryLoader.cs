using System;
using HypeDomino.Net.Api;

namespace HypeDomino.Game.Ranking
{
    public enum SeasonHistoryLoadOutcome
    {
        Loaded,
        Offline,
        SignInRequired,
        RateLimited,
        Failed,
    }

    public readonly struct SeasonHistoryLoadResult
    {
        SeasonHistoryLoadResult(SeasonHistoryLoadOutcome outcome, SeasonHistory history = null)
        {
            Outcome = outcome;
            History = history;
        }

        public SeasonHistoryLoadOutcome Outcome { get; }
        public SeasonHistory History { get; }

        public static SeasonHistoryLoadResult Loaded(SeasonHistory history) => new(SeasonHistoryLoadOutcome.Loaded, history);

        public static SeasonHistoryLoadResult Refused(ApiReply reply) => new(reply.Outcome switch
        {
            ApiOutcome.Offline => SeasonHistoryLoadOutcome.Offline,
            ApiOutcome.SignInRequired => SeasonHistoryLoadOutcome.SignInRequired,
            ApiOutcome.RateLimited => SeasonHistoryLoadOutcome.RateLimited,
            _ => SeasonHistoryLoadOutcome.Failed,
        });

        public static SeasonHistoryLoadResult Unreadable() => new(SeasonHistoryLoadOutcome.Failed);
    }

    public sealed class SeasonHistoryLoader
    {
        readonly ApiClient api;
        readonly Action<ApiReply> answered;
        Action<SeasonHistoryLoadResult> pending;

        public SeasonHistoryLoader(ApiClient api)
        {
            this.api = api;
            answered = Answered;
        }

        public bool Busy => pending != null;

        public bool Load(Action<SeasonHistoryLoadResult> done)
        {
            if (Busy || done == null)
            {
                return false;
            }
            pending = done;
            api.Get(SeasonHistory.Path, answered);
            return true;
        }

        void Answered(ApiReply reply)
        {
            var done = pending;
            pending = null;
            done(Read(reply));
        }

        static SeasonHistoryLoadResult Read(ApiReply reply)
        {
            if (reply.Outcome != ApiOutcome.Ok)
            {
                return SeasonHistoryLoadResult.Refused(reply);
            }
            return SeasonHistory.TryRead(reply.Body, out var history) ? SeasonHistoryLoadResult.Loaded(history) : SeasonHistoryLoadResult.Unreadable();
        }
    }
}
