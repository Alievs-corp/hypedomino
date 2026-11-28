using System;
using HypeDomino.Game.Ranking;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class RankingScreens
    {
        public const string NavKey = "ui.plain.nav.ranking";
        public static readonly string SoloKey = LobbyScreens.CardKeys[0];
        public static readonly string TeamsKey = LobbyScreens.CardKeys[1];

        public static ScreenView Ranking(ScreenHost host)
        {
            var screen = host.PushedScreen("Ranking", out var column);
            var modes = PlainUi.Row(column).transform;
            var solo = PlainUi.Button(modes, SoloKey, host.Strings);
            var teams = PlainUi.Button(modes, TeamsKey, host.Strings);
            var rank = Standing(screen, column);
            var leaderboard = screen.gameObject.AddComponent<LeaderboardView>();
            leaderboard.Configure(HalfHeight(PlainUi.List(column)), RowTemplate(screen.transform), Array.Empty<Sprite>());
            var history = History(screen, column, host);
            screen.gameObject.AddComponent<RankingScreenView>().Configure(rank, leaderboard, history, solo, teams);
            return screen;
        }

        static SeasonHistoryView History(ScreenView screen, Transform column, ScreenHost host)
        {
            var panel = PlainUi.Column(column);
            PlainUi.Text(panel.transform, TokenTextStyle.Title, SeasonHistoryView.TitleKey, host.Strings);
            var list = HalfHeight(PlainUi.List(panel.transform));
            var view = screen.gameObject.AddComponent<SeasonHistoryView>();
            view.Configure(panel, list, HistoryRowTemplate(screen.transform), Array.Empty<Sprite>());
            return view;
        }

        static RecyclingList HalfHeight(RecyclingList list)
        {
            var holder = (RectTransform)list.transform;
            holder.sizeDelta = new Vector2(holder.sizeDelta.x, ScreenView.ReferenceShortSide / 2f);
            list.GetComponent<LayoutElement>().preferredHeight = ScreenView.ReferenceShortSide / 2f;
            return list;
        }

        static SeasonHistoryRowSlot HistoryRowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var at = row.transform;
            var seasonName = PlainUi.Text(at, TokenTextStyle.Label);
            var badge = PlainUi.Mark(at).GetComponent<Image>();
            var lp = PlainUi.Text(at, TokenTextStyle.Counter);
            var position = PlainUi.Text(at, TokenTextStyle.Counter);
            var reward = PlainUi.Text(at, TokenTextStyle.Caption);
            var slot = row.AddComponent<SeasonHistoryRowSlot>();
            slot.Configure(seasonName, badge, lp, position, reward);
            row.SetActive(false);
            return slot;
        }

        static RankScreenView Standing(ScreenView screen, Transform column)
        {
            var panel = PlainUi.Column(column);
            var at = panel.transform;
            var seasonName = PlainUi.Text(at, TokenTextStyle.Title);
            var seasonLeft = PlainUi.Text(at, TokenTextStyle.Caption);
            var badge = PlainUi.Mark(at).GetComponent<Image>();
            var lp = PlainUi.Text(at, TokenTextStyle.Counter);
            var lpFill = Fill(at);
            var placementFill = Fill(at);
            var view = screen.gameObject.AddComponent<RankScreenView>();
            view.Configure(panel, badge, Array.Empty<Sprite>(), lpFill.gameObject, lpFill, lp, placementFill.gameObject, placementFill, seasonName, seasonLeft);
            return view;
        }

        static Image Fill(Transform parent)
        {
            var fill = PlainUi.Mark(parent).GetComponent<Image>();
            fill.color = TokenColor.RolePrimaryFill;
            fill.type = Image.Type.Filled;
            fill.fillMethod = Image.FillMethod.Horizontal;
            return fill;
        }

        static LeaderboardRowSlot RowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var at = row.transform;
            var position = PlainUi.Text(at, TokenTextStyle.Counter);
            var nickname = PlainUi.Text(at, TokenTextStyle.Label);
            var badge = PlainUi.Mark(at).GetComponent<Image>();
            var lp = PlainUi.Text(at, TokenTextStyle.Counter);
            var own = PlainUi.Mark(at);
            var slot = row.AddComponent<LeaderboardRowSlot>();
            slot.Configure(position, nickname, badge, lp, own);
            row.SetActive(false);
            return slot;
        }
    }
}
