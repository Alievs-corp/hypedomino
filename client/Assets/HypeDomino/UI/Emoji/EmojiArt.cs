using System;
using HypeDomino.UI.Motion;
using UnityEngine;

namespace HypeDomino.UI.Emoji
{
    public sealed class EmojiArt
    {
        readonly Sprite still;
        readonly Sprite[] loop;
        readonly float framesPerSecond;

        public EmojiArt(Sprite still, Sprite[] loop, float framesPerSecond)
        {
            this.still = still;
            this.loop = loop ?? Array.Empty<Sprite>();
            this.framesPerSecond = framesPerSecond;
        }

        public Sprite Frame(EffectLevel level, float elapsedSeconds)
        {
            if (level == EffectLevel.Low || loop.Length == 0 || framesPerSecond <= 0f) return still;
            return loop[(int)(Mathf.Max(0f, elapsedSeconds) * framesPerSecond) % loop.Length];
        }
    }
}
