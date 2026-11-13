using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using Google.Protobuf;
using HypeDomino.Content;
using HypeDomino.Game;
using HypeDomino.Net;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Localization;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;
using AnnouncedBundle = Hypedomino.V1.StringBundle;
using DownloadedBundle = HypeDomino.UI.Localization.StringBundle;

namespace HypeDomino.Tests.EditMode
{
    public sealed class AppBootTests
    {
        const int ChosenRate = 120;
        const int MenuRate = 60;
        const string Cdn = "https://cdn.example.com";
        const string Version = "2026.10.26-1";
        const ulong Nonce = 0x5EED_CAFE_F00D_0001;
        const string Builtin = @"{""entries"":[{""key"":""k"",""en"":""E"",""tr"":""T""}]}";
        const string AzContent = @"{""language"":""az"",""strings"":{""k"":""A""}}";
        const string EnHash = "e0";
        const string RuHash = "r0";

        static readonly string[] NoLocales = { };
        static readonly Regex MissingK = new Regex("^Strings missing key=k language=az version=0$");

        sealed class IdleCatalogSource : ICatalogSource
        {
            public List<string> Downloads { get; } = new();

            public string StoredVersion => null;

            public string StoredBaseUrl => null;

            public void LoadStored(Action<bool> done) { }

            public void Download(string url, Action<bool> done) => Downloads.Add(url);

            public void Commit(string version, string baseUrl, Action<bool> done) { }
        }

        sealed class ServerFrames : IFrameTransport
        {
            readonly List<byte[]> incoming = new(4);
            int next;

            public void Send(ReadOnlySpan<byte> frame) { }

            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                frame = default;
                if (next == incoming.Count)
                {
                    incoming.Clear();
                    next = 0;
                    return false;
                }
                frame = incoming[next++];
                return true;
            }

            public void Deliver(byte[] frame) => incoming.Add(frame);
        }

        sealed class Listened
        {
            public readonly ServerFrames Server = new();
            public readonly Session Session = new();
            public readonly AppFlow Flow = new();
            public readonly IdleCatalogSource Source = new();
            public readonly FrameRatePolicy FrameRate = AppBootTests.FrameRate();
            public readonly GameStrings Strings = AppBootTests.Strings("az");
            public readonly CatalogUpdater Catalog;

            public Listened()
            {
                Catalog = new CatalogUpdater(Source, "Android");
                Flow.Changed += state => AppBoot.FollowMatch(state, FrameRate, Catalog);
                AppBoot.Listen(Session, Flow, Catalog, Strings);
                Session.Hello(Server, new ClientHello());
            }

            public void Receive(Envelope.PayloadOneofCase payload, IMessage message)
            {
                Server.Deliver(FrameOf(payload, message));
                Session.Poll();
            }
        }

