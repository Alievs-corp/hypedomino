using System.Collections.Generic;
using HypeDomino.Game.Lobby;
using HypeDomino.Game.Login;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game
{
    public sealed class ScreenHost
    {
        readonly ScreenStack stack;
        readonly Dictionary<AppState, ScreenView> screens = new();

        public ScreenHost(ScreenStack stack)
        {
            this.stack = stack;
            screens[AppState.Login] = BuildLogin();
            screens[AppState.Lobby] = BuildLobby();
        }

        public ScreenView ScreenOf(AppState state) => screens.TryGetValue(state, out var screen) ? screen : null;

        public void Follow(AppState state) => stack.Replace(ScreenOf(state));

        ScreenView BuildLogin()
        {
            var screen = PlainUi.Screen(stack.transform, nameof(AppState.Login), out var column);
            var google = PlainUi.Button(column);
            var apple = PlainUi.Button(column);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var deletionPending = PlainUi.Column(column);
            var cancelDeletion = PlainUi.Button(deletionPending.transform);
            screen.gameObject.AddComponent<LoginScreenView>().Configure(google, apple, message, deletionPending, cancelDeletion);
            return screen;
        }

        ScreenView BuildLobby()
        {
            var screen = PlainUi.Screen(stack.transform, nameof(AppState.Lobby), out var column);
            var coins = PlainUi.Text(column, TokenTextStyle.Counter);
            var crystals = PlainUi.Text(column, TokenTextStyle.Counter);
            var level = PlainUi.Text(column, TokenTextStyle.Counter);
            var retryHeader = PlainUi.Button(column);
            var cards = new Button[PlayModes.Count];
            var busy = new GameObject[PlayModes.Count];
            for (var card = 0; card < cards.Length; card++)
            {
                cards[card] = PlainUi.Button(column);
                busy[card] = PlainUi.Mark(column);
            }
            var cancel = PlainUi.Button(column);
            var countdown = PlainUi.Text(column, TokenTextStyle.Body);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            screen.gameObject.AddComponent<LobbyScreenView>().Configure(cards, busy, cancel, countdown, message, coins, crystals, level, retryHeader);
            return screen;
        }
    }
}
