using System;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;

namespace HypeDomino.UI.Effects
{
    public sealed class ScorePopupView : MonoBehaviour, IEffectSettingsTarget
    {
        const float MillisecondsPerSecond = 1000f;

        [SerializeField] TMP_Text[] popups = Array.Empty<TMP_Text>();

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
            text.canvasRenderer.cull = true;
            text.rectTransform.position = at;
            text.fontSize = burst.ScoreSize;
            text.color = TokenColor.RoleScoreText;
            text.SetText("+{0}", points);
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
                text.enabled = true;
                text.SetText("+{0}", 0);
                text.canvasRenderer.cull = true;
            }
        }

        void Animate(int slot, float nowMs)
        {
            var text = popups[slot];
            var elapsedMs = nowMs - startedMs[slot];
            var popup = default(ScorePopup);
            var shown = pool.IsLive(slot) && elapsedMs >= 0f && ScorePopup.At(elapsedMs, holdMs[slot], reduceMotion, out popup);
            if (text.canvasRenderer.cull == shown) text.canvasRenderer.cull = !shown;
            if (!shown) return;
            text.rectTransform.localScale = new Vector3(popup.Scale, popup.Scale, 1f);
            text.alpha = popup.Alpha;
        }
    }
}
