using System;
using System.Collections.Generic;
using System.IO;
using Google.Protobuf;
using HypeDomino.Net;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools.Constraints;
using Is = UnityEngine.TestTools.Constraints.Is;

namespace HypeDomino.Tests.EditMode
{
    public sealed class ProtocolSessionTests
    {
        const ulong Nonce = 0x1122334455667788;

        static string GoldenDirectory => Path.Combine(Application.dataPath, "..", "..", "proto", "testdata", "frames");

        static byte[] Golden(string name) => File.ReadAllBytes(Path.Combine(GoldenDirectory, name + ".bin"));

        static byte[] Bytes(ArraySegment<byte> segment) => segment.AsSpan().ToArray();

        static byte[] AcceptedHello(ulong acked, ulong nonce, ulong serverTimeMs) => new Envelope
        {
            ProtocolVersion = ProtocolSession.ProtocolVersion,
            AckedSequence = acked,
            ServerHello = new ServerHello { Status = HelloStatus.Accepted, SessionNonce = nonce, ServerTimeUnixMs = serverTimeMs }
        }.ToByteArray();

        static byte[] ViewAcking(ulong acked) => new Envelope
        {
            ProtocolVersion = ProtocolSession.ProtocolVersion,
            AckedSequence = acked,
            MatchView = new MatchView { State = MatchState.Playing }
        }.ToByteArray();

        static ProtocolSession Accepted(ulong acked, ulong nonce)
        {
            var session = new ProtocolSession();
            session.Hello(new ClientHello(), 0);
            Assert.IsTrue(session.Receive(AcceptedHello(acked, nonce, 0), 0));
            return session;
        }

        [Test]
        public void EveryGoldenFrameDecodesToTheSameHeaderAndPayloadAsTheGeneratedParser()
        {
            var files = Directory.GetFiles(GoldenDirectory, "*.bin");
            Assert.IsNotEmpty(files);
            foreach (var file in files)
            {
                var bytes = File.ReadAllBytes(file);
                var envelope = Envelope.Parser.ParseFrom(bytes);
                Assert.IsTrue(FrameDecoder.TryRead(bytes, out var header), file);
                Assert.AreEqual(envelope.ProtocolVersion, header.ProtocolVersion, file);
                Assert.AreEqual(envelope.Sequence, header.Sequence, file);
                Assert.AreEqual(envelope.AckedSequence, header.AckedSequence, file);
                Assert.AreEqual(envelope.SessionNonce, header.SessionNonce, file);
                Assert.AreEqual(envelope.PayloadCase, header.Payload, file);
                var field = Envelope.Descriptor.FindFieldByNumber((int)header.Payload);
                var payload = new ArraySegment<byte>(bytes, header.PayloadStart, header.PayloadLength).AsSpan().ToArray();
                Assert.AreEqual(field.Accessor.GetValue(envelope), field.MessageType.Parser.ParseFrom(payload), file);
            }
        }

        [Test]
        public void TheDecoderReadsTheNonceAndSkipsUnknownFields()
        {
            var frame = new Envelope { ProtocolVersion = 1, SessionNonce = Nonce, PassIntent = new PassIntent() }.ToByteArray();
            var unknown = new byte[] { 0xF8, 0x07, 0x05, 0xFD, 0x07, 1, 2, 3, 4 };
            var bytes = new byte[frame.Length + unknown.Length];
            frame.CopyTo(bytes, 0);
            unknown.CopyTo(bytes, frame.Length);
            Assert.IsTrue(FrameDecoder.TryRead(bytes, out var header));
            Assert.AreEqual(Nonce, header.SessionNonce);
            Assert.AreEqual(Envelope.PayloadOneofCase.PassIntent, header.Payload);
        }

        [Test]
        public void TheDecoderRefusesTruncatedFrames()
        {
            var bytes = Golden("match_view");
            Assert.IsFalse(FrameDecoder.TryRead(new ReadOnlySpan<byte>(bytes, 0, bytes.Length - 1), out _));
            Assert.IsFalse(FrameDecoder.TryRead(new byte[] { 0x08, 0x80 }, out _));
            Assert.IsFalse(FrameDecoder.TryRead(new byte[] { 0x00 }, out _));
            Assert.IsFalse(new ProtocolSession().Receive(new byte[] { 0x5A, 0x02, 0xFF, 0xFF }, 0));
        }

        [Test]
        public void SequencedIntentsMatchTheGoldenFrames()
        {
            var session = Accepted(6, 0);
            Assert.AreEqual(Golden("place_tile_intent"), Bytes(session.PlaceTile(2, 3, BoardEnd.Left, 0)));
            Assert.AreEqual(Golden("place_tile_intent_spinner_arm"), Bytes(session.PlaceTile(3, 5, BoardEnd.Top, 0)));
            Assert.AreEqual(Golden("pass_intent"), Bytes(session.Pass(0)));
            Assert.AreEqual(Golden("draw_intent"), Bytes(session.Draw(0)));
            Assert.AreEqual(Golden("end_turn_intent"), Bytes(Accepted(8, 0).EndTurn(0)));
        }

        [Test]
        public void UnsequencedIntentsMatchTheGoldenFrames()
        {
            var session = Accepted(6, 0);
            Assert.AreEqual(Golden("loading_intent"), Bytes(session.Loading(100, true)));
            Assert.AreEqual(Golden("emoji_intent"), Bytes(session.Emoji("emoji.laugh")));
            Assert.AreEqual(Golden("quick_message_intent"), Bytes(session.QuickMessage("quick.well_played", true)));
            Assert.AreEqual(Golden("join_queue_intent"), Bytes(session.JoinQueue(Mode._2V2, QueueKind.Ranked, null)));
            Assert.AreEqual(Golden("leave_queue_intent"), Bytes(session.LeaveQueue()));
            Assert.AreEqual(6, session.Sequence);
        }

