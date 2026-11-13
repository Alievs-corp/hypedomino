using HypeDomino.UI.Motion;
using Unity.Profiling;
using UnityEngine;

namespace HypeDomino.UI.Board
{
    public sealed class TilePlaceLayer : MonoBehaviour
    {
        public const string ProfilerMarkerName = "Effect.tilePlace";

        static readonly ProfilerMarker TickMarker = new(ProfilerMarkerName);

        TweenPool pool;
        TransformTweenSink sink;
        TilePlaceQueue queue;
        EffectLevel level = EffectLevel.Medium;
        bool reduceMotion;

        public int MovingCount => pool.Count;
        public int Capacity => pool.Capacity;
        public float SettledAtMs => queue.SettledAtMs;

        public void Build(Transform[] tiles)
        {
            sink = new TransformTweenSink(tiles);
            pool = TweenPool.For(level);
        }

        public void Configure(EffectLevel effectLevel, bool reduce)
        {
            reduceMotion = reduce;
            if (effectLevel == level) return;
            level = effectLevel;
            pool.FinishAll(ref sink);
            pool = TweenPool.For(level);
        }

        public void Place(int tile, Vector3 fromPosition, float fromAngle, Vector3 toPosition, float toAngle, float nowMs) =>
            queue.Place(pool, ref sink, tile, fromPosition, fromAngle, toPosition, toAngle, nowMs, reduceMotion);

        public void Tick(float nowMs)
        {
            using (TickMarker.Auto()) pool.Step(nowMs, ref sink);
        }

        void Update()
        {
            if (pool != null) Tick(Time.unscaledTime * 1000f);
        }
    }
}