        static byte[] FrameOf(Envelope.PayloadOneofCase payload, IMessage message) =>
            new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, Nonce), payload, message).ToArray();

        static ServerHello Hello(HelloStatus status, bool inMatch) =>
            new ServerHello { Status = status, InMatch = inMatch, SessionNonce = Nonce, CatalogVersion = Version, ContentBaseUrl = Cdn };

        static MatchView View(MatchState state) => new MatchView { State = state };

        static string CatalogUrl => ContentUrls.Catalog(Cdn, "Android", Version);

        static FrameRatePolicy FrameRate() =>
            new FrameRatePolicy(new GraphicsQuality(DeviceTier.High, GraphicsLevel.High, 0, false), new[] { 60, 120 }, ChosenRate, false);

        static CatalogUpdater Catalog() => new CatalogUpdater(new IdleCatalogSource(), "Android");

        static GameStrings Strings(string language) =>
            new GameStrings(BuiltinStrings.Parse(Builtin), LanguageChoice.Load(language, () => NoLocales), false);

        static string AzHash => DownloadedBundle.Sha256Hex(Encoding.UTF8.GetBytes(AzContent));

        static ServerHello Announcing(HelloStatus status, params FallbackChain[] chains)
        {
            var hello = Hello(status, false);
            hello.StringBundles.Add(new AnnouncedBundle { Language = "az", Version = 3, Hash = AzHash });
            hello.StringBundles.Add(new AnnouncedBundle { Language = "en", Version = 4, Hash = EnHash });
            hello.StringBundles.Add(new AnnouncedBundle { Language = "ru", Version = 2, Hash = RuHash });
            hello.FallbackChains.Add(chains);
            return hello;
        }

        static FallbackChain Chain(string language, params string[] languages)
        {
            var chain = new FallbackChain { Language = language };
            chain.Languages.Add(languages);
            return chain;
        }

        static string[] LanguagesOf(List<AnnouncedBundle> bundles)
        {
            var languages = new string[bundles.Count];
            for (var i = 0; i < languages.Length; i++) languages[i] = bundles[i].Language;
            return languages;
        }

        [TestCase(AppState.Boot, false)]
        [TestCase(AppState.Login, false)]
        [TestCase(AppState.Lobby, false)]
        [TestCase(AppState.Match, true)]
        [TestCase(AppState.UpdateRequired, false)]
        [TestCase(AppState.Maintenance, false)]
        public void OnlyTheMatchStateCountsAsInMatch(AppState state, bool inMatch)
        {
            var frameRate = FrameRate();
            var catalog = Catalog();
            AppBoot.FollowMatch(state, frameRate, catalog);
            Assert.AreEqual(inMatch, frameRate.InMatch);
            Assert.AreEqual(inMatch, catalog.InMatch);
        }

        [Test]
        public void TheChosenRateAndTheCatalogHoldLastOnlyWhileInTheMatch()
        {
            var frameRate = FrameRate();
            var catalog = Catalog();
            AppBoot.FollowMatch(AppState.Match, frameRate, catalog);
            Assert.AreEqual(ChosenRate, frameRate.Target);
            AppBoot.FollowMatch(AppState.Lobby, frameRate, catalog);
            Assert.AreEqual(MenuRate, frameRate.Target);
            Assert.IsFalse(catalog.InMatch);
        }

        [Test]
        public void TheEditorHasNoCatalogAndStillFollowsTheMatch()
        {
            var frameRate = FrameRate();
            AppBoot.FollowMatch(AppState.Match, frameRate, null);
            Assert.IsTrue(frameRate.InMatch);
        }

        [Test]
        public void AnAcceptedHelloFromTheSessionOpensTheLobbyAndFetchesTheAnnouncedCatalog()
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Hello(HelloStatus.Accepted, false));
            Assert.AreEqual(AppState.Lobby, app.Flow.State);
            CollectionAssert.AreEqual(new[] { CatalogUrl }, app.Source.Downloads);
            Assert.AreEqual(Cdn, app.Catalog.BaseUrl);
        }

        [Test]
        public void AHelloIntoARunningMatchHoldsTheAnnouncedCatalogUntilTheMatchIsLeft()
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Hello(HelloStatus.Accepted, true));
            Assert.AreEqual(AppState.Match, app.Flow.State);
            Assert.IsTrue(app.FrameRate.InMatch);
            CollectionAssert.IsEmpty(app.Source.Downloads);

            app.Receive(Envelope.PayloadOneofCase.MatchView, View(MatchState.Finished));
            Assert.IsTrue(app.Flow.LeaveMatch());
            CollectionAssert.AreEqual(new[] { CatalogUrl }, app.Source.Downloads);
        }

        [TestCase(HelloStatus.UpdateRequired, AppState.UpdateRequired)]
        [TestCase(HelloStatus.Maintenance, AppState.Maintenance)]
        public void AHelloThatBlocksPlayShowsItsScreenAndFetchesNoCatalog(HelloStatus status, AppState shown)
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Hello(status, false));
            Assert.AreEqual(shown, app.Flow.State);
            CollectionAssert.IsEmpty(app.Source.Downloads);
            Assert.IsNull(app.Catalog.BaseUrl);
        }

        [Test]
        public void AMatchViewFromTheSessionTakesTheLobbyIntoTheMatch()
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Hello(HelloStatus.Accepted, false));
            app.Receive(Envelope.PayloadOneofCase.MatchView, View(MatchState.Loading));
            Assert.AreEqual(AppState.Match, app.Flow.State);
            Assert.IsTrue(app.Catalog.InMatch);
            Assert.IsFalse(app.Flow.LeaveMatch());
        }

        [Test]
        public void TheEditorWithoutACatalogStillFollowsTheHello()
        {
            var session = new Session();
            var flow = new AppFlow();
            var server = new ServerFrames();
            AppBoot.Listen(session, flow, null, Strings("az"));
            session.Hello(server, new ClientHello());
            server.Deliver(FrameOf(Envelope.PayloadOneofCase.ServerHello, Hello(HelloStatus.Accepted, false)));
            session.Poll();
            Assert.AreEqual(AppState.Lobby, flow.State);
        }

        [Test]
        public void TheBundlesToFetchAreTheChosenLanguageAndEnglishNotYetHeld()
        {
            var fetch = AppBoot.FollowStrings(Announcing(HelloStatus.Accepted), Strings("az"));
            CollectionAssert.AreEqual(new[] { "az", "en" }, LanguagesOf(fetch));
            Assert.AreEqual(3u, fetch[0].Version);
            Assert.AreEqual(AzHash, fetch[0].Hash);
            Assert.AreEqual(4u, fetch[1].Version);
            Assert.AreEqual(EnHash, fetch[1].Hash);
        }

        [Test]
        public void EnglishAloneIsFetchedWhenEnglishIsChosen() =>
            CollectionAssert.AreEqual(new[] { "en" }, LanguagesOf(AppBoot.FollowStrings(Announcing(HelloStatus.Accepted), Strings("en"))));

        [Test]
        public void ABundleWhoseHashIsAlreadyHeldIsNotFetchedAgain()
        {
            var strings = Strings("az");
            Assert.IsTrue(strings.Stage("az", 3, AzHash, Encoding.UTF8.GetBytes(AzContent)));
            CollectionAssert.AreEqual(new[] { "en" }, LanguagesOf(AppBoot.FollowStrings(Announcing(HelloStatus.Accepted), strings)));
            Assert.IsTrue(strings.ApplyStaged(false));
            CollectionAssert.AreEqual(new[] { "en" }, LanguagesOf(AppBoot.FollowStrings(Announcing(HelloStatus.Accepted), strings)));
        }

        [Test]
        public void AHelloWithoutBundlesFetchesNothing() =>
            CollectionAssert.IsEmpty(AppBoot.FollowStrings(Hello(HelloStatus.Accepted, false), Strings("az")));

        [Test]
        public void TheAnnouncedChainsReplaceTheDefaultOnes()
        {
            var strings = Strings("az");
            LogAssert.Expect(LogType.Warning, MissingK);
            Assert.AreEqual("T", strings.Text("k"));
            AppBoot.FollowStrings(Announcing(HelloStatus.Accepted, Chain("az", "az", "en")), strings);
            Assert.AreEqual("E", strings.Text("k"));
        }

        [Test]
        public void AChainNotEndingInEnglishIsRefusedAndLogged()
        {
            var strings = Strings("az");
            LogAssert.Expect(LogType.Warning, new Regex("^Strings chain refused language=az$"));
            AppBoot.FollowStrings(Announcing(HelloStatus.Accepted, Chain("az", "az", "ru")), strings);
            LogAssert.Expect(LogType.Warning, MissingK);
            Assert.AreEqual("T", strings.Text("k"));
        }

        [Test]
        public void AnAcceptedHelloFromTheSessionFollowsTheAnnouncedStrings()
        {
            var app = new Listened();
            LogAssert.Expect(LogType.Log, new Regex("^Strings fetch language=az version=3$"));
            LogAssert.Expect(LogType.Log, new Regex("^Strings fetch language=en version=4$"));
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Announcing(HelloStatus.Accepted, Chain("az", "az", "en")));
            LogAssert.Expect(LogType.Warning, MissingK);
            Assert.AreEqual("E", app.Strings.Text("k"));
        }

        [Test]
        public void AHelloThatBlocksPlayLeavesTheStringsAlone()
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Announcing(HelloStatus.Maintenance, Chain("az", "az", "en")));
            LogAssert.Expect(LogType.Warning, MissingK);
            Assert.AreEqual("T", app.Strings.Text("k"));
        }

        [Test]
        public void AMatchViewPassedToTheFlowAllocatesNothing()
        {
            var app = new Listened();
            app.Receive(Envelope.PayloadOneofCase.ServerHello, Hello(HelloStatus.Accepted, true));
            var view = FrameOf(Envelope.PayloadOneofCase.MatchView, View(MatchState.Playing));

            TestDelegate step = () =>
            {
                app.Server.Deliver(view);
                app.Session.Poll();
            };
            step();

            Assert.That(step, Allocates.Nothing);
            Assert.AreEqual(AppState.Match, app.Flow.State);
        }
    }
}
