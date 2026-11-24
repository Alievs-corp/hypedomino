using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.Lobby;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.Net.Integrity;
using HypeDomino.UI.Localization;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class LobbyScreenViewTests
    {
        const ulong Nonce = 0x5EED_0074_0000_0004;
        const ulong ServerNowMs = 1_800_000_000_000;
        const int Casual2V2 = 3;
        const string WalletJson = "{\"coins\":1500,\"crystals\":20,\"entries\":[]}";
        const string OwnJson = @"{""account_id"":""8d1f0c52-5a4e-4f7b-9a43-0c7d2b1e6f10"",""nickname"":""Leyla"",""setup_required"":false,""rename_required"":false,""level"":7,""xp"":1250,""loadout"":{},""stats"":{""games"":40,""wins"":22,""losses"":18,""win_rate_percent"":55,""streak"":2,""best_streak"":6,""modes"":{}}}";

        sealed class FakeTransport : IFrameTransport
        {
            readonly Queue<byte[]> incoming = new();

            public void Send(ReadOnlySpan<byte> frame) { }

            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                frame = default;
                if (incoming.Count == 0) return false;
                frame = incoming.Dequeue();
                return true;
            }

            public void Deliver(Envelope.PayloadOneofCase payload, IMessage message) =>
                incoming.Enqueue(new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, Nonce), payload, message).ToArray());
        }

        sealed class Sent
        {
            public string PathAndQuery;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { PathAndQuery = pathAndQuery, Done = done });

            public Sent Last(string path) => Requests.FindLast(request => request.PathAndQuery == path);
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

        GameObject root;
        Button[] cards;
        GameObject[] busy;
        Button cancel;
        TMP_Text countdown;
        TMP_Text message;
        TMP_Text coins;
        TMP_Text crystals;
        TMP_Text level;
        Button retry;
        MatchQueue queue;
        LobbyScreen screen;
        LobbyHeaderLoader loader;
        Api api;
        Session session;
        ServerClock clock;
        LanguageChoice choice;
        GameStrings strings;
        LobbyScreenView view;
        readonly List<Action<IntegrityProof>> proofs = new();

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(LobbyScreenViewTests), typeof(RectTransform));
            cards = new Button[PlayModes.Count];
            busy = new GameObject[PlayModes.Count];
            for (var card = 0; card < cards.Length; card++)
            {
                cards[card] = NewButton("Card " + card);
                busy[card] = new GameObject("Busy " + card, typeof(RectTransform));
                busy[card].transform.SetParent(cards[card].transform, false);
            }
            cancel = NewButton("Cancel");
            countdown = NewText("Countdown");
            message = NewText("Message");
            coins = NewText("Coins");
            crystals = NewText("Crystals");
            level = NewText("Level");
            retry = NewButton("Retry");
            proofs.Clear();
            queue = new MatchQueue();
            var header = new LobbyHeader();
            screen = new LobbyScreen(header, new LobbyPlay(queue, (_, done) => proofs.Add(done)), queue);
            api = new Api();
            loader = new LobbyHeaderLoader(new ApiClient(api, SignedIn()), header);
            session = AcceptedSession();
            clock = new ServerClock();
            clock.Sync(ServerNowMs, 0, 0);
            choice = LanguageChoice.Load(GameLanguage.English, () => Array.Empty<string>());
            strings = new GameStrings(BuiltinStrings.Load(), choice, false);
            view = root.AddComponent<LobbyScreenView>();
            view.Configure(cards, busy, cancel, countdown, message, coins, crystals, level, retry);
            view.Bind(screen, queue, loader, session, clock, strings);
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        Button NewButton(string name)
        {
            var button = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            button.transform.SetParent(root.transform, false);
            return button;
        }

        TMP_Text NewText(string name)
        {
            var text = new GameObject(name, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            text.transform.SetParent(root.transform, false);
            return text;
        }

        static AuthSession SignedIn()
        {
            var auth = new Auth();
            var store = new Store();
            var authSession = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            authSession.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            return authSession;
        }

        static Session AcceptedSession()
        {
            var transport = new FakeTransport();
            var accepted = new Session();
            accepted.Hello(transport, new ClientHello());
            transport.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce });
            accepted.Poll();
            Assert.IsTrue(accepted.Accepted);
            return accepted;
        }

        string Duration(long seconds)
        {
            var span = TimeSpan.FromSeconds(seconds);
            var args = new Dictionary<string, object>
            {
                ["days"] = (long)span.Days,
                ["hours"] = (long)span.Hours,
                ["minutes"] = (long)span.Minutes,
                ["seconds"] = (long)span.Seconds,
            };
            return strings.Text(MessageFormat.DurationKey(span), args);
        }

        [Test]
        public void TheFirstTickShowsReadyCardsAndAnEmptyHeaderWithRetry()
        {
            view.Tick(0);

            for (var card = 0; card < cards.Length; card++)
            {
                Assert.IsTrue(cards[card].interactable, $"card {card}");
                Assert.IsFalse(busy[card].activeSelf, $"card {card}");
            }
            Assert.IsFalse(cancel.gameObject.activeSelf);
            Assert.IsFalse(countdown.enabled);
            Assert.IsFalse(message.enabled);
            Assert.IsFalse(coins.enabled);
            Assert.IsFalse(crystals.enabled);
            Assert.IsFalse(level.enabled);
            Assert.IsTrue(retry.gameObject.activeSelf);
            Assert.IsTrue(retry.interactable);
        }

        [Test]
        public void TappingACardMarksItLocksTheOthersAndCancelReleasesThem()
        {
            view.Tick(0);

            cards[Casual2V2].onClick.Invoke();
            view.Tick(0);

            Assert.AreEqual(1, proofs.Count);
            Assert.IsTrue(busy[Casual2V2].activeSelf);
            for (var card = 0; card < cards.Length; card++) Assert.IsFalse(cards[card].interactable, $"card {card}");
            Assert.IsTrue(cancel.gameObject.activeSelf);

            cancel.onClick.Invoke();
            view.Tick(0);

            Assert.IsFalse(busy[Casual2V2].activeSelf);
            Assert.IsTrue(cards[Casual2V2].interactable);
            Assert.IsFalse(cancel.gameObject.activeSelf);
        }

        [Test]
        public void ASearchShowsItsElapsedTimeOncePerSecond()
        {
            queue.Read(new QueueState { Status = QueueStatus.Searching, Mode = Mode._1V1, Queue = QueueKind.Ranked, JoinedUnixMs = ServerNowMs, SearchStartsUnixMs = ServerNowMs });

            view.Tick(65.2);

            Assert.IsTrue(countdown.enabled);
            Assert.AreEqual(Duration(65), countdown.text);

            view.Tick(66.1);

            Assert.AreEqual(Duration(66), countdown.text);

            queue.Read(new QueueState { Status = QueueStatus.Left });
            view.Tick(67);

            Assert.IsFalse(countdown.enabled);
        }

        [Test]
        public void ARefusedJoinShowsItsTextAndTheBanTimeLeft()
        {
            queue.Read(new QueueState { Status = QueueStatus.Refused, Mode = Mode._1V1, Queue = QueueKind.Ranked, Refusal = QueueRefusal.RankedBan, RefusedUntilUnixMs = ServerNowMs + 90_000 });

            view.Tick(0);

            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(QueueRefusalKeys.RankedBan), message.text);
            Assert.IsTrue(countdown.enabled);
            Assert.AreEqual(Duration(90), countdown.text);
        }

        [Test]
        public void RetryLoadsTheHeaderOnceAndTheLoadedHeaderHidesIt()
        {
            view.Tick(0);

            retry.onClick.Invoke();
            retry.onClick.Invoke();
            view.Tick(0);

            Assert.AreEqual(2, api.Requests.Count);
            Assert.IsFalse(retry.interactable);

            api.Last("/wallet").Done(new HttpReply(200, WalletJson));
            api.Last("/profile/me").Done(new HttpReply(200, OwnJson));
            view.Tick(0);

            Assert.IsTrue(coins.enabled);
            Assert.AreEqual(MessageFormat.Number(1500, strings.Language), coins.text);
            Assert.AreEqual(MessageFormat.Number(20, strings.Language), crystals.text);
            Assert.AreEqual(MessageFormat.Number(7, strings.Language), level.text);
            Assert.IsFalse(retry.gameObject.activeSelf);
        }

        [Test]
        public void AFailedHeaderLoadLeavesRetryAvailable()
        {
            retry.onClick.Invoke();

            api.Last("/wallet").Done(new HttpReply(503, ""));
            api.Last("/profile/me").Done(new HttpReply(503, ""));
            view.Tick(0);

            Assert.IsTrue(retry.gameObject.activeSelf);
            Assert.IsTrue(retry.interactable);
        }

        [Test]
        public void TheShownTextFollowsALanguageChange()
        {
            queue.Read(new QueueState { Status = QueueStatus.Refused, Mode = Mode._1V1, Queue = QueueKind.Ranked, Refusal = QueueRefusal.RankedBan });
            view.Tick(0);

            Assert.IsTrue(choice.Choose("az"));
            view.Tick(0);

            Assert.AreEqual(BuiltinStrings.Load().Value(QueueRefusalKeys.RankedBan, "az"), message.text);
        }
    }
}
