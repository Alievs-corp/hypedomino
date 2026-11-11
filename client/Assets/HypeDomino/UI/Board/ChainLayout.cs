using System.Collections.Generic;
using UnityEngine;

namespace HypeDomino.UI.Board
{
    public sealed class ChainLayout
    {
        const float Tolerance = 0.01f;

        PlacedDomino[] placed;
        PlacedDomino[] candidate;

        public ChainLayout(int capacity)
        {
            placed = new PlacedDomino[capacity];
            candidate = new PlacedDomino[capacity];
        }

        public int Count { get; private set; }

        public Rect Bounds { get; private set; }

        public float Scale { get; private set; }

        public Vector2 Focus => Bounds.center;

        public PlacedDomino this[int index] => placed[index];

        public void Arrange(IReadOnlyList<Domino> line, int anchorIndex, IReadOnlyList<Domino> top, IReadOnlyList<Domino> bottom, Vector2 board, float maxScale)
        {
            Count = line.Count + top.Count + bottom.Count;
            Scale = 0f;
            Bounds = default;
            if (line.Count == 0) return;

            var aspect = board.y / board.x;
            var straight = Length(line) + Length(top) + Length(bottom) + 2f;
            for (var halfWidth = straight * Mathf.Max(1f, 1f / aspect); halfWidth >= 1f; halfWidth -= 1f)
            {
                Lay(line, anchorIndex, top, bottom, new Vector2(halfWidth, halfWidth * aspect));
                if (Overlaps()) continue;
                var bounds = Enclose();
                var scale = Mathf.Min(maxScale, Mathf.Min(board.x / bounds.width, board.y / bounds.height));
                if (scale <= Scale) continue;
                Scale = scale;
                Bounds = bounds;
                (placed, candidate) = (candidate, placed);
            }
        }

        static float Length(IReadOnlyList<Domino> tiles)
        {
            var length = 0f;
            for (var i = 0; i < tiles.Count; i++) length += tiles[i].IsDouble ? 1f : 2f;
            return length;
        }

        void Lay(IReadOnlyList<Domino> line, int anchorIndex, IReadOnlyList<Domino> top, IReadOnlyList<Domino> bottom, Vector2 box)
        {
            var anchor = line[anchorIndex];
            candidate[anchorIndex] = new PlacedDomino(Vector2.zero, Vector2.right, anchor.First, anchor.Second, anchor.IsDouble);
            var reach = anchor.IsDouble ? 0.5f : 1f;

            var arm = new Arm(new Vector2(reach, 0f), Vector2.right, box);
            for (var i = anchorIndex + 1; i < line.Count; i++) candidate[i] = arm.Place(line[i].First, line[i].Second, line[i].IsDouble);

            arm = new Arm(new Vector2(-reach, 0f), Vector2.left, box);
            for (var i = anchorIndex - 1; i >= 0; i--) candidate[i] = arm.Place(line[i].Second, line[i].First, line[i].IsDouble);

            LayArm(top, line.Count, new Arm(new Vector2(0f, 1f), Vector2.up, box));
            LayArm(bottom, line.Count + top.Count, new Arm(new Vector2(0f, -1f), Vector2.down, box));
        }

        void LayArm(IReadOnlyList<Domino> tiles, int offset, Arm arm)
        {
            for (var i = 0; i < tiles.Count; i++) candidate[offset + i] = arm.Place(tiles[i].First, tiles[i].Second, tiles[i].IsDouble);
        }

        bool Overlaps()
        {
            for (var i = 0; i < Count; i++)
            {
                var area = candidate[i].Area;
                for (var j = i + 1; j < Count; j++)
                {
                    var other = candidate[j].Area;
                    var width = Mathf.Min(area.xMax, other.xMax) - Mathf.Max(area.xMin, other.xMin);
                    var height = Mathf.Min(area.yMax, other.yMax) - Mathf.Max(area.yMin, other.yMin);
                    if (width > Tolerance && height > Tolerance) return true;
                }
            }
            return false;
        }

        Rect Enclose()
        {
            var min = candidate[0].Area.min;
            var max = candidate[0].Area.max;
            for (var i = 1; i < Count; i++)
            {
                min = Vector2.Min(min, candidate[i].Area.min);
                max = Vector2.Max(max, candidate[i].Area.max);
            }
            return Rect.MinMaxRect(min.x, min.y, max.x, max.y);
        }

        struct Arm
        {
            Vector2 end;
            Vector2 heading;
            readonly Vector2 box;

            public Arm(Vector2 end, Vector2 heading, Vector2 box)
            {
                this.end = end;
                this.heading = heading;
                this.box = box;
            }

            public PlacedDomino Place(byte near, byte far, bool isDouble)
            {
                if (isDouble) return Advance(near, far, 1f, true);
                if (Crosses()) return Turn(near, far);
                return Advance(near, far, 2f, false);
            }

            bool Crosses()
            {
                var limit = heading.x != 0f ? box.x : box.y;
                return Vector2.Dot(end + heading * 2f, heading) > limit + Tolerance;
            }

            PlacedDomino Advance(byte near, byte far, float length, bool across)
            {
                var center = end + heading * (length * 0.5f);
                end += heading * length;
                return new PlacedDomino(center, heading, near, far, across);
            }

            PlacedDomino Turn(byte near, byte far)
            {
                var turned = new Vector2(heading.y, -heading.x);
                var corner = end + heading * 0.5f;
                end = corner + turned * 1.5f;
                heading = turned;
                return new PlacedDomino(corner + turned * 0.5f, turned, near, far, false);
            }
        }
    }
}
