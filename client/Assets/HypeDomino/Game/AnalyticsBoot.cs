using System;
using HypeDomino.Core.Analytics;
using HypeDomino.Game.Analytics;
using HypeDomino.Net.Api;
using UnityEngine;
using Object = UnityEngine.Object;

namespace HypeDomino.Game
{
    public static class AnalyticsBoot
    {
        static AnalyticsFlushTimer timer;

        public static AnalyticsQueue Queue { get; private set; }

        public static void Start(ApiClient api)
        {
            var settings = AnalyticsSettings.Draft;
            Queue = new AnalyticsQueue(new ApiAnalyticsSender(api), DeviceTimeMs, settings.BatchSize, settings.QueueLimit);
            timer = new AnalyticsFlushTimer(Queue, settings.FlushSeconds);
            var driver = new GameObject(nameof(AnalyticsBoot)) { hideFlags = HideFlags.HideInHierarchy };
            Object.DontDestroyOnLoad(driver);
            driver.AddComponent<Driver>();
        }

        static long DeviceTimeMs() => DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();

        sealed class Driver : MonoBehaviour
        {
            void Update() => timer.Tick(Time.realtimeSinceStartupAsDouble);

            void OnApplicationPause(bool pause)
            {
                if (pause) timer.AppPaused();
            }
        }
    }
}
