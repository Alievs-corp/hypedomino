using System;
using System.Collections.Generic;
using HypeDomino.Game.Medals;
using HypeDomino.Game.Profile;
using HypeDomino.Game.Ranking;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Profile;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class PlayerProfileViewTests
    {
        const string Target = "acc-2";
        const string MatchId = "match-9";
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""}]}";
        const string PublicJson = @"{""account_id"":""acc-2"",""nickname"":""Murad"",""level"":12,""loadout"":{},""stats"":{""games"":40,""win_rate_percent"":55,""best_streak"":6},
""ranked"":{""ranks"":{""1v1"":""Gold II""}},
""medals"":[{""id"":""medal.first_win"",""nameKey"":""medal.first_win.name"",""rarityId"":""common"",""staticAddress"":""medal/first_win/static""}]}";

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

        static readonly string[] NoLocales = { };

        GameObject root;
        AvatarView avatar;
        TMP_Text nickname;
        TMP_Text[] stats;
        Image rankBadge;
        Sprite[] tierBadges;
        MedalCellSlot[] showcase;
        Button report;
        GameObject reasonPanel;
        Button[] reasons;
        Button closeReasons;
        TMP_Text message;
        PlayerProfileView view;
        Api api;
        ApiClient client;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        ProfileLoader loader;
        PlayerReport reporter;

        [SetUp]
        public void BuildView()
        {
            root = new GameObject("Player profile", typeof(RectTransform), typeof(Canvas));
            avatar = Child("Avatar", root.transform).AddComponent<AvatarView>();
            avatar.Configure(NewArt("Art", avatar.transform), NewArt("Frame", avatar.transform), null, null);
            nickname = Text("Nickname", root.transform);
            stats = new TMP_Text[(int)ProfileStat.BestStreak + 1];
            for (var at = 0; at < stats.Length; at++) stats[at] = Text("Stat " + at, root.transform);
            var summary = Child("Summary", root.transform).AddComponent<ProfileSummaryView>();
            summary.Configure(avatar, nickname, Text("Level", root.transform), stats);
            rankBadge = Child("Rank", root.transform).AddComponent<Image>();
            tierBadges = new Sprite[(int)RankTier.Challenger + 1];
            for (var tier = 1; tier < tierBadges.Length; tier++) tierBadges[tier] = Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), Vector2.zero);
            showcase = new MedalCellSlot[2];
            for (var at = 0; at < showcase.Length; at++)
            {
                var cell = Child("Cell " + at, root.transform);
                showcase[at] = cell.AddComponent<MedalCellSlot>();
                showcase[at].Configure(NewArt("Icon", cell.transform), Text("Name", cell.transform), Child("Locked", cell.transform));
            }
            report = NewButton("Report", root.transform);
            reasonPanel = Child("Reasons", root.transform);
            reasons = new Button[(int)PlayerReportReason.Other + 1];
            for (var at = 0; at < reasons.Length; at++) reasons[at] = NewButton("Reason " + at, reasonPanel.transform);
            closeReasons = NewButton("Close", reasonPanel.transform);
            message = Text("Message", root.transform);
            view = Child("View", root.transform).AddComponent<PlayerProfileView>();
            view.Configure(summary, rankBadge, tierBadges, showcase, message);
            view.ConfigureReport(report, reasonPanel, reasons, closeReasons);

            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            client = new ApiClient(api, session);
            stage = new CosmeticStage(new FakeArtSource<Sprite>(), new FakeArtSource<LoopArt>(), null, null, null, EffectLevel.Medium, false);
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var catalog));
            manifest = new ActiveCosmeticManifest(catalog);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            loader = new ProfileLoader(client);
            reporter = new PlayerReport(client);
            view.Bind(loader, reporter, stage, manifest, strings, Target, "1v1", MatchId);
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

        static Button NewButton(string name, Transform parent) => Child(name, parent).AddComponent<Image>().gameObject.AddComponent<Button>();

        static CosmeticImage NewArt(string name, Transform parent)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<AnimatedImage>().gameObject.AddComponent<CosmeticImage>();
            image.transform.SetParent(parent, false);
            return image;
        }

        void LoadProfile() => api.To(PlayerProfile.PublicPath(Target)).Done(new HttpReply(200, PublicJson));

        [Test]
        public void NothingShowsUntilTheServerAnswers()
        {
            Assert.AreEqual("GET", api.To(PlayerProfile.PublicPath(Target)).Method);
            Assert.IsFalse(avatar.gameObject.activeSelf);
            Assert.IsFalse(rankBadge.enabled);
            Assert.IsFalse(report.gameObject.activeSelf);
            Assert.IsFalse(reasonPanel.activeSelf);
        }

        [Test]
        public void ThePublicProfileShowsItsStatsRankAndShowcase()
        {
            LoadProfile();

            Assert.AreEqual("Murad", nickname.text);
            Assert.AreEqual("40", stats[(int)ProfileStat.Games].text);
            Assert.AreEqual("55", stats[(int)ProfileStat.WinRatePercent].text);
            Assert.IsFalse(stats[(int)ProfileStat.Wins].enabled);
            Assert.IsFalse(stats[(int)ProfileStat.Streak].enabled);
            Assert.IsTrue(rankBadge.enabled);
            Assert.AreSame(tierBadges[(int)RankTier.Gold], rankBadge.sprite);
            Assert.AreEqual("[!medal.first_win.name]", showcase[0].transform.Find("Name").GetComponent<TMP_Text>().text);
            Assert.IsFalse(showcase[1].gameObject.activeSelf);
            Assert.IsTrue(report.gameObject.activeSelf);
        }

        [Test]
        public void PickingAReasonSendsTheReportAndShowsTheReceipt()
        {
            LoadProfile();
            report.onClick.Invoke();
            Assert.IsTrue(reasonPanel.activeSelf);

            reasons[(int)PlayerReportReason.OffensiveNickname].onClick.Invoke();

            var sent = api.To(PlayerReport.Path);
            Assert.AreEqual("POST", sent.Method);
            Assert.AreEqual(@"{""target"":""acc-2"",""reason"":""offensive_nickname"",""match_id"":""match-9""}", sent.Body);
            Assert.IsFalse(reasons[0].interactable);
            Assert.IsFalse(closeReasons.interactable);

            sent.Done(new HttpReply(200, "{}"));

            Assert.IsFalse(reasonPanel.activeSelf);
            Assert.IsFalse(report.interactable);
            Assert.AreEqual(strings.Text(PlayerReportKeys.Received), message.text);
        }

        [Test]
        public void ARefusedReportShowsWhyAndLeavesReportingOpen()
        {
            LoadProfile();
            report.onClick.Invoke();
            reasons[(int)PlayerReportReason.Other].onClick.Invoke();

            api.To(PlayerReport.Path).Done(new HttpReply(429, "{}"));

            Assert.AreEqual(strings.Text(PlayerReportKeys.RateLimited), message.text);
            Assert.IsTrue(report.interactable);
        }

        [Test]
        public void ClosingTheReasonsSendsNothing()
        {
            LoadProfile();
            report.onClick.Invoke();

            closeReasons.onClick.Invoke();

            Assert.IsFalse(reasonPanel.activeSelf);
            Assert.IsNull(api.To(PlayerReport.Path));
        }

        [Test]
        public void ALateReplyForThePreviousPlayerIsDroppedAndTheNewPlayerIsLoaded()
        {
            const string Other = "acc-3";
            view.Bind(loader, reporter, stage, manifest, strings, Other, "1v1", MatchId);
            Assert.IsNull(api.To(PlayerProfile.PublicPath(Other)));

            LoadProfile();

            Assert.IsFalse(report.gameObject.activeSelf);
            Assert.AreNotEqual("Murad", nickname.text);
            Assert.AreEqual("GET", api.To(PlayerProfile.PublicPath(Other)).Method);

            api.To(PlayerProfile.PublicPath(Other)).Done(new HttpReply(200, PublicJson.Replace("acc-2", Other).Replace("Murad", "Leyla")));
            report.onClick.Invoke();
            reasons[(int)PlayerReportReason.Other].onClick.Invoke();

            Assert.AreEqual("Leyla", nickname.text);
            Assert.AreEqual(@"{""target"":""acc-3"",""reason"":""other"",""match_id"":""match-9""}", api.To(PlayerReport.Path).Body);
        }

        [Test]
        public void AnUnknownPlayerShowsTheLoadMessage()
        {
            api.To(PlayerProfile.PublicPath(Target)).Done(new HttpReply(404, @"{""error"":""profile.unknown""}"));

            Assert.AreEqual(strings.Text(ProfileLoadKeys.Unknown), message.text);
            Assert.IsFalse(report.gameObject.activeSelf);
        }
    }
}
