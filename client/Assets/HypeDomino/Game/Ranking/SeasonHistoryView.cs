using System;
using System.Collections.Generic;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Tokens;
using UnityEngine;

namespace HypeDomino.Game.Ranking
{
    public sealed class SeasonHistoryView : MonoBehaviour, IRowBinder
    {
        public const string TitleKey = "ui.plain.ranking.history";
        public const string LpKey = "ui.plain.ranking.history_lp";
        public const string PositionKey = "ui.plain.ranking.history_position";
        public const string RewardKey = "ui.plain.ranking.history_reward";
        public const string NoRewardKey = "ui.plain.ranking.history_no_reward";
        const string LpArgument = "lp";
        const string PositionArgument = "position";
        const string CoinsArgument = "coins";

        [SerializeField] GameObject historyPanel;
        [SerializeField] RecyclingList list;
        [SerializeField] SeasonHistoryRowSlot rowTemplate;
        [SerializeField] Sprite[] tierBadges;

        readonly Dictionary<string, object> args = new();
        GameStrings strings;
        Action<SeasonHistoryLoadResult> loaded;
        SeasonHistory history;
        string mode;
        PastSeasonMode[] shownModes = Array.Empty<PastSeasonMode>();
        string[] names = Array.Empty<string>();
        string[] lps = Array.Empty<string>();
        string[] positions = Array.Empty<string>();
        string[] rewards = Array.Empty<string>();
        int rowCount;

        public int ShownSeasonCount => rowCount;

        public void Configure(GameObject historyRoot, RecyclingList rows, SeasonHistoryRowSlot template, Sprite[] tierBadgeSprites)
        {
            Unbind();
            historyPanel = historyRoot;
            list = rows;
            rowTemplate = template;
            tierBadges = tierBadgeSprites;
        }

        public void Bind(SeasonHistoryLoader loader, string rankedMode, GameStrings texts)
        {
            Unbind();
            mode = rankedMode;
            strings = texts;
            loaded ??= OnLoaded;
            if (list.RowCount == 0)
            {
                var template = (RectTransform)rowTemplate.transform;
                list.Build(template, template.rect.height, TokenSpacing.Sm);
            }
            strings.Changed += ShowRows;
            ShowRows();
            loader.Load(loaded);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            strings.Changed -= ShowRows;
            strings = null;
        }

        void OnLoaded(SeasonHistoryLoadResult result)
        {
            if (strings == null) return;
            history = result.Outcome == SeasonHistoryLoadOutcome.Loaded ? result.History : null;
            ShowRows();
        }

        void ShowRows()
        {
            var seasons = history?.Count ?? 0;
            Reserve(seasons);
            rowCount = 0;
            for (var index = 0; index < seasons; index++)
            {
                var season = history.SeasonAt(index);
                if (!season.TryFind(mode, out var past)) continue;
                shownModes[rowCount] = past;
                names[rowCount] = season.NameKey.Length > 0 ? strings.Text(season.NameKey) : string.Empty;
                lps[rowCount] = Formatted(LpKey, LpArgument, past.LP);
                positions[rowCount] = Formatted(PositionKey, PositionArgument, past.Position);
                rewards[rowCount] = past.Rewarded ? Formatted(RewardKey, CoinsArgument, past.RewardCoins) : strings.Text(NoRewardKey);
                rowCount++;
            }
            historyPanel.SetActive(rowCount > 0);
            list.Show(rowCount, this);
        }

        void Reserve(int seasons)
        {
            if (shownModes.Length >= seasons) return;
            shownModes = new PastSeasonMode[seasons];
            names = new string[seasons];
            lps = new string[seasons];
            positions = new string[seasons];
            rewards = new string[seasons];
        }

        string Formatted(string key, string argument, long value)
        {
            args.Clear();
            args[argument] = value;
            return strings.Text(key, args);
        }

        public void Bind(RectTransform row, int index) =>
            row.GetComponent<SeasonHistoryRowSlot>().Show(names[index], shownModes[index].Badge, tierBadges, lps[index], positions[index], rewards[index]);
    }
}
