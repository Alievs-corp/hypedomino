using System;
using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Match;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MatchScreenTests
    {
        const ulong ServerNowMs = 1_000_000;

        sealed class Table : IMatchTable
        {
            public Vector2 BoardSize => new(400f, 300f);
            public float BoardMaxScale => 1f;
            public int SeatsShown;
            public int ScoresShown;
            public int BoardsShown;
            public MatchSnapshot Seats;
            public ScoreBoard Score;
            public ChainLayout Board;

            public void ShowSeats(MatchSnapshot view)
            {
                SeatsShown++;
                Seats = view;
            }

            public void ShowScore(ScoreBoard score)
            {
                ScoresShown++;
                Score = score;
            }

            public void ShowBoard(ChainLayout board)
            {
                BoardsShown++;
                Board = board;
            }
        }

        sealed class Hud : IMatchHud
        {
            public int HandsShown;
            public HandChoice Hand;
            public int TurnsShown;
            public float TurnRemaining;
            public int PassesShown;
            public TableSide PassSide;
            public float PassAlpha;
            public int PassesHidden;
            public TableSide HiddenSide;

            public void ShowHand(HandChoice hand)
            {
                HandsShown++;
                Hand = hand;
            }

            public void ShowTurn(TurnTimer timer)
            {
                TurnsShown++;
                TurnRemaining = timer.Remaining;
            }

            public void ShowPass(TableSide side, PassNotice notice)
            {
                PassesShown++;
                PassSide = side;
                PassAlpha = notice.Alpha;
            }

            public void HidePass(TableSide side)
            {
                PassesHidden++;
                HiddenSide = side;
            }
        }

        sealed class ServerFrames : IFrameTransport
        {
            byte[] pending;

            public void Send(ReadOnlySpan<byte> frame) { }

            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                frame = pending;
                var has = pending != null;
                pending = null;
                return has;
            }

            public void Deliver(Envelope.PayloadOneofCase payload, IMessage message) =>
                pending = new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, 1), payload, message).ToArray();
        }

        readonly MatchSnapshot snapshot = new();
        Table table;
        Hud hud;
        MatchScreen screen;

        [SetUp]
        public void CreateScreen()
        {
            table = new Table();
            hud = new Hud();
            screen = new MatchScreen(table, hud);
        }

        static Tile T(uint low, uint high) => new() { Low = low, High = high };

        static OrientedTile O(uint first, uint second) => new() { First = first, Second = second };

        static MatchView Playing(Seat current, params LegalMove[] legal)
        {
            var view = new MatchView
            {
                State = MatchState.Playing,
                Seat = Seat.South,
                Mode = Mode._2V2,
                Current = current,
                Scores = new TeamPoints { TeamA = 40, TeamB = 25 },
                TargetScore = 150,
                TurnDeadlineUnixMs = ServerNowMs + 10_000,
                TurnSeconds = 15,
                TurnWarningSeconds = 5,
                Board = new BoardView { Line = { O(2, 3), O(3, 6), O(6, 1) }, OpeningIndex = 1 },
            };
            view.Hand.Add(new[] { T(1, 2), T(3, 3), T(4, 6) });
            view.LegalMoves.Add(legal);
            return view;
        }

        static LegalMove Move(uint low, uint high, BoardEnd end) => new() { Tile = T(low, high), End = end };

        MatchSnapshot Read(MatchView view)
        {
            Assert.IsTrue(snapshot.TryRead(view.ToByteArray()));
            return snapshot;
        }

        void Show(MatchView view) => screen.Show(Read(view));

        static ServerClock SyncedClock()
        {
            var clock = new ServerClock();
            clock.Sync(ServerNowMs, 0d, 0d);
            return clock;
        }

        [Test]
        public void AViewReachesTheSeatsTheScoreTheBoardAndTheHand()
        {
            Show(Playing(Seat.South, Move(1, 2, BoardEnd.Right)));

            Assert.AreEqual(1, table.SeatsShown);
            Assert.AreSame(snapshot, table.Seats);
            Assert.AreEqual(1, table.ScoresShown);
            Assert.AreSame(screen.Score, table.Score);
            Assert.AreEqual(Team.A, screen.Score.Mine.Team);
            Assert.AreEqual(40u, screen.Score.Mine.Score);
            Assert.AreEqual(25u, screen.Score.Theirs.Score);
            Assert.AreEqual(150u, screen.Score.TargetScore);
            Assert.AreEqual(1, table.BoardsShown);
            Assert.AreSame(screen.Board, table.Board);
            Assert.AreEqual(3, screen.Board.Count);
            Assert.Greater(screen.Board.Scale, 0f);
            Assert.AreEqual(1, hud.HandsShown);
            Assert.AreSame(screen.Hand, hud.Hand);
        }

        [Test]
        public void TheHandHighlightsOnlyWhatTheServerListsAsLegal()
        {
            Show(Playing(Seat.South, Move(3, 3, BoardEnd.Left)));

            Assert.AreEqual(3, screen.Hand.HandCount);
            Assert.IsFalse(screen.Hand.IsSelectable(0));
            Assert.IsTrue(screen.Hand.IsSelectable(1));
            Assert.IsFalse(screen.Hand.IsSelectable(2));
        }

        [Test]
        public void NoLegalMovesFromTheServerMeansNoHighlightEvenWhenATileWouldFit()
        {
            Show(Playing(Seat.South));

            for (var i = 0; i < screen.Hand.HandCount; i++) Assert.IsFalse(screen.Hand.IsSelectable(i));
            Assert.IsTrue(screen.Hand.CanPass);
        }

        [Test]
        public void TheTimerTicksFromTheServerDeadline()
        {
            Show(Playing(Seat.South));
            screen.Tick(SyncedClock(), 4d, false);

            Assert.AreEqual(1, hud.TurnsShown);
            Assert.IsTrue(screen.Timer.Running);
            Assert.AreEqual(6f, hud.TurnRemaining, 0.001f);
            Assert.IsFalse(screen.Timer.IsWarning);

            screen.Tick(SyncedClock(), 6d, false);
            Assert.IsTrue(screen.Timer.IsWarning);
        }

        [Test]
        public void TheTimerStaysStillWithoutASyncedClock()
        {
            Show(Playing(Seat.South));
            screen.Tick(new ServerClock(), 4d, false);

            Assert.AreEqual(1, hud.TurnsShown);
            Assert.IsFalse(screen.Timer.Running);
            Assert.AreEqual(0f, hud.TurnRemaining);
        }

        [Test]
        public void APassShowsAtThePassingSeatAndLeavesAfterItsLifetime()
        {
            Show(Playing(Seat.West));
            screen.Tick(SyncedClock(), 0d, false);
            Assert.AreEqual(0, hud.PassesShown);

            Show(Playing(Seat.North));
            screen.Tick(SyncedClock(), 10d, false);
            Assert.AreEqual(1, hud.PassesShown);
            Assert.AreEqual(TableSide.Left, hud.PassSide);

            screen.Tick(SyncedClock(), 10d + Seconds(PassNotice.LifetimeMs) - 0.05d, false);
            Assert.AreEqual(2, hud.PassesShown);
            Assert.AreEqual(0, hud.PassesHidden);

            screen.Tick(SyncedClock(), 10d + Seconds(PassNotice.LifetimeMs) + 0.01d, false);
            Assert.AreEqual(1, hud.PassesHidden);
            Assert.AreEqual(TableSide.Left, hud.HiddenSide);

            screen.Tick(SyncedClock(), 20d, false);
            Assert.AreEqual(2, hud.PassesShown);
            Assert.AreEqual(1, hud.PassesHidden);
        }

        [Test]
        public void APlacementIsNotAPass()
        {
            Show(Playing(Seat.West));
            var placed = Playing(Seat.North);
            placed.Board.Line.Add(O(1, 4));
            Show(placed);
            screen.Tick(SyncedClock(), 0d, false);

            Assert.AreEqual(0, hud.PassesShown);
        }

        [Test]
        public void TheFirstViewAfterAHelloShowsNoPass()
        {
            var session = new Session();
            var frames = new ServerFrames();
            screen.Bind(session);
            session.Hello(frames, new ClientHello());
            frames.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = 1 });
            session.Poll();
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, Playing(Seat.West));
            session.Poll();
            frames.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = 1 });
            session.Poll();
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, Playing(Seat.North));
            session.Poll();
            screen.Tick(SyncedClock(), 0d, false);

            Assert.AreEqual(2, table.SeatsShown);
            Assert.AreEqual(0, hud.PassesShown);
        }

        [Test]
        public void ViewsFromTheSessionStopAfterUnbind()
        {
            var session = new Session();
            var frames = new ServerFrames();
            screen.Bind(session);
            session.Hello(frames, new ClientHello());
            frames.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = 1 });
            session.Poll();
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, Playing(Seat.South, Move(1, 2, BoardEnd.Right)));
            session.Poll();

            Assert.AreEqual(1, table.SeatsShown);
            Assert.IsTrue(screen.Hand.IsSelectable(0));

            screen.Unbind();
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, Playing(Seat.West));
            session.Poll();
            Assert.AreEqual(1, table.SeatsShown);
        }

        [Test]
        public void AViewOutsideAKnownModeShowsNoPass()
        {
            var first = Playing(Seat.West);
            first.Mode = Mode.Unspecified;
            Show(first);
            var second = Playing(Seat.North);
            second.Mode = Mode.Unspecified;
            Show(second);
            screen.Tick(SyncedClock(), 0d, false);

            Assert.AreEqual(0, hud.PassesShown);
        }

        [Test]
        public void ShowingAViewAndTickingAllocateNothing()
        {
            var clock = SyncedClock();
            var west = Playing(Seat.West, Move(1, 2, BoardEnd.Right)).ToByteArray();
            var north = Playing(Seat.North).ToByteArray();
            var now = 0d;
            TestDelegate frame = () =>
            {
                snapshot.TryRead(west);
                screen.Show(snapshot);
                screen.Tick(clock, now += 0.1d, false);
                snapshot.TryRead(north);
                screen.Show(snapshot);
                screen.Tick(clock, now += 0.1d, true);
            };
            frame();

            Assert.That(frame, Allocates.Nothing);
            Assert.Greater(hud.PassesShown, 0);
        }

        static double Seconds(float ms) => ms / 1000d;
    }
}
