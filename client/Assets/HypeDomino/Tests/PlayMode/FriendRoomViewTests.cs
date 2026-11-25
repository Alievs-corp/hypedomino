using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Friends;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Profile;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class FriendRoomViewTests
    {
        const ulong Nonce = 0x5EED_0000_0000_0848;
        const string Leader = "acc-leader";
        const string Friend = "acc-friend";
        const string Code = "K7QM3X";
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""}]}";
        const string Friends = "{\"friends\":[{\"account_id\":\"acc-third\",\"nickname\":\"Nara\",\"since\":\"2026-11-01T10:00:00Z\",\"presence\":\"online\"},{\"account_id\":\"acc-fourth\",\"nickname\":\"Elvin\",\"since\":\"2026-11-01T10:00:00Z\",\"presence\":\"offline\"}]}";

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
            public string Path;
            public Action<HttpReply> Done;
        }

        sealed class Api : IApiTransport
        {
            public readonly List<Sent> Requests = new();

            public void Send(string method, string pathAndQuery, string body, string authorization, string mac, Action<HttpReply> done) =>
                Requests.Add(new Sent { Path = pathAndQuery, Done = done });

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
        GameObject lobbyPanel;
        GameObject teamsPanel;
        RoomMemberCell[] members;
        RoomMemberCell[] teamA;
        RoomMemberCell[] teamB;
        RoomMemberCell[] waiting;
        TMP_Text code;
        TMP_Text message;
        Button leave;
        Button invite;
        Button startOneVsOne;
        Button startTwoVsTwo;
        Button startMatch;
        Button back;
        GameObject invitePanel;
        TMP_Text inviteCode;
        FriendRoomView view;
        Api api;
        FakeTransport transport;
        Session session;
        FriendRoom room;
        FriendLobbyLink link;
        GameStrings strings;
        string local;

        [SetUp]
        public void BuildView()
        {
            local = Leader;
            root = new GameObject("Friend room", typeof(RectTransform), typeof(Canvas));
            lobbyPanel = Child("Lobby", root.transform);
            teamsPanel = Child("Teams", root.transform);
            members = Cells("Member", FriendRoom.Capacity, lobbyPanel.transform);
            teamA = Cells("Team A", 2, teamsPanel.transform);
            teamB = Cells("Team B", 2, teamsPanel.transform);
            waiting = Cells("Waiting", 2, teamsPanel.transform);
            code = Text("Code", root.transform);
            message = Text("Message", root.transform);
            leave = NewButton("Leave", root.transform);
            invite = NewButton("Invite", root.transform);
            startOneVsOne = NewButton("1v1", lobbyPanel.transform);
            startTwoVsTwo = NewButton("2v2", lobbyPanel.transform);
            startMatch = NewButton("Start match", teamsPanel.transform);
            back = NewButton("Back", teamsPanel.transform);
            invitePanel = Child("Invites", root.transform);
            inviteCode = Text("Room code", invitePanel.transform);
            var invites = Child("Invite view", root.transform).AddComponent<InvitePanelView>();
            invites.Configure(invitePanel, NewList(invitePanel.transform), FriendRow(invitePanel.transform), inviteCode, NewButton("Close", invitePanel.transform));
            view = Child("View", root.transform).AddComponent<FriendRoomView>();
            view.Configure(lobbyPanel, teamsPanel, members, code, message, leave, invite, invites);
            view.ConfigureStart(startOneVsOne, startTwoVsTwo);
            view.ConfigureTeams(teamA, teamB, waiting, startMatch, back);
            var emojiRoot = Child("Emoji", lobbyPanel.transform).AddComponent<CanvasGroup>();
            var wheelPanel = Child("Wheel", emojiRoot.transform);
            var slot = NewButton("Slot", wheelPanel.transform);
            view.ConfigureEmoji(emojiRoot, NewButton("Wheel button", emojiRoot.transform), wheelPanel, new[] { slot }, new[] { NewArt("Art", slot.transform) });

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
            session.LobbyStateReceived += state => room.Read(state, local);
            var commands = new LobbyCommands();
            link = new FriendLobbyLink(room, commands, new ApiClient(api, authSession), () => 500);
            link.Follow(session);
            var stage = new CosmeticStage(new FakeArtSource<Sprite>(), new FakeArtSource<LoopArt>(), null, null, null, EffectLevel.Medium, false);
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var catalog));
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            view.Bind(room, link, commands, session, stage, new ActiveCosmeticManifest(catalog), strings);
            session.Hello(transport, new ClientHello());
            Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce });
            transport.Sent.Clear();
        }

        [TearDown]
        public void DestroyView()
        {
            link.Unfollow();
            UnityEngine.Object.Destroy(root);
        }

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

        void Deliver(Envelope.PayloadOneofCase payload, IMessage message)
        {
            transport.Deliver(payload, message);
            session.Poll();
        }

        void Lobby(LobbyRefusal refusal = LobbyRefusal.Unspecified)
        {
            var state = new LobbyState { Code = Code, LeaderAccountId = Leader, Phase = LobbyPhase.Lobby, Refusal = refusal };
            state.MemberAccountIds.Add(new[] { Leader, Friend });
            Deliver(Envelope.PayloadOneofCase.LobbyState, state);
        }

        void Teams(string seatY)
        {
            var state = new LobbyState { Code = Code, LeaderAccountId = Leader, Phase = LobbyPhase.Teams, Mode = Mode._1V1 };
            state.MemberAccountIds.Add(new[] { Leader, Friend });
            state.Seats.Add(new LobbySeat { Seat = Seat.X, AccountId = Leader });
            state.Seats.Add(new LobbySeat { Seat = Seat.Y, AccountId = seatY });
            Deliver(Envelope.PayloadOneofCase.LobbyState, state);
        }

        void LoadProfile(string account, string nickname) =>
            api.To(PlayerProfile.PublicPath(account)).Done(new HttpReply(200, "{\"account_id\":\"" + account + "\",\"nickname\":\"" + nickname + "\",\"level\":3,\"loadout\":{\"avatar\":\"avatar.fox\"},\"stats\":{},\"ranked\":{\"ranks\":{}},\"medals\":[]}"));

        LobbyIntent LastIntent()
        {
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[transport.Sent.Count - 1]);
            Assert.AreEqual(Envelope.PayloadOneofCase.LobbyIntent, envelope.PayloadCase);
            return envelope.LobbyIntent;
        }

        static bool LeaderMarked(RoomMemberCell cell) => cell.transform.Find("Leader").gameObject.activeSelf;

        static bool Selected(RoomMemberCell cell) => cell.transform.Find("Selected").gameObject.activeSelf;

        static string Nickname(RoomMemberCell cell) => cell.transform.Find("Nickname").GetComponent<TMP_Text>().text;

        [Test]
        public void OutsideARoomNeitherPanelShows()
        {
            Assert.IsFalse(lobbyPanel.activeSelf);
            Assert.IsFalse(teamsPanel.activeSelf);
            Assert.IsFalse(invitePanel.activeSelf);
        }

        [Test]
        public void TheLobbyShowsEachMemberWithTheLeaderMarked()
        {
            Lobby();
            LoadProfile(Leader, "Aysel");
            LoadProfile(Friend, "Murad");

            Assert.IsTrue(lobbyPanel.activeSelf);
            Assert.IsFalse(teamsPanel.activeSelf);
            Assert.AreEqual(Code, code.text);
            Assert.AreEqual("Aysel", Nickname(members[0]));
            Assert.AreEqual("Murad", Nickname(members[1]));
            Assert.IsTrue(LeaderMarked(members[0]));
            Assert.IsFalse(LeaderMarked(members[1]));
            Assert.IsTrue(members[0].transform.Find("Avatar").gameObject.activeSelf);
            Assert.IsFalse(members[2].gameObject.activeSelf);
            Assert.IsFalse(members[3].gameObject.activeSelf);
            Assert.IsFalse(message.enabled);
        }

        [Test]
        public void AMemberWithoutAProfileYetShowsNoAvatar()
        {
            Lobby();

            Assert.IsTrue(members[1].gameObject.activeSelf);
            Assert.IsFalse(members[1].transform.Find("Avatar").gameObject.activeSelf);
            Assert.IsFalse(members[1].transform.Find("Nickname").GetComponent<TMP_Text>().enabled);
        }

        [Test]
        public void TheLeaderStartsTeamSetupInTheChosenMode()
        {
            Lobby();
            Assert.IsTrue(startTwoVsTwo.gameObject.activeSelf);

            startTwoVsTwo.onClick.Invoke();

            var intent = LastIntent();
            Assert.AreEqual(LobbyCommand.Teams, intent.Command);
            Assert.AreEqual(Mode._2V2, intent.Mode);
        }

        [Test]
        public void AMemberWhoIsNotTheLeaderHasNoStart()
        {
            local = Friend;
            Lobby();

            Assert.IsFalse(startOneVsOne.gameObject.activeSelf);
            Assert.IsFalse(startTwoVsTwo.gameObject.activeSelf);
            startOneVsOne.onClick.Invoke();
            Assert.IsEmpty(transport.Sent);
        }

        [Test]
        public void LeaveSendsTheLeaveIntent()
        {
            Lobby();

            leave.onClick.Invoke();

            Assert.AreEqual(LobbyCommand.Leave, LastIntent().Command);
        }

        [Test]
        public void TheInvitePanelListsFriendsByStateAndInvitesOnce()
        {
            Lobby();
            api.To(FriendList.Path).Done(new HttpReply(200, Friends));

            invite.onClick.Invoke();

            Assert.IsTrue(invitePanel.activeSelf);
            Assert.AreEqual(Code, inviteCode.text);
            var rows = invitePanel.GetComponentsInChildren<FriendInviteRow>();
            var nara = Array.Find(rows, row => row.transform.Find("Nickname").GetComponent<TMP_Text>().text == "Nara");
            var elvin = Array.Find(rows, row => row.transform.Find("Nickname").GetComponent<TMP_Text>().text == "Elvin");
            Assert.IsTrue(elvin.transform.Find("State " + (int)FriendInviteState.Offline).gameObject.activeSelf);
            Assert.IsFalse(elvin.transform.Find("Invite").GetComponent<Button>().interactable);
            var naraInvite = nara.transform.Find("Invite").GetComponent<Button>();
            Assert.IsTrue(naraInvite.interactable);

            naraInvite.onClick.Invoke();

            var intent = LastIntent();
            Assert.AreEqual(LobbyCommand.Invite, intent.Command);
            Assert.AreEqual("acc-third", intent.AccountId);
            Assert.IsFalse(naraInvite.interactable);
            Assert.IsTrue(nara.transform.Find("State " + (int)FriendInviteState.Invited).gameObject.activeSelf);
        }

        [Test]
        public void TeamSetupShowsSeatsAndTheMembersWithoutOne()
        {
            Teams(string.Empty);
            LoadProfile(Leader, "Aysel");
            LoadProfile(Friend, "Murad");

            Assert.IsFalse(lobbyPanel.activeSelf);
            Assert.IsTrue(teamsPanel.activeSelf);
            Assert.AreEqual("Aysel", Nickname(teamA[0]));
            Assert.IsFalse(teamA[1].gameObject.activeSelf);
            Assert.IsTrue(teamB[0].gameObject.activeSelf);
            Assert.IsFalse(teamB[0].transform.Find("Avatar").gameObject.activeSelf);
            Assert.AreEqual("Murad", Nickname(waiting[0]));
            Assert.IsFalse(waiting[1].gameObject.activeSelf);
            Assert.IsTrue(startMatch.gameObject.activeSelf);
            Assert.IsFalse(startMatch.interactable);
            Assert.IsTrue(back.gameObject.activeSelf);
        }

        [Test]
        public void TheLeaderPlacesAWaitingMemberIntoATappedSeat()
        {
            Teams(string.Empty);

            waiting[0].Tap.onClick.Invoke();
            Assert.IsTrue(Selected(waiting[0]));
            Assert.IsEmpty(transport.Sent);

            teamB[0].Tap.onClick.Invoke();

            var intent = LastIntent();
            Assert.AreEqual(LobbyCommand.Place, intent.Command);
            Assert.AreEqual(Friend, intent.AccountId);
            Assert.AreEqual(Seat.Y, intent.Seat);
            Assert.IsFalse(Selected(waiting[0]));
        }

        [Test]
        public void StartMatchAndBackFollowTheServerState()
        {
            Teams(Friend);
            Assert.IsTrue(startMatch.interactable);

            startMatch.onClick.Invoke();
            Assert.AreEqual(LobbyCommand.Start, LastIntent().Command);

            back.onClick.Invoke();
            Assert.AreEqual(LobbyCommand.Back, LastIntent().Command);
        }

        [Test]
        public void AMemberAsksForASeatInsteadOfPlacing()
        {
            local = Friend;
            Teams(string.Empty);

            Assert.IsFalse(startMatch.gameObject.activeSelf);
            Assert.IsFalse(back.gameObject.activeSelf);
            teamB[0].Tap.onClick.Invoke();

            var intent = LastIntent();
            Assert.AreEqual(LobbyCommand.RequestSeat, intent.Command);
            Assert.AreEqual(Seat.Y, intent.Seat);
        }

        [Test]
        public void ARefusalShowsItsText()
        {
            Lobby(LobbyRefusal.SeatsOpen);

            Assert.IsTrue(message.enabled);
            Assert.AreEqual(strings.Text(LobbyRefusalKeys.SeatsOpen), message.text);
        }
    }
}
