using System;
using System.Collections.Generic;
using System.Globalization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;

namespace HypeDomino.UI.Effects
{
    public sealed class ScorePopupView : MonoBehaviour, IEffectSettingsTarget
    {
        const float MillisecondsPerSecond = 1000f;
        const string Plus = "+";

        [SerializeField] TMP_Text[] popups = Array.Empty<TMP_Text>();

        readonly Dictionary<uint, string> labels = new();
        EffectPool pool;
        float[] startedMs;
        int[] holdMs;
        EffectLevel level = EffectLevel.Medium;
        bool reduceMotion;

        public int Capacity => TokenEffect.ScorePopupPooled;

        void Awake()
        {
            if (pool == null && popups.Length > 0) Build();
        }

        public void Configure(TMP_Text[] texts)
        {
            popups = texts ?? throw new ArgumentNullException(nameof(texts));
            Build();
        }

        public void Configure(EffectLevel effectLevel, bool reduce)
        {
            level = effectLevel;
            reduceMotion = reduce;
        }

        public void Show(uint points, int step, Vector3 at, float startMs)
        {
            if (pool == null || step <= 0) return;
            var burst = ScoreCascade.Burst(step, level, reduceMotion);
            var slot = pool.Begin(startMs, 0);
            startedMs[slot] = startMs;
            holdMs[slot] = burst.HoldMs;
            var text = popups[slot];
            text.rectTransform.position = at;
            text.fontSize = burst.ScoreSize;
            text.color = TokenColor.RoleScoreText;
            text.alpha = 0f;
            text.text = LabelOf(points);
        }

        public void Tick(float nowMs)
        {
            pool.Tick(nowMs);
            for (var slot = 0; slot < popups.Length; slot++) Animate(slot, nowMs);
        }

        void LateUpdate()
        {
            if (pool != null) Tick(Time.unscaledTime * MillisecondsPerSecond);
        }

        void Build()
        {
            if (popups.Length != Capacity) throw new ArgumentException($"{popups.Length} popups for a pool of {Capacity}", nameof(popups));
            pool = new EffectPool(TokenEffectKind.ScorePopup, new ParticleBudget(level, reduceMotion));
            startedMs = new float[Capacity];
            holdMs = new int[Capacity];
            foreach (var text in popups)
            {
                text.canvasRenderer.cullTransparentMesh = true;
                text.alpha = 0f;
            }
        }

        string LabelOf(uint points)
        {
            if (labels.TryGetValue(points, out var label)) return label;
            label = Plus + points.ToString(CultureInfo.InvariantCulture);
            labels.Add(points, label);
            return label;
        }

        void Animate(int slot, float nowMs)
        {
            var text = popups[slot];
            var elapsedMs = nowMs - startedMs[slot];
            var popup = default(ScorePopup);
            var shown = pool.IsLive(slot) && elapsedMs >= 0f && ScorePopup.At(elapsedMs, holdMs[slot], reduceMotion, out popup);
            var alpha = shown ? popup.Alpha : 0f;
            if (text.alpha != alpha) text.alpha = alpha;
            if (shown) text.rectTransform.localScale = new Vector3(popup.Scale, popup.Scale, 1f);
        }
    }
}
