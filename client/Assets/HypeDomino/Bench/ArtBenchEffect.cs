using System;
using System.Collections.Generic;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Bench
{
    [Serializable]
    public sealed class ArtBenchEffectFile
    {
        public ArtBenchEffectSprite[] sprites;
        public ArtBenchEffectEntry[] effects;
    }

    [Serializable]
    public sealed class ArtBenchEffectSprite
    {
        public string name;
        public string material;
    }

    [Serializable]
    public sealed class ArtBenchEffectEntry
    {
        public string name;
        public ArtBenchEffectLevels levels;
    }

    [Serializable]
    public sealed class ArtBenchEffectLevels
    {
        public ArtBenchEffectLevel low;
        public ArtBenchEffectLevel medium;
        public ArtBenchEffectLevel high;

        public ArtBenchEffectLevel Of(EffectLevel level) => level == EffectLevel.High ? high : level == EffectLevel.Medium ? medium : low;
    }

    [Serializable]
    public sealed class ArtBenchEffectLevel
    {
        public ArtBenchEffectLayer[] sprites = Array.Empty<ArtBenchEffectLayer>();
        public ArtBenchEffectParticles[] particles = Array.Empty<ArtBenchEffectParticles>();
    }

    [Serializable]
    public sealed class ArtBenchEffectLayer
    {
        public string sprite;
        public string tint;
    }

    [Serializable]
    public sealed class ArtBenchEffectParticles
    {
        public string sprite;
        public int count;
        public string tint;
    }

    public sealed class ArtBenchEffect : MonoBehaviour
    {
        const float Reach = 420f;
        const float NearestReach = 0.35f;
        const float LayerStartScale = 0.8f;
        const float LayerEndScale = 1.2f;
        const float RepeatGapSeconds = 0.6f;
        const int Seed = 15;

        readonly List<Image> particles = new();
        readonly List<Image> layers = new();
        readonly List<Vector2> paths = new();
        readonly List<Color> particleColors = new();
        readonly List<Color> layerColors = new();
        ArtBenchEffectEntry entry;
        Func<string, Sprite> spriteOf;
        Func<string, Material> materialOf;
        TokenEffectKind kind;
        float durationSeconds;
        float startedAt;
        int liveParticles;
        int liveLayers;

        public string Name => entry.name;

        public void Configure(ArtBenchEffectEntry effect, TokenEffectKind effectKind, Func<string, Sprite> sprites, Func<string, Material> materials)
        {
            entry = effect;
            kind = effectKind;
            spriteOf = sprites;
            materialOf = materials;
            durationSeconds = TokenEffect.DurationMsOfKind[(int)effectKind] / 1000f;
            var random = new System.Random(Seed);
            var most = Math.Max(Count(effect.levels.low), Math.Max(Count(effect.levels.medium), Count(effect.levels.high)));
            for (var at = 0; at < most; at++)
            {
                particles.Add(NewImage("Particle"));
                paths.Add(Path(random));
                particleColors.Add(Color.clear);
            }
            var mostLayers = Math.Max(effect.levels.low.sprites.Length, Math.Max(effect.levels.medium.sprites.Length, effect.levels.high.sprites.Length));
            for (var at = 0; at < mostLayers; at++)
            {
                layers.Add(NewImage("Layer"));
                layerColors.Add(Color.clear);
            }
        }

        public string Play(EffectLevel level, bool reduceMotion, float now)
        {
            startedAt = now;
            var budget = new ParticleBudget(level, reduceMotion);
            liveParticles = 0;
            foreach (var group in entry.levels.Of(budget.ParticleLevel).particles)
            {
                var granted = budget.Take(kind, group.count);
                for (var at = 0; at < granted; at++) Use(particles[liveParticles], particleColors, liveParticles++, group.sprite, group.tint);
            }
            liveLayers = 0;
            foreach (var layer in entry.levels.Of(level).sprites) Use(layers[liveLayers], layerColors, liveLayers++, layer.sprite, layer.tint);
            for (var at = liveParticles; at < particles.Count; at++) particles[at].gameObject.SetActive(false);
            for (var at = liveLayers; at < layers.Count; at++) layers[at].gameObject.SetActive(false);
            Advance(now);
            return $"particles={liveParticles} particleLevel={budget.ParticleLevel} layers={liveLayers}";
        }

        void Update() => Advance(Time.unscaledTime);

        void Advance(float now)
        {
            var elapsed = (now - startedAt) % (durationSeconds + RepeatGapSeconds);
            var progress = Mathf.Clamp01(elapsed / durationSeconds);
            var travel = 1f - (1f - progress) * (1f - progress);
            var fade = 1f - progress;
            for (var at = 0; at < liveParticles; at++)
            {
                particles[at].rectTransform.anchoredPosition = paths[at] * travel;
                particles[at].color = Faded(particleColors[at], fade);
            }
            var scale = Mathf.Lerp(LayerStartScale, LayerEndScale, progress);
            for (var at = 0; at < liveLayers; at++)
            {
                layers[at].rectTransform.localScale = new Vector3(scale, scale, 1f);
                layers[at].color = Faded(layerColors[at], 1f - progress * progress);
            }
        }

        void Use(Image image, List<Color> colors, int at, string sprite, string tint)
        {
            image.sprite = spriteOf(sprite);
            image.material = materialOf(sprite);
            image.SetNativeSize();
            image.rectTransform.anchoredPosition = Vector2.zero;
            colors[at] = TintOf(tint);
            image.gameObject.SetActive(image.sprite != null);
        }

        Image NewImage(string name)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).GetComponent<Image>();
            image.raycastTarget = false;
            image.transform.SetParent(transform, false);
            image.gameObject.SetActive(false);
            return image;
        }

        static Vector2 Path(System.Random random)
        {
            var angle = (float)(random.NextDouble() * Math.PI * 2.0);
            var distance = Reach * Mathf.Lerp(NearestReach, 1f, (float)random.NextDouble());
            return new Vector2(Mathf.Cos(angle), Mathf.Sin(angle)) * distance;
        }

        static int Count(ArtBenchEffectLevel level)
        {
            var total = 0;
            foreach (var group in level.particles) total += group.count;
            return total;
        }

        static Color Faded(Color color, float alpha)
        {
            color.a = alpha;
            return color;
        }

        public static Color TintOf(string token)
        {
            switch (token)
            {
                case "color.gold.hi": return TokenColor.GoldHi;
                case "color.gold.default": return TokenColor.GoldDefault;
                case "color.role.score.fill": return TokenColor.RoleScoreFill;
                case "color.role.secondary.fill": return TokenColor.RoleSecondaryFill;
                case "team": return TokenColor.RoleTeamAFill;
                default: return Color.clear;
            }
        }
    }
}
