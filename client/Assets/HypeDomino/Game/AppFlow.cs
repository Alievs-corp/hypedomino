using System;
using Hypedomino.V1;

namespace HypeDomino.Game
{
    public sealed class AppFlow
    {
        bool matchFinished;

        public AppState State { get; private set; } = AppState.Boot;

        public event Action<AppState> Changed;

        public void SignInNeeded()
        {
            if (State == AppState.UpdateRequired) return;
            Enter(AppState.Login);
        }

        public void Hello(ServerHello hello)
        {
            switch (hello.Status)
            {
                case HelloStatus.Accepted:
                    matchFinished = false;
                    Enter(hello.InMatch ? AppState.Match : AppState.Lobby);
                    return;
                case HelloStatus.UpdateRequired:
                    Enter(AppState.UpdateRequired);
                    return;
                case HelloStatus.Maintenance:
                    Enter(AppState.Maintenance);
                    return;
            }
        }

        public void View(MatchView view)
        {
            var finished = view.State == MatchState.Finished;
            if (State == AppState.Match)
            {
                matchFinished = finished;
                return;
            }
            if (State != AppState.Lobby || finished) return;
            matchFinished = false;
            Enter(AppState.Match);
        }

        public bool LeaveMatch()
        {
            if (State != AppState.Match || !matchFinished) return false;
            Enter(AppState.Lobby);
            return true;
        }

        void Enter(AppState next)
        {
            if (next == State) return;
            State = next;
            Changed?.Invoke(next);
        }
    }
}
