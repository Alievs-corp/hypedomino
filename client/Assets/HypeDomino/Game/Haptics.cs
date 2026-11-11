using System;
using System.Runtime.InteropServices;
using HypeDomino.Game.Match;
using UnityEngine;
using UnityEngine.Android;

namespace HypeDomino.Game
{
    public static class Haptics
    {
        const string EnabledKey = "haptics.enabled";
        const int AndroidVirtualKey = 1;
        const int AndroidLongPress = 0;
        const int AndroidConfirm = 16;
        const int FirstSdkWithConfirm = 30;

        static bool? enabled;
        static AndroidView androidView;

        public static bool Enabled
        {
            get => enabled ??= PlayerPrefs.GetInt(EnabledKey, 1) == 1;
            set
            {
                if (Enabled == value) return;
                enabled = value;
                PlayerPrefs.SetInt(EnabledKey, value ? 1 : 0);
                PlayerPrefs.Save();
            }
        }

        public static void Play(HapticCue cue)
        {
            if (cue == HapticCue.None || !Enabled || Application.isEditor) return;
            if (Application.platform == RuntimePlatform.Android) PlayAndroid(cue);
            if (Application.platform == RuntimePlatform.IPhonePlayer) PlayIos(cue);
        }

        static void PlayAndroid(HapticCue cue)
        {
            androidView ??= new AndroidView();
            androidView.Perform(cue switch
            {
                HapticCue.TilePlace => AndroidVirtualKey,
                HapticCue.Score => androidView.Sdk >= FirstSdkWithConfirm ? AndroidConfirm : AndroidVirtualKey,
                _ => AndroidLongPress,
            });
        }

#if UNITY_IOS && !UNITY_EDITOR
        [DllImport("__Internal")]
        static extern void HypeDominoPlayHaptic(int cue);

        static void PlayIos(HapticCue cue) => HypeDominoPlayHaptic((int)cue);
#else
        static void PlayIos(HapticCue cue)
        {
        }
#endif

        sealed class AndroidView
        {
            readonly IntPtr view;
            readonly IntPtr performHapticFeedback;
            readonly jvalue[] constant = new jvalue[1];

            public readonly int Sdk;

            public AndroidView()
            {
                using var version = new AndroidJavaClass("android.os.Build$VERSION");
                Sdk = version.GetStatic<int>("SDK_INT");
                using var window = AndroidApplication.currentActivity.Call<AndroidJavaObject>("getWindow");
                using var decorView = window.Call<AndroidJavaObject>("getDecorView");
                view = AndroidJNI.NewGlobalRef(decorView.GetRawObject());
                var viewClass = AndroidJNI.GetObjectClass(view);
                performHapticFeedback = AndroidJNI.GetMethodID(viewClass, "performHapticFeedback", "(I)Z");
                AndroidJNI.DeleteLocalRef(viewClass);
            }

            public void Perform(int feedbackConstant)
            {
                constant[0].i = feedbackConstant;
                AndroidJNI.CallBooleanMethod(view, performHapticFeedback, constant);
            }
        }
    }
}
