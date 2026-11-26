using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class FriendRoomEmojiViewTests
    {
        const ulong Nonce = 0x5EED_0000_0000_0279;
        const string Leader = "acc-leader";
        const string Friend = "acc-friend";
        const string Code = "K7QM3X";
        const string Laugh = "emoji.laugh";
        const string Cry = "emoji.cry";
        const string LaughStill = "cosmetic/emoji/laugh/static";
        const string CryStill = "cosmetic/emoji/cry/static";
        const int SlotButtons = 3;
        const string TwoWorn = "{\"items\":[],\"loadout\":{\"emojis\":[\"emoji.laugh\",\"emoji.cry\"]},\"emoji_slots\":2}";
        const string FourWornOfFive = "{\"items\":[],\"loadout\":{\"emojis\":[\"emoji.laugh\",\"emoji.cry\",\"emoji.laugh\",\"emoji.cry\"]},\"emoji_slots\":5}";
        const string Catalog ="{\"schemaVersion\":1,\"manifestVersion\":1,\"items\":["
            + "{\"id\":\"emoji.laugh\",\"type\":\"emoji\",\"availability\":\"live\",\"staticAddress\":\"cosmetic/emoji/laugh/static\"},"
            + "{\"id\":\"emoji.cry\",\"type\":\"emoji\",\"availability\":\"live\",\"staticAddress\":\"cosmetic/emoji/cry/static\"},"
            + "{\"id\":\"emoji.gone\",\"type\":\"emoji\",\"availability\":\"disabled\",\"staticAddress\":\"cosmetic/emoji/gone/static\"},"
            + "{\"id\":\"avatar.fox\",\"type\":\"avatar\",\"rarityId\":\"common\",\"availability\":\"live\",\"staticAddress\":\"cosmetic/avatar/fox/static\"}"
            + "]}";

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

        sealed class Api : IApiTransport
        {
            public readonly Dictionary<string, Action<HttpReply>> Waiting = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) => Waiting[pathAndQuery] = done;
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

        sealed class Source<T> : ICosmeticArtSource<T> where T : class
        {
            public readonly List<string> Loads = new();

            public void Load(string address, Action<T> done) => Loads.Add(address);

            public void Release(string address)
            {
            }
        }

        static readonly string[] NoLocales = { };

        GameObject root;
        GameObject lobbyPanel;
        RoomMemberCell[] members;
        CanvasGroup emojiRoot;
        Button wheel;
        GameObject wheelPanel;
        Button[] slots;
        FriendRoomView view;
        Source<Sprite> stills;
        FakeTransport transport;
        Session session;
        FriendRoom room;
        FriendLobbyLink link;
        Api api;
        long now;

        [SetUp]
        public void BuildView()
        {
            now = 1000;
            root = new GameObject("Friend room", typeof(RectTransform), typeof(Canvas));
            lobbyPanel = Child("Lobby", root.transform);
            var teamsPanel = Child("Teams", root.transform);
            members = Cells("Member", FriendRoom.Capacity, lobbyPanel.transform);
            emojiRoot = Child("Emoji", lobbyPanel.transform).AddComponent<CanvasGroup>();
            wheel = NewButton("Wheel button", emojiRoot.transform);
            wheelPanel = Child("Wheel", emojiRoot.transform);
            slots = new Button[SlotButtons];
            var slotArt = new CosmeticImage[SlotButtons];
            for (var at = 0; at < SlotButtons; at++)
            {
                slots[at] = NewButton("Slot " + at, wheelPanel.transform);
                slotArt[at] = NewArt("Art", slots[at].transform);
            }
            var invitePanel = Child("Invites", root.transform);
            var invites = Child("Invite view", root.transform).AddComponent<InvitePanelView>();
            invites.Configure(invitePanel, NewList(invitePanel.transform), FriendRow(invitePanel.transform), Text("Room code", invitePanel.transform), NewButton("Close", invitePanel.transform));
            view = Child("View", root.transform).AddComponent<FriendRoomView>();
            view.Configure(lobbyPanel, teamsPanel, members, Text("Code", root.transform), Text("Message", root.transform), NewButton("Leave", root.transform), NewButton("Invite", root.transform), invites);
            view.ConfigureStart(NewButton("1v1", lobbyPanel.transform), NewButton("2v2", lobbyPanel.transform));
            view.ConfigureTeams(Cells("Team A", 2, teamsPanel.transform), Cells("Team B", 2, teamsPanel.transform), Cells("Waiting", 2, teamsPanel.transform), NewButton("Start match", teamsPanel.transform), NewButton("Back", teamsPanel.transform));
            var slotTemplate = NewButton("Slot template", root.transform);
            NewArt("Art", slotTemplate.transform);
            slotTemplate.gameObject.SetActive(false);
            view.ConfigureEmoji(emojiRoot, wheel, wheelPanel, slots, slotArt, slotTemplate);

            var auth = new Auth();
            var store = new Store();
            var authSession = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            authSession.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            transport = new FakeTransport();
            session = new Session();
            room = new FriendRoom();
            session.LobbyStateReceived += state => room.Read(state, Leader);
            var commands = new LobbyCommands();
            api = new Api();
            link = new FriendLobbyLink(room, commands, new ApiClient(api, authSession), Now);
            link.Follow(session);
            stills = new Source<Sprite>();
            var stage = new CosmeticStage(stills, new Source<LoopArt>(), null, null, null, EffectLevel.Low, false);
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var catalog));
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            view.Bind(room, link, commands, session, stage, new ActiveCosmeticManifest(catalog), strings);
            session.Hello(transport, new ClientHello());
            Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce });
            transport.Sent.Clear();
        }

        [TearDown]
        public void DestroyView()
        {
            link.Unfollow();
            Object.DestroyImmediate(root);
        }

        long Now() => now;

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

        static RoomMemberCell[] Cells(string name, int count, Transform parent)
        {
            var cells = new RoomMemberCell[count];
            for (var at = 0; at < count; at++)
            {
                var cell = Child(name + " " + at, parent);
                var avatar = Child("Avatar", cell.transform).AddComponent<AvatarView>();
                avatar.Configure(NewArt("Art", avatar.transform), NewArt("Frame", avatar.transform), null, null);
                cells[at] = cell.AddComponent<RoomMemberCell>();
                cells[at].Configure(avatar, Text("Nickname", cell.transform), Child("Leader", cell.transform), Child("Selected", cell.transform), NewButton("Tap", cell.transform));
                cells[at].ConfigureEmoji(NewArt("Emoji", cell.transform));
            }
            return cells;
        }

        static FriendInviteRow FriendRow(Transform parent)
        {
            var row = Child("Friend row", parent);
            ((RectTransform)row.transform).sizeDelta = new Vector2(600f, 80f);
            var marks = new GameObject[(int)FriendInviteState.RoomClosed + 1];
            for (var at = 0; at < marks.Length; at++) marks[at] = Child("State " + at, row.transform);
            var friendRow = row.AddComponent<FriendInviteRow>();
            friendRow.Configure(Text("Nickname", row.transform), NewButton("Invite", row.transform), marks);
            return friendRow;
        }

        static RecyclingList NewList(Transform parent)
        {
            var holder = new GameObject("Friends", typeof(RectTransform), typeof(ScrollRect));
            holder.transform.SetParent(parent, false);
            ((RectTransform)holder.transform).sizeDelta = new Vector2(600f, 800f);
            var viewport = Stretched("Viewport", holder.transform);
            var content = Stretched("Content", viewport);
            var scroll = holder.GetComponent<ScrollRect>();
            scroll.viewport = viewport;
            scroll.content = content;
            scroll.horizontal = false;
            return holder.AddComponent<RecyclingList>();
        }

        static RectTransform Stretched(string name, Transform parent)
        {
            var rect = new GameObject(name, typeof(RectTransform)).GetComponent<RectTransform>();
            rect.SetParent(parent, false);
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.offsetMin = Vector2.zero;
            rect.offsetMax = Vector2.zero;
            return rect;
        }

        static bool ShowsEmoji(RoomMemberCell cell) => cell.transform.Find("Emoji").gameObject.activeSelf;

        void Deliver(Envelope.PayloadOneofCase payload, IMessage message)
        {
            transport.Deliver(payload, message);
            session.Poll();
        }

        void Lobby(LobbyPhase phase = LobbyPhase.Lobby, params string[] accounts)
        {
            var state = new LobbyState { Code = Code, LeaderAccountId = Leader, Phase = phase, Mode = Mode._1V1 };
            state.MemberAccountIds.Add(accounts.Length == 0 ? new[] { Leader, Friend } : accounts);
            Deliver(Envelope.PayloadOneofCase.LobbyState, state);
        }

        void WearTwo() => api.Waiting[CosmeticCollection.InventoryPath](new HttpReply(200, TwoWorn));

        void Relay(string account, string item) =>
            Deliver(Envelope.PayloadOneofCase.LobbyRelay, new LobbyRelay { AccountId = account, ItemId = item });

        [Test]
        public void WithNoEmojiWornTheWheelButtonIsHidden()
        {
            Lobby();

            Assert.IsTrue(lobbyPanel.activeSelf);
            Assert.IsFalse(emojiRoot.gameObject.activeSelf);
        }

        [Test]
        public void TheWheelButtonShowsInTheLobbyOnceEmojiAreWorn()
        {
            Lobby();

            WearTwo();

            Assert.IsTrue(emojiRoot.gameObject.activeSelf);
            Assert.IsFalse(wheelPanel.activeSelf);
        }

        [Test]
        public void TeamSetupHidesTheWheelButton()
        {
            WearTwo();
            Lobby();

            Lobby(LobbyPhase.Teams);

            Assert.IsFalse(emojiRoot.gameObject.activeSelf);
        }

        [Test]
        public void AnOpenWheelIsClosedWhenTheLobbyPhaseEnds()
        {
            WearTwo();
            Lobby();
            wheel.onClick.Invoke();

            Lobby(LobbyPhase.Teams);
            Lobby();

            Assert.IsTrue(emojiRoot.gameObject.activeSelf);
            Assert.IsFalse(wheelPanel.activeSelf);
        }

        [Test]
        public void TheWheelButtonOpensTheWornSlotsWithTheirStaticArt()
        {
            WearTwo();
            Lobby();

            wheel.onClick.Invoke();

            Assert.IsTrue(wheelPanel.activeSelf);
            Assert.IsTrue(slots[0].gameObject.activeSelf);
            Assert.IsTrue(slots[1].gameObject.activeSelf);
            Assert.IsFalse(slots[2].gameObject.activeSelf);
            CollectionAssert.Contains(stills.Loads, LaughStill);
            CollectionAssert.Contains(stills.Loads, CryStill);

            wheel.onClick.Invoke();

            Assert.IsFalse(wheelPanel.activeSelf);
        }

        [Test]
        public void ASlotTapSendsItsEmojiClosesTheWheelAndShowsNothingYet()
        {
            WearTwo();
            Lobby();
            wheel.onClick.Invoke();

            slots[1].onClick.Invoke();

            Assert.AreEqual(1, transport.Sent.Count);
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[0]);
            Assert.AreEqual(Envelope.PayloadOneofCase.EmojiIntent, envelope.PayloadCase);
            Assert.AreEqual(Cry, envelope.EmojiIntent.ItemId);
            Assert.IsFalse(wheelPanel.activeSelf);
            Assert.IsFalse(ShowsEmoji(members[0]));
        }

        [Test]
        public void TheWheelShowsDisabledOpacityUntilTheCalmEnds()
        {
            WearTwo();
            Lobby();
            wheel.onClick.Invoke();
            slots[0].onClick.Invoke();

            view.Tick();
            Assert.AreEqual(TokenOpacity.Disabled, emojiRoot.alpha);

            now += TokenEffect.EmojiDurationMs;
            view.Tick();

            Assert.AreEqual(1f, emojiRoot.alpha);
        }

        [Test]
        public void AnEmojiShownAppearsOverTheSendersAvatarForItsDuration()
        {
            WearTwo();
            Lobby();

            Relay(Friend, Laugh);

            Assert.IsTrue(ShowsEmoji(members[1]));
            Assert.IsFalse(ShowsEmoji(members[0]));
            CollectionAssert.Contains(stills.Loads, LaughStill);

            now += TokenEffect.EmojiDurationMs - 1;
            view.Tick();
            Assert.IsTrue(ShowsEmoji(members[1]));

            now += 1;
            view.Tick();
            Assert.IsFalse(ShowsEmoji(members[1]));
        }

        [TestCase("avatar.fox")]
        [TestCase("emoji.gone")]
        [TestCase("emoji.unknown")]
        public void AnItemNotRenderedAsAnEmojiIsDropped(string item)
        {
            WearTwo();
            Lobby();

            Relay(Friend, item);

            Assert.IsFalse(ShowsEmoji(members[1]));
        }

        [Test]
        public void AMemberWhoLeavesTakesTheirEmojiAlong()
        {
            WearTwo();
            Lobby();
            Relay(Friend, Laugh);

            Lobby(LobbyPhase.Lobby, Leader);
            view.Tick();

            Assert.IsFalse(ShowsEmoji(members[1]));
        }

        [Test]
        public void MoreSlotsThanButtonsBuildTheMissingButtonsFromTheTemplateOnce()
        {
            api.Waiting[CosmeticCollection.InventoryPath](new HttpReply(200, FourWornOfFive));
            Lobby();

            wheel.onClick.Invoke();
            wheel.onClick.Invoke();
            wheel.onClick.Invoke();

            var built = wheelPanel.GetComponentsInChildren<Button>(true);
            Assert.AreEqual(5, built.Length);
            for (var at = 0; at < 4; at++) Assert.IsTrue(built[at].gameObject.activeSelf);
            Assert.IsFalse(built[4].gameObject.activeSelf);

            built[3].onClick.Invoke();

            Assert.AreEqual(1, transport.Sent.Count);
            Assert.AreEqual(Cry, Envelope.Parser.ParseFrom(transport.Sent[0]).EmojiIntent.ItemId);
            Assert.IsFalse(wheelPanel.activeSelf);
        }

        [Test]
        public void TickingTheShownEmojiAllocatesNothing()
        {
            WearTwo();
            Lobby();
            Relay(Friend, Laugh);
            TestDelegate tick = view.Tick;
            tick();

            Assert.That(tick, Allocates.Nothing);
        }
    }
}
