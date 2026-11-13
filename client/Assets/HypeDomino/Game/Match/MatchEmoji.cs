using HypeDomino.Net;
using HypeDomino.UI.Emoji;
using HypeDomino.UI.Match;
using Hypedomino.V1;

namespace HypeDomino.Game.Match
{
    public sealed class MatchEmoji
    {
        readonly EmojiWheel wheel;
        readonly EmojiIntent intent = new();
        Seat local;
        int viewer;
        int seatCount;

        public MatchEmoji(EmojiWheel wheel) => this.wheel = wheel;

        public void Begin(Seat localSeat, Mode mode)
        {
            local = localSeat;
            seatCount = mode == Mode._1V1 || mode == Mode._2V2 ? SeatOrder.Size(mode) : 0;
            viewer = seatCount == 0 ? 0 : SeatOrder.Viewer(localSeat, mode);
        }

        public bool TrySend(Session session, int slot, long nowMs)
        {
            if (!session.Accepted || !wheel.TrySend(slot, nowMs, out var item)) return false;
            intent.ItemId = item;
            session.Send(Envelope.PayloadOneofCase.EmojiIntent, intent);
            return true;
        }

        public bool TryPlace(Seat sender, long nowMs, out TableSide side)
        {
            side = TableSide.Bottom;
            if (seatCount == 0 || sender < Seat.North || sender > Seat.Y) return false;
            if (IsTwoSeat(sender) != (seatCount == 2)) return false;
            if (sender == local) wheel.OwnEmojiShown(nowMs);
            side = TableSides.Of(viewer, SeatOrder.Clockwise(sender), seatCount);
            return true;
        }

        static bool IsTwoSeat(Seat seat) => seat == Seat.X || seat == Seat.Y;
    }
}
