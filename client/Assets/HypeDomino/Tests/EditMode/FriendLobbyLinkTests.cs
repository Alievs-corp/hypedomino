using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Friends;
using HypeDomino.Game.Inventory;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Profile;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class FriendLobbyLinkTests
    {
        const ulong Nonce = 0x5EED_0000_0000_0279;
        const string Leader = "acc-leader";
        const string Friend = "acc-friend";
        const string OnlineFriends = "{\"friends\":[{\"account_id\":\"acc-third\",\"nickname\":\"Nara\",\"since\":\"2026-11-01T10:00:00Z\",\"presence\":\"online\"}]}";
        const string TwoWorn = @"{""items"":[],""loadout"":{""emojis"":[""emoji.laugh"",""emoji.cry""]},""emoji_slots"":2}";

        sealed class FakeTransport : IFrameTransport
        {
            readonly Queue<byte[]> incoming = new();

            public List<byte[]> Sent { get; } = new();

            public void Send(ReadOnlySpan<byte> frame) => Sent.Add(frame.ToArray());

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

            public int CountTo(string path) => Requests.FindAll(request => request.PathAndQuery == path).Count;

            public Sent LastTo(string path) => Requests.FindLast(request => request.PathAndQuery == path);
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

        Api api;
        FakeTransport transport;
        Session session;
        FriendRoom room;
        FriendLobbyLink link;
        List<MemberEmoji> shown;
        List<InventoryLoadOutcome> outcomes;
        long nowMs;

        [SetUp]
        public void Build()
        {
            nowMs = 500;
            api = new Api();
            var auth = new Auth();
            var store = new Store();
            var authSession = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            authSession.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            transport = new FakeTransport();
            session = new Session();
            room = new FriendRoom();
            session.LobbyStateReceived += ReadRoom;
            link = new FriendLobbyLink(room, new LobbyCommands(), new ApiClient(api, authSession), Now);
            shown = new List<MemberEmoji>();
            outcomes = new List<InventoryLoadOutcome>();
            link.EmojiShown += shown.Add;
            link.InventoryLoaded += Loaded;
            link.Follow(session);
        }

        [TearDown]
        public void Release() => link.Unfollow();

        long Now() => nowMs;

        void ReadRoom(LobbyState state) => room.Read(state, Leader);

        void Loaded(InventoryLoadResult result) => outcomes.Add(result.Outcome);

        void Hello(HelloStatus status = HelloStatus.Accepted)
        {
            session.Hello(transport, new ClientHello());
            Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = status, SessionNonce = Nonce });
            transport.Sent.Clear();
        }

        void Deliver(Envelope.PayloadOneofCase payload, IMessage message)
        {
            transport.Deliver(payload, message);
            session.Poll();
        }

        void Enter(string code = "K7QM3X")
        {
            var state = new LobbyState { Code = code, LeaderAccountId = Leader, Phase = LobbyPhase.Lobby, Mode = Mode._1V1 };
            state.MemberAccountIds.Add(new[] { Leader, Friend });
            Deliver(Envelope.PayloadOneofCase.LobbyState, state);
        }

        void Leave() => Deliver(Envelope.PayloadOneofCase.LobbyState, new LobbyState());

        void Relay(string account, string item) =>
            Deliver(Envelope.PayloadOneofCase.LobbyRelay, new LobbyRelay { AccountId = account, ItemId = item });

        void AnswerInventory(string json) => api.LastTo(CosmeticCollection.InventoryPath).Done(new HttpReply(200, json));

        void FailInventory(int status, string body = "", int retryAfterSeconds = 0) =>
            api.LastTo(CosmeticCollection.InventoryPath).Done(new HttpReply(status, body, retryAfterSeconds));

        int InventoryRequests() => api.CountTo(CosmeticCollection.InventoryPath);

        void Wait(long ms)
        {
            nowMs += ms;
            link.Tick();
        }

        static string Worn(int slots, params string[] emojis) =>
            "{\"items\":[],\"loadout\":{\"emojis\":[" + (emojis.Length == 0 ? "" : "\"" + string.Join("\",\"", emojis) + "\"") + "]},\"emoji_slots\":" + slots + "}";

        [Test]
        public void AnAcceptedHelloLoadsTheInventory()
        {
            Hello();

            Assert.AreEqual(1, api.CountTo(CosmeticCollection.InventoryPath));
            AnswerInventory(TwoWorn);
            Assert.AreEqual(new[] { InventoryLoadOutcome.Loaded }, outcomes);
            Assert.IsNotNull(link.Wheel);
            Assert.IsNotNull(link.Emoji);
            CollectionAssert.AreEqual(new[] { "emoji.laugh", "emoji.cry" }, link.Worn.loadout.emojis);
            Assert.AreEqual(2, link.Worn.emoji_slots);
        }

        [TestCase(HelloStatus.UpdateRequired)]
        [TestCase(HelloStatus.Maintenance)]
        public void AHelloThatIsNotAcceptedLoadsNothing(HelloStatus status)
        {
            Hello(status);

            Assert.AreEqual(0, api.CountTo(CosmeticCollection.InventoryPath));
        }

        [Test]
        public void EachRoomStateReachesTheMembersPanel()
        {
            Hello();
            Enter();

            Assert.AreEqual(2, link.Members.All.Count);
            Assert.AreEqual(Leader, link.Members.All[0].Account);
            Assert.IsTrue(link.Members.All[0].IsLeader);
            Assert.IsTrue(link.Members.All[0].IsLocal);
            Assert.AreEqual(Friend, link.Members.All[1].Account);
            Assert.AreEqual(1, api.CountTo(PlayerProfile.PublicPath(Leader)));

            Leave();

            Assert.AreEqual(0, link.Members.All.Count);
        }

        [Test]
        public void ThereIsNoWheelBeforeTheInventory()
        {
            Hello();
            Enter();

            Assert.IsNull(link.Wheel);
            Assert.IsNull(link.Emoji);
            Assert.IsFalse(link.TrySend(0));
            Assert.IsEmpty(transport.Sent);
        }

        [TestCase(404, "{\"error\":\"not_found\"}", InventoryLoadOutcome.Failed)]
        [TestCase(403, "{\"error\":\"forbidden\"}", InventoryLoadOutcome.Failed)]
        [TestCase(401, "{\"error\":\"access_rejected\"}", InventoryLoadOutcome.SignInRequired)]
        [TestCase(200, "not json", InventoryLoadOutcome.Failed)]
        public void ARejectedInventoryLoadIsReportedAtOnceAndNeverRetried(int status, string body, InventoryLoadOutcome outcome)
        {
            Hello();

            FailInventory(status, body);
            Wait(60000);

            Assert.AreEqual(new[] { outcome }, outcomes);
            Assert.AreEqual(1, InventoryRequests());
            Assert.IsNull(link.Wheel);
            Assert.IsNull(link.Worn);
        }

        [TestCase(HttpReply.NoResponse, InventoryLoadOutcome.Offline)]
        [TestCase(500, InventoryLoadOutcome.Failed)]
        [TestCase(503, InventoryLoadOutcome.Failed)]
        [TestCase(429, InventoryLoadOutcome.RateLimited)]
        public void AFailedLoadIsRetriedThreeTimesAfterTwoFourAndEightSeconds(int status, InventoryLoadOutcome outcome)
        {
            Hello();

            foreach (var backoffMs in new long[] { 2000, 4000, 8000 })
            {
                var asked = InventoryRequests();
                FailInventory(status);
                Wait(backoffMs - 1);
                Assert.AreEqual(asked, InventoryRequests());
                Wait(1);
                Assert.AreEqual(asked + 1, InventoryRequests());
                Assert.IsEmpty(outcomes);
            }
            FailInventory(status);
            Wait(60000);

            Assert.AreEqual(new[] { outcome }, outcomes);
            Assert.AreEqual(4, InventoryRequests());
            Assert.IsNull(link.Wheel);
        }

        [Test]
        public void AnInventoryArrivingOnARetryStopsRetrying()
        {
            Hello();
            FailInventory(503);
            Wait(2000);

            AnswerInventory(TwoWorn);
            Wait(60000);

            Assert.AreEqual(new[] { InventoryLoadOutcome.Loaded }, outcomes);
            Assert.AreEqual(2, InventoryRequests());
            Assert.IsNotNull(link.Wheel);
        }

        [Test]
        public void ARetryAfterLongerThanTheBackoffIsWaitedOut()
        {
            Hello();

            FailInventory(429, "{\"error\":\"rate_limited\"}", 30);
            Wait(29999);
            Assert.AreEqual(1, InventoryRequests());
            Wait(1);

            Assert.AreEqual(2, InventoryRequests());
        }

        [Test]
        public void ARetryAfterShorterThanTheBackoffKeepsTheBackoff()
        {
            Hello();

            FailInventory(429, "{\"error\":\"rate_limited\"}", 1);
            Wait(1999);
            Assert.AreEqual(1, InventoryRequests());
            Wait(1);

            Assert.AreEqual(2, InventoryRequests());
        }

        [Test]
        public void AfterTheRetriesTheNextAcceptedHelloLoadsAgainWithAFreshBackoff()
        {
            Hello();
            FailInventory(503);
            Wait(2000);
            FailInventory(503);
            Wait(4000);
            FailInventory(503);
            Wait(8000);
            FailInventory(503);

            Hello();
            Assert.AreEqual(5, InventoryRequests());
            FailInventory(503);
            Wait(2000);

            Assert.AreEqual(6, InventoryRequests());
            AnswerInventory(TwoWorn);
            Assert.AreEqual(new[] { InventoryLoadOutcome.Failed, InventoryLoadOutcome.Loaded }, outcomes);
        }

        [Test]
        public void AnAcceptedHelloDuringTheBackoffLoadsAtOnce()
        {
            Hello();
            FailInventory(503);
            Wait(1000);

            Hello();

            Assert.AreEqual(2, InventoryRequests());
        }

        [Test]
        public void SigningOutCancelsThePendingRetry()
        {
            Hello();
            FailInventory(503);

            link.SignedOut();
            Wait(60000);

            Assert.AreEqual(1, InventoryRequests());
            Assert.IsEmpty(outcomes);
        }

        [Test]
        public void ARefreshOfTheSameAccountKeepsThePendingRetry()
        {
            link.SignedIn("acc-1");
            Hello();
            FailInventory(503);

            link.SignedIn("acc-1");
            Wait(2000);

            Assert.AreEqual(2, InventoryRequests());
        }

        [Test]
        public void AnotherAccountCancelsThePendingRetry()
        {
            link.SignedIn("acc-1");
            Hello();
            FailInventory(503);

            link.SignedIn("acc-2");
            Wait(60000);

            Assert.AreEqual(1, InventoryRequests());
            Assert.IsEmpty(outcomes);
        }

        [Test]
        public void AnInventoryInFlightAtSignOutIsDropped()
        {
            Hello();

            link.SignedOut();
            AnswerInventory(TwoWorn);

            Assert.IsEmpty(outcomes);
            Assert.IsNull(link.Worn);
        }

        [Test]
        public void AHelloAfterSignOutLoadsAgainOnceTheDroppedAnswerArrives()
        {
            Hello();
            link.SignedOut();
            Hello();
            Assert.AreEqual(1, InventoryRequests());

            AnswerInventory(TwoWorn);
            Assert.AreEqual(2, InventoryRequests());
            AnswerInventory(TwoWorn);

            Assert.AreEqual(new[] { InventoryLoadOutcome.Loaded }, outcomes);
        }

        [Test]
        public void ATickWithNoRetryDueAllocatesNothing()
        {
            Hello();
            FailInventory(503);
            link.Tick();

            Assert.That(link.Tick, Allocates.Nothing);
            link.SignedOut();
            Assert.That(link.Tick, Allocates.Nothing);
        }

        [Test]
        public void EnteringARoomPutsTheWornEmojiOnTheWheel()
        {
            Hello();
            AnswerInventory(TwoWorn);
            Assert.AreEqual(0, link.Wheel.Count);

            Enter();

            Assert.AreEqual(2, link.Wheel.Count);
            Assert.AreEqual("emoji.laugh", link.Wheel.Slot(0));
            Assert.AreEqual("emoji.cry", link.Wheel.Slot(1));
            Assert.IsTrue(link.Emoji.CanSend);
        }

        [Test]
        public void TheWheelIsFilledBeforeTheMembersPanelHearsOfTheRoom()
        {
            Hello();
            AnswerInventory(TwoWorn);
            var countAtChange = -1;
            link.Members.Changed += () => countAtChange = link.Wheel.Count;

            Enter();

            Assert.AreEqual(2, countAtChange);
        }

        [Test]
        public void NowReadsTheLinkClock() => Assert.AreEqual(Now(), link.NowMs);

        [Test]
        public void AnInventoryArrivingInsideTheRoomFillsTheWheelAtOnce()
        {
            Hello();
            Enter();

            AnswerInventory(TwoWorn);

            Assert.AreEqual(2, link.Wheel.Count);
            Assert.AreEqual("emoji.laugh", link.Wheel.Slot(0));
        }

        [Test]
        public void ATapSendsTheSlotOnTheFollowedSession()
        {
            Hello();
            AnswerInventory(TwoWorn);
            Enter();
            link.Wheel.Toggle();

            Assert.IsTrue(link.TrySend(1));

            Assert.AreEqual(1, transport.Sent.Count);
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[0]);
            Assert.AreEqual(Envelope.PayloadOneofCase.EmojiIntent, envelope.PayloadCase);
            Assert.AreEqual("emoji.cry", envelope.EmojiIntent.ItemId);
            Assert.IsTrue(link.Wheel.Calm(Now()));
        }

        [Test]
        public void AMembersRelayIsShownThroughTheLink()
        {
            Hello();
            AnswerInventory(TwoWorn);
            Enter();

            Relay(Friend, "emoji.party");

            Assert.AreEqual(1, shown.Count);
            Assert.AreEqual(Friend, shown[0].Account);
            Assert.AreEqual("emoji.party", shown[0].ItemId);
        }

        [Test]
        public void AnEquipReplyReachesTheWheelInTheNextRoom()
        {
            Hello();
            AnswerInventory(TwoWorn);
            Enter();

            link.Wear(new EquipReply { loadout = new HypeDomino.UI.Cosmetics.LoadoutReply { emojis = new[] { "emoji.party" } } });
            Assert.AreEqual("emoji.laugh", link.Wheel.Slot(0), "the server keeps the emoji the member joined with");

            Leave();
            Enter("Z23456");
            Assert.AreEqual(1, link.Wheel.Count);
            Assert.AreEqual("emoji.party", link.Wheel.Slot(0));
        }

        [Test]
        public void MoreSlotsBuildALargerWheelThatAloneFollowsTheSession()
        {
            Hello();
            AnswerInventory(Worn(1, "emoji.laugh", "emoji.cry", "emoji.party"));
            var first = link.Wheel;

            link.Wear(InventoryReply.FromJson(Worn(3, "emoji.laugh", "emoji.cry", "emoji.party")));
            Enter();

            Assert.AreNotSame(first, link.Wheel);
            Assert.AreEqual(0, first.Count);
            Assert.AreEqual(3, link.Wheel.Count);
            Relay(Friend, "emoji.party");
            Assert.AreEqual(1, shown.Count);
        }

        [Test]
        public void FewerSlotsKeepTheWheel()
        {
            Hello();
            AnswerInventory(TwoWorn);
            var first = link.Wheel;

            link.Wear(InventoryReply.FromJson(Worn(1, "emoji.laugh", "emoji.cry")));
            Enter();

            Assert.AreSame(first, link.Wheel);
            Assert.AreEqual(1, link.Wheel.Count);
        }

        [Test]
        public void UnfollowingStopsReadingTheSession()
        {
            Hello();
            AnswerInventory(TwoWorn);
            link.Unfollow();

            Enter();
            Relay(Friend, "emoji.party");

            Assert.AreEqual(0, link.Members.All.Count);
            Assert.AreEqual(0, link.Wheel.Count);
            Assert.IsEmpty(shown);
            Assert.IsFalse(link.TrySend(0));
        }

        void AnswerFriends(string json) => api.LastTo(FriendList.Path).Done(new HttpReply(200, json));

        [Test]
        public void EnteringARoomLoadsTheFriendsOnceForTheInvitePanel()
        {
            var friendOutcomes = new List<FriendListLoadOutcome>();
            link.FriendsLoaded += friendOutcomes.Add;
            Hello();

            Assert.AreEqual(0, api.CountTo(FriendList.Path));

            Enter();
            Enter();

            Assert.AreEqual(1, api.CountTo(FriendList.Path));
            AnswerFriends(OnlineFriends);
            Assert.AreEqual(new[] { FriendListLoadOutcome.Loaded }, friendOutcomes);
            Assert.AreEqual(1, link.Invites.Friends.Count);
            Assert.AreEqual("K7QM3X", link.Invites.Code);

            Leave();
            Enter();

            Assert.AreEqual(2, api.CountTo(FriendList.Path));
        }

        [Test]
        public void AnInviteIsSentOnceUntilTheRoomChanges()
        {
            Hello();
            Enter();
            AnswerFriends(OnlineFriends);
            var nara = link.Invites.Friends[0];
            var sent = transport.Sent.Count;

            Assert.IsTrue(link.TryInvite(nara));
            Assert.AreEqual(sent + 1, transport.Sent.Count);
            Assert.AreEqual(FriendInviteState.Invited, link.Invites.StateOf(nara));
            Assert.IsFalse(link.TryInvite(nara));
            Assert.AreEqual(sent + 1, transport.Sent.Count);

            Enter();

            Assert.AreEqual(FriendInviteState.Invitable, link.Invites.StateOf(nara));
        }

        [Test]
        public void NoInviteIsSentWithoutASession()
        {
            Hello();
            Enter();
            AnswerFriends(OnlineFriends);
            var nara = link.Invites.Friends[0];
            link.Unfollow();

            Assert.IsFalse(link.TryInvite(nara));
            Assert.AreEqual(FriendInviteState.Invitable, link.Invites.StateOf(nara));
        }
    }
}
