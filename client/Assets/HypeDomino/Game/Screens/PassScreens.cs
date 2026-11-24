using HypeDomino.Game.Pass;
using HypeDomino.Game.Shop;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Shop;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class PassScreens
    {
        public const int PremiumOffers = 2;

        public static ScreenView Pass(ScreenHost host)
        {
            var screen = host.PushedScreen("Pass", out var column);
            var level = PlainUi.Text(column, TokenTextStyle.Counter);
            var levelFill = PlainUi.Mark(column).GetComponent<Image>();
            levelFill.color = TokenColor.RolePrimaryFill;
            levelFill.type = Image.Type.Filled;
            levelFill.fillMethod = Image.FillMethod.Horizontal;
            var timeLeft = PlainUi.Text(column, TokenTextStyle.Caption);
            var levels = PlainUi.List(column);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var coinLevel = LevelPurchase(column, host, CurrencyKeys.Coin, out var coinPrice);
            var crystalLevel = LevelPurchase(column, host, CurrencyKeys.Crystal, out var crystalPrice);
            var premium = new ShopOfferRow[PremiumOffers];
            for (var i = 0; i < premium.Length; i++) premium[i] = ShopScreens.OfferRow(column);
            var preview = PlainUi.Column(column);
            var title = PlainUi.Text(preview.transform, TokenTextStyle.Title);
            var price = PlainUi.Text(preview.transform, TokenTextStyle.Counter);
            var currency = PlainUi.Text(preview.transform, TokenTextStyle.Label);
            var confirm = PlainUi.Button(preview.transform, ShopScreens.BuyKey, host.Strings);
            var close = PlainUi.Button(preview.transform, ShopScreens.CloseKey, host.Strings);
            var view = screen.gameObject.AddComponent<PassScreenView>();
            view.Configure(level, levelFill, timeLeft, levels, RowTemplate(screen.transform), message);
            view.ConfigureLevelPurchase(coinLevel, coinPrice, crystalLevel, crystalPrice);
            view.ConfigurePremium(premium, preview, title, price, currency, confirm, close);
            return screen;
        }

        static Button LevelPurchase(Transform column, ScreenHost host, string currencyKey, out TMP_Text price)
        {
            var row = PlainUi.Row(column).transform;
            PlainUi.Text(row, TokenTextStyle.Label, currencyKey, host.Strings);
            var buy = PlainUi.Button(row);
            price = PlainUi.Label(buy);
            return buy;
        }

        static PassLevelRow RowTemplate(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var number = PlainUi.Text(row.transform, TokenTextStyle.Counter);
            var template = row.AddComponent<PassLevelRow>();
            template.Configure(number, Cell(row.transform), Cell(row.transform));
            row.SetActive(false);
            return template;
        }

        static PassCellView Cell(Transform row)
        {
            var cell = PlainUi.Row(row);
            var coins = PlainUi.Text(cell.transform, TokenTextStyle.Caption);
            var crystals = PlainUi.Text(cell.transform, TokenTextStyle.Caption);
            var view = cell.AddComponent<PassCellView>();
            view.Configure(coins, crystals, PlainUi.Mark(cell.transform), PlainUi.Mark(cell.transform), PlainUi.Mark(cell.transform));
            return view;
        }
    }
}
