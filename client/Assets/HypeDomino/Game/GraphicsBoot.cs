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
        const string ProbeDropsKey = "graphics.probeDrops";
        const string ProbeSettledKey = "graphics.probeSettled";
        const string ProbeVersionKey = "graphics.probeVersion";
        const int AutoChoice = -1;

        public static GraphicsQuality Quality { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            var device = new DeviceFacts(SystemInfo.systemMemorySize, SystemInfo.graphicsDeviceName, HighestRefreshHz());
            var sameVersion = PlayerPrefs.GetString(ProbeVersionKey) == Application.version;
            Quality = new GraphicsQuality(
                DeviceClass.TierOf(device),
                LoadChoice(),
                sameVersion ? PlayerPrefs.GetInt(ProbeDropsKey, 0) : 0,
                sameVersion && PlayerPrefs.GetInt(ProbeSettledKey, 0) == 1);
            Quality.Changed += OnChanged;
            var renderScale = ApplyRenderScale();
            Debug.Log($"Graphics tier={Quality.Tier} level={Quality.Level} effects={Quality.Effects} renderScale={renderScale:F2} probeDrops={Quality.ProbeDrops} " +
                      $"memoryMb={device.MemoryMb} gpu=\"{device.GpuName}\" refreshHz={device.MaxRefreshHz} shortSide={NativeShortSide()}");
        }

        public static void SettleProbe(double workTimeP95Ms)
        {
            var gateMs = Quality.WorkTimeGateMs;
            var result = Quality.SettleProbe(workTimeP95Ms);
            SaveState();
            Debug.Log($"GraphicsProbe workP95={workTimeP95Ms:F2}ms gate={gateMs:F2}ms result={result} level={Quality.Level} probeDrops={Quality.ProbeDrops}");
        }

        static void OnChanged()
        {
            SaveState();
            ApplyRenderScale();
        }

        static void SaveState()
        {
            PlayerPrefs.SetInt(ChoiceKey, Quality.Choice.HasValue ? (int)Quality.Choice.Value : AutoChoice);
            PlayerPrefs.SetInt(ProbeDropsKey, Quality.ProbeDrops);
            PlayerPrefs.SetInt(ProbeSettledKey, Quality.ProbeSettled ? 1 : 0);
            PlayerPrefs.SetString(ProbeVersionKey, Application.version);
            PlayerPrefs.Save();
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
