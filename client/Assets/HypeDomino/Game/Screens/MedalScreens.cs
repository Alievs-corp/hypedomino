using HypeDomino.Game.Medals;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;

namespace HypeDomino.Game.Screens
{
    public static class MedalScreens
    {
        public const string NavKey = "ui.plain.nav.medals";
        public const int Columns = 3;

        public static ScreenView Medals(ScreenHost host)
        {
            var screen = host.PushedScreen("Medals", out var column);
            screen.gameObject.AddComponent<MedalGridView>().Configure(PlainUi.List(column), RowTemplate(screen.transform));
            return screen;
        }

        static MedalRowSlot RowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var cells = new MedalCellSlot[Columns];
            for (var column = 0; column < cells.Length; column++) cells[column] = Cell(row.transform);
            var template = row.AddComponent<MedalRowSlot>();
            template.Configure(cells);
            row.SetActive(false);
            return template;
        }

        static MedalCellSlot Cell(Transform row)
        {
            var cell = PlainUi.Row(row);
            var icon = PlainUi.Icon(cell.transform);
            var medalName = PlainUi.Text(cell.transform, TokenTextStyle.Caption);
            var locked = PlainUi.Mark(cell.transform);
            var slot = cell.AddComponent<MedalCellSlot>();
            slot.Configure(icon, medalName, locked);
            return slot;
        }
    }
}
