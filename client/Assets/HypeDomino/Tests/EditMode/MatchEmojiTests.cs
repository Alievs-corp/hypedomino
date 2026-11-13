using System;
using System.Collections.Generic;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Emoji;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Match;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MatchEmojiTests
    {
        const ulong Nonce = 0x5EED_E770_0000_0151;
        static readonly string[] Equipped = { "emoji.laugh", "emoji.cry" };
        static readonly string[] NoLocales = Array.Empty<string>();
        const string Builtin = @"{""entries"":[{""key"":""quick.well_played"",""en"":""Well played"",""az"":""Əla oyun""}]}";

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

            public void Deliver(byte[] frame) => incoming.Enqueue(frame);
        }

        EmojiWheel wheel;
        SeatMessages messages;
        LanguageChoice language;
        MatchEmoji emoji;

        [SetUp]
        public void CreateWheel()
        {
            wheel = new EmojiWheel(Equipped.Length);
            wheel.Equip(Equipped);
            messages = new SeatMessages();
            language = LanguageChoice.Load("az", () => NoLocales);
            emoji = new MatchEmoji(wheel, messages, new GameStrings(BuiltinStrings.Parse(Builtin), language, false));
        }

        static Session AcceptedSession(FakeTransport transport)
        {
            var session = new Session();
            session.Hello(transport, new ClientHello());
            var hello = new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce };
            transport.Deliver(new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, Nonce), Envelope.PayloadOneofCase.ServerHello, hello).ToArray());
            session.Poll();
            Assert.IsTrue(session.Accepted);
            transport.Sent.Clear();
            return session;
        }

        [Test]
        public void ATapOnAnOpenWheelSendsTheSlotAsAnUnsequencedIntent()
        {
            var transport = new FakeTransport();
            var session = AcceptedSession(transport);
            wheel.Toggle();

            Assert.IsTrue(emoji.TrySend(session, 1, 500));

            Assert.AreEqual(1, transport.Sent.Count);
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[0]);
            Assert.AreEqual(Envelope.PayloadOneofCase.EmojiIntent, envelope.PayloadCase);
            Assert.AreEqual("emoji.cry", envelope.EmojiIntent.ItemId);
            Assert.AreEqual(0ul, envelope.Sequence);
            Assert.IsTrue(wheel.Calm(500));
        }

        [Test]
        public void NothingIsSentWhileCalmOrWithoutAnAcceptedSession()
        {
            var transport = new FakeTransport();
            wheel.Toggle();
            Assert.IsFalse(emoji.TrySend(new Session(), 0, 0));
            Assert.IsTrue(wheel.IsOpen, "a tap without a session keeps the wheel open and calm free");
            Assert.IsFalse(wheel.Calm(0));

            var session = AcceptedSession(transport);
            Assert.IsTrue(emoji.TrySend(session, 0, 0));
            wheel.Toggle();
            Assert.IsFalse(emoji.TrySend(session, 0, TokenEffect.EmojiDurationMs - 1));
            Assert.AreEqual(1, transport.Sent.Count);
        }

        [TestCase(Seat.South, Seat.South, TableSide.Bottom)]
        [TestCase(Seat.South, Seat.West, TableSide.Left)]
        [TestCase(Seat.South, Seat.North, TableSide.Top)]
        [TestCase(Seat.South, Seat.East, TableSide.Right)]
        [TestCase(Seat.East, Seat.South, TableSide.Left)]
        [TestCase(Seat.Unspecified, Seat.North, TableSide.Top)]
        public void AFourSeatEmojiPopsAtTheSendersSide(Seat local, Seat sender, TableSide expected)
        {
            emoji.Begin(local, Mode._2V2);
            Assert.IsTrue(emoji.TryPlace(sender, 0, out var side));
            Assert.AreEqual(expected, side);
        }

        [TestCase(Seat.X, Seat.X, TableSide.Bottom)]
        [TestCase(Seat.X, Seat.Y, TableSide.Top)]
        [TestCase(Seat.Y, Seat.X, TableSide.Top)]
        public void ATwoSeatEmojiPopsAtTheSendersSide(Seat local, Seat sender, TableSide expected)
        {
            emoji.Begin(local, Mode._1V1);
            Assert.IsTrue(emoji.TryPlace(sender, 0, out var side));
            Assert.AreEqual(expected, side);
        }

        [TestCase(Mode._1V1, Seat.North)]
        [TestCase(Mode._2V2, Seat.Y)]
        [TestCase(Mode._2V2, Seat.Unspecified)]
        [TestCase(Mode._2V2, (Seat)7)]
        [TestCase(Mode.Unspecified, Seat.North)]
        public void ASeatOutsideTheModeIsDropped(Mode mode, Seat sender)
        {
            emoji.Begin(Seat.Unspecified, mode);
            Assert.IsFalse(emoji.TryPlace(sender, 0, out _));
        }

        [Test]
        public void OnlyTheOwnEmojiArrivingRestartsTheCalm()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            emoji.TryPlace(Seat.North, 400, out _);
            Assert.IsFalse(wheel.Calm(400), "another seat's emoji leaves the wheel free");

            emoji.TryPlace(Seat.South, 400, out _);
            Assert.IsTrue(wheel.Calm(400 + TokenEffect.EmojiDurationMs - 1));
            Assert.IsFalse(wheel.Calm(400 + TokenEffect.EmojiDurationMs));
        }

        [Test]
        public void PlacingAnEmojiDoesNotAllocate()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            emoji.TryPlace(Seat.East, 0, out _);
            Assert.That(() =>
            {
                for (var at = 0; at < 100; at++)
                {
                    emoji.TryPlace(Seat.East, at, out _);
                    emoji.TryPlace(Seat.South, at, out _);
                }
            }, Allocates.Nothing);
        }

        [Test]
        public void AQuickMessageIsSentAsAnUnsequencedIntentOnlyOnAnAcceptedSession()
        {
            Assert.IsFalse(emoji.TrySendMessage(new Session(), "quick.well_played", false));

            var transport = new FakeTransport();
            var session = AcceptedSession(transport);
            Assert.IsTrue(emoji.TrySendMessage(session, "quick.well_played", true));

            Assert.AreEqual(1, transport.Sent.Count);
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[0]);
            Assert.AreEqual(Envelope.PayloadOneofCase.QuickMessageIntent, envelope.PayloadCase);
            Assert.AreEqual("quick.well_played", envelope.QuickMessageIntent.Key);
            Assert.IsTrue(envelope.QuickMessageIntent.TeamOnly);
            Assert.AreEqual(0ul, envelope.Sequence);
        }

        [Test]
        public void AQuickMessageShowsAtTheSendersSideInTheViewersLanguageForTheEmojiDuration()
        {
            emoji.Begin(Seat.South, Mode._2V2);

            Assert.IsTrue(emoji.TryShowMessage(Message(Seat.West), 1000));

            Assert.IsTrue(messages.TryGet(TableSide.Left, 1000 + TokenEffect.EmojiDurationMs - 1, out var text));
            Assert.AreEqual("Əla oyun", text);
            Assert.IsFalse(messages.TryGet(TableSide.Left, 1000 + TokenEffect.EmojiDurationMs, out text));
            Assert.IsNull(text);
            Assert.IsFalse(messages.TryGet(TableSide.Top, 1000, out _));

            language.Choose("ru");
            Assert.IsTrue(emoji.TryShowMessage(Message(Seat.North), 2000));
            Assert.IsTrue(messages.TryGet(TableSide.Top, 2000, out text));
            Assert.AreEqual("Well played", text);
        }

        [Test]
        public void ANewQuickMessageReplacesTheSeatsMessageFromTheStart()
        {
            emoji.Begin(Seat.X, Mode._1V1);
            emoji.TryShowMessage(Message(Seat.Y), 0);
            emoji.TryShowMessage(Message(Seat.Y), 1000);

            Assert.IsTrue(messages.TryGet(TableSide.Top, 1000 + TokenEffect.EmojiDurationMs - 1, out _));
            Assert.IsFalse(messages.TryGet(TableSide.Top, 1000 + TokenEffect.EmojiDurationMs, out _));
        }

        [TestCase(Mode._1V1, Seat.North)]
        [TestCase(Mode._2V2, Seat.X)]
        [TestCase(Mode._2V2, Seat.Unspecified)]
        [TestCase(Mode.Unspecified, Seat.South)]
        public void AQuickMessageFromASeatOutsideTheModeIsDropped(Mode mode, Seat sender)
        {
            emoji.Begin(Seat.Unspecified, mode);
            Assert.IsFalse(emoji.TryShowMessage(Message(sender), 0));
            for (var side = 0; side < TableSides.Count; side++) Assert.IsFalse(messages.TryGet((TableSide)side, 0, out _));
        }

        [Test]
        public void TheOwnQuickMessageLeavesTheEmojiWheelFree()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            Assert.IsTrue(emoji.TryShowMessage(Message(Seat.South), 0));
            Assert.IsTrue(messages.TryGet(TableSide.Bottom, 0, out _));
            Assert.IsFalse(wheel.Calm(0));
        }

        [Test]
        public void OnlyTheOtherSeatsOfTheMatchCanBeMuted()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            Assert.IsFalse(emoji.CanMute(Seat.South));
            Assert.IsTrue(emoji.CanMute(Seat.North));
            Assert.IsTrue(emoji.CanMute(Seat.East));
            Assert.IsFalse(emoji.CanMute(Seat.X));
            Assert.IsFalse(emoji.ToggleMute(Seat.South));
            Assert.IsFalse(emoji.IsMuted(Seat.South));

            emoji.Begin(Seat.Unspecified, Mode._1V1);
            Assert.IsTrue(emoji.CanMute(Seat.X));
            Assert.IsTrue(emoji.CanMute(Seat.Y));
            Assert.IsFalse(emoji.CanMute(Seat.Unspecified));
        }

        [Test]
        public void AMutedSeatsEmojiAndQuickMessagesAreDroppedUntilUnmuted()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            Assert.AreEqual(MatchEmoji.MuteKey, emoji.MuteLabelKey(Seat.East));

            Assert.IsTrue(emoji.ToggleMute(Seat.East));

            Assert.IsTrue(emoji.IsMuted(Seat.East));
            Assert.AreEqual(MatchEmoji.UnmuteKey, emoji.MuteLabelKey(Seat.East));
            Assert.IsFalse(emoji.TryPlace(Seat.East, 0, out _));
            Assert.IsFalse(emoji.TryShowMessage(Message(Seat.East), 0));
            Assert.IsFalse(messages.TryGet(TableSide.Right, 0, out _));
            Assert.IsTrue(emoji.TryPlace(Seat.West, 0, out _), "other seats stay heard");
            Assert.IsTrue(emoji.TryShowMessage(Message(Seat.North), 0));

            Assert.IsTrue(emoji.ToggleMute(Seat.East));
            Assert.IsFalse(emoji.IsMuted(Seat.East));
            Assert.IsTrue(emoji.TryPlace(Seat.East, 0, out _));
            Assert.IsTrue(emoji.TryShowMessage(Message(Seat.East), 0));
        }

        [Test]
        public void MutesLastOnlyForTheMatch()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            emoji.ToggleMute(Seat.East);
            emoji.End();
            Assert.IsFalse(emoji.IsMuted(Seat.East));
            Assert.IsTrue(emoji.TryPlace(Seat.East, 0, out _));

            emoji.ToggleMute(Seat.East);
            emoji.TryShowMessage(Message(Seat.North), 0);
            emoji.Begin(Seat.South, Mode._2V2);
            Assert.IsFalse(emoji.IsMuted(Seat.East));
            Assert.IsFalse(messages.TryGet(TableSide.Top, 0, out _), "a new match starts without old messages");
        }

        [Test]
        public void DroppingAMutedSeatDoesNotAllocate()
        {
            emoji.Begin(Seat.South, Mode._2V2);
            emoji.ToggleMute(Seat.East);
            var shown = Message(Seat.East);
            emoji.TryShowMessage(shown, 0);
            Assert.That(() =>
            {
                for (var at = 0; at < 100; at++)
                {
                    emoji.TryPlace(Seat.East, at, out _);
                    emoji.TryShowMessage(shown, at);
                    emoji.IsMuted(Seat.East);
                    emoji.MuteLabelKey(Seat.East);
                }
            }, Allocates.Nothing);
        }

        static QuickMessageShown Message(Seat sender) => new() { Seat = sender, Key = "quick.well_played" };
    }
}
