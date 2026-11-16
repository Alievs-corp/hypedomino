using System;
using System.IO;
using System.Linq;
using HypeDomino.UI.Audio;
using UnityEditor;
using UnityEngine;

namespace HypeDomino.Editor
{
    public static class SoundContent
    {
        public const string Root = "Assets/HypeDomino/UI/Audio";
        public const string SfxFolder = Root + "/Clips/Sfx";
        public const string MusicFolder = Root + "/Clips/Music";
        public const string ResourcesFolder = Root + "/Resources";
        const string NormalizeProperty = "normalize";
        public const string BankPath = ResourcesFolder + "/" + SoundBank.ResourceName + ".asset";

        [MenuItem("HypeDomino/Import Sounds")]
        public static void ImportAll()
        {
            foreach (var path in SfxPaths()) ImportClip(path, false);
            foreach (var path in MusicPaths()) ImportClip(path, true);
            SaveBank();
            AssetDatabase.SaveAssets();
            Debug.Log("HypeDomino sounds imported");
        }

        public static void ImportAllAndExit() => EditorMethod.Run(ImportAll);

        public static string[] SfxPaths() => WavPaths(SfxFolder);

        public static string[] MusicPaths() => WavPaths(MusicFolder);

        public static string TilePlacePath(int variant) => $"{SfxFolder}/tile_place_{variant + 1:00}.wav";

        public static string SoundPath(Sound sound) => $"{SfxFolder}/{SoundName(sound)}.wav";

        public static string TrackPath(Music music) => $"{MusicFolder}/{TrackName(music)}.wav";

        public static string SoundName(Sound sound) => sound switch
        {
            Sound.TileDraw => "tile_draw",
            Sound.Pass => "pass",
            Sound.TurnWarning => "turn_warning",
            Sound.Score => "score",
            Sound.HandEnd => "hand_end",
            Sound.MatchWin => "match_win",
            Sound.MatchLose => "match_lose",
            Sound.UiClick => "ui_click",
            _ => throw new ArgumentOutOfRangeException(nameof(sound), sound, "tile_place has one file per variant"),
        };

        public static string TrackName(Music music) => music switch
        {
            Music.Lobby => "lobby",
            Music.Match => "match",
            _ => throw new ArgumentOutOfRangeException(nameof(music), music, "no track"),
        };

        static string[] WavPaths(string folder) =>
            Directory.GetFiles(folder, "*.wav").Select(path => path.Replace('\\', '/')).OrderBy(path => path, StringComparer.Ordinal).ToArray();

        static void ImportClip(string path, bool music)
        {
            var importer = (AudioImporter)AssetImporter.GetAtPath(path);
            importer.forceToMono = !music;
            importer.loadInBackground = false;
            var settings = importer.defaultSampleSettings;
            settings.loadType = music ? AudioClipLoadType.Streaming : AudioClipLoadType.DecompressOnLoad;
            settings.compressionFormat = music ? AudioCompressionFormat.Vorbis : AudioCompressionFormat.ADPCM;
            settings.preloadAudioData = !music;
            settings.sampleRateSetting = AudioSampleRateSetting.PreserveSampleRate;
            importer.defaultSampleSettings = settings;
            KeepLevels(importer);
            importer.SaveAndReimport();
        }

        public static bool Normalizes(AudioImporter importer) => new SerializedObject(importer).FindProperty(NormalizeProperty).boolValue;

        static void KeepLevels(AudioImporter importer)
        {
            var serialized = new SerializedObject(importer);
            serialized.FindProperty(NormalizeProperty).boolValue = false;
            serialized.ApplyModifiedPropertiesWithoutUndo();
        }

        static void SaveBank()
        {
            var tilePlace = new AudioClip[SoundBank.TilePlaceVariants];
            for (var i = 0; i < tilePlace.Length; i++) tilePlace[i] = Load(TilePlacePath(i));
            var sounds = new AudioClip[SoundBank.SoundCount];
            for (var i = 0; i < sounds.Length; i++)
                if ((Sound)i != Sound.TilePlace) sounds[i] = Load(SoundPath((Sound)i));
            var tracks = new AudioClip[SoundBank.MusicCount];
            for (var i = 0; i < tracks.Length; i++)
                if ((Music)i != Music.None) tracks[i] = Load(TrackPath((Music)i));
            var bank = AssetDatabase.LoadAssetAtPath<SoundBank>(BankPath);
            if (bank == null)
            {
                if (!AssetDatabase.IsValidFolder(ResourcesFolder)) AssetDatabase.CreateFolder(Root, "Resources");
                bank = ScriptableObject.CreateInstance<SoundBank>();
                bank.Set(tilePlace, sounds, tracks);
                AssetDatabase.CreateAsset(bank, BankPath);
                return;
            }
            bank.Set(tilePlace, sounds, tracks);
            EditorUtility.SetDirty(bank);
        }

        static AudioClip Load(string path) =>
            AssetDatabase.LoadAssetAtPath<AudioClip>(path) ?? throw new FileNotFoundException("sound missing", path);
    }
}
