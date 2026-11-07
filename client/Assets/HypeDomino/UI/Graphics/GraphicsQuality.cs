using System;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;

namespace HypeDomino.UI.Graphics
{
    public sealed class GraphicsQuality
    {
        public GraphicsQuality(DeviceTier tier, GraphicsLevel? savedChoice)
        {
            Tier = tier;
            Choice = savedChoice > Highest ? Highest : savedChoice;
        }

        public event Action Changed;

        public DeviceTier Tier { get; }

        public GraphicsLevel? Choice { get; private set; }

        public GraphicsLevel Auto => (GraphicsLevel)TokenGraphics.AutoLevelOfTier[(int)Tier];

        public GraphicsLevel Highest => (GraphicsLevel)TokenGraphics.HighestLevelOfTier[(int)Tier];

        public GraphicsLevel Level => Choice ?? Auto;

        public EffectLevel Effects => (EffectLevel)TokenGraphics.EffectLevelOfLevel[(int)Level];

        public bool CanChoose(GraphicsLevel level) => level <= Highest;

        public bool Choose(GraphicsLevel? level)
        {
            if (level.HasValue && !CanChoose(level.Value)) return false;
            if (level == Choice) return true;
            Choice = level;
            Changed?.Invoke();
            return true;
        }

        public float RenderScale(int nativeShortSide)
        {
            if (nativeShortSide <= 0) return 1f;
            return Math.Min(1f, (float)TokenGraphics.MaxRenderShortSideOfLevel[(int)Level] / nativeShortSide);
        }
    }
}
