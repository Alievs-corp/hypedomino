using System;
using System.Collections.Generic;
using HypeDomino.Core.Analytics;
using HypeDomino.Game.Medals;
using HypeDomino.Game.Profile;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Medals;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Profile;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ProfileScreenViewTests
    {
        const string FoxStatic = "cosmetic/avatar/fox/static";
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""}]}";
        const string OwnJson = @"{""account_id"":""acc-1"",""nickname"":""Leyla"",""setup_required"":false,""rename_required"":false,""level"":7,""xp"":1250,
""loadout"":{""avatar"":""avatar.fox""},""stats"":{""games"":1234,""wins"":22,""losses"":18,""win_rate_percent"":55,""streak"":3,""best_streak"":6,""modes"":{}}}";
        const string RenamedJson = @"{""account_id"":""acc-1"",""nickname"":""Leyla07"",""setup_required"":false,""rename_required"":false,""level"":7,""xp"":1250,
""loadout"":{""avatar"":""avatar.fox""},""stats"":{""games"":1234,""wins"":22,""losses"":18,""win_rate_percent"":55,""streak"":3,""best_streak"":6,""modes"":{}}}";
        const string MedalsJson = @"{""version"":7,""slots"":3,""medals"":[
{""id"":""medal.first_win"",""nameKey"":""medal.first_win.name"",""rarityId"":""common"",""staticAddress"":""medal/first_win/static""}],
""owned"":[{""medal"":""medal.first_win"",""earned_at"":""2026-11-01T10:00:00Z""}],
""showcase"":[""medal.first_win""]}";

        sealed class Sent
        {
            public string Method;
            public string Path;
            public string Body;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { Method = method, Path = pathAndQuery, Body = body, Done = done });

            public Sent To(string path) => Requests.FindLast(request => request.Path == path);
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

        sealed class AnalyticsSink : IAnalyticsSender
        {
            public void Send(string analyticsId, IReadOnlyList<AnalyticsEvent> batch, Action<bool> done) => done(true);
        }

        static readonly string[] NoLocales = { };

        GameObject root;
        AvatarView avatar;
        TMP_Text nickname;
        TMP_Text level;
        TMP_Text[] stats;
        MedalCellSlot[] showcase;
        TMP_InputField nicknameInput;
        Button save;
        TMP_Text message;
        ProfileScreenView view;
        Api api;
        ApiClient client;
        AnalyticsQueue events;
        FakeArtSource<Sprite> stills;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;

        [SetUp]
        public void BuildView()
        {
            root = new GameObject("Profile screen", typeof(RectTransform), typeof(Canvas));
            avatar = Child("Avatar", root.transform).AddComponent<AvatarView>();
            avatar.Configure(NewArt("Art", avatar.transform), NewArt("Frame", avatar.transform), null, null);
            nickname = Text("Nickname", root.transform);
            level = Text("Level", root.transform);
            stats = new TMP_Text[(int)ProfileStat.BestStreak + 1];
            for (var at = 0; at < stats.Length; at++) stats[at] = Text("Stat " + at, root.transform);
            var summary = Child("Summary", root.transform).AddComponent<ProfileSummaryView>();
            summary.Configure(avatar, nickname, level, stats);
            showcase = new MedalCellSlot[2];
            for (var at = 0; at < showcase.Length; at++)
            {
                var cell = Child("Cell " + at, root.transform);
                showcase[at] = cell.AddComponent<MedalCellSlot>();
                showcase[at].Configure(NewArt("Icon", cell.transform), Text("Name", cell.transform), Child("Locked", cell.transform));
            }
            nicknameInput = Child("Nickname input", root.transform).AddComponent<TMP_InputField>();
            save = Child("Save", root.transform).AddComponent<Image>().gameObject.AddComponent<Button>();
            message = Text("Message", root.transform);
            view = Child("View", root.transform).AddComponent<ProfileScreenView>();
            view.Configure(summary, showcase, nicknameInput, save, message);

            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            client = new ApiClient(api, session);
            events = new AnalyticsQueue(new AnalyticsSink(), () => 1700000000000, 10, 10);
            stills = new FakeArtSource<Sprite>();
            stage = new CosmeticStage(stills, new FakeArtSource<LoopArt>(), null, null, null, EffectLevel.Medium, false);
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var catalog));
            manifest = new ActiveCosmeticManifest(catalog);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            view.Bind(new ProfileLoader(client), new MedalLoader(client), new NicknameSaver(client, events), stage, manifest, strings);
        }

        [TearDown]
        public void DestroyView() => UnityEngine.Object.Destroy(root);

        static GameObject Child(string name, Transform parent)
        {
            var child = new GameObject(name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child;
        }

        static TMP_Text Text(string name, Transform parent) => Child(name, parent).AddComponent<TextMeshProUGUI>();

        static CosmeticImage NewArt(string name, Transform parent)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<AnimatedImage>().gameObject.AddComponent<CosmeticImage>();
            image.transform.SetParent(parent, false);
            return image;
        }

        void Answer(string path, int status, string body) => api.To(path).Done(new HttpReply(status, body));

        [Test]
        public void NothingShowsUntilTheServerAnswers()
        {
            Assert.IsNotNull(api.To(PlayerProfile.OwnPath));
            Assert.IsNotNull(api.To(MedalCollection.CollectionPath));
            Assert.IsFalse(avatar.gameObject.activeSelf);
            Assert.IsFalse(nickname.enabled);
            Assert.IsFalse(stats[0].enabled);
            Assert.IsFalse(save.interactable);
            Assert.IsFalse(message.enabled);
            Assert.IsFalse(showcase[0].gameObject.activeSelf);
        }

        [Test]
        public void TheOwnProfileShowsTheServersNicknameLevelStatsAndAvatar()
        {
            Answer(PlayerProfile.OwnPath, 200, OwnJson);

            Assert.IsTrue(avatar.gameObject.activeSelf);
            Assert.AreEqual("Leyla", nickname.text);
            Assert.AreEqual("7", level.text);
            Assert.AreEqual("1.234", stats[(int)ProfileStat.Games].text);
            Assert.AreEqual("55", stats[(int)ProfileStat.WinRatePercent].text);
            Assert.AreEqual("6", stats[(int)ProfileStat.BestStreak].text);
            Assert.IsTrue(stats[(int)ProfileStat.Losses].enabled);
            CollectionAssert.Contains(stills.Loads, FoxStatic);
            Assert.IsTrue(save.interactable);
        }

        [Test]
        public void TheShowcaseShowsTheChosenMedalsOnly()
        {
            Answer(MedalCollection.CollectionPath, 200, MedalsJson);

            Assert.IsTrue(showcase[0].gameObject.activeSelf);
            Assert.AreEqual("[!medal.first_win.name]", showcase[0].transform.Find("Name").GetComponent<TMP_Text>().text);
            Assert.IsFalse(showcase[0].transform.Find("Locked").gameObject.activeSelf);
            Assert.IsFalse(showcase[1].gameObject.activeSelf);
        }

        [Test]
        public void SavingANicknameShowsTheNameTheServerReturns()
        {
            Answer(PlayerProfile.OwnPath, 200, OwnJson);
            nicknameInput.text = "Leyla07";

            save.onClick.Invoke();

            var sent = api.To(NicknameRename.Path);
            Assert.AreEqual("PUT", sent.Method);
            Assert.AreEqual(@"{""nickname"":""Leyla07""}", sent.Body);
            Assert.IsFalse(save.interactable);
            Assert.AreEqual("Leyla", nickname.text);

            sent.Done(new HttpReply(200, RenamedJson));

            Assert.AreEqual("Leyla07", nickname.text);
            Assert.IsTrue(save.interactable);
            Assert.IsFalse(message.enabled);
        }

        [Test]
        public void ARefusedNicknameShowsTheServersReason()
        {
            Answer(PlayerProfile.OwnPath, 200, OwnJson);
            nicknameInput.text = "Taken";
            save.onClick.Invoke();

            api.To(NicknameRename.Path).Done(new HttpReply(409, @"{""error"":""nickname.taken""}"));

            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(NicknameSaveKeys.Taken), message.text);
            Assert.AreEqual("Leyla", nickname.text);
        }

        [Test]
        public void AFailedLoadShowsTheLoadMessage()
        {
            Answer(PlayerProfile.OwnPath, 429, "{}");

            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(ProfileLoadKeys.RateLimited), message.text);
            Assert.IsFalse(avatar.gameObject.activeSelf);
            Assert.IsFalse(save.interactable);
        }
    }
}
