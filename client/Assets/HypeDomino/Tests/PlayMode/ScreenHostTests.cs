using System.Collections;
using HypeDomino.Content;
using HypeDomino.Game;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Inventory;
using HypeDomino.Game.Lobby;
using HypeDomino.Game.Login;
using HypeDomino.Game.Medals;
using HypeDomino.Game.Missions;
using HypeDomino.Game.Pass;
using HypeDomino.Game.Privacy;
using HypeDomino.Game.Profile;
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
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.TestTools;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScreenHostTests
    {
        const string StoreUrl = "https://store.test/app";
        const int NavButtons = 9;
        const int LoginButtons = 3;
        const int SettingsLabels = 18;
        const int RankingLabels = 3;
        const int RoomEntryButtons = 2;
        const int RoomEntryLabels = 3;
        const int InvitePromptButtons = 2;
        const int InvitePromptLabels = 2;
        const int RoomLabels = 20;
        const int ProfileLabels = 9;
        static readonly string[] NoLocales = { };
        static readonly int[] Panel120 = { 60, 120 };

        sealed class SizedStore : ICacheStore
        {
            long bytes;

            public int Reads { get; private set; }

            public long OccupiedBytes
            {
                get
                {
                    Reads++;
                    return bytes;
                }
                set => bytes = value;
            }

            public bool ClearAtStart { get; set; }

            public void ApplyLimits(long limitBytes, int unusedForSeconds)
            {
            }

            public bool TryClear()
            {
                bytes = 0;
                return true;
            }
        }

        GameObject root;
        ScreenStack stack;
        ScreenHost host;
        LanguageChoice choice;
        GameStrings strings;
        SettingsModel settings;
        SizedStore cacheStore;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Host");
            stack = root.AddComponent<ScreenStack>();
            stack.enabled = false;
            cacheStore = new SizedStore();
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
                new ContentCache(cacheStore, () => false));
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
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons + RoomEntryButtons + InvitePromptButtons, lobby.GetComponentsInChildren<Button>(true).Length);
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
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons + RoomEntryLabels + InvitePromptLabels, lobby.Length);
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
        public void BuildingTheHostReadsNoCacheSize()
        {
            Assert.AreEqual(0, cacheStore.Reads);
        }

        [UnityTest]
        public IEnumerator TheSettingsScreenReadsTheCacheSizeOncePerOpeningAndAfterAClear()
        {
            cacheStore.OccupiedBytes = 12_400_000;
            var screen = Pushed(SettingsScreens.NavKey);
            yield return null;
            yield return null;
            Assert.AreEqual(1, cacheStore.Reads);
            Assert.IsTrue(HasText(screen, "Keş: 12,4 MB"));

            stack.Pop();
            stack.Advance(TokenDurationMs.Base);
            Assert.IsFalse(screen.IsShown);
            yield return null;
            cacheStore.OccupiedBytes = 3_000_000;
            NavButton(SettingsScreens.NavKey).onClick.Invoke();
            yield return null;
            Assert.AreEqual(2, cacheStore.Reads);
            Assert.IsTrue(HasText(screen, "Keş: 3,0 MB"));

            ButtonOf(screen, SettingsScreens.ClearCacheKey).onClick.Invoke();
            Assert.AreEqual(3, cacheStore.Reads);
            Assert.IsTrue(HasText(screen, "Keş: 0,0 MB"));

            choice.Choose(GameLanguage.English);
            Assert.AreEqual(3, cacheStore.Reads);
            Assert.IsTrue(HasText(screen, "Cache: 0.0 MB"));
        }

        [Test]
        public void ASettledSettingsScreenFrameAllocatesNothing()
        {
            var view = Pushed(SettingsScreens.NavKey).GetComponent<SettingsScreenView>();
            view.RefreshWhenShown();

            Assert.That(view.RefreshWhenShown, Allocates.Nothing);
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

        [Test]
        public void TheProfileScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = Pushed(ProfileScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<ProfileScreenView>());
            Assert.IsNotNull(screen.GetComponent<ProfileSummaryView>());
            Assert.IsNotNull(screen.GetComponentInChildren<AvatarView>(true));
            Assert.IsNotNull(screen.GetComponentInChildren<TMP_InputField>(true).textComponent);
            Assert.IsNotNull(screen.GetComponentInChildren<ScrollRect>(true));
            var showcase = screen.GetComponentsInChildren<MedalCellSlot>(true);
            Assert.AreEqual(ProfileScreens.ShowcaseCells, showcase.Length);
            foreach (var cell in showcase) Assert.IsFalse(cell.gameObject.activeSelf);
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(ProfileLabels, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
            AssertShowsItsString(TextOf(screen, ProfileScreens.LevelKey));
            foreach (var key in ProfileScreens.StatKeys) AssertShowsItsString(TextOf(screen, key));
            AssertShowsItsString(TextOf(screen, ProfileScreens.SaveNicknameKey));
        }

        [Test]
        public void TheProfileNavButtonPushesItAndBackPopsIt()
        {
            var lobby = host.ScreenOf(AppState.Lobby);

            var screen = Pushed(ProfileScreens.NavKey);
            Assert.AreNotSame(lobby, screen);
            Assert.AreEqual(2, stack.Count);
            Assert.IsTrue(screen.TakesInput);
            Assert.IsFalse(lobby.TakesInput);

            ButtonOf(screen, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
            Assert.IsFalse(screen.TakesInput);
        }

        [Test]
        public void TheMissionsScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = Pushed(MissionScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<MissionListView>());
            Assert.IsNotNull(screen.GetComponentInChildren<RecyclingList>(true));
            var row = screen.GetComponentInChildren<MissionRowSlot>(true);
            Assert.IsNotNull(row);
            Assert.IsFalse(row.gameObject.activeSelf);
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(MissionScreens.TabKeys.Length + 1, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
        }

        [Test]
        public void TheMedalsScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = Pushed(MedalScreens.NavKey);

            Assert.IsNotNull(screen.GetComponent<MedalGridView>());
            Assert.IsNotNull(screen.GetComponentInChildren<RecyclingList>(true));
            var row = screen.GetComponentInChildren<MedalRowSlot>(true);
            Assert.IsNotNull(row);
            Assert.IsFalse(row.gameObject.activeSelf);
            Assert.AreEqual(MedalScreens.Columns, row.Columns);
            Assert.AreEqual(MedalScreens.Columns, row.GetComponentsInChildren<CosmeticImage>(true).Length);
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(1, labels.Length);
            foreach (var label in labels) AssertShowsItsString(label);
        }

        [TestCase(MissionScreens.NavKey)]
        [TestCase(MedalScreens.NavKey)]
        public void TheMissionsAndMedalsNavButtonsPushTheirScreenAndBackPopsIt(string navKey)
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            var screen = Pushed(navKey);

            Assert.AreNotSame(lobby, screen);
            Assert.AreEqual(2, stack.Count);
            Assert.IsTrue(screen.TakesInput);
            Assert.IsFalse(lobby.TakesInput);

            ButtonOf(screen, ScreenHost.BackKey).onClick.Invoke();
            Assert.AreSame(lobby, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(lobby.TakesInput);
        }

        [Test]
        public void TheUpdateAndMaintenanceStatesShowTheirScreens()
        {
            var update = host.ScreenOf(AppState.UpdateRequired);
            var maintenance = host.ScreenOf(AppState.Maintenance);
            Assert.IsFalse(update.IsShown);
            Assert.IsFalse(maintenance.IsShown);

            host.Follow(AppState.UpdateRequired);
            stack.Advance(TokenDurationMs.Slow);
            Assert.AreSame(update, stack.Top);
            Assert.IsTrue(update.IsShown);
            Assert.IsTrue(update.TakesInput);

            host.Follow(AppState.Maintenance);
            Assert.AreSame(maintenance, stack.Top);
            Assert.AreEqual(1, stack.Count);
            Assert.IsTrue(maintenance.TakesInput);
            Assert.IsFalse(update.TakesInput);
            stack.Advance(TokenDurationMs.Slow);
            Assert.IsTrue(maintenance.IsShown);
            Assert.IsFalse(update.IsShown);
        }

        [Test]
        public void TheUpdateRequiredScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = host.ScreenOf(AppState.UpdateRequired);

            Assert.IsNotNull(screen.GetComponent<UpdateRequiredView>());
            Assert.AreEqual(1, screen.GetComponentsInChildren<Button>(true).Length);
            AssertEveryLabelShowsItsString(screen, NoticeScreens.UpdateMessageKey, NoticeScreens.UpdateKey);
        }

        [Test]
        public void TheBoundUpdateRequiredScreenShowsItsPanelAndOpensTheStoreLink()
        {
            var screen = host.ScreenOf(AppState.UpdateRequired);
            var flow = new AppFlow();
            string opened = null;
            screen.GetComponent<UpdateRequiredView>().Bind(flow, StoreUrl, link => opened = link);
            var panel = TextOf(screen, NoticeScreens.UpdateMessageKey).transform.parent.gameObject;
            Assert.IsFalse(panel.activeSelf);

            flow.Hello(new ServerHello { Status = HelloStatus.UpdateRequired });
            ButtonOf(screen, NoticeScreens.UpdateKey).onClick.Invoke();

            Assert.IsTrue(panel.activeSelf);
            Assert.AreEqual(StoreUrl, opened);
        }

        [Test]
        public void TheMaintenanceScreenBuildsWithItsViewAndEveryLabelShowsItsString()
        {
            var screen = host.ScreenOf(AppState.Maintenance);

            Assert.IsNotNull(screen.GetComponent<MaintenanceScreenView>());
            Assert.AreEqual(1, screen.GetComponentsInChildren<Button>(true).Length);
            AssertEveryLabelShowsItsString(screen, NoticeScreens.MaintenanceMessageKey, NoticeScreens.RetryKey);
        }

        [Test]
        public void TheBoundMaintenanceScreenShowsItsPanelAndRetries()
        {
            var screen = host.ScreenOf(AppState.Maintenance);
            var dials = 0;
            var maintenance = new MaintenanceScreen(() => dials++, () => 0);
            screen.GetComponent<MaintenanceScreenView>().Bind(maintenance, strings);
            var panel = TextOf(screen, NoticeScreens.MaintenanceMessageKey).transform.parent.gameObject;
            var retry = ButtonOf(screen, NoticeScreens.RetryKey);
            Assert.IsFalse(panel.activeSelf);

            maintenance.Read(new ServerHello { Status = HelloStatus.Maintenance });
            retry.onClick.Invoke();

            Assert.IsTrue(panel.activeSelf);
            Assert.AreEqual(1, dials);
            Assert.IsFalse(retry.interactable);
        }

        [Test]
        public void TheBoundMaintenanceScreenShowsTheServerEndTimeInUtcOnlyWhenSent()
        {
            const ulong endsMs = 1_800_000_060_000;
            var screen = host.ScreenOf(AppState.Maintenance);
            var maintenance = new MaintenanceScreen(() => { }, () => 0);
            screen.GetComponent<MaintenanceScreenView>().Bind(maintenance, strings);
            var shown = strings.Text(MaintenanceScreenView.EndsKey, new System.Collections.Generic.Dictionary<string, object> { [MaintenanceScreenView.EndArgument] = "15.01.2027 08:01" });

            maintenance.Read(new ServerHello { Status = HelloStatus.Maintenance, ServerTimeUnixMs = endsMs - 60_000, MaintenanceEndsUnixMs = endsMs });
            var ends = System.Array.Find(screen.GetComponentsInChildren<TMP_Text>(true), text => text.text == shown);

            Assert.IsNotNull(ends);
            Assert.IsTrue(ends.gameObject.activeSelf);
            StringAssert.Contains("UTC", shown);

            maintenance.Read(new ServerHello { Status = HelloStatus.Maintenance, ServerTimeUnixMs = endsMs });

            Assert.IsFalse(ends.gameObject.activeSelf);
        }

        void AssertEveryLabelShowsItsString(ScreenView screen, params string[] keys)
        {
            var labels = screen.GetComponentsInChildren<PlainLabel>(true);
            Assert.AreEqual(keys.Length, labels.Length);
            foreach (var key in keys) AssertShowsItsString(LabelOf(screen, key));
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
        public void TheLobbyHasTheInvitePromptHiddenUntilAnInviteArrives()
        {
            var lobby = host.ScreenOf(AppState.Lobby);
            var accept = ButtonOf(lobby, RoomScreens.AcceptInviteKey);
            var decline = ButtonOf(lobby, RoomScreens.DeclineInviteKey);

            Assert.IsNotNull(lobby.GetComponent<InvitePromptView>());
            AssertShowsItsString(LabelOf(lobby, RoomScreens.AcceptInviteKey));
            AssertShowsItsString(LabelOf(lobby, RoomScreens.DeclineInviteKey));
            Assert.AreSame(accept.transform.parent, decline.transform.parent);
            Assert.IsFalse(accept.transform.parent.parent.gameObject.activeSelf);
        }

        [Test]
        public void TheRoomEmojiWheelStartsWithoutSlotsAndKeepsItsSlotTemplateHidden()
        {
            var room = host.ScreenOf(AppState.Room);
            var wheel = ButtonOf(room, RoomScreens.EmojiKey).transform.parent;

            Assert.AreEqual(1, wheel.GetComponentsInChildren<Button>(true).Length);
            var template = System.Array.Find(room.GetComponentsInChildren<Button>(true), button => button.transform.parent == room.transform);
            Assert.IsNotNull(template);
            Assert.IsFalse(template.gameObject.activeSelf);
            Assert.IsNotNull(template.GetComponentInChildren<CosmeticImage>(true));
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

        static bool HasText(ScreenView screen, string text) =>
            System.Array.Exists(screen.GetComponentsInChildren<TMP_Text>(true), shown => shown.text == text);

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
