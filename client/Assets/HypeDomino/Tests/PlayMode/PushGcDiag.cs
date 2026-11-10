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

        static long Bytes() => GC.GetAllocatedBytesForCurrentThread();

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
                long b;
                b = Bytes(); stack.Pop(); Note($" {at} pop={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($" {at} popUpdate={Bytes() - b}");
                b = Bytes(); stack.Advance(10000f); Note($" {at} popSettle(hide)={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($" {at} hideUpdate={Bytes() - b}");
                yield return null;
                b = Bytes(); stack.Push(shop); Note($" {at} push={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($" {at} pushUpdate={Bytes() - b}");
                b = Bytes(); stack.Advance(10000f); Note($" {at} pushSettle={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($" {at} settleUpdate={Bytes() - b}");
            }
            var raw = new GameObject("Raw", typeof(Canvas)).GetComponent<Canvas>();
            raw.renderMode = RenderMode.ScreenSpaceOverlay;
            for (var at = 0; at < 60; at++) new GameObject("I", typeof(RectTransform), typeof(Image)).transform.SetParent(raw.transform, false);
            Canvas.ForceUpdateCanvases();
            yield return null;
            for (var at = 0; at < 3; at++)
            {
                long b;
                b = Bytes(); raw.enabled = false; Note($"raw {at} disable={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($"raw {at} disableUpdate={Bytes() - b}");
                yield return null;
                b = Bytes(); raw.enabled = true; Note($"raw {at} enable={Bytes() - b}");
                b = Bytes(); Canvas.ForceUpdateCanvases(); Note($"raw {at} enableUpdate={Bytes() - b}");
                yield return null;
                b = Bytes(); raw.gameObject.SetActive(false); raw.gameObject.SetActive(true); Note($"raw {at} goToggle={Bytes() - b}");
                yield return null;
            }
            yield return Toggle("noRaycast", false, false);
            yield return Toggle("raycast", true, false);
            yield return Toggle("nested", true, true);
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
                var b = Bytes(); root.enabled = false; Canvas.ForceUpdateCanvases(); var off = Bytes() - b;
                if (nested) Note($"{label} innerRoot={inner.isRootCanvas} innerEnabled={inner.isActiveAndEnabled} innerRootCanvas={inner.rootCanvas.name}");
                yield return null;
                b = Bytes(); root.enabled = true; Canvas.ForceUpdateCanvases(); var on = Bytes() - b;
                Note($"{label} {at} off={off} on={on}");
                yield return null;
            }
        }
    }
}
