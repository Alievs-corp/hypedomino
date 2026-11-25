using HypeDomino.Game.Promo;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;

namespace HypeDomino.Game.Screens
{
    public static class PromoScreens
    {
        public const string NavKey = "ui.plain.nav.promo";

        public static ScreenView Promo(ScreenHost host)
        {
            var screen = host.PushedScreen("Promo", out var column);
            var field = PlainUi.Input(column);
            var paste = PlainUi.Button(column, "ui.plain.promo.paste", host.Strings);
            var redeem = PlainUi.Button(column, "ui.plain.promo.redeem", host.Strings);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            screen.gameObject.AddComponent<PromoCodeView>().Configure(field, paste, redeem, message);
            return screen;
        }
    }
}
