using System;
using System.Globalization;
using HypeDomino.Game;
using HypeDomino.UI.Localization;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class CacheSizeTextTests
    {
        [TestCase(0L, "0.0")]
        [TestCase(4_096L, "0.0")]
        [TestCase(49_999L, "0.0")]
        [TestCase(50_000L, "0.1")]
        [TestCase(12_000_000L, "12.0")]
        [TestCase(12_345_678L, "12.3")]
        [TestCase(12_350_000L, "12.4")]
        [TestCase(1_000_000_000L, "1000.0")]
        public void TheSizeIsInMegabytesWithOneDecimal(long bytes, string megabytes)
        {
            Assert.AreEqual(megabytes, CacheSizeText.Megabytes(bytes).ToString(CultureInfo.InvariantCulture));
        }

        [TestCase(GameLanguage.English, 12_400_000L, "Cache: 12.4 MB")]
        [TestCase("az", 12_400_000L, "Keş: 12,4 MB")]
        [TestCase(GameLanguage.English, 1_000_000_000L, "Cache: 1,000.0 MB")]
        [TestCase("az", 1_000_000_000L, "Keş: 1.000,0 MB")]
        public void TheTextUsesTheNumberFormatOfTheLanguage(string language, long bytes, string text)
        {
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(language, Array.Empty<string>), false);

            Assert.AreEqual(text, CacheSizeText.Of(strings, bytes));
        }

        [Test]
        public void TheShippedBundleHasTheSizeKeyInAzerbaijaniAndEnglish()
        {
            var strings = BuiltinStrings.Load();

            Assert.IsNotNull(strings.Value(CacheSizeText.Key, GameLanguage.English));
            Assert.IsNotNull(strings.Value(CacheSizeText.Key, "az"));
        }
    }
}
