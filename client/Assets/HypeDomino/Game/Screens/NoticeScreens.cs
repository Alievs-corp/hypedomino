using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;

namespace HypeDomino.Game.Screens
{
    public static class NoticeScreens
    {
        public const string UpdateMessageKey = "ui.plain.update.message";
        public const string UpdateKey = "ui.plain.update.update";
        public const string MaintenanceMessageKey = "ui.plain.maintenance.message";
        public const string RetryKey = "ui.plain.lobby.retry";

        public static ScreenView UpdateRequired(ScreenHost host)
        {
            var screen = host.Screen(nameof(AppState.UpdateRequired), out var column);
            var panel = PlainUi.Column(column);
            PlainUi.Text(panel.transform, TokenTextStyle.Body, UpdateMessageKey, host.Strings);
            var update = PlainUi.Button(panel.transform, UpdateKey, host.Strings);
            screen.gameObject.AddComponent<UpdateRequiredView>().Configure(panel, update);
            return screen;
        }

        public static ScreenView Maintenance(ScreenHost host)
        {
            var screen = host.Screen(nameof(AppState.Maintenance), out var column);
            var panel = PlainUi.Column(column);
            PlainUi.Text(panel.transform, TokenTextStyle.Body, MaintenanceMessageKey, host.Strings);
            var retry = PlainUi.Button(panel.transform, RetryKey, host.Strings);
            screen.gameObject.AddComponent<MaintenanceScreenView>().Configure(panel, retry);
            return screen;
        }
    }
}
