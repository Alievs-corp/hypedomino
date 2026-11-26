using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using UnityEngine;

namespace HypeDomino.UI.Board
{
    public readonly struct BoardTileSink : ITweenSink
    {
        static readonly Vector3 RestingShadowOffset = new Vector3(0f, -TokenElevation.E1OffsetY, 0f);

        readonly Transform[] tiles;
        readonly Transform[] shadows;

        public BoardTileSink(Transform[] tiles, Transform[] shadows)
        {
            this.tiles = tiles;
            this.shadows = shadows;
        }

        public void Apply(int target, TweenChannel channel, Vector3 value)
        {
            Move(tiles[target], channel, value);
            if (shadows != null && target < shadows.Length && shadows[target] != null)
                Move(shadows[target], channel, channel == TweenChannel.Position ? value + RestingShadowOffset : value);
        }

        static void Move(Transform moved, TweenChannel channel, Vector3 value)
        {
            switch (channel)
            {
                case TweenChannel.Position:
                    moved.localPosition = value;
                    return;
                case TweenChannel.Rotation:
                    moved.localRotation = Quaternion.Euler(value);
                    return;
                case TweenChannel.Scale:
                    moved.localScale = value;
                    return;
            }
        }
    }
}
