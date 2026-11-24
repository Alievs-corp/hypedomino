using System;
using System.Collections.Generic;
using HypeDomino.Core.Analytics;
using HypeDomino.Core.Diagnostics;
using HypeDomino.Game.Privacy;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Settings;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class FirstLaunchConsentViewTests
    {
        const string PolicyVersion = "2026-11-01";
        const string AppVersion = "0.4.2";

        sealed class Sent
        {
            public string PathAndQuery;
            public string Body;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { PathAndQuery = pathAndQuery, Body = body });
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

        GameObject root;
        Api api;
        ConsentStartup startup;
        Toggle analytics;
        Toggle diagnostics;
        Button proceed;
        int continued;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(FirstLaunchConsentViewTests), typeof(RectTransform));
            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            var client = new ApiClient(api, session);
            var gates = new ConsentGates(new AnalyticsQueue(new AnalyticsSender(), () => 1000, 2, 4), new ErrorReports(new ErrorSender()), PrivacySettings.RandomAnalyticsId);
            startup = new ConsentStartup(client, gates);
            analytics = Add<Toggle>();
            diagnostics = Add<Toggle>();
            proceed = Add<Button>();
            continued = 0;
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        T Add<T>() where T : Component
        {
            var child = new GameObject(typeof(T).Name, typeof(RectTransform));
            child.transform.SetParent(root.transform, false);
            return child.AddComponent<T>();
        }

        SettingToggleRow Row(Toggle toggle)
        {
            var row = Add<SettingToggleRow>();
            row.Configure(toggle);
            return row;
        }

        FirstLaunchConsents Show()
        {
            analytics.isOn = true;
            diagnostics.isOn = true;
            var choices = new FirstLaunchConsents(PolicyVersion, AppVersion);
            var view = root.AddComponent<FirstLaunchConsentView>();
            view.Configure(Row(analytics), Row(diagnostics), proceed);
            view.Bind(choices, startup, () => continued++);
            return choices;
        }

        Sent Last => api.Requests[api.Requests.Count - 1];

        [Test]
        public void NoPurposeStartsTickedAndNothingIsHeldOrSent()
        {
            var choices = Show();

            Assert.IsFalse(analytics.isOn);
            Assert.IsFalse(diagnostics.isOn);
            Assert.AreEqual(0, choices.Held);
            Assert.AreEqual(0, api.Requests.Count);
        }

        [Test]
        public void ContinuingHandsTheChoicesOverInTheOrderThePlayerMadeThem()
        {
            Show();

            diagnostics.isOn = true;
            analytics.isOn = true;
            proceed.onClick.Invoke();

            Assert.AreEqual(1, continued);
            Assert.AreEqual(0, api.Requests.Count);

            Assert.IsTrue(startup.SignedIn("acc-1"));

            Assert.AreEqual(FirstLaunchConsents.Path, Last.PathAndQuery);
            Assert.AreEqual("{\"choices\":[{\"purpose\":\"diagnostics\",\"granted\":true},{\"purpose\":\"analytics\",\"granted\":true}],\"policyVersion\":\"2026-11-01\",\"appVersion\":\"0.4.2\"}", Last.Body);
        }

        [Test]
        public void ATurnedBackChoiceIsHeldAsTwoChanges()
        {
            var choices = Show();

            analytics.isOn = true;
            analytics.isOn = false;
            proceed.onClick.Invoke();
            startup.SignedIn("acc-1");

            Assert.AreEqual(2, choices.Held);
            Assert.AreEqual("{\"choices\":[{\"purpose\":\"analytics\",\"granted\":true},{\"purpose\":\"analytics\",\"granted\":false}],\"policyVersion\":\"2026-11-01\",\"appVersion\":\"0.4.2\"}", Last.Body);
        }

        [Test]
        public void ContinuingWithoutAChoiceReadsTheServerRecordAfterSignIn()
        {
            Show();

            proceed.onClick.Invoke();
            startup.SignedIn("acc-1");

            Assert.AreEqual(1, continued);
            Assert.AreEqual("/consents", Last.PathAndQuery);
        }
    }
}
