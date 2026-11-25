using HypeDomino.Game.Missions;
using HypeDomino.UI.Missions;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class MissionScreens
    {
        public const string NavKey = "ui.plain.nav.missions";

        public static readonly string[] TabKeys =
        {
            "ui.plain.missions.daily",
            "ui.plain.missions.weekly",
            "ui.plain.missions.achievement",
        };

        public static ScreenView Missions(ScreenHost host)
        {
            var screen = host.PushedScreen("Missions", out var column);
            var view = screen.gameObject.AddComponent<MissionListView>();
            var tabs = PlainUi.Row(column).transform;
            for (var tab = 0; tab < TabKeys.Length; tab++) Tab(tabs, host, view, (MissionTab)tab);
            view.Configure(PlainUi.List(column), RowTemplate(screen.transform));
            return screen;
        }

        static void Tab(Transform tabs, ScreenHost host, MissionListView view, MissionTab tab) =>
            PlainUi.Button(tabs, TabKeys[(int)tab], host.Strings).onClick.AddListener(() => view.ShowTab(tab));

        static MissionRowSlot RowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var missionName = PlainUi.Text(row.transform, TokenTextStyle.Label);
            var progress = PlainUi.Text(row.transform, TokenTextStyle.Counter);
            var target = PlainUi.Text(row.transform, TokenTextStyle.Counter);
            var fill = PlainUi.Mark(row.transform).GetComponent<Image>();
            fill.color = TokenColor.RolePrimaryFill;
            fill.type = Image.Type.Filled;
            fill.fillMethod = Image.FillMethod.Horizontal;
            var slot = row.AddComponent<MissionRowSlot>();
            slot.Configure(missionName, progress, target, fill, PlainUi.Mark(row.transform));
            row.SetActive(false);
            return slot;
        }
    }
}
