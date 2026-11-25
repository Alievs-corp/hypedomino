using HypeDomino.Content;
using HypeDomino.Game;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Inventory;
using HypeDomino.Game.Lobby;
using HypeDomino.Game.Login;
using HypeDomino.Game.Pass;
using HypeDomino.Game.Privacy;
using HypeDomino.Game.Promo;
using HypeDomino.Game.Ranking;
using HypeDomino.Game.Screens;
using HypeDomino.Game.Shop;
using HypeDomino.UI.Audio;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Settings;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScreenHostTests
    {
        const int NavButtons = 6;
        const int LoginButtons = 3;
        const int SettingsLabels = 18;
        const int RankingLabels = 3;
        const int RoomEntryButtons = 2;
        const int RoomEntryLabels = 3;
        const int RoomLabels = 20;
        static readonly string[] NoLocales = { };
        static readonly int[] Panel120 = { 60, 120 };

        GameObject root;
        ScreenStack stack;
        ScreenHost host;
        LanguageChoice choice;
        GameStrings strings;
        SettingsModel settings;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Host");
            stack = root.AddComponent<ScreenStack>();
            stack.enabled = false;
            choice = LanguageChoice.Load("az", () => NoLocales);
            strings = new GameStrings(BuiltinStrings.Load(), choice, true);
            var graphics = new GraphicsQuality(DeviceTier.High, GraphicsLevel.High, 0, false);
            settings = new SettingsModel(
                choice,
                new AudioVolumes(AudioVolumes.Full, AudioVolumes.Full, AudioVolumes.Full),
                new MotionPreference(null, false),
                graphics,
                new FrameRatePolicy(graphics, Panel120, null, false),
                new HapticsPreference(true),
                new ContentCache(new NoCacheStore(), () => false));
            host = new ScreenHost(stack, strings, settings);
        }

        [TearDown]
        public void DestroyHost()
        {
            Object.Destroy(root);
            settings.Dispose();
        }

        [Test]
        public void TheHostBuildsTheLoginAndLobbyScreensHidden()
        {
            var login = host.ScreenOf(AppState.Login);
            var lobby = host.ScreenOf(AppState.Lobby);

            Assert.IsNotNull(login.GetComponent<LoginScreenView>());
            Assert.IsNotNull(lobby.GetComponent<LobbyScreenView>());
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons + RoomEntryButtons, lobby.GetComponentsInChildren<Button>(true).Length);
            Assert.IsFalse(login.IsShown);
            Assert.IsFalse(lobby.IsShown);
            Assert.IsNull(host.ScreenOf(AppState.Match));
            Assert.AreEqual(0, stack.Count);
        }

        [Test]
        public void TheShownScreenFollowsTheAppState()
        {
            var login = host.ScreenOf(AppState.Login);
            var lobby = host.ScreenOf(AppState.Lobby);

            host.Follow(AppState.Login);
            stack.Advance(TokenDurationMs.Slow);
            Assert.AreSame(login, stack.Top);
            Assert.IsTrue(login.IsShown);
            Assert.IsTrue(login.TakesInput);

            host.Follow(AppState.Lobby);
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
            Assert.IsFalse(login.TakesInput);
            stack.Advance(TokenDurationMs.Slow);
            Assert.IsTrue(lobby.IsShown);
            Assert.IsFalse(login.IsShown);

            host.Follow(AppState.Match);
            Assert.AreEqual(0, stack.Count);
            Assert.IsFalse(lobby.TakesInput);
            stack.Advance(TokenDurationMs.Base);
            Assert.IsFalse(lobby.IsShown);
        }

        [Test]
        public void EveryLoginAndLobbyButtonShowsItsGameString()
        {
            var login = host.ScreenOf(AppState.Login).GetComponentsInChildren<PlainLabel>(true);
            var lobby = host.ScreenOf(AppState.Lobby).GetComponentsInChildren<PlainLabel>(true);

            Assert.AreEqual(LoginButtons, login.Length);
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons + RoomEntryLabels, lobby.Length);
            foreach (var label in login) AssertShowsItsString(label);
            foreach (var label in lobby) AssertShowsItsString(label);
        }

        [Test]
        public void ALabelFollowsALanguageChange()
        {
            var back = TextOf(Pushed("ui.plain.nav.shop"), ScreenHost.BackKey);
            Assert.AreEqual("Geri", back.text);

            choice.Choose(GameLanguage.English);

            Assert.AreEqual("Back", back.text);
        }

        [Test]
        public void ANavButtonPushesItsScreenAndBackPopsIt()
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            host.Follow(AppState.Lobby);

            NavButton("ui.plain.nav.shop").onClick.Invoke();
            var shop = stack.Top;
            Assert.AreNotSame(lobby, shop);
            Assert.AreEqual(2, stack.Count);
            Assert.IsTrue(shop.TakesInput);
            Assert.IsFalse(lobby.TakesInput);

            ButtonOf(shop, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
            Assert.IsFalse(shop.TakesInput);
        }

        [Test]
        public void AStateChangeLeavesAPushedScreen()
        {
            host.Follow(AppState.Lobby);
            NavButton("ui.plain.nav.pass").onClick.Invoke();
            var pass = stack.Top;

            host.Follow(AppState.Login);

            Assert.AreEqual(1, stack.Count);
            Assert.AreSame(host.ScreenOf(AppState.Login), stack.Top);
            Assert.IsFalse(pass.TakesInput);
        }

        [Test]
        public void TheShopScreenBuildsWithItsView()
        {
            var shop = Pushed("ui.plain.nav.shop");

            Assert.IsNotNull(shop.GetComponent<ShopScreenView>());
            Assert.IsNotNull(shop.GetComponentInChildren<RecyclingList>(true));
            Assert.IsNotNull(shop.GetComponentInChildren<ShopOfferRow>(true));
            Assert.IsFalse(shop.GetComponentInChildren<ShopOfferRow>(true).gameObject.activeSelf);
            Assert.AreEqual(ShopSectionKeys.Count + 5, shop.GetComponentsInChildren<Button>(true).Length);
            AssertShowsItsString(TextOf(shop, "ui.plain.shop.buy"));
            AssertShowsItsString(TextOf(shop, "ui.plain.close"));
            AssertShowsItsString(TextOf(shop, PromoScreens.NavKey));
        }

        [Test]
        public void TheShopOpensTheSamePromoScreenAsTheLobbyAndBackReturnsToTheShop()
        {
            var shop = Pushed("ui.plain.nav.shop");

            ButtonOf(shop, PromoScreens.NavKey).onClick.Invoke();
            Assert.AreSame(host.Promo, stack.Top);
            Assert.AreEqual(3, stack.Count);
            Assert.IsTrue(host.Promo.TakesInput);
            Assert.IsFalse(shop.TakesInput);

            ButtonOf(host.Promo, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(shop, stack.Top);
            Assert.IsTrue(shop.TakesInput);

            ButtonOf(shop, ScreenHost.BackKey).onClick.Invoke();
            NavButton(PromoScreens.NavKey).onClick.Invoke();
            Assert.AreSame(host.Promo, stack.Top);
            Assert.AreEqual(2, stack.Count);
        }

        [Test]
        public void ThePassScreenBuildsWithItsView()
        {
            var pass = Pushed("ui.plain.nav.pass");

            Assert.IsNotNull(pass.GetComponent<PassScreenView>());
            Assert.IsNotNull(pass.GetComponentInChildren<RecyclingList>(true));
            Assert.IsNotNull(pass.GetComponentInChildren<PassLevelRow>(true));
            Assert.AreEqual(2, pass.GetComponentsInChildren<PassCellView>(true).Length);
            Assert.AreEqual(2, pass.GetComponentsInChildren<ShopOfferRow>(true).Length);
            AssertShowsItsString(TextOf(pass, "currency.coin.name"));
            AssertShowsItsString(TextOf(pass, "currency.crystal.name"));
        }

        [Test]
        public void ThePromoScreenBuildsWithItsView()
        {
            var promo = Pushed("ui.plain.nav.promo");

            Assert.IsNotNull(promo.GetComponent<PromoCodeView>());
            var field = promo.GetComponentInChildren<TMP_InputField>(true);
            Assert.IsNotNull(field.textComponent);
            Assert.IsNotNull(field.textViewport);
            AssertShowsItsString(TextOf(promo, "ui.plain.promo.paste"));
            AssertShowsItsString(TextOf(promo, "ui.plain.promo.redeem"));
        }

        [Test]
        public void TheSettingsScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = Pushed(SettingsScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<SettingsScreenView>());
            Assert.IsNotNull(screen.GetComponentInChildren<LanguageRow>(true));
            Assert.AreEqual(SettingsScreens.VolumeKeys.Length, screen.GetComponentsInChildren<VolumeSliderRow>(true).Length);
            Assert.AreEqual(4, screen.GetComponentsInChildren<SettingToggleRow>(true).Length);
            Assert.AreEqual(2, screen.GetComponentsInChildren<TMP_Dropdown>(true).Length);
            Assert.IsNotNull(screen.GetComponentInChildren<ScrollRect>(true));
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(SettingsLabels, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
        }

        [Test]
        public void TheSettingsRowsStartAtTheModelAndChangeIt()
        {
            var screen = Pushed(SettingsScreens.NavKey);
            var haptics = ToggleOf(screen, SettingsScreens.HapticsKey);
            var music = SliderOf(screen, SettingsScreens.VolumeKeys[(int)AudioGroup.Music]);

            Assert.IsTrue(haptics.isOn);
            Assert.AreEqual(AudioVolumes.Full, music.value);

            haptics.isOn = false;
            music.value = 0f;

            Assert.IsFalse(settings.Haptics.Enabled);
            Assert.AreEqual(0f, settings.Volumes.Level(AudioGroup.Music));
        }

        [Test]
        public void ClearingTheCacheShowsItsOutcome()
        {
            var screen = Pushed(SettingsScreens.NavKey);

            ButtonOf(screen, SettingsScreens.ClearCacheKey).onClick.Invoke();

            var cleared = strings.Text(CacheClearKeys.Cleared);
            Assert.IsTrue(System.Array.Exists(screen.GetComponentsInChildren<TMP_Text>(true), text => text.text == cleared));
        }

        [Test]
        public void SettingsOpensTheSamePromoScreenAndBackReturnsToSettings()
        {
            var screen = Pushed(SettingsScreens.NavKey);

            ButtonOf(screen, PromoScreens.NavKey).onClick.Invoke();
            Assert.AreSame(host.Promo, stack.Top);
            Assert.AreEqual(3, stack.Count);
            Assert.IsFalse(screen.TakesInput);

            ButtonOf(host.Promo, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(screen, stack.Top);
            Assert.IsTrue(screen.TakesInput);
        }

        [Test]
        public void TheSettingsScreenHasTheAccountRowsWithTheWarningClosed()
        {
            var screen = Pushed(SettingsScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<AccountSettingsView>());
            var warning = TextOf(screen, SettingsScreens.DeleteWarningKey).transform.parent.gameObject;
            Assert.IsFalse(warning.activeSelf);
            Assert.AreSame(warning.transform, ButtonOf(screen, SettingsScreens.DeleteWithGoogleKey).transform.parent);
            Assert.AreSame(warning.transform, ButtonOf(screen, SettingsScreens.DeleteWithAppleKey).transform.parent);
            Assert.AreSame(warning.transform, ButtonOf(screen, SettingsScreens.KeepAccountKey).transform.parent);
            Assert.AreNotSame(warning.transform, ButtonOf(screen, SettingsScreens.ExportDataKey).transform.parent);
            Assert.AreNotSame(warning.transform, ButtonOf(screen, SettingsScreens.DeleteAccountKey).transform.parent);
        }

        [Test]
        public void TheRankingScreenBuildsWithItsViewsAndEveryLabelShowsItsString()
        {
            var screen = Pushed(RankingScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<RankingScreenView>());
            Assert.IsNotNull(screen.GetComponent<RankScreenView>());
            Assert.IsNotNull(screen.GetComponent<LeaderboardView>());
            Assert.IsNotNull(screen.GetComponentInChildren<RecyclingList>(true));
            Assert.IsFalse(screen.GetComponentInChildren<LeaderboardRowSlot>(true).gameObject.activeSelf);
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(RankingLabels, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
            Assert.IsNotNull(ButtonOf(screen, RankingScreens.SoloKey));
            Assert.IsNotNull(ButtonOf(screen, RankingScreens.TeamsKey));
        }

        [Test]
        public void TheRankingNavButtonPushesItsScreenAndBackPopsIt()
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            var ranking = Pushed(RankingScreens.NavKey);

            Assert.AreNotSame(lobby, ranking);
            Assert.AreEqual(2, stack.Count);
            Assert.IsTrue(ranking.TakesInput);
            Assert.IsFalse(lobby.TakesInput);

            ButtonOf(ranking, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
            Assert.IsFalse(ranking.TakesInput);
        }

        [Test]
        public void TheLobbyHasTheRoomEntry()
        {
            var lobby = host.ScreenOf(AppState.Lobby);

            Assert.IsNotNull(lobby.GetComponent<RoomEntryView>());
            Assert.IsNotNull(lobby.GetComponentInChildren<TMP_InputField>(true));
            Assert.IsNotNull(ButtonOf(lobby, RoomScreens.CreateKey));
            Assert.IsNotNull(ButtonOf(lobby, RoomScreens.JoinKey));
            AssertShowsItsString(TextOf(lobby, RoomScreens.CodeKey));
        }

        [Test]
        public void TheRoomScreenBuildsWithItsViewsAndEveryLabelShowsItsString()
        {
            var room = host.ScreenOf(AppState.Room);

            Assert.IsNotNull(room.GetComponent<FriendRoomView>());
            Assert.IsNotNull(room.GetComponent<InvitePanelView>());
            Assert.IsNotNull(room.GetComponentInChildren<RecyclingList>(true));
            Assert.IsFalse(room.GetComponentInChildren<FriendInviteRow>(true).gameObject.activeSelf);
            Assert.AreEqual(FriendRoom.Capacity + 2 * RoomScreens.TeamSeats + RoomScreens.WaitingSeats, room.GetComponentsInChildren<RoomMemberCell>(true).Length);
            Assert.AreEqual((int)FriendInviteState.RoomClosed + 1, RoomScreens.FriendStateKeys.Length);
            Assert.IsFalse(room.IsShown);
            var labels = room.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(RoomLabels, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
        }

        [Test]
        public void TheCollectionScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = Pushed(CollectionScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<CollectionScreenView>());
            Assert.IsNotNull(screen.GetComponentInChildren<RecyclingList>(true));
            var template = screen.GetComponentInChildren<CollectionRowSlot>(true);
            Assert.IsFalse(template.gameObject.activeSelf);
            Assert.AreEqual(CollectionScreens.Columns, template.Columns);
            Assert.AreEqual(CollectionScreens.Columns + 1, screen.GetComponentsInChildren<CosmeticImage>(true).Length);
            Assert.IsFalse(ButtonOf(screen, CollectionScreens.EquipKey).transform.parent.gameObject.activeSelf);
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(CollectionScreens.FilterKeys.Length + 5, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
        }

        [Test]
        public void TheRoomStateShowsTheRoomScreen()
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            var room = host.ScreenOf(AppState.Room);
            host.Follow(AppState.Lobby);
            stack.Advance(TokenDurationMs.Slow);

            host.Follow(AppState.Room);

            Assert.AreSame(room, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(room.TakesInput);
            Assert.IsFalse(lobby.TakesInput);
            stack.Advance(TokenDurationMs.Slow);
            Assert.IsTrue(room.IsShown);
            Assert.IsFalse(lobby.IsShown);
        }

        [Test]
        public void TheCollectionNavButtonPushesItAndBackPopsIt()
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            var screen = Pushed(CollectionScreens.NavKey);

            Assert.AreNotSame(lobby, screen);
            Assert.AreEqual(2, stack.Count);
            Assert.IsTrue(screen.TakesInput);

            ButtonOf(screen, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsFalse(screen.TakesInput);
        }

        static Toggle ToggleOf(ScreenView screen, string key) => LabelOf(screen, key).transform.parent.GetComponentInChildren<Toggle>(true);

        static Slider SliderOf(ScreenView screen, string key) => LabelOf(screen, key).transform.parent.GetComponentInChildren<Slider>(true);

        ScreenView Pushed(string navKey)
        {
            host.Follow(AppState.Lobby);
            NavButton(navKey).onClick.Invoke();
            return stack.Top;
        }

        Button NavButton(string key) => ButtonOf(host.ScreenOf(AppState.Lobby), key);

        static Button ButtonOf(ScreenView screen, string key) => LabelOf(screen, key).GetComponent<Button>();

        static TMP_Text TextOf(ScreenView screen, string key) => LabelOf(screen, key).GetComponentInChildren<TMP_Text>(true);

        static PlainLabel LabelOf(ScreenView screen, string key)
        {
            foreach (var label in screen.GetComponentsInChildren<PlainLabel>(true))
                if (label.Key == key)
                    return label;
            Assert.Fail("No label " + key);
            return null;
        }

        void AssertShowsItsString(PlainLabel label) => AssertShowsItsString(label.GetComponentInChildren<TMP_Text>(true));

        void AssertShowsItsString(TMP_Text text)
        {
            var key = text.GetComponentInParent<PlainLabel>(true).Key;
            Assert.AreEqual(strings.Text(key), text.text);
            StringAssert.DoesNotStartWith("[!", text.text);
        }
    }
}
