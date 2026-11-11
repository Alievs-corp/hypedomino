using System;
using System.Collections.Generic;

namespace HypeDomino.UI.Localization
{
    public static class GameLanguage
    {
        public const string English = "en";

        static readonly string[] Launch = { "az", English, "tr", "ru" };
        static readonly char[] SubtagSeparators = { '-', '_' };
        static readonly string[] EnglishChain = { English };

        static readonly Dictionary<string, string[]> DefaultChains = new()
        {
            ["az"] = new[] { "az", "tr", English },
            ["tr"] = new[] { "tr", English },
            ["ru"] = new[] { "ru", English },
            [English] = EnglishChain,
        };

        public static bool IsLaunch(string language) => Array.IndexOf(Launch, language) >= 0;

        public static string Pick(string saved, IReadOnlyList<string> deviceLocales)
        {
            if (IsLaunch(saved)) return saved;
            for (var i = 0; i < deviceLocales.Count; i++)
            {
                var language = LanguageSubtag(deviceLocales[i]);
                if (IsLaunch(language)) return language;
            }
            return English;
        }

        public static string[] DefaultChain(string language) =>
            DefaultChains.TryGetValue(language ?? string.Empty, out var chain) ? chain : EnglishChain;

        static string LanguageSubtag(string locale)
        {
            if (string.IsNullOrEmpty(locale)) return null;
            var end = locale.IndexOfAny(SubtagSeparators);
            return (end < 0 ? locale : locale.Substring(0, end)).ToLowerInvariant();
        }
    }
}
