using HypeDomino.UI.Effects;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using TMPro;
using UnityEngine;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScorePopupViewTests
    {
        GameObject root;
        ScorePopupView view;
        TMP_Text[] texts;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Screen", typeof(RectTransform), typeof(Canvas));
            view = new GameObject("Popups", typeof(RectTransform)).AddComponent<ScorePopupView>();
            view.transform.SetParent(root.transform, false);
            texts = new TMP_Text[TokenEffect.ScorePopupPooled];
            for (var at = 0; at < texts.Length; at++)
            {
                texts[at] = new GameObject("Popup " + at, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
                texts[at].transform.SetParent(view.transform, false);
            }
            view.Configure(texts);
            view.Configure(EffectLevel.High, false);
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        static bool Shown(TMP_Text text) => text.alpha > 0f;

        [Test]
        public void APopupShowsPlusThePointsInTheScoreRoleSizedByItsStep()
        {
            view.Show(15, 3, new Vector3(10f, 20f, 0f), 0f);
            view.Tick(TokenDurationMs.Fast);

            var text = texts[0];
            Assert.IsTrue(Shown(text));
            Assert.AreEqual("+15", text.text);
            Assert.AreEqual((float)ScoreCascade.Burst(3, EffectLevel.High, false).ScoreSize, text.fontSize);
            Assert.AreEqual((Color)TokenColor.RoleScoreText, (Color)(Color32)text.color);
            Assert.AreEqual(new Vector3(10f, 20f, 0f), text.rectTransform.position);
            for (var at = 1; at < texts.Length; at++) Assert.IsFalse(Shown(texts[at]));
        }

        [Test]
        public void APopupHidesBeforeItStartsAndAfterItsLifetime()
        {
            view.Show(5, 1, Vector3.zero, 100f);

            view.Tick(50f);
            Assert.IsFalse(Shown(texts[0]));

            view.Tick(100f + ScoreCascade.Burst(1, EffectLevel.High, false).LifetimeMs);
            Assert.IsFalse(Shown(texts[0]));
        }

        [Test]
        public void ReduceMotionFadesInsteadOfScaling()
        {
            view.Configure(EffectLevel.High, true);
            view.Show(10, 2, Vector3.zero, 0f);

            view.Tick(TokenDurationMs.Fast / 2f);

            Assert.AreEqual(1f, texts[0].rectTransform.localScale.x, 1e-5f);
            Assert.Less(texts[0].alpha, 1f);
        }

        [Test]
        public void AFullPoolReusesTheOldestPopup()
        {
            for (var at = 0; at < texts.Length; at++) view.Show((uint)(5 * (at + 1)), 1, Vector3.zero, at);

            view.Show(30, 6, Vector3.zero, texts.Length);

            Assert.AreEqual("+30", texts[0].text);
        }

        [Test]
        public void AStepOfZeroShowsNothing()
        {
            view.Show(0, 0, Vector3.zero, 0f);
            view.Tick(TokenDurationMs.Fast);

            foreach (var text in texts)
            {
                Assert.IsFalse(Shown(text));
                Assert.IsTrue(text.canvasRenderer.cullTransparentMesh);
            }
        }

        [Test]
        public void ShowingAndAnimatingAPopupAllocatesNothing()
        {
            for (var at = 0; at < texts.Length; at++)
            {
                view.Show(at % 2 == 0 ? 25u : 35u, 5 + at % 2, Vector3.one, at);
                view.Tick(TokenDurationMs.Fast + at);
            }
            view.Tick(TokenEffect.ScorePopupDurationMs * 2f);
            Canvas.ForceUpdateCanvases();

            Assert.That(() =>
            {
                view.Show(35, 6, Vector3.zero, 3000f);
                view.Tick(3000f + TokenDurationMs.Fast);
                view.Tick(3000f + TokenDurationMs.Fast + 20f);
                view.Tick(3000f + TokenEffect.ScorePopupDurationMs);
            }, Allocates.Nothing);
        }
    }
}
