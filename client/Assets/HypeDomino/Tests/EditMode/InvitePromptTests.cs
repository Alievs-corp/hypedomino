using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Profile;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class InvitePromptTests
    {
        const ulong Nonce = 0x5EED_0000_0000_0848;
        const string Local = "acc-local";
        const string Leader = "acc-leader";
        const string Friend = "acc-friend";
        const string Code = "K7QM3X";
        const string OtherCode = "Z23456";

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

        static readonly string[] NoLocales = { };

        GameObject root;
        GameObject panel;
        TMP_Text message;
        Button accept;
        Button decline;
        FakeTransport transport;
        Session session;
        FriendLobbyLink link;
        GameStrings strings;
        Api api;

        [SetUp]
        public void BuildPrompt()
        {
            root = new GameObject("Lobby", typeof(RectTransform), typeof(Canvas));
            panel = Child("Invite", root.transform);
            message = Child("Message", panel.transform).AddComponent<TextMeshProUGUI>();
            accept = NewButton("Accept", panel.transform);
            decline = NewButton("Decline", panel.transform);
            var view = Child("View", root.transform).AddComponent<InvitePromptView>();
            view.Configure(panel, message, accept, decline);

            var auth = new Auth();
            var store = new Store();
            var authSession = new AuthSession(auth, new Identity(), store);
            store.Write(AuthSession.RefreshTokenKey, "refresh-0");
            authSession.Resume(_ => { });
            auth.Refreshes[0](new HttpReply(200, "{\"access_token\":\"access-1\",\"refresh_token\":\"refresh-1\",\"expires_in\":900,\"request_key\":\"" + Convert.ToBase64String(new byte[32]) + "\",\"account_id\":\"acc-1\"}"));
            transport = new FakeTransport();
            session = new Session();
            var room = new FriendRoom();
            session.LobbyStateReceived += state => room.Read(state, Local);
            var commands = new LobbyCommands();
            api = new Api();
            link = new FriendLobbyLink(room, commands, new ApiClient(api, authSession), () => 0);
            link.Follow(session);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load("az", () => NoLocales), true);
            view.Bind(link.Incoming, commands, session, strings);
            session.Hello(transport, new ClientHello());
            Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce });
            transport.Sent.Clear();
        }

        [TearDown]
        public void DestroyPrompt()
        {
            link.Unfollow();
            Object.DestroyImmediate(root);
        }

        static GameObject Child(string name, Transform parent)
        {
            var child = new GameObject(name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child;
        }

        static Button NewButton(string name, Transform parent) => Child(name, parent).AddComponent<Image>().gameObject.AddComponent<Button>();

        static string PublicProfile(string account, string nickname) =>
            "{\"account_id\":\"" + account + "\",\"nickname\":\"" + nickname + "\",\"level\":3,\"loadout\":{\"avatar\":\"avatar.fox\",\"frame\":\"frame.gold\"},"
            + "\"stats\":{\"games\":4,\"win_rate_percent\":50,\"best_streak\":2},\"ranked\":{\"ranks\":{\"1v1\":\"Gold II\"}}}";

        void Deliver(Envelope.PayloadOneofCase payload, IMessage message)
        {
            transport.Deliver(payload, message);
            session.Poll();
        }

        void Invite(string code, string from) =>
            Deliver(Envelope.PayloadOneofCase.LobbyInvite, new LobbyInvite { Code = code, FromAccountId = from });

        void Room(string code)
        {
            var state = new LobbyState { Code = code };
            if (code.Length > 0)
            {
                state.LeaderAccountId = Leader;
                state.Phase = LobbyPhase.Lobby;
                state.MemberAccountIds.Add(new[] { Leader, Local });
            }
            Deliver(Envelope.PayloadOneofCase.LobbyState, state);
        }

        void Profile(string account, string nickname) =>
            api.Waiting[PlayerProfile.PublicPath(account)](new HttpReply(200, PublicProfile(account, nickname)));

        string Named(string nickname) =>
            strings.Text(InvitePromptView.FromKey, new Dictionary<string, object> { [InvitePromptView.NicknameArgument] = nickname });

        LobbyIntent SentIntent() => Envelope.Parser.ParseFrom(transport.Sent[0]).LobbyIntent;

        [Test]
        public void ThePromptIsHiddenUntilAnInviteArrives()
        {
            Assert.IsFalse(panel.activeSelf);
            Assert.IsFalse(link.Incoming.Shown);
        }

        [Test]
        public void AnInviteShowsTheUnnamedLineUntilTheInvitersNicknameArrives()
        {
            Invite(Code, Leader);

            Assert.IsTrue(panel.activeSelf);
            Assert.IsTrue(accept.interactable);
            Assert.AreEqual(strings.Text(InvitePromptView.UnnamedKey), message.text);

            Profile(Leader, "Nara");

            Assert.AreEqual("Nara", link.Incoming.Nickname);
            Assert.AreEqual(Named("Nara"), message.text);
            StringAssert.Contains("Nara", message.text);
        }

        [Test]
        public void AProfileThatFailsKeepsTheUnnamedLine()
        {
            Invite(Code, Leader);

            api.Waiting[PlayerProfile.PublicPath(Leader)](new HttpReply(500, "{}"));

            Assert.IsTrue(panel.activeSelf);
            Assert.IsNull(link.Incoming.Nickname);
            Assert.AreEqual(strings.Text(InvitePromptView.UnnamedKey), message.text);
        }

        [Test]
        public void AcceptSendsJoinWithTheInvitedCodeAndHidesThePrompt()
        {
            Invite(Code, Leader);

            accept.onClick.Invoke();

            Assert.AreEqual(1, transport.Sent.Count);
            Assert.AreEqual(new LobbyIntent { Command = LobbyCommand.Join, Code = Code }, SentIntent());
            Assert.IsFalse(panel.activeSelf);
        }

        [Test]
        public void DeclineSendsNothingAndShowsTheEarlierInviteWithItsInviter()
        {
            Invite(Code, Leader);
            Profile(Leader, "Nara");
            Invite(OtherCode, Friend);
            Assert.AreEqual(OtherCode, link.Incoming.Current.Code);
            Profile(Friend, "Lale");
            Assert.AreEqual(Named("Lale"), message.text);

            decline.onClick.Invoke();

            Assert.IsEmpty(transport.Sent);
            Assert.IsTrue(panel.activeSelf);
            Assert.AreEqual(Code, link.Incoming.Current.Code);
            Assert.AreEqual(strings.Text(InvitePromptView.UnnamedKey), message.text);

            Profile(Leader, "Nara");
            Assert.AreEqual(Named("Nara"), message.text);

            decline.onClick.Invoke();

            Assert.IsFalse(panel.activeSelf);
        }

        [Test]
        public void ANicknameForAnInviterNoLongerShownIsDroppedAndTheShownOneIsAsked()
        {
            Invite(Code, Leader);
            Invite(OtherCode, Friend);

            Profile(Leader, "Nara");

            Assert.IsNull(link.Incoming.Nickname);
            Profile(Friend, "Lale");
            Assert.AreEqual("Lale", link.Incoming.Nickname);
        }

        [Test]
        public void InsideARoomAcceptIsDisabledUntilThePlayerLeaves()
        {
            Room(OtherCode);
            Invite(Code, Leader);

            Assert.IsTrue(panel.activeSelf);
            Assert.IsFalse(accept.interactable);
            accept.onClick.Invoke();
            Assert.IsEmpty(transport.Sent);

            Room(string.Empty);

            Assert.IsTrue(accept.interactable);
        }

        [Test]
        public void EnteringTheInvitedRoomDropsTheInvite()
        {
            Invite(Code, Leader);

            Room(Code);

            Assert.IsFalse(panel.activeSelf);
        }
    }
}
