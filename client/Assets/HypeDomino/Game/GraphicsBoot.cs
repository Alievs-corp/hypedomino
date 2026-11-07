using System;
using HypeDomino.UI.Graphics;
using UnityEngine;
using UnityEngine.Rendering;
using UnityEngine.Rendering.Universal;

namespace HypeDomino.Game
{
    public static class GraphicsBoot
    {
        const string ChoiceKey = "graphics.choice";
        const int AutoChoice = -1;

        public static GraphicsQuality Quality { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            var device = new DeviceFacts(SystemInfo.systemMemorySize, SystemInfo.graphicsDeviceName, HighestRefreshHz());
            Quality = new GraphicsQuality(DeviceClass.TierOf(device), LoadChoice());
            Quality.Changed += OnChanged;
            var renderScale = ApplyRenderScale();
            Debug.Log($"Graphics tier={Quality.Tier} level={Quality.Level} effects={Quality.Effects} renderScale={renderScale:F2} " +
                      $"memoryMb={device.MemoryMb} gpu=\"{device.GpuName}\" refreshHz={device.MaxRefreshHz} shortSide={NativeShortSide()}");
        }

        static void OnChanged()
        {
            PlayerPrefs.SetInt(ChoiceKey, Quality.Choice.HasValue ? (int)Quality.Choice.Value : AutoChoice);
            PlayerPrefs.Save();
            ApplyRenderScale();
        }

        static GraphicsLevel? LoadChoice()
        {
            var saved = PlayerPrefs.GetInt(ChoiceKey, AutoChoice);
            if (saved < (int)GraphicsLevel.Low || saved > (int)GraphicsLevel.Ultra) return null;
            return (GraphicsLevel)saved;
        }

        static float ApplyRenderScale()
        {
            var renderScale = Quality.RenderScale(NativeShortSide());
            if (Application.isEditor) return renderScale;
            if (GraphicsSettings.currentRenderPipeline is UniversalRenderPipelineAsset pipeline) pipeline.renderScale = renderScale;
            return renderScale;
        }

        static int NativeShortSide() => Math.Min(Display.main.systemWidth, Display.main.systemHeight);

        static int HighestRefreshHz()
        {
            var highest = Screen.currentResolution.refreshRateRatio.value;
            foreach (var resolution in Screen.resolutions)
                highest = Math.Max(highest, resolution.refreshRateRatio.value);
            return (int)Math.Round(highest);
        }
    }
}
