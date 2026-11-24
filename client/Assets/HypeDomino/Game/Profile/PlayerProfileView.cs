using System;
using HypeDomino.Game.Medals;
using HypeDomino.Game.Ranking;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Profile;
using TMPro;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.UI;

namespace HypeDomino.Game.Profile
{
    public sealed class PlayerProfileView : MonoBehaviour
    {
        [SerializeField] ProfileSummaryView summary;
        [SerializeField] Image rankBadge;
        [SerializeField] Sprite[] tierBadges;
        [SerializeField] MedalCellSlot[] showcase;
        [SerializeField] Button report;
        [SerializeField] GameObject reasonPanel;
        [SerializeField] Button[] reasons;
        [SerializeField] Button closeReasons;
        [SerializeField] TMP_Text message;

        Action<ProfileLoadResult> loaded;
        Action<PlayerReportResult> reported;
        UnityAction[] reasonPicks;
        ProfileLoader profiles;
        PlayerReport reporter;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        PlayerProfile profile;
        string accountId;
        string rankedMode;
        string matchId;
        string messageKey;
        bool reportSent;
        bool loadWaiting;
        int binding;

        public void Configure(ProfileSummaryView summaryView, Image rankBadgeImage, Sprite[] tierBadgeSprites, MedalCellSlot[] showcaseCells, TMP_Text messageLine)
        {
            Unbind();
            summary = summaryView;
            rankBadge = rankBadgeImage;
            tierBadges = tierBadgeSprites;
            showcase = showcaseCells;
            message = messageLine;
        }

        public void ConfigureReport(Button reportButton, GameObject reasonRoot, Button[] reasonButtons, Button closeButton)
        {
            Unbind();
            report = reportButton;
            reasonPanel = reasonRoot;
            reasons = reasonButtons;
            closeReasons = closeButton;
            reasonPicks = null;
        }

        public void Bind(ProfileLoader loader, PlayerReport playerReport, CosmeticStage cosmetics, ActiveCosmeticManifest cosmeticManifest, GameStrings texts,
            string playerAccountId, string mode, string currentMatchId)
        {
            Unbind();
            var current = ++binding;
            profiles = loader;
            reporter = playerReport;
            stage = cosmetics;
            manifest = cosmeticManifest;
            strings = texts;
            accountId = playerAccountId;
            rankedMode = mode;
            matchId = currentMatchId;
            profile = null;
            messageKey = null;
            reportSent = false;
            loaded = result => OnLoaded(current, result);
            reported = result => OnReported(current, result);
            reasonPicks ??= ReasonPicks();
            for (var at = 0; at < reasonPicks.Length; at++) reasons[at].onClick.AddListener(reasonPicks[at]);
            report.onClick.AddListener(OnReport);
            closeReasons.onClick.AddListener(OnCloseReasons);
            strings.Changed += Redraw;
            manifest.Changed += Redraw;
            reasonPanel.SetActive(false);
            Redraw();
            loadWaiting = !profiles.LoadPublic(accountId, loaded);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            binding++;
            for (var at = 0; at < reasonPicks.Length; at++) reasons[at].onClick.RemoveListener(reasonPicks[at]);
            report.onClick.RemoveListener(OnReport);
            closeReasons.onClick.RemoveListener(OnCloseReasons);
            strings.Changed -= Redraw;
            manifest.Changed -= Redraw;
            strings = null;
            manifest = null;
            reporter = null;
            profiles = null;
        }

        UnityAction[] ReasonPicks()
        {
            var count = Math.Min(reasons.Length, (int)PlayerReportReason.Other + 1);
            var picks = new UnityAction[count];
            for (var at = 0; at < count; at++)
            {
                var reason = (PlayerReportReason)at;
                picks[at] = () => OnReason(reason);
            }
            return picks;
        }

        void OnLoaded(int from, ProfileLoadResult result)
        {
            if (from != binding)
            {
                RequestWaitingLoad();
                return;
            }
            if (result.Outcome == ProfileLoadOutcome.Loaded) profile = result.Profile;
            messageKey = result.MessageKey;
            Redraw();
        }

        void RequestWaitingLoad()
        {
            if (!loadWaiting || strings == null) return;
            loadWaiting = !profiles.LoadPublic(accountId, loaded);
        }

        void OnReport()
        {
            reasonPanel.SetActive(true);
            ShowReport();
        }

        void OnCloseReasons()
        {
            if (reporter.Busy) return;
            reasonPanel.SetActive(false);
        }

        void OnReason(PlayerReportReason reason)
        {
            if (profile == null || !reporter.Send(profile.AccountId, reason, matchId, reported)) return;
            ShowReport();
        }

        void OnReported(int from, PlayerReportResult result)
        {
            if (from != binding) return;
            reportSent = result.Outcome == PlayerReportOutcome.Received;
            messageKey = result.MessageKey;
            reasonPanel.SetActive(false);
            Redraw();
        }

        void Redraw()
        {
            summary.Show(profile, stage, manifest.Current, strings);
            ShowRank();
            ShowMedals();
            ShowReport();
            message.enabled = messageKey != null;
            message.text = messageKey == null ? string.Empty : strings.Text(messageKey);
        }

        void ShowRank()
        {
            var badge = default(RankBadge);
            for (var at = 0; profile != null && at < profile.RankCount; at++)
            {
                var rank = profile.Rank(at);
                if (rank.Mode == rankedMode && RankBadge.TryRead(rank.Label, out badge)) break;
            }
            RankBadgeArt.Show(rankBadge, tierBadges, badge);
        }

        void ShowMedals()
        {
            for (var at = 0; at < showcase.Length; at++)
            {
                if (profile == null || at >= profile.MedalCount)
                {
                    showcase[at].Hide();
                    continue;
                }
                var medal = profile.Medal(at);
                showcase[at].Show(stage, medal.StaticAddress, strings.Text(medal.NameKey), true);
            }
        }

        void ShowReport()
        {
            report.gameObject.SetActive(profile != null);
            report.interactable = !reportSent;
            var sending = reporter.Busy;
            for (var at = 0; at < reasons.Length; at++) reasons[at].interactable = !sending && at < reasonPicks.Length;
            closeReasons.interactable = !sending;
        }
    }
}
