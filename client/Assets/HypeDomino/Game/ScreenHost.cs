using System;
using System.Collections.Generic;
using HypeDomino.Game.Analytics;
using HypeDomino.Game.Screens;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Screens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game
{
    public sealed class ScreenHost
    {
        public const string BackKey = "ui.plain.back";

        readonly ScreenStack stack;
        readonly Dictionary<AppState, ScreenView> screens = new();
        readonly Dictionary<ScreenView, string> screenIds = new();
        readonly Transform nav;

        public ScreenHost(ScreenStack stack, GameStrings strings, SettingsModel settings)
        {
            this.stack = stack;
            Strings = strings;
            stack.TopShown += TopShown;
            screens[AppState.Login] = LoginScreens.Login(this);
            screens[AppState.Lobby] = LobbyScreens.Lobby(this, out nav);
            screens[AppState.Room] = RoomScreens.Room(this);
            screens[AppState.UpdateRequired] = NoticeScreens.UpdateRequired(this);
            screens[AppState.Maintenance] = NoticeScreens.Maintenance(this);
            PrivacyNotice = FirstLaunchScreens.PrivacyNotice(this);
            Promo = PromoScreens.Promo(this);
            Nav("ui.plain.nav.pass", PassScreens.Pass(this));
            Nav(PromoScreens.NavKey, Promo);
            Nav("ui.plain.nav.shop", ShopScreens.Shop(this));
            Nav(SettingsScreens.NavKey, SettingsScreens.Settings(this, settings));
            Nav(RankingScreens.NavKey, RankingScreens.Ranking(this));
            Nav(CollectionScreens.NavKey, CollectionScreens.Collection(this));
            Nav(ProfileScreens.NavKey, ProfileScreens.Profile(this));
            Nav(MissionScreens.NavKey, MissionScreens.Missions(this));
            Nav(MedalScreens.NavKey, MedalScreens.Medals(this));
        }

        public event Action<string> ScreenShown;

        public GameStrings Strings { get; }

        public ScreenView Promo { get; }

        public ScreenView PrivacyNotice { get; }

        public ScreenView ScreenOf(AppState state) => screens.TryGetValue(state, out var screen) ? screen : null;

        public void Follow(AppState state) => stack.Replace(ScreenOf(state));

        public string ScreenIdOf(ScreenView screen) => screenIds.TryGetValue(screen, out var id) ? id : null;

        public ScreenView Screen(string name, out Transform column)
        {
            var screen = PlainUi.Screen(stack.transform, name, out column);
            screenIds[screen] = ScreenViewIds.Of(name);
            return screen;
        }

        public ScreenView PushedScreen(string name, out Transform column)
        {
            var screen = Screen(name, out column);
            PlainUi.Button(column, BackKey, Strings).onClick.AddListener(Back);
            return screen;
        }

        public void Back() => stack.Pop();

        public void Open(ScreenView screen) => stack.Push(screen);

        public Button PushButton(Transform parent, string labelKey, ScreenView screen)
        {
            var button = PlainUi.Button(parent, labelKey, Strings);
            button.onClick.AddListener(() => Open(screen));
            return button;
        }

        void Nav(string labelKey, ScreenView screen) => PushButton(nav, labelKey, screen);

        void TopShown(ScreenView screen)
        {
            if (screenIds.TryGetValue(screen, out var id)) ScreenShown?.Invoke(id);
        }
    }
}