        [Test]
        public void TheHelloResumesFromTheLastAcknowledgedSequenceWithoutANonce()
        {
            var session = Accepted(40, Nonce);
            Assert.IsTrue(session.Receive(ViewAcking(41), 0));
            var hello = Envelope.Parser.ParseFrom(Bytes(session.Hello(new ClientHello { SessionToken = "sess_01HZX" }, 0)));
            Assert.AreEqual(41, hello.Sequence);
            Assert.AreEqual(41, hello.ClientHello.LastAckedSequence);
            Assert.AreEqual("sess_01HZX", hello.ClientHello.SessionToken);
            Assert.AreEqual(0, hello.SessionNonce);
        }

        [Test]
        public void IntentsCarryTheNonceOfTheAcceptedHelloAndCountOnFromItsAck()
        {
            var session = Accepted(41, Nonce);
            var pass = Envelope.Parser.ParseFrom(Bytes(session.Pass(0)));
            Assert.AreEqual(42, pass.Sequence);
            Assert.AreEqual(Nonce, pass.SessionNonce);
            var emoji = Envelope.Parser.ParseFrom(Bytes(session.Emoji("emoji.wave")));
            Assert.AreEqual(0, emoji.Sequence);
            Assert.AreEqual(Nonce, emoji.SessionNonce);
        }

        [Test]
        public void AReconnectReplacesTheNonceAndRestartsFromTheServerAck()
        {
            var session = Accepted(5, Nonce);
            session.Pass(0);
            session.Pass(0);
            session.Hello(new ClientHello(), 0);
            Assert.IsTrue(session.Receive(AcceptedHello(6, 99, 0), 0));
            var draw = Envelope.Parser.ParseFrom(Bytes(session.Draw(0)));
            Assert.AreEqual(7, draw.Sequence);
            Assert.AreEqual(99, draw.SessionNonce);
        }

        [Test]
        public void ARefusedHelloKeepsTheSession()
        {
            var session = Accepted(5, Nonce);
            session.Hello(new ClientHello(), 0);
            var refused = new Envelope { ServerHello = new ServerHello { Status = HelloStatus.UpdateRequired } }.ToByteArray();
            Assert.IsTrue(session.Receive(refused, 0));
            Assert.AreEqual(5, session.Acked);
            Assert.AreEqual(Nonce, Envelope.Parser.ParseFrom(Bytes(session.Pass(0))).SessionNonce);
        }

        [Test]
        public void TheHelloSetsTheServerClockFromTheMidpointOfTheRoundTrip()
        {
            var session = new ProtocolSession();
            session.Hello(new ClientHello(), 1000);
            Assert.IsTrue(session.Receive(AcceptedHello(0, Nonce, 50000), 1100));
            Assert.AreEqual(100, session.RoundTripMs);
            Assert.AreEqual(50050, session.ServerNowMs(1100));
            Assert.AreEqual(51050, session.ServerNowMs(2100));
        }

        [Test]
        public void TheRoundTripOfAnIntentEndsWithTheViewThatAcknowledgesIt()
        {
            var session = Accepted(3, Nonce);
            session.PlaceTile(1, 2, BoardEnd.Left, 5000);
            Assert.IsTrue(session.Receive(ViewAcking(3), 5030));
            Assert.AreEqual(0, session.RoundTripMs);
            Assert.IsTrue(session.Receive(ViewAcking(4), 5070));
            Assert.AreEqual(70, session.RoundTripMs);
            Assert.IsTrue(session.Receive(ViewAcking(4), 5200));
            Assert.AreEqual(70, session.RoundTripMs);
            Assert.AreEqual(4, session.Acked);
        }

        [Test]
        public void EveryServerPayloadReachesItsEvent()
        {
            var session = new ProtocolSession();
            var seen = new List<string>();
            session.HelloReceived += hello => seen.Add(hello.CatalogVersion);
            session.ViewReceived += view => seen.Add(view.State.ToString());
            session.ErrorReceived += error => seen.Add(error.Code.ToString());
            session.EmojiReceived += shown => seen.Add(shown.ItemId);
            session.QuickMessageReceived += shown => seen.Add(shown.Key);
            session.QueueReceived += state => seen.Add(state.Status.ToString());
            foreach (var name in new[] { "server_hello_accepted", "match_view", "error_illegal_intent", "emoji_shown", "quick_message_shown", "queue_state_searching" })
                Assert.IsTrue(session.Receive(Golden(name), 0), name);
            CollectionAssert.AreEqual(new[] { "2026.10.26-1", "Playing", "IllegalIntent", "emoji.laugh", "quick.well_played", "Searching" }, seen);
            Assert.AreEqual(8, session.Acked);
        }

        [Test]
        public void EncodingAWarmIntentAndReadingAHeaderDoNotAllocate()
        {
            var session = Accepted(1, Nonce);
            var view = Golden("match_view");
            session.PlaceTile(1, 2, BoardEnd.Left, 0);
            session.Emoji("emoji.laugh");
            Assert.That(() =>
            {
                session.PlaceTile(1, 2, BoardEnd.Left, 0);
                session.Pass(0);
                session.Emoji("emoji.laugh");
                FrameDecoder.TryRead(view, out _);
            }, Is.Not.AllocatingGCMemory());
        }
    }
}
