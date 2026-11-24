using System;
using System.Collections.Generic;
using System.Linq;
using HypeDomino.Game.Inventory;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class CollectionScreenViewTests
    {
        const int Columns = 2;
        const float RowHeight = 100f;
        const float ViewportHeight = 1000f;
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.starter"",""type"":""avatar"",""nameKey"":""cosmetic.avatar.starter.name"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/starter/static"",""sources"":[{""type"":""default""}]},
{""id"":""emoji.smile"",""type"":""emoji"",""nameKey"":""cosmetic.emoji.smile.name"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/emoji/smile/static"",""sources"":[{""type"":""default""}]},
{""id"":""avatar.comet_fox"",""type"":""avatar"",""nameKey"":""cosmetic.avatar.comet_fox.name"",""rarityId"":""rare"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/comet_fox/static"",""sources"":[{""type"":""shop"",""ref"":""offer.avatar_comet_fox""}]},
{""id"":""avatar.moon_owl"",""type"":""avatar"",""nameKey"":""cosmetic.avatar.moon_owl.name"",""rarityId"":""rare"",""availability"":""retired"",""staticAddress"":""cosmetic/avatar/moon_owl/static"",""sources"":[{""type"":""event"",""ref"":""event.moon""}]}]}";
        const string InventoryJson = @"{""items"":[{""id"":""avatar.starter""},{""id"":""emoji.smile""},{""id"":""avatar.moon_owl""}],
""loadout"":{""avatar"":""avatar.starter"",""emojis"":[]},""emoji_slots"":4}";

        sealed class Api : IApiTransport
        {
            public readonly List<string> Methods = new();
            public readonly List<string> Paths = new();
            public readonly List<string> Bodies = new();
            public readonly List<Action<HttpReply>> Replies = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done)
            {
                Methods.Add(method);
                Paths.Add(pathAndQuery);
                Bodies.Add(body);
                Replies.Add(done);
            }
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

        static readonly string[] NoLocales = { };

        GameObject root;
        ScrollRect scroll;
        CollectionScreenView view;
        Api api;
        InventoryLoader loader;
        LoadoutEquipper equipper;
        FakeArtSource<Sprite> stills;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        Button[] filters;
        TMP_Text message;
        GameObject preview;
        TMP_Text previewName;
        GameObject previewLocked;
        GameObject previewEquipped;
        Button equip;
        Button close;

        [SetUp]
        public void BuildView()
        {
            root = new GameObject("Collection screen", typeof(RectTransform), typeof(Canvas));
            var holder = new GameObject("List", typeof(RectTransform), typeof(ScrollRect));
            holder.transform.SetParent(root.transform, false);
            ((RectTransform)holder.transform).sizeDelta = new Vector2(600f, ViewportHeight);
            var viewport = Stretched("Viewport", holder.transform);
            viewport.gameObject.AddComponent<RectMask2D>();
            var content = Stretched("Content", viewport);
            scroll = holder.GetComponent<ScrollRect>();
            scroll.viewport = viewport;
            scroll.content = content;
            scroll.horizontal = false;
            var list = holder.AddComponent<RecyclingList>();

            var template = new GameObject("Row", typeof(RectTransform));
            template.transform.SetParent(root.transform, false);
            ((RectTransform)template.transform).sizeDelta = new Vector2(600f, RowHeight);
            var cells = new CollectionCellSlot[Columns];
            for (var column = 0; column < Columns; column++)
            {
                var cell = Child("Cell " + column, template.transform);
                cells[column] = cell.AddComponent<CollectionCellSlot>();
                cells[column].Configure(Art("Icon", cell.transform), Text("Name", cell.transform), Child("Locked", cell.transform), Child("Equipped", cell.transform), cell.AddComponent<Button>());
            }
            var row = template.AddComponent<CollectionRowSlot>();
            row.Configure(cells);

            filters = new[] { ButtonChild("Avatars"), ButtonChild("Emoji") };
            message = Text("Message", root.transform);
            preview = Child("Preview", root.transform);
            previewName = Text("Preview name", preview.transform);
            previewLocked = Child("Preview locked", preview.transform);
            previewEquipped = Child("Preview equipped", preview.transform);
            equip = ButtonChild("Equip");
            close = ButtonChild("Close");
            view = Child("View", root.transform).AddComponent<CollectionScreenView>();
            view.Configure(filters, new[] { CosmeticAddresses.Avatar, CosmeticAddresses.Emoji }, list, row, message);
            view.ConfigurePreview(preview, Art("Preview art", preview.transform), previewName, previewLocked, previewEquipped, equip, close);

            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            var client = new ApiClient(api, session);
            loader = new InventoryLoader(client);
            equipper = new LoadoutEquipper(client);
            stills = new FakeArtSource<Sprite>();
            stage = new CosmeticStage(stills, new FakeArtSource<LoopArt>(), null, null, null, EffectLevel.Medium, false);
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var catalog));
            manifest = new ActiveCosmeticManifest(catalog);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            view.Bind(loader, equipper, stage, manifest, strings);
        }

        [TearDown]
        public void DestroyView() => UnityEngine.Object.Destroy(root);

        static GameObject Child(string name, Transform parent)
        {
            var child = new GameObject(name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child;
        }

        Button ButtonChild(string name) => Child(name, root.transform).AddComponent<Button>();

        static TMP_Text Text(string name, Transform parent) => Child(name, parent).AddComponent<TextMeshProUGUI>();

        static CosmeticImage Art(string name, Transform parent)
        {
            var art = new GameObject(name, typeof(RectTransform), typeof(Image), typeof(AnimatedImage)).AddComponent<CosmeticImage>();
            art.transform.SetParent(parent, false);
            return art;
        }

        static RectTransform Stretched(string name, Transform parent)
        {
            var rect = new GameObject(name, typeof(RectTransform)).GetComponent<RectTransform>();
            rect.SetParent(parent, false);
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.sizeDelta = Vector2.zero;
            return rect;
        }

        void Answer(int status, string body) => api.Replies[api.Replies.Count - 1](new HttpReply(status, body));

        List<Transform> ShownCells() =>
            scroll.content.Cast<Transform>().Where(row => row.gameObject.activeSelf).OrderBy(row => -((RectTransform)row).anchoredPosition.y)
                .SelectMany(row => row.Cast<Transform>()).Where(cell => cell.gameObject.activeSelf).ToList();

        string[] ShownNames() => ShownCells().Select(cell => cell.Find("Name").GetComponent<TMP_Text>().text).ToArray();

        bool[] Shown(string look) => ShownCells().Select(cell => cell.Find(look).gameObject.activeSelf).ToArray();

        void PickCell(int at) => ShownCells()[at].GetComponent<Button>().onClick.Invoke();

        [Test]
        public void NothingIsShownUntilTheServerInventoryArrives()
        {
            Assert.AreEqual(CosmeticCollection.InventoryPath, api.Paths[0]);
            Assert.AreEqual(0, ShownCells().Count);
            Assert.IsFalse(preview.activeSelf);
        }

        [Test]
        public void TheFirstFilterShowsItsItemsInManifestOrderWithUnownedItemsLocked()
        {
            Answer(200, InventoryJson);

            Assert.AreEqual(new[] { strings.Text("cosmetic.avatar.starter.name"), strings.Text("cosmetic.avatar.comet_fox.name"), strings.Text("cosmetic.avatar.moon_owl.name") }, ShownNames());
            Assert.AreEqual(new[] { false, true, false }, Shown("Locked"));
            Assert.AreEqual(new[] { true, false, false }, Shown("Equipped"));
            Assert.IsFalse(filters[0].interactable);
            Assert.IsTrue(filters[1].interactable);
        }

        [Test]
        public void AnotherFilterShowsItsOwnType()
        {
            Answer(200, InventoryJson);

            filters[1].onClick.Invoke();

            Assert.AreEqual(new[] { strings.Text("cosmetic.emoji.smile.name") }, ShownNames());
            Assert.IsTrue(filters[0].interactable);
        }

        [Test]
        public void EachCellShowsTheStillOfItsItem()
        {
            Answer(200, InventoryJson);

            CollectionAssert.AreEquivalent(new[] { "cosmetic/avatar/starter/static", "cosmetic/avatar/comet_fox/static", "cosmetic/avatar/moon_owl/static" }, stills.Loads);
        }

        [Test]
        public void AnUnownedItemPreviewsLockedAndCannotBeEquipped()
        {
            Answer(200, InventoryJson);

            PickCell(1);

            Assert.IsTrue(preview.activeSelf);
            Assert.AreEqual(strings.Text("cosmetic.avatar.comet_fox.name"), previewName.text);
            Assert.IsTrue(previewLocked.activeSelf);
            Assert.IsFalse(equip.interactable);
        }

        [Test]
        public void EquippingSendsTheIntentAndShowsWhatTheServerConfirms()
        {
            Answer(200, InventoryJson);
            PickCell(2);
            Assert.IsTrue(equip.interactable);

            equip.onClick.Invoke();

            Assert.AreEqual("PUT", api.Methods[1]);
            Assert.AreEqual(CosmeticCollection.EquipPath, api.Paths[1]);
            Assert.AreEqual(@"{""slot"":""avatar"",""items"":[""avatar.moon_owl""]}", api.Bodies[1]);
            Assert.IsFalse(equip.interactable);
            Assert.AreEqual(new[] { true, false, false }, Shown("Equipped"));

            Answer(200, @"{""loadout"":{""avatar"":""avatar.moon_owl""}}");

            Assert.AreEqual(new[] { false, false, true }, Shown("Equipped"));
            Assert.IsTrue(previewEquipped.activeSelf);
            Assert.IsFalse(equip.interactable);
            Assert.IsFalse(message.enabled);
        }

        [Test]
        public void ARefusedEquipShowsTheMessageAndReloadsTheInventory()
        {
            Answer(200, InventoryJson);
            PickCell(2);
            equip.onClick.Invoke();

            Answer(403, "{\"error\":\"item.not_owned\"}");

            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(EquipKeys.NotOwned), message.text);
            Assert.AreEqual(3, api.Paths.Count);
            Assert.AreEqual(CosmeticCollection.InventoryPath, api.Paths[2]);
            Assert.AreEqual(new[] { true, false, false }, Shown("Equipped"));
        }

        [Test]
        public void AFailedLoadShowsTheMessageAndNoItems()
        {
            Answer(500, "{}");

            Assert.AreEqual(0, ShownCells().Count);
            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(InventoryLoadKeys.Failed), message.text);
        }

        [Test]
        public void ClosingThePreviewHidesIt()
        {
            Answer(200, InventoryJson);
            PickCell(0);

            close.onClick.Invoke();

            Assert.IsFalse(preview.activeSelf);
        }
    }
}
