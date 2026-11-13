using HypeDomino.UI.Audio;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class AudioBoot
    {
        const string MusicKey = "audio.musicVolume";
        const string SfxKey = "audio.sfxVolume";
        const string UiKey = "audio.uiVolume";

        public static AudioVolumes Volumes { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            Volumes = new AudioVolumes(
                PlayerPrefs.GetFloat(MusicKey, AudioVolumes.Full),
                PlayerPrefs.GetFloat(SfxKey, AudioVolumes.Full),
                PlayerPrefs.GetFloat(UiKey, AudioVolumes.Full));
            Volumes.Changed += OnChanged;
            Log();
            if (Application.isEditor) return;
            ApplyOutput();
            var driver = new GameObject(nameof(AudioBoot)) { hideFlags = HideFlags.HideInHierarchy };
            Object.DontDestroyOnLoad(driver);
            driver.AddComponent<Driver>();
        }

        static void OnChanged()
        {
            PlayerPrefs.SetFloat(MusicKey, Volumes.Level(AudioGroup.Music));
            PlayerPrefs.SetFloat(SfxKey, Volumes.Level(AudioGroup.Sfx));
            PlayerPrefs.SetFloat(UiKey, Volumes.Level(AudioGroup.Ui));
            PlayerPrefs.Save();
            Log();
            if (!Application.isEditor) ApplyOutput();
        }

        static void ApplyOutput()
        {
            var started = AudioSettings.Mobile.audioOutputStarted;
            if (Volumes.Silent && started) AudioSettings.Mobile.StopAudioOutput();
            if (!Volumes.Silent && !started) AudioSettings.Mobile.StartAudioOutput();
        }

        static void Log() => Debug.Log($"Audio music={Volumes.Level(AudioGroup.Music)} sfx={Volumes.Level(AudioGroup.Sfx)} ui={Volumes.Level(AudioGroup.Ui)} silent={Volumes.Silent}");

        sealed class Driver : MonoBehaviour
        {
            void OnApplicationFocus(bool hasFocus)
            {
                if (hasFocus) ApplyOutput();
            }
        }
    }
}
