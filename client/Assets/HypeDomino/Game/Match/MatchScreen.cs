using System;
using Google.Protobuf;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Match;
using Hypedomino.V1;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public interface IMatchTable
    {
        Vector2 BoardSize { get; }
        float BoardMaxScale { get; }
        void ShowBackground(MatchSnapshot view);
        void ShowSeats(MatchSnapshot view);
        void ShowScore(ScoreBoard score);
        void ShowBoard(ChainLayout board);
    }

    public interface IMatchHud
    {
        void ShowHand(HandChoice hand);
        void ShowTurn(TurnTimer timer);
        void ShowPass(TableSide side, PassNotice notice);
        void HidePass(TableSide side);
    }

    public sealed class MatchScreen
    {
        readonly IMatchTable table;
        readonly IMatchHud hud;
        readonly BoardChain board = new();
        readonly SoundCues cues = new();
        readonly Action<MatchSnapshot> onView;
        readonly Action<ServerHello> onHello;
        readonly PlaceTileIntent placeIntent = new() { Tile = new Tile() };
        readonly PassIntent passIntent = new();
        readonly DrawIntent drawIntent = new();
        readonly EndTurnIntent endTurnIntent = new();
        Session session;
        Seat lastCurrent;
        TableSide passSide;
        bool passPending;
        bool passShowing;
        double passStartedAt;

        public MatchScreen(IMatchTable table, IMatchHud hud)
        {
            this.table = table ?? throw new ArgumentNullException(nameof(table));
            this.hud = hud ?? throw new ArgumentNullException(nameof(hud));
            onView = Show;
            onHello = _ => StartConnection();
        }

        public ScoreBoard Score { get; } = new();
        public HandChoice Hand { get; } = new();
        public TurnTimer Timer { get; } = new();
        public ChainLayout Board => board.Layout;
        bool CanSend => session != null && session.Accepted;

        public void Bind(Session source)
        {
            Unbind();
            session = source ?? throw new ArgumentNullException(nameof(source));
            session.HelloReceived += onHello;
            session.ViewReceived += onView;
            StartConnection();
        }

        public void Unbind()
        {
            if (session == null) return;
            session.HelloReceived -= onHello;
            session.ViewReceived -= onView;
            session = null;
        }

        public void Show(MatchSnapshot view)
        {
            var passed = (cues.Read(view) & MatchSounds.Pass) != 0;
            var passer = lastCurrent;
            lastCurrent = view.Current;
            table.ShowBackground(view);
            table.ShowSeats(view);
            Score.Read(view);
            table.ShowScore(Score);
            board.Arrange(view, table.BoardSize, table.BoardMaxScale);
            table.ShowBoard(board.Layout);
            Timer.Read(view);
            Hand.Read(view);
            hud.ShowHand(Hand);
            if (passed && TrySideOf(view, passer, out var side)) StartPass(side);
        }

        public void Tick(ServerClock clock, double nowSeconds, bool reduceMotion)
        {
            Timer.Tick(clock, nowSeconds);
            hud.ShowTurn(Timer);
            TickPass(nowSeconds, reduceMotion);
        }

        public HandTap TapTile(int index)
        {
            if (!CanSend) return HandTap.Ignored;
            var tap = Hand.Tap(index, out var placement);
            if (tap == HandTap.Placed) SendPlacement(placement);
            else if (tap != HandTap.Ignored) hud.ShowHand(Hand);
            return tap;
        }

        public bool ChooseEnd(BoardEnd end) =>
            CanSend && Hand.TryChooseEnd(end, out var placement) && SendPlacement(placement);

        public bool DropTile(int index, ReadOnlySpan<Vector2> endPositions, Vector2 drop) =>
            CanSend && Hand.TryDrop(index, endPositions, drop, out var placement) && SendPlacement(placement);

        public bool EndTurn() => CanSend && Hand.TryEndTurn() && SendIntent(Envelope.PayloadOneofCase.EndTurnIntent, endTurnIntent);

        public bool Draw() => CanSend && Hand.TryDraw() && SendIntent(Envelope.PayloadOneofCase.DrawIntent, drawIntent);

        public bool Pass() => CanSend && Hand.TryPass() && SendIntent(Envelope.PayloadOneofCase.PassIntent, passIntent);

        bool SendPlacement(TilePlacement placement)
        {
            placeIntent.Tile.Low = placement.Tile.Low;
            placeIntent.Tile.High = placement.Tile.High;
            placeIntent.End = placement.End;
            return SendIntent(Envelope.PayloadOneofCase.PlaceTileIntent, placeIntent);
        }

        bool SendIntent(Envelope.PayloadOneofCase intent, IMessage message)
        {
            session.Send(intent, message);
            hud.ShowHand(Hand);
            return true;
        }

        void StartConnection()
        {
            cues.StartConnection();
            lastCurrent = Seat.Unspecified;
            passPending = false;
            if (!passShowing) return;
            passShowing = false;
            hud.HidePass(passSide);
        }

        void StartPass(TableSide side)
        {
            if (passShowing) hud.HidePass(passSide);
            passShowing = false;
            passSide = side;
            passPending = true;
        }

        void TickPass(double nowSeconds, bool reduceMotion)
        {
            if (passPending)
            {
                passPending = false;
                passShowing = true;
                passStartedAt = nowSeconds;
            }
            if (!passShowing) return;
            if (PassNotice.At((float)((nowSeconds - passStartedAt) * 1000d), reduceMotion, out var notice))
            {
                hud.ShowPass(passSide, notice);
                return;
            }
            passShowing = false;
            hud.HidePass(passSide);
        }

        static bool TrySideOf(MatchSnapshot view, Seat seat, out TableSide side)
        {
            side = TableSide.Bottom;
            if (view.Mode != Mode._1V1 && view.Mode != Mode._2V2) return false;
            var seatCount = SeatOrder.Size(view.Mode);
            if (!SeatOrder.IsIn(seat, seatCount)) return false;
            if (view.Seat != Seat.Unspecified && !SeatOrder.IsIn(view.Seat, seatCount)) return false;
            side = TableSides.Of(SeatOrder.Viewer(view.Seat, view.Mode), SeatOrder.Clockwise(seat), seatCount);
            return true;
        }
    }
}
