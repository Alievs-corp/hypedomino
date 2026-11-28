using System;
using System.Collections.Generic;
using HypeDomino.Content;
using HypeDomino.Game;
using HypeDomino.Game.Screens;
using HypeDomino.Game.Shop;
using HypeDomino.Game.Tutorial;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.Net.Integrity;
using HypeDomino.UI.Audio;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Settings;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ShopPurchaseTapsTests
    {
        const int PurchaseTapLimit = 3;
        const string ShopNavKey = "ui.plain.nav.shop";
        const string CoinOfferId = "offer.table_night_market";
        const string CoinOfferNameKey = "shop.offer.table_night_market";
        const string StoreOfferId = "offer.starter";
        const string StoreOfferNameKey = "shop.offer.starter";
        const string StoreProductId = "offer_starter";
        const string ShopJson = @"{""version"":12,""sections"":[
{""section"":""featured"",""offers"":[{""id"":""offer.table_night_market"",""kind"":""cosmetic"",""name_key"":""shop.offer.table_night_market"",""price_coins"":3000,""cosmetics"":[""table.night_market""],""limit"":1,""bought"":0,""owned"":[],""compensation"":{},""purchasable"":true},{""id"":""offer.starter"",""kind"":""bundle"",""name_key"":""shop.offer.starter"",""product_id"":""offer_starter"",""crystals"":300,""limit"":1,""bought"":0,""owned"":[],""compensation"":{},""purchasable"":true}]}]}";
        const string WalletJson = "{\"coins\":3200,\"crystals\":40}";
        const string StarterCheckoutJson = "{\"offer\":\"offer.starter\",\"product_id\":\"offer_starter\",\"account_token\":\"account-token-test\"}";
        static readonly string[] NoLocales = { };
        static readonly int[] Panel120 = { 60, 120 };

        sealed class Sent
        {
            public string PathAndQuery;
            public string Body;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { PathAndQuery = pathAndQuery, Body = body, Done = done });

            public void Answer(int status, string body) => Requests[Requests.Count - 1].Done(new HttpReply(status, body));
        }

        sealed class Auth : IAuthTransport
        {
            public readonly List<Action<HttpReply>> Refreshes = new();

            public void Post(string path, string json, Action<HttpReply> done) => Refreshes.Add(done);
        }

        sealed class Identity : IIdentityProvider
        {
            public bool Supports(AuthProvider provider) => false;

            public void RequestIdToken(AuthProvider provider, string nonce, Action<IdTokenResult> done) => throw new NotSupportedException();
        }

        sealed class Store : ISecureStore
        {
            readonly Dictionary<string, string> values = new();

            public string Read(string key) => values.TryGetValue(key, out var value) ? value : null;

            public void Write(string key, string value) => values[key] = value;

            public void Delete(string key) => values.Remove(key);
        }

        sealed class Billing : IStoreBilling
        {
            public readonly List<string> Products = new();

            public bool CanPurchase => true;

            public void Purchase(string productId, string accountToken, Action<StoreBillingResult> done) => Products.Add(productId);

            public void Finish(StoreReceipt receipt, bool finishedByServer)
            {
            }
        }

        static void NoProof(IntegrityPurpose purpose, Action<IntegrityProof> done) => done(null);

        GameObject root;
        ScreenStack stack;
        ScreenHost host;
        GameStrings strings;
        SettingsModel settings;
        Api api;
        Billing billing;
        ApiClient client;
        int tapsInTheShop;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Host");
            stack = root.AddComponent<ScreenStack>();
            stack.enabled = false;
            var choice = LanguageChoice.Load("az", () => NoLocales);
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
            host = new ScreenHost(stack, strings, settings, new TutorialOffer());
            api = new Api();
            billing = new Billing();
            client = SignedInClient(api);
        }

        [TearDown]
        public void DestroyHost()
        {
            Object.Destroy(root);
            settings.Dispose();
        }

        static ApiClient SignedInClient(Api transport)
        {
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            return new ApiClient(transport, session);
        }

        [Test]
        public void ACoinOfferReachesTheCheckoutInFewerThanThreeTapsFromTheShop()
        {
            var shop = OpenTheShopFromTheLobby();

            Tap(shop, OfferButton(shop, CoinOfferNameKey));
            Assert.AreEqual(2, api.Requests.Count);
            Tap(shop, BuyButton(shop));

            Assert.AreEqual(3, api.Requests.Count);
            Assert.AreEqual(ShopBuyer.Path, api.Requests[2].PathAndQuery);
            Assert.AreEqual(CoinOfferId, JsonUtility.FromJson<ShopBuyRequest>(api.Requests[2].Body).offer);
            Assert.Less(tapsInTheShop, PurchaseTapLimit);
        }

        [Test]
        public void AStoreOfferReachesTheStoreInFewerThanThreeTapsFromTheShop()
        {
            var shop = OpenTheShopFromTheLobby();

            Tap(shop, OfferButton(shop, StoreOfferNameKey));
            Tap(shop, BuyButton(shop));

            Assert.AreEqual(3, api.Requests.Count);
            Assert.AreEqual(ShopPurchase.BeginPath, api.Requests[2].PathAndQuery);
            Assert.AreEqual(StoreOfferId, JsonUtility.FromJson<ShopBeginRequest>(api.Requests[2].Body).offer);
            api.Answer(200, StarterCheckoutJson);
            CollectionAssert.AreEqual(new[] { StoreProductId }, billing.Products);
            Assert.Less(tapsInTheShop, PurchaseTapLimit);
        }

        ScreenView OpenTheShopFromTheLobby()
        {
            host.Follow(AppState.Lobby);
            var lobby = host.ScreenOf(AppState.Lobby);
            Assert.IsTrue(lobby.TakesInput);
            ButtonOf(lobby, ShopNavKey).onClick.Invoke();
            var shop = stack.Top;
            Assert.AreNotSame(lobby, shop);
            var checkout = new ShopCheckout(new ShopBuyer(client), new ShopPurchase(client, billing, NoProof));
            shop.GetComponent<ShopScreenView>().Bind(new ShopLoader(client), checkout, new ServerClock(), strings);
            api.Answer(200, ShopJson);
            api.Answer(200, WalletJson);
            Assert.AreEqual(2, api.Requests.Count);
            tapsInTheShop = 0;
            return shop;
        }

        void Tap(ScreenView screen, Button button)
        {
            Assert.IsTrue(screen.TakesInput);
            Assert.IsTrue(button.gameObject.activeInHierarchy);
            Assert.IsTrue(button.interactable);
            button.onClick.Invoke();
            tapsInTheShop++;
        }

        Button OfferButton(ScreenView shop, string nameKey)
        {
            var title = strings.Text(nameKey);
            foreach (var row in shop.GetComponentsInChildren<ShopOfferRow>())
            {
                var button = row.GetComponentInChildren<Button>();
                if (button.GetComponentInChildren<TMP_Text>().text == title) return button;
            }
            Assert.Fail("No shown offer " + nameKey);
            return null;
        }

        static Button BuyButton(ScreenView shop) => ButtonOf(shop, ShopScreens.BuyKey);

        static Button ButtonOf(ScreenView screen, string key)
        {
            foreach (var label in screen.GetComponentsInChildren<PlainLabel>(true))
                if (label.Key == key)
                    return label.GetComponent<Button>();
            Assert.Fail("No button " + key);
            return null;
        }
    }
}
