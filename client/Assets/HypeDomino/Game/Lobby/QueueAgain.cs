using HypeDomino.Net;
using Hypedomino.V1;

namespace HypeDomino.Game.Lobby
{
    public sealed class QueueAgain
    {
        const int NoCard = -1;

        readonly MatchQueue queue;
        readonly LobbyPlay play;
        readonly AppFlow flow;
        int foundCard = NoCard;

        public QueueAgain(MatchQueue queue, LobbyPlay play, AppFlow flow)
        {
            this.queue = queue;
            this.play = play;
            this.flow = flow;
            queue.Changed += OnQueueChanged;
            flow.Changed += OnFlowChanged;
        }

        public bool Offered => foundCard != NoCard && flow.State == AppState.Match;

        public bool Tap(Session session)
        {
            if (!Offered || !session.Accepted) return false;
            var card = foundCard;
            if (!flow.LeaveMatch() || flow.State != AppState.Lobby) return false;
            return play.Play(session, card);
        }

        void OnQueueChanged()
        {
            if (queue.Phase == QueuePhase.MatchFound) foundCard = CardOf(queue.Mode, queue.Kind);
        }

        void OnFlowChanged(AppState state)
        {
            if (state != AppState.Match) foundCard = NoCard;
        }

        static int CardOf(Mode mode, QueueKind kind)
        {
            for (var card = 0; card < PlayModes.Count; card++)
                if (PlayModes.At(card).Is(mode, kind)) return card;
            return NoCard;
        }
    }
}
