using System;
using System.Collections;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class PushGcDiag
    {
        readonly System.Text.StringBuilder report = new();

        void Note(string line) => report.AppendLine(line);

        static string Cost(Action action)
        {
            var recorder = UnityEngine.Profiling.Recorder.Get("GC.Alloc");
            recorder.enabled = false;
            recorder.FilterToCurrentThread();
            var before = GC.GetTotalMemory(false);
            recorder.enabled = true;
            action();
            recorder.enabled = false;
            var after = GC.GetTotalMemory(false);
            recorder.CollectFromAllThreads();
            return $"{recorder.sampleBlockCount}x/{after - before}B";
        }

        static ScreenView Screen(string name, int images, int texts)
        {
            var screen = new GameObject(name, typeof(RectTransform)).AddComponent<ScreenView>();
            var content = new GameObject("Content", typeof(RectTransform), typeof(CanvasGroup));
            content.transform.SetParent(screen.transform, false);
            var font = Resources.GetBuiltinResource<Font>("LegacyRuntime.ttf");
            for (var at = 0; at < images; at++)
            {
                var image = new GameObject("I", typeof(RectTransform), typeof(Image));
                image.transform.SetParent(content.transform, false);
            }
            for (var at = 0; at < texts; at++)
            {
                var text = new GameObject("T", typeof(RectTransform), typeof(Text)).GetComponent<Text>();
                text.transform.SetParent(content.transform, false);
                text.font = font;
                text.text = "Item " + at;
            }
            screen.Configure(content.GetComponent<CanvasGroup>(), null);
            return screen;
        }

        [UnityTest]
        public IEnumerator Measure()
        {
            new GameObject("Camera", typeof(Camera));
            var stack = new GameObject("Stack").AddComponent<ScreenStack>();
            stack.enabled = false;
            var root = Screen("Root", 40, 40);
            var shop = Screen("Shop", 30, 30);
            stack.Push(root);
            stack.Push(shop);
            for (var at = 0; at < 3; at++)
            {
                stack.Advance(10000f);
                Canvas.ForceUpdateCanvases();
                yield return null;
                yield return null;
                Note($" {at} pop={Cost(() => { stack.Pop(); })}");
                Note($" {at} popUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
                Note($" {at} popSettle(hide)={Cost(() => { stack.Advance(10000f); })}");
                Note($" {at} hideUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
                yield return null;
                Note($" {at} push={Cost(() => { stack.Push(shop); })}");
                Note($" {at} pushUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
                Note($" {at} pushSettle={Cost(() => { stack.Advance(10000f); })}");
                Note($" {at} settleUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
            }
            var raw = new GameObject("Raw", typeof(Canvas)).GetComponent<Canvas>();
            raw.renderMode = RenderMode.ScreenSpaceOverlay;
            for (var at = 0; at < 60; at++) new GameObject("I", typeof(RectTransform), typeof(Image)).transform.SetParent(raw.transform, false);
            Canvas.ForceUpdateCanvases();
            yield return null;
            for (var at = 0; at < 3; at++)
            {
                Note($"raw {at} disable={Cost(() => { raw.enabled = false; })}");
                Note($"raw {at} disableUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
                yield return null;
                Note($"raw {at} enable={Cost(() => { raw.enabled = true; })}");
                Note($"raw {at} enableUpdate={Cost(() => { Canvas.ForceUpdateCanvases(); })}");
                yield return null;
                Note($"raw {at} goToggle={Cost(() => { raw.gameObject.SetActive(false); raw.gameObject.SetActive(true); })}");
                yield return null;
            }
            yield return Toggle("noRaycast", false, false);
            yield return Toggle("raycast", true, false);
            yield return Toggle("nested", true, true);
            Note($"control={Cost(() => { var junk = new int[256]; junk[0] = 1; })}");
            Assert.Fail(report.ToString());
        }

        IEnumerator Toggle(string label, bool raycast, bool nested)
        {
            var root = new GameObject(label, typeof(Canvas), typeof(GraphicRaycaster)).GetComponent<Canvas>();
            root.renderMode = RenderMode.ScreenSpaceOverlay;
            var holder = root.transform;
            Canvas inner = null;
            if (nested)
            {
                inner = new GameObject("Inner", typeof(RectTransform), typeof(Canvas), typeof(GraphicRaycaster)).GetComponent<Canvas>();
                inner.transform.SetParent(root.transform, false);
                holder = inner.transform;
            }
            for (var at = 0; at < 60; at++)
            {
                var image = new GameObject("I", typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                image.transform.SetParent(holder, false);
                image.raycastTarget = raycast;
            }
            Canvas.ForceUpdateCanvases();
            yield return null;
            for (var at = 0; at < 3; at++)
            {
                var off = Cost(() => { root.enabled = false; Canvas.ForceUpdateCanvases(); });
                if (nested) Note($"{label} innerRoot={inner.isRootCanvas} innerEnabled={inner.isActiveAndEnabled} innerRootCanvas={inner.rootCanvas.name}");
                yield return null;
                var on = Cost(() => { root.enabled = true; Canvas.ForceUpdateCanvases(); });
                Note($"{label} {at} off={off} on={on}");
                yield return null;
            }
        }
    }
}
