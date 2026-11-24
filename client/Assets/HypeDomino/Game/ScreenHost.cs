using System;
using System.Collections.Generic;
using HypeDomino.Game.Screens;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Screens;
using UnityEngine;

namespace HypeDomino.Game
{
    public sealed class ScreenHost
    {
        public const string BackKey = "ui.plain.back";

        readonly ScreenStack stack;
        readonly Dictionary<AppState, ScreenView> screens = new();
        readonly Transform nav;

        public ScreenHost(ScreenStack stack, GameStrings strings)
        {
            this.stack = stack;
            Strings = strings;
            screens[AppState.Login] = LoginScreens.Login(this);
            screens[AppState.Lobby] = LobbyScreens.Lobby(this, out nav);
            Nav("ui.plain.nav.pass", PassScreens.Pass);
            Nav("ui.plain.nav.promo", PromoScreens.Promo);
            Nav("ui.plain.nav.shop", ShopScreens.Shop);
        }

        public GameStrings Strings { get; }

        public ScreenView ScreenOf(AppState state) => screens.TryGetValue(state, out var screen) ? screen : null;

        public void Follow(AppState state) => stack.Replace(ScreenOf(state));

        public ScreenView Screen(string name, out Transform column) => PlainUi.Screen(stack.transform, name, out column);

        public ScreenView PushedScreen(string name, out Transform column)
        {
            var screen = Screen(name, out column);
            PlainUi.Button(column, BackKey, Strings).onClick.AddListener(Back);
            return screen;
        }

        public void Back() => stack.Pop();

        void Nav(string labelKey, Func<ScreenHost, ScreenView> build)
        {
            var screen = build(this);
            PlainUi.Button(nav, labelKey, Strings).onClick.AddListener(() => stack.Push(screen));
        }
    }
}
