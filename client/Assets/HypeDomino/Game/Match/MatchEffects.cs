using System;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Effects;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public readonly struct EffectCue
    {
        public readonly int StepA;
        public readonly int StepB;
        public readonly uint PointsA;
        public readonly uint PointsB;
        public readonly HandEnding HandEnding;
        public readonly Team HandWinner;
        public readonly Team MatchWinner;

        public EffectCue(int stepA, int stepB, uint pointsA, uint pointsB, HandEnding handEnding, Team handWinner, Team matchWinner)
        {
            StepA = stepA;
            StepB = stepB;
            PointsA = pointsA;
            PointsB = pointsB;
            HandEnding = handEnding;
            HandWinner = handWinner;
            MatchWinner = matchWinner;
        }

        public bool HandEnded => HandEnding != HandEnding.Unspecified;

        public int StepOf(Team team) => team == Team.A ? StepA : team == Team.B ? StepB : 0;
    }

    public sealed class EffectCues
    {
        bool seen;
        MatchState state;
        uint hand;
        uint scoreA;
        uint scoreB;

        public void StartConnection() => seen = false;

        public EffectCue Read(MatchSnapshot view)
        {
            var cue = seen ? CueOf(view) : default;
            seen = true;
            state = view.State;
            hand = view.HandsPlayed;
            scoreA = view.Scores.TeamA;
            scoreB = view.Scores.TeamB;
            return cue;
        }

        EffectCue CueOf(MatchSnapshot view)
        {
            var handEnded = view.HandsPlayed != hand && view.HasLastResult;
            var ending = handEnded ? view.LastResult.Ending : HandEnding.Unspecified;
            var handWinner = handEnded ? view.LastResult.Winner : Team.Unspecified;
            var matchWinner = view.State == MatchState.Finished && state != MatchState.Finished ? view.Winner : Team.Unspecified;
            var placements = (int)view.TurnPlacements;
            var scores = view.Scores;
            return new EffectCue(StepOf(scores.TeamA, scoreA, handEnded, placements), StepOf(scores.TeamB, scoreB, handEnded, placements),
                GainOf(scores.TeamA, scoreA), GainOf(scores.TeamB, scoreB), ending, handWinner, matchWinner);
        }

        static uint GainOf(uint score, uint before) => score > before ? score - before : 0;

        public static int StepOf(uint score, uint before, bool handEnded, int placements)
        {
            if (score <= before) return 0;
            var points = score - before;
            return handEnded ? ScoreCascade.StepOf(points) : ScoreCascade.ChainStepOf(points, placements);
        }
    }

    public sealed class MatchEffects
    {
        readonly EffectPlayer player;
        readonly TilePlaceLayer tiles;
        readonly RectTransform board;
        readonly RectTransform banner;
        readonly RectTransform meterMine;
        readonly RectTransform meterTheirs;
        readonly ScorePopupView popups;
        readonly EffectCues cues = new();
        Team mine = Team.A;

        public MatchEffects(EffectPlayer player, TilePlaceLayer tiles, RectTransform board, RectTransform banner, RectTransform meterMine, RectTransform meterTheirs, ScorePopupView popups)
        {
            this.popups = popups ? popups : throw new ArgumentNullException(nameof(popups));
            this.player = player;
            this.tiles = tiles;
            this.board = board;
            this.banner = banner;
            this.meterMine = meterMine;
            this.meterTheirs = meterTheirs;
        }

        public void StartConnection() => cues.StartConnection();

        public EffectCue Read(MatchSnapshot view, float nowMs)
        {
            var cue = cues.Read(view);
            mine = view.Seat == Seat.Unspecified ? Team.A : SeatOrder.TeamOf(view.Seat);
            Play(in cue, nowMs);
            return cue;
        }

        void Play(in EffectCue cue, float nowMs)
        {
            var settledMs = Mathf.Max(nowMs, tiles != null ? tiles.SettledAtMs : nowMs);
            var lastTile = tiles != null ? tiles.LastPlaced as RectTransform : null;
            var handSpot = lastTile != null ? lastTile : board;
            var presentedMs = settledMs;
            if (cue.HandEnding == HandEnding.Domino)
            {
                player.Play(EffectRequest.On(TokenEffectKind.Domino, handSpot, settledMs).ForTeam(ScorePanel.ColourOf(cue.HandWinner)));
                presentedMs += TokenEffect.DominoDurationMs;
            }
            if (cue.HandEnding == HandEnding.Block)
            {
                player.Play(EffectRequest.On(TokenEffectKind.Block, board, settledMs));
                presentedMs += TokenEffect.BlockDurationMs;
            }
            var pointsMs = cue.HandEnded ? settledMs + TokenDurationMs.Emphasis : settledMs;
            PlayPoints(Team.A, cue.StepA, cue.PointsA, cue.HandEnded, lastTile, pointsMs);
            PlayPoints(Team.B, cue.StepB, cue.PointsB, cue.HandEnded, lastTile, pointsMs);
            if (cue.MatchWinner == Team.Unspecified) return;
            player.Play(EffectRequest.On(TokenEffectKind.MatchWin, banner, presentedMs).ForTeam(ScorePanel.ColourOf(cue.MatchWinner)));
        }

        void PlayPoints(Team team, int step, uint points, bool handEnded, RectTransform lastTile, float startMs)
        {
            if (step == 0) return;
            var meter = team == mine ? meterMine : meterTheirs;
            var request = handEnded || lastTile == null
                ? EffectRequest.On(TokenEffectKind.ScoreCascade, meter, startMs)
                : EffectRequest.On(TokenEffectKind.ScoreCascade, lastTile, startMs).TrailTo(meter.position);
            player.Play(request.AtStep(step));
            popups.Show(points, step, request.At, startMs);
        }
    }
}
