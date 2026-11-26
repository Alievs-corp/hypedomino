using System;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Match;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public sealed class MatchBoard : IDisposable
    {
        const float HandAngle = 0f;

        readonly BoardTilesView tiles;
        readonly TilePlaceLayer placing;
        readonly Transform[] origins;
        readonly MatchEffects effects;
        readonly Func<float> clockMs;
        readonly Action<MatchSounds> playSounds;
        readonly Action<int> settled;

        public MatchBoard(BoardTilesView tiles, TilePlaceLayer placing, Transform[] origins, MatchEffects effects, Func<float> clockMs, Action<MatchSounds> playSounds)
        {
            this.tiles = tiles ? tiles : throw new ArgumentNullException(nameof(tiles));
            this.placing = placing ? placing : throw new ArgumentNullException(nameof(placing));
            this.origins = origins ?? throw new ArgumentNullException(nameof(origins));
            this.clockMs = clockMs ?? throw new ArgumentNullException(nameof(clockMs));
            this.playSounds = playSounds ?? throw new ArgumentNullException(nameof(playSounds));
            this.effects = effects;
            settled = PlaySettled;
            placing.Settled += settled;
        }

        public void Show(ChainLayout board, BoardPlacements placements)
        {
            if (placements.Change == BoardChange.Snap) placing.Land();
            tiles.Show(board);
            var nowMs = clockMs();
            Follow(board, nowMs);
            if (placements.Change != BoardChange.Placed || !placements.HasPlacer) return;
            var origin = OriginOf(placements.PlacerSide);
            if (origin == null) return;
            var placed = placements.PlacedTiles;
            for (var at = 0; at < placed.Length; at++) Place(placed[at], origin, nowMs);
        }

        public void ShowEffects(MatchSnapshot view) => effects?.Read(view, clockMs());

        public void StartConnection() => effects?.StartConnection();

        public void Dispose()
        {
            placing.Settled -= settled;
            placing.Land();
        }

        void Follow(ChainLayout board, float nowMs)
        {
            for (var at = 0; at < board.Count; at++)
            {
                var slot = BoardTilesView.SlotOf(board[at]);
                if (slot >= tiles.Capacity || !placing.IsSettling(slot)) continue;
                var tile = tiles.TileAt(slot);
                placing.Follow(slot, tile.localPosition, tile.localEulerAngles.z, nowMs);
            }
        }

        void Place(int slot, Transform origin, float nowMs)
        {
            if (slot >= tiles.Capacity) return;
            var tile = tiles.TileAt(slot);
            var from = tile.parent.InverseTransformPoint(origin.position);
            placing.Place(slot, from, HandAngle, tile.localPosition, tile.localEulerAngles.z, nowMs);
        }

        Transform OriginOf(TableSide side)
        {
            var at = (int)side;
            return at < origins.Length ? origins[at] : null;
        }

        void PlaySettled(int slot) => playSounds(MatchSounds.TilePlace);
    }
}
