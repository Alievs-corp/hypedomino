using System.Collections.Generic;
using System.IO;
using HypeDomino.Core.DeviceRisk;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class DeviceRiskProbe
    {
        public static void Collect(List<string> signals)
        {
            if (Application.platform == RuntimePlatform.Android) DeviceRiskSignals.CollectAndroid(ReadAndroidFacts(), PathExists, signals);
            else if (Application.platform == RuntimePlatform.IPhonePlayer) DeviceRiskSignals.CollectIos(PathExists, signals);
        }

        static bool PathExists(string path) => File.Exists(path) || Directory.Exists(path);

        static DeviceFacts ReadAndroidFacts()
        {
            using (var build = new AndroidJavaClass("android.os.Build"))
            {
                return new DeviceFacts(
                    build.GetStatic<string>("FINGERPRINT"),
                    build.GetStatic<string>("MODEL"),
                    build.GetStatic<string>("MANUFACTURER"),
                    build.GetStatic<string>("HARDWARE"),
                    build.GetStatic<string>("PRODUCT"),
                    build.GetStatic<string>("TAGS"));
            }
        }
    }
}
