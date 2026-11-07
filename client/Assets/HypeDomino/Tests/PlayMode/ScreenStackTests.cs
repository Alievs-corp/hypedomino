using System.Collections;
using System.Collections.Generic;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScreenStackTests
    {
        readonly List<GameObject> created = new();
        ScreenStack stack;

        [SetUp]
        public void CreateStack()
        {
            stack = Track(new GameObject("Stack")).AddComponent<ScreenStack>();
            stack.enabled = false;
        }

        [TearDown]
        public void DestroyAll()
        {
            foreach (var item in created) Object.Destroy(item);
            created.Clear();
        }

        GameObject Track(GameObject item)
        {
            created.Add(item);
            return item;
        }

        ScreenView NewScreen(string name, bool modal = false)
        {
            var screen = Track(new GameObject(name, typeof(RectTransform))).AddComponent<ScreenView>();
            var content = new GameObject("Content", typeof(RectTransform), typeof(CanvasGroup));
            content.transform.SetParent(screen.transform, false);
            Image scrim = null;
            if (modal)
            {
                scrim = new GameObject("Scrim", typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                scrim.transform.SetParent(screen.transform, false);
                scrim.transform.SetAsFirstSibling();
            }
            screen.Configure(content.GetComponent<CanvasGroup>(), scrim);
            return screen;
        }

        static Transform ContentOf(ScreenView screen) => screen.GetComponentInChildren<CanvasGroup>().transform;

        static float ScrimAlpha(ScreenView screen) => screen.GetComponentInChildren<Image>().canvasRenderer.GetAlpha();

        ScreenView SettledRoot()
        {
            var root = NewScreen("Root");
            stack.Push(root);
            stack.Advance(TokenDurationMs.Slow);
            return root;
        }

        [Test]
        public void APushedScreenTakesInputFromItsFirstFrame()
        {
            var root = SettledRoot();
            var next = NewScreen("Next");

            stack.Push(next);

            Assert.AreSame(next, stack.Top);
            Assert.IsTrue(next.IsShown);
            Assert.IsTrue(next.TakesInput);
            Assert.IsFalse(root.TakesInput);
            Assert.AreEqual(0f, next.Visibility);
            Assert.Greater(next.SortingOrder, root.SortingOrder);
        }

        [Test]
        public void TheScreenBelowStaysUntilThePushEndsAndThenHides()
        {
            var root = SettledRoot();
            var next = NewScreen("Next");
            stack.Push(next);

            stack.Advance(TokenDurationMs.Slow - 1);
            Assert.IsTrue(root.IsShown);
            Assert.Less(next.Visibility, 1f);
            Assert.Greater(ContentOf(next).localScale.x, TokenMotion.PanelEnterScale);

            stack.Advance(1);
            Assert.AreEqual(1f, next.Visibility);
            Assert.AreEqual(1f, ContentOf(next).localScale.x);
            Assert.IsFalse(root.IsShown);
        }

        [Test]
        public void PopHandsInputBackAtOnceAndHidesTheTopAfterItsExit()
        {
            var root = SettledRoot();
            var next = NewScreen("Next");
            stack.Push(next);
            stack.Advance(TokenDurationMs.Slow);

            Assert.IsTrue(stack.Pop());

            Assert.AreSame(root, stack.Top);
            Assert.IsTrue(root.IsShown);
            Assert.IsTrue(root.TakesInput);
            Assert.IsFalse(next.TakesInput);
            Assert.IsTrue(next.IsShown);

            stack.Advance(TokenDurationMs.Base);
            Assert.AreEqual(0f, next.Visibility);
            Assert.AreEqual(TokenMotion.PanelEnterScale, ContentOf(next).localScale.x, 0.0001f);
            Assert.IsFalse(next.IsShown);
        }

        [Test]
        public void ThePopOfAPushInProgressReversesFromItsCurrentValue()
        {
            SettledRoot();
            var next = NewScreen("Next");
            stack.Push(next);
            stack.Advance(TokenDurationMs.Slow / 4f);
            var reached = next.Visibility;
            Assert.That(reached, Is.GreaterThan(0f).And.LessThan(1f));

            stack.Pop();
            stack.Advance(0.01f);
            Assert.AreEqual(reached, next.Visibility, 0.01f);

            stack.Advance(TokenDurationMs.Base * reached);
            Assert.AreEqual(0f, next.Visibility);
            Assert.IsFalse(next.IsShown);
        }

        [Test]
        public void PushingAScreenBackDuringItsExitReversesIt()
        {
            SettledRoot();
            var next = NewScreen("Next");
            stack.Push(next);
            stack.Advance(TokenDurationMs.Slow);
            stack.Pop();
            stack.Advance(TokenDurationMs.Base / 2f);
            var reached = next.Visibility;

            stack.Push(next);
            stack.Advance(0.01f);

            Assert.AreEqual(reached, next.Visibility, 0.01f);
            Assert.IsTrue(next.TakesInput);
            stack.Advance(TokenDurationMs.Slow);
            Assert.AreEqual(1f, next.Visibility);
            Assert.IsTrue(next.IsShown);
        }

        [Test]
        public void AModalKeepsTheScreenBelowVisibleBehindTheScrim()
        {
            var root = SettledRoot();
            var modal = NewScreen("Modal", true);

            stack.Push(modal);
            stack.Advance(TokenDurationMs.Moderate);

            Assert.IsTrue(root.IsShown);
            Assert.IsFalse(root.TakesInput);
            Assert.AreEqual(1f, modal.Visibility);
            Assert.AreEqual(TokenOpacity.Scrim, ScrimAlpha(modal), 0.0001f);

            stack.Pop();
            stack.Advance(TokenDurationMs.Base);
            Assert.AreEqual(0f, ScrimAlpha(modal), 0.0001f);
            Assert.IsFalse(modal.IsShown);
        }

        [Test]
        public void PoppingBackToAModalShowsTheScreenUnderItToo()
        {
            var root = SettledRoot();
            var modal = NewScreen("Modal", true);
            var next = NewScreen("Next");
            stack.Push(modal);
            stack.Advance(TokenDurationMs.Moderate);
            stack.Push(next);
            stack.Advance(TokenDurationMs.Slow);
            Assert.IsFalse(root.IsShown);
            Assert.IsFalse(modal.IsShown);

            stack.Pop();

            Assert.IsTrue(modal.IsShown);
            Assert.IsTrue(modal.TakesInput);
            Assert.IsTrue(root.IsShown);
            Assert.IsFalse(root.TakesInput);
        }

        [Test]
        public void TheRootScreenIsNeverPopped()
        {
            var root = SettledRoot();
            Assert.IsFalse(stack.Pop());
            Assert.AreSame(root, stack.Top);
            Assert.IsTrue(root.TakesInput);
        }

        [Test]
        public void ReduceMotionFadesWithoutScaling()
        {
            SettledRoot();
            stack.ReduceMotion = true;
            var next = NewScreen("Next");
            stack.Push(next);

            stack.Advance(TokenDurationMs.Slow / 2f);

            Assert.That(next.Visibility, Is.GreaterThan(0f).And.LessThan(1f));
            Assert.AreEqual(1f, ContentOf(next).localScale.x);
        }

        [UnityTest]
        public IEnumerator TheScreenCanvasMapsItsShortSideTo1080Units()
        {
            var root = SettledRoot();
            yield return null;

            var expected = Mathf.Min(Screen.width, Screen.height) / 1080f;
            Assert.AreEqual(expected, root.GetComponent<Canvas>().scaleFactor, 0.001f);
        }
    }
}
