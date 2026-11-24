using HypeDomino.UI.Match;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public static class ScorePanel
    {
        public static void Show(ScorePanelView panel, ScoreBoard board)
        {
            panel.ShowTarget(board.TargetScore);
            Show(panel.Mine, board.Mine, board.LockMark);
            Show(panel.Theirs, board.Theirs, board.LockMark);
        }

        public static Color32 ColourOf(Team team) => team == Team.A ? TokenColor.RoleTeamAFill : TokenColor.RoleTeamBFill;

        static void Show(TeamScoreView view, TeamMeter meter, float lockMark)
        {
            view.Show(ColourOf(meter.Team), meter.Score, meter.Fill, meter.Locked);
            view.ShowLockMark(lockMark);
        }
    }
}
