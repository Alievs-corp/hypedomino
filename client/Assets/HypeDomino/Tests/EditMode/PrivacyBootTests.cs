using System;
using System.Collections.Generic;
using HypeDomino.Core.Analytics;
using HypeDomino.Core.Diagnostics;
using HypeDomino.Game;
using HypeDomino.Game.Privacy;
using HypeDomino.Game.Push;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class PrivacyBootTests
    {
        const string AppVersion = "0.4.2";

        sealed class Sent
        {
            public string Method;
            public string PathAndQuery;
            public string Body;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { Method = method, PathAndQuery = pathAndQuery, Body = body, Done = done });
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

        Api api;
        ApiClient client;
        AnalyticsQueue analytics;
        ErrorReports diagnostics;
        PolicyVersions versions;
        ConsentStartup startup;
        FirstLaunchConsents firstLaunch;
        PrivacySettings privacy;
        PushSettings push;

        [SetUp]
        public void SignIn()
        {
            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var session = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            session.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            client = new ApiClient(api, session);
            analytics = new AnalyticsQueue(new AnalyticsSender(), () => 1000, 2, 4);
            diagnostics = new ErrorReports(new ErrorSender());
            versions = new PolicyVersions();
            startup = new ConsentStartup(client, new ConsentGates(analytics, diagnostics, PrivacySettings.RandomAnalyticsId));
            Func<string> current = () => versions.PrivacyPolicy;
            firstLaunch = new FirstLaunchConsents(current, AppVersion);
            privacy = new PrivacySettings(client, analytics, diagnostics, PrivacySettings.RandomAnalyticsId, current, AppVersion);
            push = new PushSettings(client, current, AppVersion);
        }

        static ServerHello Hello(HelloStatus status, PrivacyVersions announced) => new ServerHello { Status = status, PrivacyVersions = announced };

        static PrivacyVersions Announced(string policy, string terms) => new PrivacyVersions { PrivacyPolicyVersion = policy, TermsVersion = terms };

        Sent Last => api.Requests[api.Requests.Count - 1];

        [Test]
        public void NoVersionIsKnownBeforeAnAcceptedHello()
        {
            Assert.AreEqual(string.Empty, versions.PrivacyPolicy);
            Assert.AreEqual(string.Empty, versions.Terms);
        }

        [TestCase(HelloStatus.UpdateRequired)]
        [TestCase(HelloStatus.Maintenance)]
        public void AHelloThatIsNotAcceptedChangesNothingAndSendsNothing(HelloStatus status)
        {
            firstLaunch.Choose(ConsentPurpose.Analytics, true);
            startup.Hold(firstLaunch);
            Assert.IsFalse(PrivacyBoot.Follow(Hello(status, Announced("2026-11-01", "2026-10-01")), versions, startup, "acc-1"));
            Assert.AreEqual(string.Empty, versions.PrivacyPolicy);
            Assert.AreEqual(0, api.Requests.Count);
        }

        [Test]
        public void AnAcceptedHelloWithoutVersionsKeepsTheKnownOnes()
        {
            versions.Read(Hello(HelloStatus.Accepted, Announced("2026-11-01", "2026-10-01")));
            versions.Read(Hello(HelloStatus.Accepted, null));
            Assert.AreEqual("2026-11-01", versions.PrivacyPolicy);
            Assert.AreEqual("2026-10-01", versions.Terms);
        }

        [Test]
        public void TheAnnouncedPolicyVersionGoesWithTheFirstLaunchChoices()
        {
            firstLaunch.Choose(ConsentPurpose.PromotionalPush, true);
            firstLaunch.Choose(ConsentPurpose.Diagnostics, false);
            startup.Hold(firstLaunch);

            Assert.IsTrue(PrivacyBoot.Follow(Hello(HelloStatus.Accepted, Announced("2026-11-01", "2026-10-01")), versions, startup, "acc-1"));

            Assert.AreEqual("2026-10-01", versions.Terms);
            Assert.AreEqual("POST", Last.Method);
            Assert.AreEqual(FirstLaunchConsents.Path, Last.PathAndQuery);
            Assert.AreEqual("{\"choices\":[{\"purpose\":\"promotional_push\",\"granted\":true},{\"purpose\":\"diagnostics\",\"granted\":false}],\"policyVersion\":\"2026-11-01\",\"appVersion\":\"0.4.2\"}", Last.Body);
        }

        [Test]
        public void WithoutAPublishedPolicyVersionTheChoicesWaitAndTheStoredRecordIsRead()
        {
            firstLaunch.Choose(ConsentPurpose.Analytics, true);
            startup.Hold(firstLaunch);

            Assert.IsTrue(PrivacyBoot.Follow(Hello(HelloStatus.Accepted, Announced("", "")), versions, startup, "acc-1"));

            Assert.AreEqual(1, api.Requests.Count);
            Assert.AreEqual("GET", Last.Method);
            Assert.AreEqual(PrivacySettings.Path, Last.PathAndQuery);
            Assert.AreEqual(1, firstLaunch.Held);
        }

        [Test]
        public void SettingsChangesCarryTheLatestAnnouncedPolicyVersion()
        {
            PrivacyBoot.Follow(Hello(HelloStatus.Accepted, Announced("2026-11-01", "2026-10-01")), versions, startup, "acc-1");
            Last.Done(new HttpReply(200, "{\"analytics\":false,\"diagnostics\":false}"));
            privacy.Load(_ => { });
            Last.Done(new HttpReply(200, "{\"analytics\":false,\"diagnostics\":false}"));
            push.Load(_ => { });
            Last.Done(new HttpReply(200, "{\"gameplay\":true,\"promotional\":false}"));

            versions.Read(Hello(HelloStatus.Accepted, Announced("2026-12-01", "2026-10-01")));

            Assert.IsTrue(privacy.SetAnalytics(true, _ => { }));
            Assert.AreEqual("{\"analytics\":true,\"diagnostics\":false,\"policyVersion\":\"2026-12-01\",\"appVersion\":\"0.4.2\"}", Last.Body);
            Last.Done(new HttpReply(200, "{\"analytics\":true,\"diagnostics\":false}"));
            Assert.IsTrue(push.SetPromotional(true, _ => { }));
            Assert.AreEqual("{\"gameplay\":true,\"promotional\":true,\"policyVersion\":\"2026-12-01\",\"appVersion\":\"0.4.2\"}", Last.Body);
        }
    }
}
