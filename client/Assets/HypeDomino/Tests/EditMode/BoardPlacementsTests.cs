using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Match;
using Hypedomino.V1;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class BoardPlacementsTests
    {
        readonly MatchSnapshot snapshot = new();
        BoardPlacements placements;

        [SetUp]
        public void Build() => placements = new BoardPlacements();

        static OrientedTile O(uint first, uint second) => new() { First = first, Second = second };

        static MatchView View(uint hand, uint south, uint west, uint north, uint east, params OrientedTile[] line)
        {
            var view = new MatchView { State = MatchState.Playing, Mode = Mode._2V2, Seat = Seat.South, HandsPlayed = hand, Board = new BoardView() };
            view.Board.Line.Add(line);
            view.Seats.Add(new SeatView { Seat = Seat.North, HandSize = north });
            view.Seats.Add(new SeatView { Seat = Seat.East, HandSize = east });
            view.Seats.Add(new SeatView { Seat = Seat.South, HandSize = south });
            view.Seats.Add(new SeatView { Seat = Seat.West, HandSize = west });
            return view;
        }

        BoardChange Read(MatchView view)
        {
            Assert.IsTrue(snapshot.TryRead(view.ToByteArray()));
            return placements.Read(snapshot);
        }

        [Test]
        public void TheFirstViewOfAConnectionSnaps()
        {
            Assert.AreEqual(BoardChange.Snap, Read(View(1, 6, 6, 6, 6, O(2, 3), O(3, 6))));
            Assert.AreEqual(0, placements.PlacedTiles.Length);

            placements.StartConnection();

            Assert.AreEqual(BoardChange.Snap, Read(View(1, 6, 6, 6, 6, O(2, 3), O(3, 6))));
        }

        [Test]
        public void ANewTileIsPlacedFromTheSeatWhoseHandShrank()
        {
            Read(View(1, 6, 6, 6, 6, O(2, 3), O(3, 6)));

            Assert.AreEqual(BoardChange.Placed, Read(View(1, 6, 5, 6, 6, O(1, 2), O(2, 3), O(3, 6))));

            CollectionAssert.AreEqual(new[] { TileFaces.Index(new Domino(1, 2)) }, placements.PlacedTiles.ToArray());
            Assert.IsTrue(placements.HasPlacer);
            Assert.AreEqual(TableSide.Left, placements.PlacerSide);
        }

        [Test]
        public void EverySeatPlacesFromItsOwnSide()
        {
            Read(View(1, 6, 6, 6, 6, O(3, 3)));

            Read(View(1, 5, 6, 6, 6, O(3, 3), O(3, 4)));
            Assert.AreEqual(TableSide.Bottom, placements.PlacerSide);
            Read(View(1, 5, 6, 5, 6, O(3, 3), O(3, 4), O(4, 4)));
            Assert.AreEqual(TableSide.Top, placements.PlacerSide);
            Read(View(1, 5, 6, 5, 5, O(0, 3), O(3, 3), O(3, 4), O(4, 4)));
            Assert.AreEqual(TableSide.Right, placements.PlacerSide);
            CollectionAssert.AreEqual(new[] { TileFaces.Index(new Domino(0, 3)) }, placements.PlacedTiles.ToArray());
        }

        [Test]
        public void ARepeatedViewPlacesNothing()
        {
            Read(View(1, 6, 6, 6, 6, O(2, 3)));
            Read(View(1, 6, 5, 6, 6, O(2, 3), O(3, 6)));

            Assert.AreEqual(BoardChange.None, Read(View(1, 6, 5, 6, 6, O(2, 3), O(3, 6))));
            Assert.AreEqual(0, placements.PlacedTiles.Length);
        }

        [Test]
        public void ANewHandSnaps()
        {
            Read(View(1, 1, 6, 6, 6, O(2, 3), O(3, 6)));

            Assert.AreEqual(BoardChange.Snap, Read(View(2, 7, 7, 7, 7)));
            Assert.AreEqual(0, placements.PlacedTiles.Length);
        }

        [Test]
        public void ABoardThatLostATileSnaps()
        {
            Read(View(1, 6, 6, 6, 6, O(2, 3), O(3, 6)));

            Assert.AreEqual(BoardChange.Snap, Read(View(1, 6, 5, 6, 6, O(1, 2), O(2, 3))));
            Assert.AreEqual(0, placements.PlacedTiles.Length);
        }

        [Test]
        public void AFinishedMatchSnaps()
        {
            Read(View(1, 6, 6, 6, 6, O(2, 3)));
            var finished = View(1, 6, 5, 6, 6, O(2, 3), O(3, 6));
            finished.State = MatchState.Finished;

            Assert.AreEqual(BoardChange.Snap, Read(finished));
        }

        [Test]
        public void APlacementWithoutAShrunkenHandHasNoPlacer()
        {
            Read(View(1, 6, 6, 6, 6, O(2, 3)));

            Assert.AreEqual(BoardChange.Placed, Read(View(1, 6, 6, 6, 6, O(2, 3), O(3, 6))));
            Assert.IsFalse(placements.HasPlacer);
        }

        [Test]
        public void ReadingAViewAllocatesNothing()
        {
            var before = View(1, 6, 6, 6, 6, O(2, 3)).ToByteArray();
            var after = View(1, 6, 5, 6, 6, O(2, 3), O(3, 6)).ToByteArray();
            snapshot.TryRead(before);
            placements.Read(snapshot);

            Assert.That(() =>
            {
                snapshot.TryRead(after);
                placements.Read(snapshot);
                snapshot.TryRead(before);
                placements.Read(snapshot);
            }, Allocates.Nothing);
        }
    }
}
