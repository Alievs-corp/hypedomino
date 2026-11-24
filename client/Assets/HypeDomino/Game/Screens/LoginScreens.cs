using HypeDomino.Game.Login;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;

namespace HypeDomino.Game.Screens
{
    public static class LoginScreens
    {
        public static ScreenView Login(ScreenHost host)
        {
            var screen = host.Screen(nameof(AppState.Login), out var column);
            var google = PlainUi.Button(column, "ui.plain.login.google", host.Strings);
            var apple = PlainUi.Button(column, "ui.plain.login.apple", host.Strings);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var deletionPending = PlainUi.Column(column);
            var cancelDeletion = PlainUi.Button(deletionPending.transform, "ui.plain.login.cancel_deletion", host.Strings);
            screen.gameObject.AddComponent<LoginScreenView>().Configure(google, apple, message, deletionPending, cancelDeletion);
            return screen;
        }
    }
}
