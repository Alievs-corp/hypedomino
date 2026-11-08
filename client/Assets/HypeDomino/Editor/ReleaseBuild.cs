using System;
using System.IO;
using System.Text.RegularExpressions;
using HypeDomino.Content;
using UnityEditor;
using UnityEditor.AddressableAssets;
using UnityEditor.Build;
using UnityEditor.Build.Reporting;

namespace HypeDomino.Editor
{
    public static class ReleaseBuild
    {
        public const string CatalogVersionArgument = "-catalogVersion";
        public const string AndroidRemoteFolder = "ServerData/Android";
        const string OutputPathArgument = "-customBuildPath";
        const string DefaultAndroidOutput = "Builds/hypedomino.aab";

        static readonly Regex CatalogVersionPattern = new Regex(@"^[A-Za-z0-9][A-Za-z0-9._-]{0,63}\z");

        public static void BuildAndroid() => EditorMethod.Run(() => BuildAndroid(Argument(CatalogVersionArgument), Argument(OutputPathArgument) ?? DefaultAndroidOutput));

        public static void BuildAndroid(string catalogVersion, string output)
        {
            if (!IsCatalogVersion(catalogVersion)) throw new ArgumentException($"{CatalogVersionArgument} must be 1 to 64 letters, digits, '.', '_' or '-' starting with a letter or digit, got '{catalogVersion}'");
            var settings = AddressableAssetSettingsDefaultObject.Settings ?? throw new BuildFailedException("Addressables settings are missing; run ContentSetup.Apply");
            var previousVersion = settings.OverridePlayerVersion;
            var previousAppBundle = EditorUserBuildSettings.buildAppBundle;
            try
            {
                settings.OverridePlayerVersion = catalogVersion;
                EditorUserBuildSettings.buildAppBundle = true;
                if (Directory.Exists(AndroidRemoteFolder)) Directory.Delete(AndroidRemoteFolder, true);
                Directory.CreateDirectory(Path.GetDirectoryName(Path.GetFullPath(output)));
                var report = BuildPipeline.BuildPlayer(new BuildPlayerOptions
                {
                    scenes = EditorBuildSettingsScene.GetActiveSceneList(EditorBuildSettings.scenes),
                    locationPathName = output,
                    target = BuildTarget.Android,
                    options = BuildOptions.None,
                });
                if (report.summary.result != BuildResult.Succeeded) throw new BuildFailedException($"Android release build {report.summary.result}");
                var catalog = Path.Combine(AndroidRemoteFolder, ContentUrls.CatalogFile(catalogVersion));
                if (!File.Exists(catalog)) throw new BuildFailedException($"{catalog} was not built");
            }
            finally
            {
                settings.OverridePlayerVersion = previousVersion;
                EditorUserBuildSettings.buildAppBundle = previousAppBundle;
            }
        }

        public static bool IsCatalogVersion(string version) => version != null && CatalogVersionPattern.IsMatch(version);

        static string Argument(string name)
        {
            var arguments = Environment.GetCommandLineArgs();
            var at = Array.IndexOf(arguments, name);
            return at >= 0 && at + 1 < arguments.Length ? arguments[at + 1] : null;
        }
    }
}
