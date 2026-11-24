using HypeDomino.Game.Shop;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class ShopScreens
    {
        public const string BuyKey = "ui.plain.shop.buy";
        public const string CloseKey = "ui.plain.close";

        public static ScreenView Shop(ScreenHost host)
        {
            var screen = host.PushedScreen("Shop", out var column);
            var coins = PlainUi.Text(column, TokenTextStyle.Counter);
            var crystals = PlainUi.Text(column, TokenTextStyle.Counter);
            var tabs = PlainUi.Row(column).transform;
            var categories = new Button[ShopSectionKeys.Count];
            var titles = new TMP_Text[ShopSectionKeys.Count];
            for (var i = 0; i < categories.Length; i++)
            {
                categories[i] = PlainUi.Button(tabs);
                titles[i] = PlainUi.Label(categories[i]);
            }
            var timeLeft = PlainUi.Text(column, TokenTextStyle.Caption);
            var offers = PlainUi.List(column);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var preview = PlainUi.Column(column);
            var title = PlainUi.Text(preview.transform, TokenTextStyle.Title);
            var price = PlainUi.Text(preview.transform, TokenTextStyle.Counter);
            var currency = PlainUi.Text(preview.transform, TokenTextStyle.Label);
            var previewTimeLeft = PlainUi.Text(preview.transform, TokenTextStyle.Caption);
            var confirm = PlainUi.Button(preview.transform, BuyKey, host.Strings);
            var close = PlainUi.Button(preview.transform, CloseKey, host.Strings);
            var view = screen.gameObject.AddComponent<ShopScreenView>();
            view.Configure(categories, titles, timeLeft, offers, RowTemplate(screen.transform), coins, crystals, message);
            view.ConfigurePreview(preview, title, price, currency, previewTimeLeft, confirm, close);
            return screen;
        }

        public static ShopOfferRow OfferRow(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var open = PlainUi.Button(row.transform);
            var title = PlainUi.Label(open);
            var price = PlainUi.Text(row.transform, TokenTextStyle.Counter);
            var currency = PlainUi.Text(row.transform, TokenTextStyle.Label);
            var offer = row.AddComponent<ShopOfferRow>();
            offer.Configure(title, price, currency, open);
            return offer;
        }

        static ShopOfferRow RowTemplate(Transform parent)
        {
            var template = OfferRow(parent);
            template.gameObject.SetActive(false);
            return template;
        }
    }
}
