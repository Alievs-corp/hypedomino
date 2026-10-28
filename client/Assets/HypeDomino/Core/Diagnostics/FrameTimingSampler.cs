using System;
using UnityEngine;

namespace HypeDomino.Core.Diagnostics
{
    public sealed class FrameTimingSampler : MonoBehaviour
    {
        const int WindowFrames = 1200;
        const double MissedVsyncFactor = 1.5;

        readonly FrameTiming[] latestTiming = new FrameTiming[1];
        readonly double[] workTimesMs = new double[WindowFrames];
        int sampledFrames;
        int missedVsyncs;
        double missedVsyncThresholdSeconds;

        void Awake()
        {
            var refreshRate = HighestRefreshRate();
            Application.targetFrameRate = refreshRate;
            missedVsyncThresholdSeconds = MissedVsyncFactor / refreshRate;
        }

        void Update()
        {
            FrameTimingManager.CaptureFrameTimings();
            if (FrameTimingManager.GetLatestTimings(1, latestTiming) == 0) return;

            workTimesMs[sampledFrames++] = Math.Max(latestTiming[0].cpuFrameTime, latestTiming[0].gpuFrameTime);
            if (Time.unscaledDeltaTime > missedVsyncThresholdSeconds) missedVsyncs++;
            if (sampledFrames < WindowFrames) return;

            Report();
            sampledFrames = 0;
            missedVsyncs = 0;
        }

        void Report()
        {
            Array.Sort(workTimesMs);
            var p50 = workTimesMs[WindowFrames / 2];
            var p95 = workTimesMs[WindowFrames * 95 / 100];
            var missedPercent = 100.0 * missedVsyncs / WindowFrames;
            Debug.Log($"FrameTiming target={Application.targetFrameRate} frames={WindowFrames} workP50={p50:F2}ms workP95={p95:F2}ms missedVsync={missedPercent:F2}%");
        }

        static int HighestRefreshRate()
        {
            var highest = Screen.currentResolution.refreshRateRatio.value;
            foreach (var resolution in Screen.resolutions)
                highest = Math.Max(highest, resolution.refreshRateRatio.value);
            return (int)Math.Round(highest);
        }
    }
}
