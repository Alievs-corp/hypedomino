using HypeDomino.Game.Inventory;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class CollectionScreens
    {
        public const string NavKey = "ui.plain.nav.collection";
        public const string EquipKey = "ui.plain.collection.equip";
        public const string LockedKey = "ui.plain.collection.locked";
        public const string EquippedKey = "ui.plain.collection.equipped";
        public const int Columns = 3;
        const int ArtHeight = TokenSize.TouchTargetMin * 2;

        public static readonly string[] FilterTypes =
        {
            CosmeticAddresses.Avatar,
            CosmeticAddresses.Frame,
            CosmeticAddresses.Tile,
            CosmeticAddresses.Table,
            CosmeticAddresses.Emoji,
        };

        public static readonly string[] FilterKeys =
        {
            "ui.plain.collection.avatar",
            "ui.plain.collection.frame",
            "ui.plain.collection.tile",
            "ui.plain.collection.table",
            "ui.plain.collection.emoji",
        };

        public static ScreenView Collection(ScreenHost host)
        {
            var screen = host.PushedScreen("Collection", out var column);
            var tabs = PlainUi.Row(column).transform;
            var filters = new Button[FilterKeys.Length];
            for (var i = 0; i < filters.Length; i++) filters[i] = PlainUi.Button(tabs, FilterKeys[i], host.Strings);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var preview = PlainUi.Column(column);
            var art = Art(preview.transform, ArtHeight);
            var title = PlainUi.Text(preview.transform, TokenTextStyle.Title);
            var locked = PlainUi.Text(preview.transform, TokenTextStyle.Caption, LockedKey, host.Strings).gameObject;
            var equipped = PlainUi.Text(preview.transform, TokenTextStyle.Caption, EquippedKey, host.Strings).gameObject;
            var equip = PlainUi.Button(preview.transform, EquipKey, host.Strings);
            var close = PlainUi.Button(preview.transform, ShopScreens.CloseKey, host.Strings);
            preview.SetActive(false);
            var grid = PlainUi.List(column);
            var view = screen.gameObject.AddComponent<CollectionScreenView>();
            view.Configure(filters, FilterTypes, grid, RowTemplate(screen.transform), message);
            view.ConfigurePreview(preview, art, title, locked, equipped, equip, close);
            return screen;
        }

        static CollectionRowSlot RowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            ((RectTransform)row.transform).sizeDelta = new Vector2(0f, ArtHeight + TokenSize.TouchTargetMin);
            var cells = new CollectionCellSlot[Columns];
            for (var i = 0; i < cells.Length; i++) cells[i] = Cell(row.transform);
            var template = row.AddComponent<CollectionRowSlot>();
            template.Configure(cells);
            row.SetActive(false);
            return template;
        }

        static CollectionCellSlot Cell(Transform row)
        {
            var select = PlainUi.Button(row);
            select.targetGraphic.color = TokenColor.SurfaceDefault;
            var stack = select.gameObject.AddComponent<VerticalLayoutGroup>();
            stack.spacing = TokenSpacing.Xs;
            stack.childControlWidth = true;
            stack.childControlHeight = true;
            stack.childForceExpandHeight = false;
            var art = Art(select.transform, 0);
            var name = PlainUi.Text(select.transform, TokenTextStyle.Caption);
            var locked = PlainUi.Mark(select.transform);
            var equipped = PlainUi.Mark(select.transform);
            equipped.GetComponent<Image>().color = TokenColor.RoleSuccessFill;
            var cell = select.gameObject.AddComponent<CollectionCellSlot>();
            cell.Configure(art, name, locked, equipped, select);
            return cell;
        }

        static CosmeticImage Art(Transform parent, int height)
        {
            var holder = new GameObject("Art", typeof(RectTransform), typeof(Image), typeof(AnimatedImage), typeof(LayoutElement));
            holder.transform.SetParent(parent, false);
            var layout = holder.GetComponent<LayoutElement>();
            layout.preferredHeight = height;
            layout.flexibleHeight = 1f;
            var art = holder.AddComponent<CosmeticImage>();
            var image = holder.GetComponent<Image>();
            image.preserveAspect = true;
            image.raycastTarget = false;
            image.enabled = false;
            return art;
        }
    }
}
