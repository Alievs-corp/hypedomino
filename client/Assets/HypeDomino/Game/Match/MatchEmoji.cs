using System;
using HypeDomino.Net;
using HypeDomino.UI.Emoji;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Match;
using Hypedomino.V1;

namespace HypeDomino.Game.Match
{
    public sealed class MatchEmoji
    {
        public const string MuteKey = "match.seat.mute";
        public const string UnmuteKey = "match.seat.unmute";

        readonly EmojiWheel wheel;
        readonly SeatMessages messages;
        readonly GameStrings strings;
        readonly EmojiIntent intent = new();
        readonly QuickMessageIntent messageIntent = new();
        readonly bool[] muted = new bool[(int)Seat.Y + 1];
        Seat local;
        int viewer;
        int seatCount;

        public MatchEmoji(EmojiWheel wheel, SeatMessages messages, GameStrings strings)
        {
            this.wheel = wheel;
            this.messages = messages;
            this.strings = strings;
        }

        public void Begin(Seat localSeat, Mode mode)
        {
            local = localSeat;
            seatCount = mode == Mode._1V1 || mode == Mode._2V2 ? SeatOrder.Size(mode) : 0;
            viewer = seatCount == 0 ? 0 : SeatOrder.Viewer(localSeat, mode);
            Array.Clear(muted, 0, muted.Length);
            messages.Clear();
        }

        public void End() => Array.Clear(muted, 0, muted.Length);

        public bool TrySend(Session session, int slot, long nowMs)
        {
            if (!session.Accepted || !wheel.TrySend(slot, nowMs, out var item)) return false;
            intent.ItemId = item;
            session.Send(Envelope.PayloadOneofCase.EmojiIntent, intent);
            return true;
        }

        public bool TrySendMessage(Session session, string key, bool teamOnly)
        {
            if (!session.Accepted) return false;
            messageIntent.Key = key;
            messageIntent.TeamOnly = teamOnly;
            session.Send(Envelope.PayloadOneofCase.QuickMessageIntent, messageIntent);
            return true;
        }

        public bool TryPlace(Seat sender, long nowMs, out TableSide side)
        {
            if (!TryRoute(sender, out side)) return false;
            if (sender == local) wheel.OwnEmojiShown(nowMs);
            return true;
        }

        public bool TryShowMessage(QuickMessageShown shown, long nowMs)
        {
            if (!TryRoute(shown.Seat, out var side)) return false;
            messages.Show(side, strings.Text(shown.Key), nowMs);
            return true;
        }

        public bool CanMute(Seat seat) => InMatch(seat) && seat != local;

        public bool IsMuted(Seat seat) => InMatch(seat) && muted[(int)seat];

        public string MuteLabelKey(Seat seat) => IsMuted(seat) ? UnmuteKey : MuteKey;

        public bool ToggleMute(Seat seat)
        {
            if (!CanMute(seat)) return false;
            muted[(int)seat] = !muted[(int)seat];
            return true;
        }

        bool TryRoute(Seat sender, out TableSide side)
        {
            side = TableSide.Bottom;
            if (!InMatch(sender) || muted[(int)sender]) return false;
            side = TableSides.Of(viewer, SeatOrder.Clockwise(sender), seatCount);
            return true;
        }

        bool InMatch(Seat seat) =>
            seatCount != 0 && seat >= Seat.North && seat <= Seat.Y && IsTwoSeat(seat) == (seatCount == 2);

        static bool IsTwoSeat(Seat seat) => seat == Seat.X || seat == Seat.Y;
    }
}
