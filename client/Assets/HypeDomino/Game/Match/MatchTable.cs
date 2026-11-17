using System;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Match;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public sealed class MatchTable : IMatchTable
    {
        readonly SeatAvatars avatars;
        readonly CosmeticManifest manifest;
        readonly ScorePanelView score;
        readonly BoardTilesView tiles;
        readonly MatchLayoutView layout;

        public MatchTable(SeatAvatars avatars, CosmeticManifest manifest, ScorePanelView score, BoardTilesView tiles, MatchLayoutView layout, float boardMaxScale)
        {
            this.avatars = avatars ?? throw new ArgumentNullException(nameof(avatars));
            this.manifest = manifest;
            this.score = score ? score : throw new ArgumentNullException(nameof(score));
            this.tiles = tiles ? tiles : throw new ArgumentNullException(nameof(tiles));
            this.layout = layout ? layout : throw new ArgumentNullException(nameof(layout));
            BoardMaxScale = boardMaxScale;
        }

        public Vector2 BoardSize => layout.Layout.Board.size;
        public float BoardMaxScale { get; }

        public void ShowSeats(MatchSnapshot view) => avatars.Show(view, manifest);

        public void ShowScore(ScoreBoard board) => ScorePanel.Show(score, board);

        public void ShowBoard(ChainLayout board) => tiles.Show(board);
    }
}
