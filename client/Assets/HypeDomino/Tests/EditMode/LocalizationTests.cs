using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using HypeDomino.UI.Localization;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace HypeDomino.Tests.EditMode
{
    public sealed class LocalizationTests
    {
        const string Builtin = @"{""entries"":[
{""key"":""k"",""en"":""E"",""tr"":""T""},
{""key"":""coins"",""en"":""{count, plural, one {# coin} other {# coins}}"",""az"":""{count, plural, one {# sikkə} other {# sikkə}}""},
{""key"":""broken"",""en"":""fine"",""az"":""{count""},
{""key"":""common.text.unavailable"",""en"":""Unavailable""}]}";

        static readonly string[] NoLocales = { };

        [TestCase(null, new[] { "de-DE", "az-Latn-AZ", "en-US" }, "az")]
        [TestCase(null, new[] { "tr_TR" }, "tr")]
        [TestCase(null, new[] { "RU" }, "ru")]
        [TestCase(null, new[] { "en-GB", "ru-RU" }, "en")]
        [TestCase(null, new[] { "fr-FR", "de" }, "en")]
        [TestCase("ru", new[] { "az-AZ" }, "ru")]
        [TestCase("de", new[] { "tr-TR" }, "tr")]
        public void TheLanguageIsTheSavedOneElseTheFirstMappedDeviceLocale(string saved, string[] locales, string expected) =>
            Assert.AreEqual(expected, GameLanguage.Pick(saved, locales));

        [Test]
        public void WithoutLocalesTheLanguageIsEnglish() => Assert.AreEqual("en", GameLanguage.Pick(null, NoLocales));

        [TestCase("az-AZ,en-US", new[] { "az-AZ", "en-US" })]
        [TestCase("ru-RU", new[] { "ru-RU" })]
        [TestCase("tr-TR,,en", new[] { "tr-TR", "en" })]
        [TestCase("", new string[0])]
        [TestCase(null, new string[0])]
        public void DeviceLocaleListsSplitAtCommas(string joined, string[] locales) =>
            CollectionAssert.AreEqual(locales, GameLanguage.SplitLocales(joined));

        [Test]
        public void ASavedLanguageIsKeptWithoutReadingTheDevice()
        {
            var reads = 0;
            var choice = LanguageChoice.Load("ru", () =>
            {
                reads++;
                return new[] { "az-AZ" };
            });
            Assert.AreEqual("ru", choice.Language);
            Assert.AreEqual(0, reads);
        }

        [TestCase(null)]
        [TestCase("")]
        [TestCase("de")]
        public void WithoutASavedLanguageTheDeviceLocalesPick(string saved) =>
            Assert.AreEqual("tr", LanguageChoice.Load(saved, () => new[] { "fr-FR", "tr-TR", "en-US" }).Language);

        [Test]
        public void ChoosingALanguageRaisesChangedOnce()
        {
            var choice = LanguageChoice.Load("en", () => NoLocales);
            var changes = 0;
            choice.Changed += () => changes++;
            Assert.IsTrue(choice.Choose("az"));
            Assert.IsTrue(choice.Choose("az"));
            Assert.AreEqual("az", choice.Language);
            Assert.AreEqual(1, changes);
        }

        [TestCase("de")]
        [TestCase("AZ")]
        [TestCase(null)]
        public void ALanguageOutsideTheLaunchSetIsRefused(string language)
        {
            var choice = LanguageChoice.Load("en", () => NoLocales);
            var changes = 0;
            choice.Changed += () => changes++;
            Assert.IsFalse(choice.Choose(language));
            Assert.AreEqual("en", choice.Language);
            Assert.AreEqual(0, changes);
        }

        [TestCase("ru", 1, "one")]
        [TestCase("ru", 21, "one")]
        [TestCase("ru", 11, "many")]
        [TestCase("ru", 2, "few")]
        [TestCase("ru", 24, "few")]
        [TestCase("ru", 12, "many")]
        [TestCase("ru", 14, "many")]
        [TestCase("ru", 0, "many")]
        [TestCase("ru", 25, "many")]
        [TestCase("ru", 111, "many")]
        [TestCase("ru", -1, "one")]
        [TestCase("en", 1, "one")]
        [TestCase("en", 0, "other")]
        [TestCase("az", 1, "one")]
        [TestCase("tr", 5, "other")]
        public void PluralCategoriesFollowTheLanguage(string language, long number, string category) =>
            Assert.AreEqual(category, PluralRules.Category(language, number));

        [TestCase(1234567L, "en", "1,234,567")]
        [TestCase(1234567L, "az", "1.234.567")]
        [TestCase(1234567L, "tr", "1.234.567")]
        [TestCase(1234L, "ru", "1\u00A0234")]
        [TestCase(-1234L, "en", "-1,234")]
        [TestCase(999L, "en", "999")]
        public void NumbersAreGroupedForTheLanguage(long number, string language, string text) =>
            Assert.AreEqual(text, MessageFormat.Number(number, language));

        [Test]
        public void PlaceholdersPluralsAndSelectsFormat()
        {
            var args = new Dictionary<string, object> { ["name"] = "Aysel", ["count"] = 1234L, ["kind"] = "gift" };
            Assert.AreEqual("Aysel: 1\u00A0234 монеты", Format("{name}: {count, plural, one {# монета} few {# монеты} many {# монет} other {# монеты}}", "ru", args));
            Assert.AreEqual("none", Format("{count, plural, =1234 {none} one {# one} other {# other}}", "en", args));
            Assert.AreEqual("a gift for Aysel", Format("{kind, select, gift {a gift for {name}} other {other}}", "en", args));
            Assert.AreEqual("other", Format("{kind, select, chest {chest} other {other}}", "en", args));
            Assert.AreEqual("#1", Format("#{count, plural, other {{kind, select, other {1}}}}", "en", args));
        }

        [TestCase("{name")]
        [TestCase("{name}}")]
        [TestCase("{count, plural, one {#}")]
        [TestCase("{count, choice, one {#} other {#}}")]
        [TestCase("{count, plural, one {#}}")]
        [TestCase("{missing}")]
        [TestCase("{name, plural, other {#}}")]
        public void ABrokenMessageOrAMissingArgumentDoesNotFormat(string pattern) =>
            Assert.IsFalse(MessageFormat.TryFormat(pattern, "en", "en", new Dictionary<string, object> { ["name"] = "A", ["count"] = 2L }, out _));

        [Test]
        public void TheBundleHashIsSha256Hex() =>
            Assert.AreEqual("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", StringBundle.Sha256Hex(Encoding.ASCII.GetBytes("abc")));

        [Test]
        public void APublishedBundleIsReadOnlyWhenItsHashMatches()
        {
            var content = Encoding.UTF8.GetBytes(@"{""language"":""ru"",""strings"":{""a"":""\u003cb\u003e Ё"",""b"":""x\ny\""""}}");
            var hash = StringBundle.Sha256Hex(content);
            Assert.IsFalse(StringBundle.TryRead(content, 3, StringBundle.Sha256Hex(Encoding.UTF8.GetBytes("other")), out _));
            Assert.IsTrue(StringBundle.TryRead(content, 3, hash, out var bundle));
            Assert.AreEqual("ru", bundle.Language);
            Assert.AreEqual(3, bundle.Version);
            Assert.IsTrue(bundle.TryGet("a", out var a));
            Assert.AreEqual("<b> Ё", a);
            Assert.IsTrue(bundle.TryGet("b", out var b));
            Assert.AreEqual("x\ny\"", b);
        }

        [TestCase(@"{""language"":""ru""}")]
        [TestCase(@"{""language"":""ru"",""strings"":{""a"":1}}")]
        [TestCase(@"{""language"":""ru"",""strings"":{},""extra"":""x""}")]
        [TestCase(@"{""language"":""ru"",""strings"":{""a"":""\q""}}")]
        [TestCase(@"{""language"":""ru"",""strings"":{}} x")]
        public void AMalformedBundleIsRefused(string json)
        {
            var content = Encoding.UTF8.GetBytes(json);
            Assert.IsFalse(StringBundle.TryRead(content, 1, StringBundle.Sha256Hex(content), out _));
        }

        [Test]
        public void ALookupFollowsTheDefaultChain()
        {
            var choice = Choice("az");
            var strings = Strings(choice);
            Assert.AreEqual("T", strings.Text("k"));
            choice.Choose("ru");
            Assert.AreEqual("E", strings.Text("k"));
        }

        [Test]
        public void AChosenLanguageAppliesAtOnceAndAsksLiveScreensToRenderAgain()
        {
            var choice = Choice("az");
            var strings = Strings(choice);
            var renders = 0;
            strings.Changed += () => renders++;
            Assert.AreEqual("az", strings.Language);
            Assert.IsTrue(choice.Choose("ru"));
            Assert.AreEqual("ru", strings.Language);
            Assert.AreEqual("E", strings.Text("k"));
            Assert.AreEqual(1, renders);
            Assert.IsTrue(choice.Choose("ru"));
            Assert.IsFalse(choice.Choose("de"));
            Assert.AreEqual(1, renders);
        }

        [Test]
        public void TheServerChainReplacesTheDefaultOneWhenItEndsInEnglish()
        {
            var strings = Strings("az");
            Assert.IsFalse(strings.SetChain("az", new[] { "az", "tr" }));
            Assert.IsFalse(strings.SetChain("az", new[] { "tr", "en" }));
            Assert.IsTrue(strings.SetChain("az", new[] { "az", "en" }));
            Assert.AreEqual("E", strings.Text("k"));
        }

        [Test]
        public void PluralsUseTheValueLanguageAndNumbersTheSelectedOne()
        {
            var choice = Choice("az");
            var strings = Strings(choice);
            var args = new Dictionary<string, object> { ["count"] = 1500L };
            Assert.AreEqual("1.500 sikkə", strings.Text("coins", args));
            choice.Choose("ru");
            Assert.AreEqual("1\u00A0500 coins", strings.Text("coins", args));
        }

        [Test]
        public void AValueThatDoesNotParseFallsToTheNextLanguage()
        {
            var strings = Strings("az");
            Assert.AreEqual("fine", strings.Text("broken"));
        }

        [Test]
        public void ADownloadedBundleIsUsedFromTheNextScreenOutsideAMatch()
        {
            var strings = Strings("ru");
            var renders = 0;
            strings.Changed += () => renders++;
            var content = Encoding.UTF8.GetBytes(@"{""language"":""ru"",""strings"":{""k"":""К""}}");
            var hash = StringBundle.Sha256Hex(content);
            Assert.IsFalse(strings.Holds("ru", hash));
            Assert.IsFalse(strings.Stage("tr", 2, hash, content));
            Assert.IsFalse(strings.Stage("ru", 2, "0", content));
            Assert.IsTrue(strings.Stage("ru", 2, hash, content));
            Assert.IsTrue(strings.Holds("ru", hash));
            Assert.AreEqual("E", strings.Text("k"));
            Assert.IsFalse(strings.ApplyStaged(true));
            Assert.AreEqual("E", strings.Text("k"));
            Assert.AreEqual(0, renders);
            Assert.IsTrue(strings.ApplyStaged(false));
            Assert.AreEqual("К", strings.Text("k"));
            Assert.AreEqual(1, renders);
            Assert.IsTrue(strings.Holds("ru", hash));
            Assert.IsFalse(strings.ApplyStaged(false));
            Assert.AreEqual(1, renders);
        }

        [Test]
        public void AMissingKeyIsLoggedOnceAndNeverShownRawInARelease()
        {
            var release = new GameStrings(BuiltinStrings.Parse(Builtin), Choice("az"), false);
            LogAssert.Expect(LogType.Warning, new Regex("^Strings missing key=nope language=az version=0$"));
            Assert.AreEqual("Unavailable", release.Text("nope"));
            Assert.AreEqual("Unavailable", release.Text("nope"));
            LogAssert.NoUnexpectedReceived();
            var debug = new GameStrings(BuiltinStrings.Parse(Builtin), Choice("az"), true);
            LogAssert.Expect(LogType.Warning, new Regex("key=nope"));
            Assert.AreEqual("[!nope]", debug.Text("nope"));
            var bare = new GameStrings(BuiltinStrings.Parse(@"{""entries"":[]}"), Choice("en"), false);
            LogAssert.Expect(LogType.Warning, new Regex("key=nope"));
            Assert.AreEqual(string.Empty, bare.Text("nope"));
        }

        static LanguageChoice Choice(string language) => LanguageChoice.Load(language, () => NoLocales);

        static GameStrings Strings(string language) => Strings(Choice(language));

        static GameStrings Strings(LanguageChoice choice) => new(BuiltinStrings.Parse(Builtin), choice, false);

        static string Format(string pattern, string language, IReadOnlyDictionary<string, object> args)
        {
            Assert.IsTrue(MessageFormat.TryFormat(pattern, language, language, args, out var text), pattern);
            return text;
        }
    }
}
