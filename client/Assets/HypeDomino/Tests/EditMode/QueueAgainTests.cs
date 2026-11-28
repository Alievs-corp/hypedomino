using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game;
using HypeDomino.Game.Lobby;
using HypeDomino.Net;
using HypeDomino.Net.Integrity;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class QueueAgainTests
    {
        const ulong Nonce = 0x5EED_0086_0000_0001;
        const string RoomCode = "ABC234";

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

        sealed class ProofDesk
        {
            public readonly List<IntegrityPurpose> Purposes = new();

            public void Request(IntegrityPurpose purpose, Action<IntegrityProof> done)
            {
                Purposes.Add(purpose);
                done(null);
            }
        }

        FakeTransport transport;
        Session session;
        AppFlow flow;
        MatchQueue queue;
        ProofDesk proofs;
        LobbyPlay play;
        QueueAgain again;

        [SetUp]
        public void Build()
        {
            transport = new FakeTransport();
            session = new Session();
            flow = new AppFlow();
            queue = new MatchQueue();
            queue.Follow(session, flow);
            proofs = new ProofDesk();
            play = new LobbyPlay(queue, proofs.Request);
            again = new QueueAgain(queue, play, flow);
            Accept(false);
        }

        void Accept(bool inMatch)
        {
            var hello = new ServerHello { Status = HelloStatus.Accepted, SessionNonce = Nonce, InMatch = inMatch };
            session.Hello(transport, new ClientHello());
            transport.Deliver(Envelope.PayloadOneofCase.ServerHello, hello);
            session.Poll();
            Assert.IsTrue(session.Accepted);
            flow.Hello(hello);
            transport.Sent.Clear();
        }

        void FindMatch(Mode mode, QueueKind kind)
        {
            queue.Read(new QueueState { Status = QueueStatus.Searching, Mode = mode, Queue = kind });
            queue.Read(new QueueState { Status = QueueStatus.MatchFound });
            flow.View(MatchState.Playing);
            Assert.AreEqual(AppState.Match, flow.State);
        }

        void FinishQueueMatch(Mode mode, QueueKind kind)
        {
            FindMatch(mode, kind);
            flow.View(MatchState.Finished);
        }

        void FinishRoomMatch()
        {
            flow.Room(new LobbyState { Code = RoomCode, Phase = LobbyPhase.Teams });
            flow.Room(new LobbyState { Code = RoomCode, Phase = LobbyPhase.Match });
            flow.View(MatchState.Playing);
            flow.View(MatchState.Finished);
            Assert.AreEqual(AppState.Match, flow.State);
        }

        JoinQueueIntent SentJoin()
        {
            Assert.AreEqual(1, transport.Sent.Count);
            var envelope = Envelope.Parser.ParseFrom(transport.Sent[0]);
            Assert.AreEqual(Envelope.PayloadOneofCase.JoinQueueIntent, envelope.PayloadCase);
            return envelope.JoinQueueIntent;
        }

        [TestCase(Mode._1V1, QueueKind.Ranked, IntegrityPurpose.RankedMatch)]
        [TestCase(Mode._2V2, QueueKind.Ranked, IntegrityPurpose.RankedMatch)]
        [TestCase(Mode._1V1, QueueKind.Casual, IntegrityPurpose.CasualMatch)]
        [TestCase(Mode._2V2, QueueKind.Casual, IntegrityPurpose.CasualMatch)]
        public void PlayAgainJoinsTheQueueAndModeOfTheMatchThatEnded(Mode mode, QueueKind kind, IntegrityPurpose purpose)
        {
            FinishQueueMatch(mode, kind);
            Assert.IsTrue(again.Offered);

            Assert.IsTrue(again.Tap(session));

            Assert.AreEqual(AppState.Lobby, flow.State);
            var join = SentJoin();
            Assert.AreEqual(mode, join.Mode);
            Assert.AreEqual(kind, join.Queue);
            CollectionAssert.AreEqual(new[] { purpose }, proofs.Purposes);
            Assert.IsTrue(queue.AwaitingReply);
            Assert.IsFalse(again.Offered);
        }

        [Test]
        public void PlayAgainWaitsForTheMatchToFinish()
        {
            FindMatch(Mode._1V1, QueueKind.Ranked);
            Assert.IsTrue(again.Offered);

            Assert.IsFalse(again.Tap(session));

            Assert.AreEqual(AppState.Match, flow.State);
            Assert.IsEmpty(transport.Sent);
        }

        [Test]
        public void AReconnectDuringTheMatchKeepsPlayAgain()
        {
            FindMatch(Mode._2V2, QueueKind.Casual);
            Accept(true);
            flow.View(MatchState.Finished);

            Assert.IsTrue(again.Offered);
            Assert.IsTrue(again.Tap(session));
            Assert.AreEqual(Mode._2V2, SentJoin().Mode);
        }

        [Test]
        public void AMatchFoundWithoutAKnownQueueOffersNothing()
        {
            queue.Read(new QueueState { Status = QueueStatus.Searching, Mode = Mode._1V1, Queue = QueueKind.Ranked });
            Accept(false);
            queue.Read(new QueueState { Status = QueueStatus.MatchFound });
            flow.View(MatchState.Playing);
            flow.View(MatchState.Finished);

            Assert.IsFalse(again.Offered);
            Assert.IsFalse(again.Tap(session));
            Assert.AreEqual(AppState.Match, flow.State);
            Assert.IsEmpty(transport.Sent);
        }

        [Test]
        public void AFriendRoomMatchOffersNoPlayAgainAndReturnsToTheRoom()
        {
            FinishRoomMatch();

            Assert.IsFalse(again.Offered);
            Assert.IsFalse(again.Tap(session));
            Assert.IsEmpty(transport.Sent);

            flow.Room(new LobbyState { Code = RoomCode, Phase = LobbyPhase.Lobby });
            Assert.AreEqual(AppState.Room, flow.State);
        }

        [Test]
        public void AFriendRoomMatchAfterAQueueMatchOffersNoPlayAgain()
        {
            FinishQueueMatch(Mode._1V1, QueueKind.Ranked);
            Assert.IsTrue(flow.LeaveMatch());
            Assert.IsFalse(again.Offered);

            FinishRoomMatch();

            Assert.IsFalse(again.Offered);
            Assert.IsFalse(again.Tap(session));
            Assert.IsEmpty(transport.Sent);
        }

        [TestCase(HelloStatus.UpdateRequired, AppState.UpdateRequired)]
        [TestCase(HelloStatus.Maintenance, AppState.Maintenance)]
        public void AnUpdateOrMaintenanceDuringTheMatchDropsPlayAgain(HelloStatus status, AppState shown)
        {
            FinishQueueMatch(Mode._1V1, QueueKind.Ranked);

            flow.Hello(new ServerHello { Status = status });

            Assert.AreEqual(shown, flow.State);
            Assert.IsFalse(again.Offered);
            Assert.IsFalse(again.Tap(session));
            Assert.AreEqual(shown, flow.State);
            Assert.IsEmpty(transport.Sent);
        }

        [Test]
        public void PlayAgainWithoutAnAcceptedSessionStaysOnTheResult()
        {
            FinishQueueMatch(Mode._1V1, QueueKind.Ranked);

            Assert.IsFalse(again.Tap(new Session()));

            Assert.AreEqual(AppState.Match, flow.State);
            Assert.IsTrue(again.Offered);
            Assert.IsEmpty(transport.Sent);
        }

        [Test]
        public void ARankedBanRefusalOfPlayAgainShowsInTheLobby()
        {
            FinishQueueMatch(Mode._1V1, QueueKind.Ranked);
            Assert.IsTrue(again.Tap(session));

            queue.Read(new QueueState { Status = QueueStatus.Refused, Mode = Mode._1V1, Queue = QueueKind.Ranked, Refusal = QueueRefusal.RankedBan });

            Assert.AreEqual(AppState.Lobby, flow.State);
            Assert.AreEqual(QueuePhase.Refused, queue.Phase);
            Assert.AreEqual("lobby.queue.refused.ranked_ban", queue.RefusalKey);
            Assert.AreEqual(PlayCardState.Ready, play.StateOf(2));
        }

        [Test]
        public void ReadingPlayAgainAllocatesNothing()
        {
            FinishQueueMatch(Mode._2V2, QueueKind.Ranked);
            var offered = false;
            TestDelegate read = () => offered |= again.Offered;
            read();
            Assert.That(read, Allocates.Nothing);
            Assert.IsTrue(offered);
        }
    }
}
