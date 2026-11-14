using System;
using System.Collections.Generic;
using HypeDomino.UI.Cosmetics;

namespace HypeDomino.UI.Profile
{
    public enum ProfileStat
    {
        Games,
        Wins,
        Losses,
        WinRatePercent,
        Streak,
        BestStreak,
    }

    public readonly struct ProfileStatValue
    {
        public ProfileStatValue(ProfileStat stat, int value)
        {
            Stat = stat;
            Value = value;
        }

        public ProfileStat Stat { get; }
        public int Value { get; }
    }

    public readonly struct ShowcaseMedal
    {
        public ShowcaseMedal(ShowcaseMedalReply reply)
        {
            Id = reply.id;
            NameKey = reply.nameKey;
            RarityId = reply.rarityId;
            StaticAddress = reply.staticAddress;
        }

        public string Id { get; }
        public string NameKey { get; }
        public string RarityId { get; }
        public string StaticAddress { get; }
    }

    public readonly struct ShowcaseEmblem
    {
        public ShowcaseEmblem(ShowcaseEmblemReply reply)
        {
            Id = reply.id;
            Season = reply.season;
            Mode = reply.mode;
            Board = reply.board;
            Bracket = reply.bracket;
            Position = reply.position;
        }

        public string Id { get; }
        public string Season { get; }
        public string Mode { get; }
        public string Board { get; }
        public int Bracket { get; }
        public int Position { get; }
    }

    public sealed class PlayerProfile
    {
        public const string OwnPath = "/profile/me";

        const string PublicPathPrefix = "/profile/";

        readonly ProfileStatValue[] stats;
        readonly ShowcaseMedal[] medals;
        readonly ShowcaseEmblem[] emblems;

        PlayerProfile(string accountId, string nickname, int level, long xp, bool isOwn, LoadoutReply loadout, ProfileStatValue[] stats, ShowcaseMedal[] medals, ShowcaseEmblem[] emblems)
        {
            AccountId = accountId;
            Nickname = nickname;
            Level = level;
            Xp = xp;
            IsOwn = isOwn;
            AvatarId = loadout?.avatar;
            FrameId = loadout?.frame;
            this.stats = stats;
            this.medals = medals;
            this.emblems = emblems;
        }

        public string AccountId { get; }
        public string Nickname { get; }
        public int Level { get; }
        public long Xp { get; }
        public bool IsOwn { get; }
        public string AvatarId { get; }
        public string FrameId { get; }
        public int StatCount => stats.Length;
        public int MedalCount => medals.Length;
        public int EmblemCount => emblems.Length;

        public ProfileStatValue Stat(int index) => stats[index];

        public ShowcaseMedal Medal(int index) => medals[index];

        public ShowcaseEmblem Emblem(int index) => emblems[index];

        public AvatarLook Look(CosmeticManifest manifest) => AvatarLook.Of(manifest, AvatarId, FrameId);

        public static string PublicPath(string accountId) => PublicPathPrefix + Uri.EscapeDataString(accountId);

        public static bool TryFromOwn(OwnProfileReply reply, out PlayerProfile profile)
        {
            profile = null;
            if (reply == null || reply.setup_required || !Named(reply.account_id, reply.nickname))
            {
                return false;
            }
            profile = new PlayerProfile(reply.account_id, reply.nickname, reply.level, reply.xp, true, reply.loadout, OwnStats(reply.stats ?? new OwnStatsReply()),
                Array.Empty<ShowcaseMedal>(), Array.Empty<ShowcaseEmblem>());
            return true;
        }

        public static bool TryFromPublic(PublicProfileReply reply, out PlayerProfile profile)
        {
            profile = null;
            if (reply == null || !Named(reply.account_id, reply.nickname))
            {
                return false;
            }
            profile = new PlayerProfile(reply.account_id, reply.nickname, reply.level, 0, false, reply.loadout, PublicStats(reply.stats ?? new PublicStatsReply()),
                Medals(reply.medals), Emblems(reply.ranked?.emblems));
            return true;
        }

        static bool Named(string accountId, string nickname) => !string.IsNullOrEmpty(accountId) && !string.IsNullOrEmpty(nickname);

        static ProfileStatValue[] OwnStats(OwnStatsReply reply) => new[]
        {
            new ProfileStatValue(ProfileStat.Games, reply.games),
            new ProfileStatValue(ProfileStat.Wins, reply.wins),
            new ProfileStatValue(ProfileStat.Losses, reply.losses),
            new ProfileStatValue(ProfileStat.WinRatePercent, reply.win_rate_percent),
            new ProfileStatValue(ProfileStat.Streak, reply.streak),
            new ProfileStatValue(ProfileStat.BestStreak, reply.best_streak),
        };

        static ProfileStatValue[] PublicStats(PublicStatsReply reply) => new[]
        {
            new ProfileStatValue(ProfileStat.Games, reply.games),
            new ProfileStatValue(ProfileStat.WinRatePercent, reply.win_rate_percent),
            new ProfileStatValue(ProfileStat.BestStreak, reply.best_streak),
        };

        static ShowcaseMedal[] Medals(ShowcaseMedalReply[] replies)
        {
            var shown = new List<ShowcaseMedal>();
            foreach (var reply in replies ?? Array.Empty<ShowcaseMedalReply>())
            {
                if (!string.IsNullOrEmpty(reply?.id))
                {
                    shown.Add(new ShowcaseMedal(reply));
                }
            }
            return shown.ToArray();
        }

        static ShowcaseEmblem[] Emblems(ShowcaseEmblemReply[] replies)
        {
            var shown = new List<ShowcaseEmblem>();
            foreach (var reply in replies ?? Array.Empty<ShowcaseEmblemReply>())
            {
                if (!string.IsNullOrEmpty(reply?.id))
                {
                    shown.Add(new ShowcaseEmblem(reply));
                }
            }
            return shown.ToArray();
        }
    }
}
