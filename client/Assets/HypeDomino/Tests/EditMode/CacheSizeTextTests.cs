using System;
using HypeDomino.Game;
using HypeDomino.UI.Localization;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class CacheSizeTextTests
    {
        [TestCase(GameLanguage.English, 0L, "Cache: < 1 MB")]
        [TestCase("az", 0L, "Keş: < 1 MB")]
        [TestCase(GameLanguage.English, 999_999L, "Cache: < 1 MB")]
        [TestCase("az", 999_999L, "Keş: < 1 MB")]
        [TestCase(GameLanguage.English, 1_000_000L, "Cache: 1.0 MB")]
        [TestCase("az", 1_000_000L, "Keş: 1,0 MB")]
        [TestCase(GameLanguage.English, 12_345_678L, "Cache: 12.3 MB")]
        [TestCase("az", 12_345_678L, "Keş: 12,3 MB")]
        [TestCase(GameLanguage.English, 12_350_000L, "Cache: 12.4 MB")]
        [TestCase("az", 12_350_000L, "Keş: 12,4 MB")]
        [TestCase(GameLanguage.English, 12_400_000L, "Cache: 12.4 MB")]
        [TestCase("az", 12_400_000L, "Keş: 12,4 MB")]
        [TestCase(GameLanguage.English, 999_949_999L, "Cache: 999.9 MB")]
        [TestCase("az", 999_949_999L, "Keş: 999,9 MB")]
        [TestCase(GameLanguage.English, 999_950_000L, "Cache: 1.0 GB")]
        [TestCase("az", 999_950_000L, "Keş: 1,0 GB")]
        [TestCase(GameLanguage.English, 999_999_999L, "Cache: 1.0 GB")]
        [TestCase("az", 999_999_999L, "Keş: 1,0 GB")]
        [TestCase(GameLanguage.English, 1_000_000_000L, "Cache: 1.0 GB")]
        [TestCase("az", 1_000_000_000L, "Keş: 1,0 GB")]
        [TestCase(GameLanguage.English, 1_300_000_000L, "Cache: 1.3 GB")]
        [TestCase("az", 1_300_000_000L, "Keş: 1,3 GB")]
        [TestCase(GameLanguage.English, 1_234_567_890_123L, "Cache: 1,234.6 GB")]
        [TestCase("az", 1_234_567_890_123L, "Keş: 1.234,6 GB")]
        [TestCase(GameLanguage.English, long.MaxValue, "Cache: 9,223,372,036.9 GB")]
        [TestCase("az", long.MaxValue, "Keş: 9.223.372.036,9 GB")]
        public void TheSizeUsesTheUnitAndNumberFormatOfTheLanguage(string language, long bytes, string text)
        {
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(language, Array.Empty<string>), false);

            Assert.AreEqual(text, CacheSizeText.Of(strings, bytes));
        }

        [TestCase(CacheSizeText.BelowOneMegabyteKey)]
        [TestCase(CacheSizeText.MegabytesKey)]
        [TestCase(CacheSizeText.GigabytesKey)]
        public void TheShippedBundleHasEverySizeKeyInAzerbaijaniAndEnglish(string key)
        {
            var strings = BuiltinStrings.Load();

            Assert.IsNotNull(strings.Value(key, GameLanguage.English));
            Assert.IsNotNull(strings.Value(key, "az"));
        }
    }
}
