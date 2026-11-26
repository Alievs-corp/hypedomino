using System;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Match;
using Hypedomino.V1;

namespace HypeDomino.Game.Match
{
    public enum BoardChange
    {
        None,
        Snap,
        Placed,
    }

    public sealed class BoardPlacements
    {
        readonly bool[] laid = new bool[TileFaces.SetSize];
        readonly int[] placed = new int[TileFaces.SetSize];
        readonly Seat[] seats = new Seat[MatchSnapshot.MaxSeats];
        readonly uint[] sizes = new uint[MatchSnapshot.MaxSeats];
        int laidCount;
        int placedCount;
        int seatCount;
        bool seen;
        uint hand;

        public BoardChange Change { get; private set; }
        public ReadOnlySpan<int> PlacedTiles => new(placed, 0, placedCount);
        public bool HasPlacer { get; private set; }
        public TableSide PlacerSide { get; private set; }

        public void StartConnection() => seen = false;

        public BoardChange Read(MatchSnapshot view)
        {
            placedCount = 0;
            HasPlacer = false;
            var sameHand = seen && hand == view.HandsPlayed && IsInHand(view.State);
            Change = sameHand ? Compare(view) : BoardChange.Snap;
            if (Change == BoardChange.Placed) FindPlacer(view);
            else placedCount = 0;
            Keep(view);
            return Change;
        }

        BoardChange Compare(MatchSnapshot view)
        {
            var still = Collect(view.Line) + Collect(view.Top) + Collect(view.Bottom);
            if (still < laidCount) return BoardChange.Snap;
            return placedCount > 0 ? BoardChange.Placed : BoardChange.None;
        }

        int Collect(ReadOnlySpan<OrientedTileFace> tiles)
        {
            var still = 0;
            for (var at = 0; at < tiles.Length; at++)
            {
                var tile = IndexOf(tiles[at]);
                if (laid[tile]) still++;
                else if (placedCount < placed.Length) placed[placedCount++] = tile;
            }
            return still;
        }

        void FindPlacer(MatchSnapshot view)
        {
            var states = view.Seats;
            for (var at = 0; at < states.Length; at++)
            {
                if (states[at].HandSize >= SizeBefore(states[at].Seat)) continue;
                HasPlacer = MatchScreen.TrySideOf(view, states[at].Seat, out var side);
                PlacerSide = side;
                return;
            }
        }

        void Keep(MatchSnapshot view)
        {
            seen = true;
            hand = view.HandsPlayed;
            Array.Clear(laid, 0, laid.Length);
            laidCount = Mark(view.Line) + Mark(view.Top) + Mark(view.Bottom);
            var states = view.Seats;
            seatCount = Math.Min(states.Length, seats.Length);
            for (var at = 0; at < seatCount; at++)
            {
                seats[at] = states[at].Seat;
                sizes[at] = states[at].HandSize;
            }
        }

        int Mark(ReadOnlySpan<OrientedTileFace> tiles)
        {
            for (var at = 0; at < tiles.Length; at++) laid[IndexOf(tiles[at])] = true;
            return tiles.Length;
        }

        uint SizeBefore(Seat seat)
        {
            for (var at = 0; at < seatCount; at++)
                if (seats[at] == seat) return sizes[at];
            return 0;
        }

        static int IndexOf(OrientedTileFace tile) => TileFaces.Index(new Domino((byte)tile.First, (byte)tile.Second));

        static bool IsInHand(MatchState state) => state == MatchState.Opening || state == MatchState.Playing;
    }
}
