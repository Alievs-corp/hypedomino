using System;
using System.IO;
using HypeDomino.Bench;
using UnityEditor;
using UnityEditor.Build;
using UnityEditor.Build.Reporting;
using UnityEditor.SceneManagement;
using UnityEngine;

namespace HypeDomino.Editor
{
    public static class AnimationBenchBuild
    {
        const string ArtFolder = "Assets/HypeDomino/Bench/Generated";
        const string ScenePath = ArtFolder + "/AnimationBench.unity";
        const string DefaultOutputPath = "Builds/animation-bench.apk";
        const string OutputPathArgument = "-customBuildPath";
        const string BenchIdentifier = "com.hypedomino.animationbench";
        const int AvatarCount = 10;
        const int AtlasSize = 768;
        const int CellSize = 192;
        const int FramesPerAvatar = 16;

        public static void BuildAndroid()
        {
            Directory.CreateDirectory(ArtFolder);
            var output = OutputPath();
            Directory.CreateDirectory(Path.GetDirectoryName(output));
            var atlases = new Texture2D[AvatarCount];
            for (var avatar = 0; avatar < AvatarCount; avatar++) atlases[avatar] = ImportedAtlas(avatar);

            var scene = EditorSceneManager.NewScene(NewSceneSetup.EmptyScene, NewSceneMode.Single);
            new GameObject("Bench").AddComponent<AnimationBench>().Configure(atlases, CellSize, FramesPerAvatar);
            EditorSceneManager.SaveScene(scene, ScenePath);

            PlayerSettings.SetApplicationIdentifier(NamedBuildTarget.Android, BenchIdentifier);
            PlayerSettings.productName = "Animation Bench";
            EditorUserBuildSettings.buildAppBundle = false;
            var report = BuildPipeline.BuildPlayer(new BuildPlayerOptions
            {
                scenes = new[] { ScenePath },
                locationPathName = output,
                target = BuildTarget.Android,
                options = BuildOptions.Development,
            });
            Debug.Log($"AnimationBenchBuild result={report.summary.result} bytes={report.summary.totalSize}");
            if (Application.isBatchMode) EditorApplication.Exit(report.summary.result == BuildResult.Succeeded ? 0 : 1);
        }

        static string OutputPath()
        {
            var arguments = Environment.GetCommandLineArgs();
            var at = Array.IndexOf(arguments, OutputPathArgument);
            return at >= 0 && at + 1 < arguments.Length ? arguments[at + 1] : DefaultOutputPath;
        }

        static Texture2D ImportedAtlas(int avatar)
        {
            var path = $"{ArtFolder}/avatar_{avatar}.png";
            File.WriteAllBytes(path, Atlas(avatar).EncodeToPNG());
            AssetDatabase.ImportAsset(path);
            var importer = (TextureImporter)AssetImporter.GetAtPath(path);
            importer.textureType = TextureImporterType.Sprite;
            importer.spriteImportMode = SpriteImportMode.Single;
            importer.npotScale = TextureImporterNPOTScale.None;
            importer.mipmapEnabled = false;
            importer.alphaIsTransparency = true;
            importer.maxTextureSize = AtlasSize;
            var settings = importer.GetPlatformTextureSettings("Android");
            settings.overridden = true;
            settings.maxTextureSize = AtlasSize;
            settings.format = TextureImporterFormat.ASTC_6x6;
            importer.SetPlatformTextureSettings(settings);
            importer.SaveAndReimport();
            return AssetDatabase.LoadAssetAtPath<Texture2D>(path);
        }

        static Texture2D Atlas(int avatar)
        {
            var atlas = new Texture2D(AtlasSize, AtlasSize, TextureFormat.RGBA32, false);
            var pixels = new Color32[AtlasSize * AtlasSize];
            var columns = AtlasSize / CellSize;
            for (var frame = 0; frame < FramesPerAvatar; frame++)
                DrawFrame(pixels, frame % columns * CellSize, frame / columns * CellSize, avatar, frame);
            atlas.SetPixels32(pixels);
            atlas.Apply();
            return atlas;
        }

        static void DrawFrame(Color32[] pixels, int left, int bottom, int avatar, int frame)
        {
            var phase = frame * Mathf.PI * 2f / FramesPerAvatar;
            var center = CellSize / 2f + Mathf.Sin(phase) * 6f;
            var body = Color.HSVToRGB(avatar / (float)AvatarCount, 0.75f, 0.95f);
            for (var y = 0; y < CellSize; y++)
            for (var x = 0; x < CellSize; x++)
            {
                var distance = Mathf.Sqrt((x - CellSize / 2f) * (x - CellSize / 2f) + (y - center) * (y - center));
                pixels[(bottom + y) * AtlasSize + left + x] = Pixel(distance, x, y, body, phase);
            }
        }

        static Color32 Pixel(float distance, int x, int y, Color body, float phase)
        {
            const float Radius = 84f;
            const float Outline = 6f;
            if (distance > Radius) return new Color32(0, 0, 0, 0);
            if (distance > Radius - Outline) return new Color32(16, 12, 32, 255);
            var shade = 0.7f + 0.3f * Mathf.Sin(x * 0.11f + phase) * Mathf.Cos(y * 0.13f - phase);
            return new Color(body.r * shade, body.g * shade, body.b * shade, 1f);
        }
    }
}
