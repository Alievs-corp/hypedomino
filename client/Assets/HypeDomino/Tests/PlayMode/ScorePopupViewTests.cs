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

        [Test]
        public void APopupShowsPlusThePointsInTheScoreRoleSizedByItsStep()
        {
            view.Show(15, 3, new Vector3(10f, 20f, 0f), 0f);
            view.Tick(TokenDurationMs.Fast);

            var text = texts[0];
            Assert.IsTrue(text.enabled);
            Assert.AreEqual("+15", text.text);
            Assert.AreEqual((float)ScoreCascade.Burst(3, EffectLevel.High, false).ScoreSize, text.fontSize);
            Assert.AreEqual((Color)TokenColor.RoleScoreText, (Color)(Color32)text.color);
            Assert.AreEqual(new Vector3(10f, 20f, 0f), text.rectTransform.position);
            for (var at = 1; at < texts.Length; at++) Assert.IsFalse(texts[at].enabled);
        }

        [Test]
        public void APopupHidesBeforeItStartsAndAfterItsLifetime()
        {
            view.Show(5, 1, Vector3.zero, 100f);

            view.Tick(50f);
            Assert.IsFalse(texts[0].enabled);

            view.Tick(100f + ScoreCascade.Burst(1, EffectLevel.High, false).LifetimeMs);
            Assert.IsFalse(texts[0].enabled);
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

            foreach (var text in texts) Assert.IsFalse(text.enabled);
        }

        [Test]
        public void ShowingAndAnimatingAPopupAllocatesNothing()
        {
            view.Show(25, 5, Vector3.zero, 0f);
            view.Tick(TokenDurationMs.Fast);
            view.Show(35, 6, Vector3.one, 10f);
            view.Tick(TokenDurationMs.Fast + 10f);

            Assert.That(() =>
            {
                view.Show(25, 5, Vector3.zero, 20f);
                view.Tick(TokenDurationMs.Fast + 20f);
                view.Tick(TokenDurationMs.Fast + 40f);
            }, Allocates.Nothing);
        }
    }
}
