using System;
using HypeDomino.UI.Match;

namespace HypeDomino.Game.Match
{
    public sealed class MatchHud : IMatchHud
    {
        const int SideCount = (int)TableSide.Right + 1;

        readonly bool[] passShown = new bool[SideCount];
        readonly PassNotice[] passes = new PassNotice[SideCount];

        public HandChoice Hand { get; private set; }
        public TurnTimer Timer { get; private set; }
        public MatchNotices Notices { get; private set; }
        public HandArrivals Arrivals { get; private set; }
        public int HandVersion { get; private set; }
        public int ArrivalsVersion { get; private set; }
        public int NoticesVersion { get; private set; }
        public int PassVersion { get; private set; }
        public HandSummary Summary { get; private set; }
        public int SummaryVersion { get; private set; }

        public void ShowHand(HandChoice hand)
        {
            Hand = hand;
            HandVersion++;
        }

        public void ShowArrivals(HandArrivals arrivals)
        {
            Arrivals = arrivals;
            ArrivalsVersion++;
        }

        public void ShowTurn(TurnTimer timer) => Timer = timer;

        public void ShowPass(TableSide side, PassNotice notice)
        {
            passes[(int)side] = notice;
            passShown[(int)side] = true;
            PassVersion++;
        }

        public void HidePass(TableSide side)
        {
            passShown[(int)side] = false;
            PassVersion++;
        }

        public void ShowNotices(MatchNotices notices)
        {
            Notices = notices;
            NoticesVersion++;
        }

        public void ShowSummary(HandSummary summary)
        {
            Summary = summary;
            SummaryVersion++;
        }

        public bool TryGetPass(TableSide side, out PassNotice notice)
        {
            notice = passes[(int)side];
            return passShown[(int)side];
        }

        public void Clear()
        {
            Hand = null;
            Arrivals = null;
            Timer = null;
            Notices = null;
            Summary = null;
            Array.Clear(passShown, 0, passShown.Length);
            HandVersion++;
            ArrivalsVersion++;
            NoticesVersion++;
            PassVersion++;
            SummaryVersion++;
        }
    }
}
