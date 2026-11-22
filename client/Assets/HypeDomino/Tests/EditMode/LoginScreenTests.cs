using System;
using System.Collections.Generic;
using HypeDomino.Game.Login;
using HypeDomino.Game.Privacy;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Localization;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class LoginScreenTests
    {
        const string NonceJson = "{\"nonce\":\"nonce-1\"}";
        const string TokensJson = "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"AAECAw==\",\"account_id\":\"acc-1\",\"created\":true}";

        sealed class Sent
        {
            public string Path;
            public Action<HttpReply> Done;
        }

        sealed class Transport : IAuthTransport
        {
            public readonly List<Sent> Requests = new();

            public void Post(string path, string json, Action<HttpReply> done) => Requests.Add(new Sent { Path = path, Done = done });

            public void Answer(int status, string body)
            {
                var pending = Requests[Requests.Count - 1];
                pending.Done(new HttpReply(status, body));
            }
        }

        sealed class Identity : IIdentityProvider
        {
            public bool Supported = true;
            public readonly List<AuthProvider> Asked = new();
            public readonly List<Action<IdTokenResult>> Pending = new();

            public bool Supports(AuthProvider provider) => Supported;

            public void RequestIdToken(AuthProvider provider, string nonce, Action<IdTokenResult> done)
            {
                Asked.Add(provider);
                Pending.Add(done);
            }

            public void Answer(IdTokenOutcome outcome) => Pending[Pending.Count - 1](new IdTokenResult(outcome, outcome == IdTokenOutcome.Token ? "id-token" : null));
        }

        sealed class Store : ISecureStore
        {
            readonly Dictionary<string, string> values = new();

            public string Read(string key) => values.TryGetValue(key, out var value) ? value : null;

            public void Write(string key, string value) => values[key] = value;

            public void Delete(string key) => values.Remove(key);
        }

        Transport transport;
        Identity identity;
        LoginScreen screen;
        int signedIn;
        int changes;

        [SetUp]
        public void Build()
        {
            transport = new Transport();
            identity = new Identity();
            var auth = new AuthSession(transport, identity, new Store());
            screen = new LoginScreen(auth, new AccountDeletion(transport, identity), new[] { AuthProvider.Google }, () => signedIn++);
            screen.Changed += () => changes++;
            signedIn = 0;
            changes = 0;
        }

        void SignInAnswering(int status, string body)
        {
            Assert.IsTrue(screen.TrySignIn(AuthProvider.Google));
            transport.Answer(200, NonceJson);
            identity.Answer(IdTokenOutcome.Token);
            transport.Answer(status, body);
        }

        [Test]
        public void TheScreenNeedsItsParts()
        {
            var auth = new AuthSession(transport, identity, new Store());
            var deletion = new AccountDeletion(transport, identity);
            Assert.Throws<ArgumentNullException>(() => new LoginScreen(null, deletion, SignInOptions.For(UnityEngine.RuntimePlatform.Android), () => { }));
            Assert.Throws<ArgumentNullException>(() => new LoginScreen(auth, null, SignInOptions.For(UnityEngine.RuntimePlatform.Android), () => { }));
            Assert.Throws<ArgumentNullException>(() => new LoginScreen(auth, deletion, null, () => { }));
            Assert.Throws<ArgumentNullException>(() => new LoginScreen(auth, deletion, SignInOptions.For(UnityEngine.RuntimePlatform.Android), null));
        }

        [Test]
        public void ASignInStartsTheConnectionOnceAndShowsNoMessage()
        {
            SignInAnswering(200, TokensJson);

            Assert.AreEqual(1, signedIn);
            Assert.IsFalse(screen.Busy);
            Assert.IsNull(screen.MessageKey);
            Assert.AreEqual(2, changes);
            Assert.AreEqual(AuthSession.SignInPath, transport.Requests[1].Path);
        }

        [Test]
        public void TheScreenIsBusyUntilTheAnswerAndRefusesASecondTap()
        {
            Assert.IsTrue(screen.TrySignIn(AuthProvider.Google));

            Assert.IsTrue(screen.Busy);
            Assert.IsFalse(screen.TrySignIn(AuthProvider.Google));
            Assert.AreEqual(1, transport.Requests.Count);
        }

        [Test]
        public void AProviderTheScreenDoesNotOfferStartsNothing()
        {
            Assert.IsFalse(screen.TrySignIn(AuthProvider.Apple));

            Assert.IsFalse(screen.Busy);
            Assert.AreEqual(0, transport.Requests.Count);
            Assert.AreEqual(0, changes);
        }

        [Test]
        public void ClosingTheAccountPickerShowsNoMessageAndAllowsAnotherTry()
        {
            Assert.IsTrue(screen.TrySignIn(AuthProvider.Google));
            transport.Answer(200, NonceJson);
            identity.Answer(IdTokenOutcome.Cancelled);

            Assert.IsFalse(screen.Busy);
            Assert.IsNull(screen.MessageKey);
            Assert.AreEqual(0, signedIn);
            Assert.IsTrue(screen.TrySignIn(AuthProvider.Google));
        }

        [Test]
        public void ADeviceWithoutTheProviderSaysSo()
        {
            identity.Supported = false;

            Assert.IsTrue(screen.TrySignIn(AuthProvider.Google));

            Assert.IsFalse(screen.Busy);
            Assert.AreEqual(LoginKeys.Unsupported, screen.MessageKey);
            Assert.AreEqual(0, signedIn);
        }

        [TestCase(0, "", "account.deletion.offline")]
        [TestCase(429, "", "account.deletion.rate_limited")]
        [TestCase(503, "", LoginKeys.Unavailable)]
        [TestCase(403, "{\"error\":\"banned\"}", LoginKeys.Banned)]
        [TestCase(410, "{\"error\":\"deleted\"}", "account.deletion.deleted")]
        public void ARefusedSignInShowsItsMessage(int status, string body, string key)
        {
            SignInAnswering(status, body);

            Assert.AreEqual(key, screen.MessageKey);
            Assert.IsFalse(screen.Busy);
            Assert.IsFalse(screen.DeletionPending);
            Assert.AreEqual(0, signedIn);
        }

        [Test]
        public void AnAccountWaitingForDeletionShowsTheDateAndCanBeRestored()
        {
            SignInAnswering(409, "{\"error\":\"deletion_pending\",\"delete_at\":\"2026-12-21T09:00:00Z\"}");

            Assert.IsTrue(screen.DeletionPending);
            Assert.AreEqual(AccountDeletionKeys.Pending, screen.MessageKey);
            Assert.AreEqual(new DateTimeOffset(2026, 12, 21, 9, 0, 0, TimeSpan.Zero), screen.DeleteAt);
            Assert.AreEqual(0, signedIn);

            Assert.IsTrue(screen.TryCancelDeletion());
            Assert.IsTrue(screen.Busy);
            Assert.IsFalse(screen.TryCancelDeletion());
            transport.Answer(200, NonceJson);
            identity.Answer(IdTokenOutcome.Token);
            Assert.AreEqual(AccountDeletion.CancelPath, transport.Requests[transport.Requests.Count - 1].Path);
            transport.Answer(204, "");

            Assert.IsFalse(screen.Busy);
            Assert.IsFalse(screen.DeletionPending);
            Assert.AreEqual(AccountDeletionKeys.Restored, screen.MessageKey);
            Assert.AreEqual(AuthProvider.Google, identity.Asked[identity.Asked.Count - 1]);
        }

        [Test]
        public void CancellingDeletionNeedsAPendingAccount()
        {
            Assert.IsFalse(screen.TryCancelDeletion());

            SignInAnswering(503, "");

            Assert.IsFalse(screen.TryCancelDeletion());
            Assert.AreEqual(2, transport.Requests.Count);
        }

        [Test]
        public void EveryMessageIsInTheShippedBundleInAzerbaijaniAndEnglish()
        {
            var strings = BuiltinStrings.Load();
            foreach (AuthOutcome outcome in Enum.GetValues(typeof(AuthOutcome)))
            {
                var key = LoginKeys.Of(outcome);
                if (key == null) continue;
                Assert.IsNotNull(strings.Value(key, GameLanguage.English), key);
                Assert.IsNotNull(strings.Value(key, "az"), key);
            }
            Assert.AreEqual("Bu hesab bloklanıb", strings.Value(LoginKeys.Banned, "az"));
        }
    }
}
