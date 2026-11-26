using System;
using HypeDomino.UI.Effects;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using Unity.Profiling;
using UnityEngine;

namespace HypeDomino.UI.Board
{
    public sealed class TilePlaceLayer : MonoBehaviour, IEffectSettingsTarget
    {
        public const string ProfilerMarkerName = "Effect.tilePlace";

        static readonly ProfilerMarker TickMarker = new(ProfilerMarkerName);

        TweenPool pool;
        BoardTileSink sink;
        TilePlaceQueue queue;
        Transform[] placed;
        bool[] settling = Array.Empty<bool>();
        float[] settlesAtMs = Array.Empty<float>();
        float[] fromAngles = Array.Empty<float>();
        EffectPlayer effects;
        EffectLevel level = EffectLevel.Medium;
        bool reduceMotion;

        public event Action<int> Settled;

        public int MovingCount => pool.Count;
        public int Capacity => pool.Capacity;
        public float SettledAtMs => queue.SettledAtMs;
        public Transform LastPlaced { get; private set; }

        public void Build(Transform[] tiles, Transform[] shadows = null)
        {
            placed = tiles;
            sink = new BoardTileSink(tiles, shadows);
            pool = TweenPool.For(level);
            settling = new bool[tiles.Length];
            settlesAtMs = new float[tiles.Length];
            fromAngles = new float[tiles.Length];
        }

        public void Configure(EffectLevel effectLevel, bool reduce)
        {
            reduceMotion = reduce;
            if (effectLevel == level) return;
            level = effectLevel;
            pool.FinishAll(ref sink);
            SettleAllBy(float.NegativeInfinity);
            pool = TweenPool.For(level);
        }

        public void Attach(EffectPlayer player) => effects = player;

        public bool IsSettling(int tile) => settling[tile];

        public void Place(int tile, Vector3 fromPosition, float fromAngle, Vector3 toPosition, float toAngle, float nowMs)
        {
            queue.Place(pool, ref sink, tile, fromPosition, fromAngle, toPosition, toAngle, nowMs, reduceMotion);
            LastPlaced = placed[tile];
            if (queue.SettledAtMs <= nowMs) SettleAllBy(nowMs);
            settling[tile] = true;
            settlesAtMs[tile] = queue.SettledAtMs;
            fromAngles[tile] = fromAngle;
            if (effects == null || queue.SettledAtMs <= nowMs) return;
            var board = placed[tile].parent;
            var startMs = queue.SettledAtMs - TileTravel.PlaceMs(reduceMotion);
            var reach = placed[tile] is RectTransform rect ? EffectRequest.ReachOf(rect) : Vector2.zero;
            var at = board != null ? board.TransformPoint(toPosition) : toPosition;
            var from = board != null ? board.TransformPoint(fromPosition) : fromPosition;
            effects.Play(EffectRequest.On(TokenEffectKind.TilePlace, at, reach, startMs).TrailFrom(from, TileTravel.DurationMs(reduceMotion)));
        }

        public void Follow(int tile, Vector3 toPosition, float toAngle, float nowMs)
        {
            if (!settling[tile]) return;
            var turnedTo = fromAngles[tile] + Mathf.DeltaAngle(fromAngles[tile], toAngle);
            pool.Retarget(tile, TweenChannel.Position, toPosition, nowMs, ref sink);
            pool.Retarget(tile, TweenChannel.Rotation, new Vector3(0f, 0f, turnedTo), nowMs, ref sink);
        }

        public void Land()
        {
            pool.FinishAll(ref sink);
            Array.Clear(settling, 0, settling.Length);
        }

        public void Tick(float nowMs)
        {
            using (TickMarker.Auto()) pool.Step(nowMs, ref sink);
            for (var tile = 0; tile < settling.Length; tile++)
            {
                if (!settling[tile] || settlesAtMs[tile] > nowMs) continue;
                settling[tile] = false;
                Settled?.Invoke(tile);
            }
        }

        void SettleAllBy(float nowMs)
        {
            for (var tile = 0; tile < settling.Length; tile++)
                if (settling[tile] && settlesAtMs[tile] > nowMs) settlesAtMs[tile] = nowMs;
        }

        void Update()
        {
            if (pool != null) Tick(Time.unscaledTime * 1000f);
        }
    }
}
