using System;
using System.Collections.Generic;
using HypeDomino.Core.Analytics;
using HypeDomino.Core.Diagnostics;
using HypeDomino.Game.Privacy;
using HypeDomino.Game.Push;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Settings;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ConsentSettingsShownTests
    {
        const string PolicyVersion = "2026-11-01";
        const string AppVersion = "0.4.2";

        sealed class Api : IApiTransport
        {
            public readonly List<string> Paths = new();
            public readonly List<Action<HttpReply>> Replies = new();

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

        sealed class AnalyticsSender : IAnalyticsSender
        {
            public void Send(string analyticsId, IReadOnlyList<AnalyticsEvent> batch, Action<bool> done) => done(true);
        }

        sealed class ErrorSender : IErrorSender
        {
            public void Send(ErrorReport report)
            {
            }
        }

        readonly List<GameObject> created = new();
        ScreenStack stack;
        ScreenView settings;
        ConsentSettingsView view;
        Api api;

        [SetUp]
        public void Build()
        {
            stack = Track(new GameObject("Stack")).AddComponent<ScreenStack>();
            stack.enabled = false;
            stack.Push(NewScreen("Lobby", out _));
            stack.Advance(TokenDurationMs.Slow);
            settings = NewScreen("Settings", out var content);
            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            var client = new ApiClient(api, session);
            var privacy = new PrivacySettings(client, new AnalyticsQueue(new AnalyticsSender(), () => 1000, 2, 4), new ErrorReports(new ErrorSender()), PrivacySettings.RandomAnalyticsId, () => PolicyVersion, AppVersion);
            var push = new PushSettings(client, () => PolicyVersion, AppVersion);
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(GameLanguage.English, Array.Empty<string>), false);
            view = settings.gameObject.AddComponent<ConsentSettingsView>();
            view.Configure(Row(content), Row(content), Row(content), Row(content), Child<TextMeshProUGUI>(content));
            view.Bind(privacy, push, strings);
        }

        [TearDown]
        public void DestroyAll()
        {
            foreach (var item in created) Object.Destroy(item);
            created.Clear();
        }

        GameObject Track(GameObject item)
        {
            created.Add(item);
            return item;
        }

        ScreenView NewScreen(string name, out Transform content)
        {
            var screen = Track(new GameObject(name, typeof(RectTransform))).AddComponent<ScreenView>();
            var group = new GameObject("Content", typeof(RectTransform), typeof(CanvasGroup));
            group.transform.SetParent(screen.transform, false);
            screen.Configure(group.GetComponent<CanvasGroup>(), null);
            content = group.transform;
            return screen;
        }

        static T Child<T>(Transform parent) where T : Component
        {
            var child = new GameObject(typeof(T).Name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child.AddComponent<T>();
        }

        static SettingToggleRow Row(Transform parent)
        {
            var toggle = Child<Toggle>(parent);
            var row = toggle.gameObject.AddComponent<SettingToggleRow>();
            row.Configure(toggle);
            return row;
        }

        void Open()
        {
            stack.Push(settings);
            view.LoadWhenShown();
        }

        void Close()
        {
            stack.Pop();
            stack.Advance(TokenDurationMs.Slow);
            view.LoadWhenShown();
        }

        void Answer(int request, int status, string body) => api.Replies[request](new HttpReply(status, body));

        [Test]
        public void AViewBoundOnAHiddenScreenLoadsNothing()
        {
            view.LoadWhenShown();

            Assert.IsFalse(settings.IsShown);
            Assert.AreEqual(0, api.Paths.Count);
        }

        [Test]
        public void OpeningTheScreenLoadsBothOnce()
        {
            Open();
            view.LoadWhenShown();

            CollectionAssert.AreEqual(new[] { PrivacySettings.Path, PushSettings.Path }, api.Paths);
        }

        [Test]
        public void EveryOpeningLoadsBothAgainSoAFailedLoadIsRetried()
        {
            Open();
            Answer(0, 200, "{\"analytics\":true,\"diagnostics\":false}");
            Answer(1, 503, "{\"error\":\"unavailable\"}");
            Close();

            Open();

            Assert.IsTrue(settings.IsShown);
            CollectionAssert.AreEqual(new[] { PrivacySettings.Path, PushSettings.Path, PrivacySettings.Path, PushSettings.Path }, api.Paths);
        }

        [Test]
        public void ASettledConsentFrameAllocatesNothing()
        {
            Open();
            Answer(0, 200, "{\"analytics\":true,\"diagnostics\":false}");
            Answer(1, 200, "{\"gameplay\":true,\"promotional\":false}");
            view.LoadWhenShown();

            Assert.That(view.LoadWhenShown, Allocates.Nothing);
        }
    }
}
