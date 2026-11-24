using HypeDomino.Game.Lobby;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class LobbyScreens
    {
        public static readonly string[] CardKeys =
        {
            "ui.plain.lobby.ranked_1v1",
            "ui.plain.lobby.ranked_2v2",
            "ui.plain.lobby.casual_1v1",
            "ui.plain.lobby.casual_2v2",
        };

        public static ScreenView Lobby(ScreenHost host, out Transform nav)
        {
            var screen = host.Screen(nameof(AppState.Lobby), out var column);
            var coins = PlainUi.Text(column, TokenTextStyle.Counter);
            var crystals = PlainUi.Text(column, TokenTextStyle.Counter);
            var level = PlainUi.Text(column, TokenTextStyle.Counter);
            var retryHeader = PlainUi.Button(column, "ui.plain.lobby.retry", host.Strings);
            var cards = new Button[PlayModes.Count];
            var busy = new GameObject[PlayModes.Count];
            for (var card = 0; card < cards.Length; card++)
            {
                cards[card] = PlainUi.Button(column, CardKeys[card], host.Strings);
                busy[card] = PlainUi.Mark(column);
            }
            var cancel = PlainUi.Button(column, "ui.plain.lobby.cancel", host.Strings);
            var countdown = PlainUi.Text(column, TokenTextStyle.Body);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            nav = PlainUi.Column(column).transform;
            screen.gameObject.AddComponent<LobbyScreenView>().Configure(cards, busy, cancel, countdown, message, coins, crystals, level, retryHeader);
            return screen;
        }
    }
}
