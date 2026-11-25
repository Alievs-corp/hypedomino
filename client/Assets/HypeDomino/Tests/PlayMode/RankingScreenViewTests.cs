using System;
using System.Collections.Generic;
using HypeDomino.Game.Ranking;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class RankingScreenViewTests
    {
        sealed class Api : IApiTransport
        {
            public readonly List<Action<HttpReply>> Replies = new();
            public readonly List<string> Paths = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done)
            {
                Paths.Add(pathAndQuery);
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
        Button solo;
        Button teams;
        RankingScreenView view;
        Api api;
        ApiClient client;
        GameStrings strings;

        [SetUp]
        public void BuildView()
        {
            root = new GameObject("Ranking screen", typeof(RectTransform), typeof(Canvas));
            solo = Child("Solo", root.transform).AddComponent<Button>();
            teams = Child("Teams", root.transform).AddComponent<Button>();
            view = root.AddComponent<RankingScreenView>();
            view.Configure(RankView(), BoardView(), solo, teams);

            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            client = new ApiClient(api, session);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), false);
        }

        [TearDown]
        public void DestroyView() => UnityEngine.Object.Destroy(root);

        static GameObject Child(string name, Transform parent)
        {
            var child = new GameObject(name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child;
        }

        static RectTransform Stretched(string name, Transform parent)
        {
            var rect = (RectTransform)Child(name, parent).transform;
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.sizeDelta = Vector2.zero;
            return rect;
        }

        static TMP_Text Text(string name, Transform parent) => Child(name, parent).AddComponent<TextMeshProUGUI>();

        RankScreenView RankView()
        {
            var panel = Child("Standing", root.transform);
            var lpFill = Child("LP fill", panel.transform).AddComponent<Image>();
            var placementFill = Child("Placement fill", panel.transform).AddComponent<Image>();
            var rank = root.AddComponent<RankScreenView>();
            rank.Configure(panel, Child("Badge", panel.transform).AddComponent<Image>(), Array.Empty<Sprite>(), lpFill.gameObject, lpFill, Text("LP", panel.transform),
                placementFill.gameObject, placementFill, Text("Season", panel.transform), Text("Season left", panel.transform));
            return rank;
        }

        LeaderboardView BoardView()
        {
            var holder = Child("List", root.transform);
            ((RectTransform)holder.transform).sizeDelta = new Vector2(600f, 1000f);
            var viewport = Stretched("Viewport", holder.transform);
            var content = Stretched("Content", viewport);
            var scroll = holder.AddComponent<ScrollRect>();
            scroll.viewport = viewport;
            scroll.content = content;
            scroll.horizontal = false;
            var template = Child("Row", root.transform);
            ((RectTransform)template.transform).sizeDelta = new Vector2(600f, 100f);
            var slot = template.AddComponent<LeaderboardRowSlot>();
            slot.Configure(Text("Position", template.transform), Text("Nickname", template.transform), Child("Badge", template.transform).AddComponent<Image>(),
                Text("LP", template.transform), Child("Own", template.transform));
            var leaderboard = root.AddComponent<LeaderboardView>();
            leaderboard.Configure(holder.AddComponent<RecyclingList>(), slot, Array.Empty<Sprite>());
            return leaderboard;
        }

        void Bind() => view.Bind(new RankStandingLoader(client), new LeaderboardLoader(client), new ServerClock(), strings);

        [Test]
        public void TheScreenOpensOnTheSoloStandingAndBoard()
        {
            Bind();

            Assert.AreEqual(new[] { RankStanding.Path, LeaderboardStanding.PathOf(RankingScreenView.SoloMode, LeaderboardStanding.BoardGlobal) }, api.Paths.ToArray());
            Assert.IsFalse(solo.interactable);
            Assert.IsTrue(teams.interactable);
        }

        [Test]
        public void TheTeamsButtonLoadsTheTeamsBoardAndTurnsItselfOff()
        {
            Bind();
            api.Replies[0](new HttpReply(500, "{}"));
            api.Replies[1](new HttpReply(500, "{}"));

            teams.onClick.Invoke();

            Assert.AreEqual(4, api.Paths.Count);
            Assert.AreEqual(RankStanding.Path, api.Paths[2]);
            Assert.AreEqual(LeaderboardStanding.PathOf(RankingScreenView.TeamsMode, LeaderboardStanding.BoardGlobal), api.Paths[3]);
            Assert.IsTrue(solo.interactable);
            Assert.IsFalse(teams.interactable);
        }
    }
}
