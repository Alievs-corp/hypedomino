using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Match;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MatchRootTests
    {
        const string NeonTable = "cosmetic/table/neon/static";
        const string FoxAvatar = "cosmetic/avatar/fox/static";
        const ulong ServerNowMs = 1_000_000;
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""epic"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""},
{""id"":""table.neon"",""type"":""table"",""rarityId"":""epic"",""availability"":""live"",""staticAddress"":""cosmetic/table/neon/static""}]}";

        sealed class Table : IMatchTable
        {
            public Vector2 BoardSize => new(400f, 300f);
            public float BoardMaxScale => 1f;
            public int BackgroundsShown;

            public void ShowBackground(MatchSnapshot view) => BackgroundsShown++;

            public void ShowSeats(MatchSnapshot view)
            {
            }

            public void ShowScore(ScoreBoard score)
            {
            }

            public void ShowBoard(ChainLayout board)
            {
            }
        }

        sealed class Hud : IMatchHud
        {
            public int TurnsShown;
            public float TurnRemaining;

            public void ShowHand(HandChoice hand)
            {
            }

            public void ShowTurn(TurnTimer timer)
            {
                TurnsShown++;
                TurnRemaining = timer.Remaining;
            }

            public void ShowPass(TableSide side, PassNotice notice)
            {
            }

            public void HidePass(TableSide side)
            {
            }
        }

        sealed class LoadingCardsScreen : ILoadingScreen
        {
            public int CardsShown;
            public int Closed;

            public void ShowCards(MatchSnapshot view, LoadingCards cards) => CardsShown++;

            public void Close() => Closed++;
        }

        sealed class Source : ICosmeticArtSource<Sprite>
        {
            readonly Dictionary<string, Action<Sprite>> pending = new();

            public readonly List<string> Releases = new();
            public Sprite Art;

            public void Load(string address, Action<Sprite> done) => pending[address] = done;

            public void Release(string address) => Releases.Add(address);

            public void FinishAll()
            {
                var done = new List<Action<Sprite>>(pending.Values);
                pending.Clear();
                foreach (var finish in done) finish(Art);
            }
        }

        sealed class ServerFrames : IFrameTransport
        {
            readonly Queue<byte[]> pending = new();

            public readonly List<LoadingIntent> Reports = new();

            public void Send(ReadOnlySpan<byte> frame)
            {
                if (!FrameReader.TryRead(frame, out var read) || read.Payload != Envelope.PayloadOneofCase.LoadingIntent) return;
                Reports.Add(LoadingIntent.Parser.ParseFrom(read.PayloadBytes));
            }

            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                var has = pending.Count > 0;
                frame = has ? pending.Dequeue() : null;
                return has;
            }

            public void Deliver(Envelope.PayloadOneofCase payload, IMessage message) =>
                pending.Enqueue(new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, 1), payload, message).ToArray());
        }

        Table table;
        Hud hud;
        LoadingCardsScreen loadingScreen;
        Source source;
        ServerFrames frames;
        Session session;
        ServerClock clock;
        MatchRoot root;
        Texture2D texture;
        double now;

        [SetUp]
        public void CreateRoot()
        {
            table = new Table();
            hud = new Hud();
            loadingScreen = new LoadingCardsScreen();
            texture = new Texture2D(4, 4);
            source = new Source { Art = Sprite.Create(texture, new Rect(0, 0, 4, 4), Vector2.zero) };
            frames = new ServerFrames();
            session = new Session();
            clock = new ServerClock();
            clock.Sync(ServerNowMs, 0d, 0d);
            now = 0d;
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var manifest));
            root = new MatchRoot(table, hud, loadingScreen, new CosmeticArtCache<Sprite>(source), new ActiveCosmeticManifest(manifest));
            root.Bind(session);
            session.Hello(frames, new ClientHello());
            frames.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = 1 });
            session.Poll();
        }

        [TearDown]
        public void UnbindRoot()
        {
            root.Unbind();
            Object.DestroyImmediate(source.Art);
            Object.DestroyImmediate(texture);
        }

        static MatchView View(MatchState state)
        {
            var view = new MatchView
            {
                State = state,
                Mode = Mode._1V1,
                Seat = Seat.X,
                Current = Seat.X,
                Table = "table.neon",
                TurnDeadlineUnixMs = ServerNowMs + 10_000,
                TurnSeconds = 15,
                TurnWarningSeconds = 5,
            };
            view.Roster.Add(new RosterEntry { Seat = Seat.X, Avatar = "avatar.fox" });
            view.Roster.Add(new RosterEntry { Seat = Seat.Y, Avatar = "avatar.fox" });
            view.Seats.Add(new SeatView { Seat = Seat.X });
            view.Seats.Add(new SeatView { Seat = Seat.Y });
            return view;
        }

        void Receive(MatchView view)
        {
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, view);
            session.Poll();
        }

        void Tick(bool sceneReady) => root.Tick(clock, now += 0.1d, false, sceneReady);

        void TickReady() => Tick(true);

        [Test]
        public void OneBindFeedsEveryViewToTheLoadingScreenAndTheMatchTable()
        {
            Receive(View(MatchState.Loading));

            Assert.IsTrue(root.Loading.Open);
            Assert.AreEqual(1, loadingScreen.CardsShown);
            Assert.AreEqual(1, table.BackgroundsShown);
        }

        [Test]
        public void ATickReportsLoadingAndTicksTheTurnTimer()
        {
            Receive(View(MatchState.Loading));
            source.FinishAll();

            Tick(true);

            Assert.AreEqual(1, frames.Reports.Count);
            Assert.IsTrue(frames.Reports[0].Ready);
            Assert.AreEqual(1, hud.TurnsShown);
        }

        [Test]
        public void ThePrefetchedArtIsHeldWhileTheLoadingScreenIsOpen()
        {
            Receive(View(MatchState.Loading));
            source.FinishAll();

            Tick(true);
            Tick(true);

            Assert.IsEmpty(source.Releases);
        }

        [Test]
        public void ThePrefetchedArtIsReleasedOnceOnTheTickAfterTheLoadingScreenCloses()
        {
            Receive(View(MatchState.Loading));
            source.FinishAll();
            Tick(true);

            Receive(View(MatchState.Opening));
            Assert.AreEqual(1, loadingScreen.Closed);
            Assert.AreEqual(2, table.BackgroundsShown);
            Assert.IsEmpty(source.Releases);

            Tick(true);
            Tick(true);

            CollectionAssert.AreEquivalent(new[] { NeonTable, FoxAvatar }, source.Releases);
        }

        [Test]
        public void AMatchJoinedPastTheLoadingPhaseReleasesNothing()
        {
            Receive(View(MatchState.Playing));

            Tick(true);

            Assert.AreEqual(0, loadingScreen.CardsShown);
            Assert.AreEqual(1, table.BackgroundsShown);
            Assert.IsEmpty(source.Releases);
        }

        [Test]
        public void TheTurnTimerRunsOnTheServerClock()
        {
            Receive(View(MatchState.Playing));

            root.Tick(clock, 4d, false, true);

            Assert.AreEqual(6f, hud.TurnRemaining, 1e-3f);
        }

        [Test]
        public void UnbindStopsBothTheLoadingScreenAndTheMatchTable()
        {
            root.Unbind();

            Receive(View(MatchState.Loading));

            Assert.AreEqual(0, loadingScreen.CardsShown);
            Assert.AreEqual(0, table.BackgroundsShown);
        }

        [Test]
        public void TickingWhileLoadingAllocatesNothing()
        {
            Receive(View(MatchState.Loading));
            Tick(true);

            Assert.That(TickReady, Allocates.Nothing);
        }

        [Test]
        public void TickingDuringPlayAllocatesNothing()
        {
            Receive(View(MatchState.Playing));
            Tick(true);

            Assert.That(TickReady, Allocates.Nothing);
        }
    }
}
