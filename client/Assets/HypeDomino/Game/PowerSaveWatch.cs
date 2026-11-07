using UnityEngine;
using UnityEngine.Android;

namespace HypeDomino.Game
{
    public sealed class PowerSaveWatch : MonoBehaviour
    {
        AndroidJavaObject powerManager;

        public static void StartWatching()
        {
            if (Application.platform != RuntimePlatform.Android) return;
            var watch = new GameObject(nameof(PowerSaveWatch)).AddComponent<PowerSaveWatch>();
            DontDestroyOnLoad(watch.gameObject);
        }

        void Awake()
        {
            powerManager = AndroidApplication.currentContext.Call<AndroidJavaObject>("getSystemService", "power");
            ReadSystemSaver();
        }

        void OnApplicationFocus(bool hasFocus)
        {
            if (hasFocus) ReadSystemSaver();
        }

        void OnDestroy() => powerManager?.Dispose();

        void ReadSystemSaver() => GraphicsBoot.FrameRate?.SetSystemSaver(powerManager.Call<bool>("isPowerSaveMode"));
    }
}
