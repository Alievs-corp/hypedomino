using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using UnityEngine;

namespace HypeDomino.UI.Board
{
    public static class TileTravel
    {
        public static float DurationMs(bool reduceMotion) => reduceMotion ? TokenDurationMs.Fast : TokenDurationMs.Base;

        public static void Start<TSink>(TweenPool pool, ref TSink sink, int tile, Vector3 fromPosition, float fromAngle, Vector3 toPosition, float toAngle, float nowMs, bool reduceMotion)
            where TSink : ITweenSink
        {
            var duration = DurationMs(reduceMotion);
            var turnedTo = fromAngle + Mathf.DeltaAngle(fromAngle, toAngle);
            pool.Start(new Tween(tile, TweenChannel.Position, fromPosition, toPosition, nowMs, duration, TokenEasing.Standard, false), ref sink);
            pool.Start(new Tween(tile, TweenChannel.Rotation, new Vector3(0f, 0f, fromAngle), new Vector3(0f, 0f, turnedTo), nowMs, duration, TokenEasing.Standard, false), ref sink);
        }
    }
}
