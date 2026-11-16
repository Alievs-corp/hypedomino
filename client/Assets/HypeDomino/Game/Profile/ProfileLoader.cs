using System;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.UI.Profile;

namespace HypeDomino.Game.Profile
{
    public enum ProfileLoadOutcome
    {
        Loaded,
        SetupRequired,
        Offline,
        SignInRequired,
        RateLimited,
        Failed,
    }

    public static class ProfileLoadKeys
    {
        public const string Offline = "profile.load.offline";
        public const string RateLimited = "profile.load.rate_limited";
        public const string Failed = "notice.generic";

        public static string Of(ProfileLoadOutcome outcome) => outcome switch
        {
            ProfileLoadOutcome.Loaded => null,
            ProfileLoadOutcome.SetupRequired => null,
            ProfileLoadOutcome.Offline => Offline,
            ProfileLoadOutcome.RateLimited => RateLimited,
            ProfileLoadOutcome.SignInRequired => null,
            _ => Failed,
        };
    }

    public readonly struct ProfileLoadResult
    {
        ProfileLoadResult(ProfileLoadOutcome outcome, PlayerProfile profile = null)
        {
            Outcome = outcome;
            Profile = profile;
        }

        public ProfileLoadOutcome Outcome { get; }
        public PlayerProfile Profile { get; }
        public string MessageKey => ProfileLoadKeys.Of(Outcome);

        public static ProfileLoadResult Loaded(PlayerProfile profile) => new(ProfileLoadOutcome.Loaded, profile);

        public static ProfileLoadResult Refused(ApiReply reply) => new(reply.Outcome switch
        {
            ApiOutcome.Offline => ProfileLoadOutcome.Offline,
            ApiOutcome.SignInRequired => ProfileLoadOutcome.SignInRequired,
            ApiOutcome.RateLimited => ProfileLoadOutcome.RateLimited,
            _ => ProfileLoadOutcome.Failed,
        });

        public static ProfileLoadResult Unread(string body) => new(SetupRequired(body) ? ProfileLoadOutcome.SetupRequired : ProfileLoadOutcome.Failed);

        static bool SetupRequired(string body) => JsonMap.TryParse(body, out _) && OwnProfileReply.FromJson(body).setup_required;
    }

    public sealed class ProfileLoader
    {
        readonly ApiClient api;

        public ProfileLoader(ApiClient api)
        {
            this.api = api;
        }

        public bool Busy { get; private set; }

        public bool Load(Action<ProfileLoadResult> done)
        {
            if (Busy)
            {
                return false;
            }
            Busy = true;
            api.Get(PlayerProfile.OwnPath, reply => Answered(reply, done));
            return true;
        }

        void Answered(ApiReply reply, Action<ProfileLoadResult> done)
        {
            Busy = false;
            done(Read(reply));
        }

        static ProfileLoadResult Read(ApiReply reply)
        {
            if (reply.Outcome != ApiOutcome.Ok)
            {
                return ProfileLoadResult.Refused(reply);
            }
            return PlayerProfile.TryFromOwn(reply.Body, out var profile) ? ProfileLoadResult.Loaded(profile) : ProfileLoadResult.Unread(reply.Body);
        }
    }
}
