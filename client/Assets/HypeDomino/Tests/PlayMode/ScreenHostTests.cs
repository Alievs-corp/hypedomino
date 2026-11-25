using HypeDomino.Game;
using HypeDomino.Game.Lobby;
using HypeDomino.Game.Login;
using HypeDomino.Game.Pass;
using HypeDomino.Game.Promo;
using HypeDomino.Game.Screens;
using HypeDomino.Game.Shop;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScreenHostTests
    {
        const int NavButtons = 3;
        const int LoginButtons = 3;
        static readonly string[] NoLocales = { };

        GameObject root;
        ScreenStack stack;
        ScreenHost host;
        LanguageChoice choice;
        GameStrings strings;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Host");
            stack = root.AddComponent<ScreenStack>();
            stack.enabled = false;
            choice = LanguageChoice.Load("az", () => NoLocales);
            strings = new GameStrings(BuiltinStrings.Load(), choice, true);
            host = new ScreenHost(stack, strings);
        }

        [TearDown]
        public void DestroyHost() => Object.Destroy(root);

        [Test]
        public void TheHostBuildsTheLoginAndLobbyScreensHidden()
        {
            var login = host.ScreenOf(AppState.Login);
            var lobby = host.ScreenOf(AppState.Lobby);

            Assert.IsNotNull(login.GetComponent<LoginScreenView>());
            Assert.IsNotNull(lobby.GetComponent<LobbyScreenView>());
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons, lobby.GetComponentsInChildren<Button>(true).Length);
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
            Assert.AreEqual(PlayModes.Count + 2 + NavButtons, lobby.Length);
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
