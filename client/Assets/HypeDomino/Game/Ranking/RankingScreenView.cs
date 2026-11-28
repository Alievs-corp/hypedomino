using HypeDomino.Net;
using HypeDomino.UI.Localization;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Ranking
{
    public sealed class RankingScreenView : MonoBehaviour
    {
        public const string SoloMode = "1v1";
        public const string TeamsMode = "2v2";

        [SerializeField] RankScreenView rank;
        [SerializeField] LeaderboardView leaderboard;
        [SerializeField] SeasonHistoryView history;
        [SerializeField] Button solo;
        [SerializeField] Button teams;

        RankStandingLoader standings;
        LeaderboardLoader boards;
        SeasonHistoryLoader pastSeasons;
        ServerClock clock;
        GameStrings strings;

        public void Configure(RankScreenView rankView, LeaderboardView leaderboardView, SeasonHistoryView historyView, Button soloButton, Button teamsButton)
        {
            Unbind();
            rank = rankView;
            leaderboard = leaderboardView;
            history = historyView;
            solo = soloButton;
            teams = teamsButton;
        }

        public void Bind(RankStandingLoader standingLoader, LeaderboardLoader leaderboardLoader, SeasonHistoryLoader historyLoader, ServerClock serverClock, GameStrings texts)
        {
            Unbind();
            standings = standingLoader;
            boards = leaderboardLoader;
            pastSeasons = historyLoader;
            clock = serverClock;
            strings = texts;
            solo.onClick.AddListener(ShowSolo);
            teams.onClick.AddListener(ShowTeams);
            Show(SoloMode);
        }

        void Start()
        {
            if (strings == null && NetBoot.Api != null)
                Bind(new RankStandingLoader(NetBoot.Api), new LeaderboardLoader(NetBoot.Api), new SeasonHistoryLoader(NetBoot.Api), NetBoot.Connection.Clock, LanguageBoot.Strings);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            solo.onClick.RemoveListener(ShowSolo);
            teams.onClick.RemoveListener(ShowTeams);
            standings = null;
            boards = null;
            pastSeasons = null;
            clock = null;
            strings = null;
        }

        void ShowSolo() => Show(SoloMode);

        void ShowTeams() => Show(TeamsMode);

        void Show(string mode)
        {
            rank.Bind(standings, mode, clock, strings);
            leaderboard.Bind(boards, mode, LeaderboardStanding.BoardGlobal, strings);
            history.Bind(pastSeasons, mode, strings);
            solo.interactable = mode != SoloMode;
            teams.interactable = mode != TeamsMode;
        }
    }
}
