using HypeDomino.UI.Localization;
using HypeDomino.UI.Profile;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class PlayerNamesTests
    {
        static readonly string[] NoLocales = { };

        [TestCase("az", "Oyunçu")]
        [TestCase("en", "Player")]
        public void ThePlaceholderIsTheShippedTextOfTheChosenLanguage(string language, string expected)
        {
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(language, () => NoLocales), false);

            Assert.AreEqual(expected, PlayerNames.Placeholder(strings));
        }

        [Test]
        public void ThePlaceholderFollowsALanguageChange()
        {
            var choice = LanguageChoice.Load("az", () => NoLocales);
            var strings = new GameStrings(BuiltinStrings.Load(), choice, false);

            choice.Choose(GameLanguage.English);

            Assert.AreEqual("Player", PlayerNames.Placeholder(strings));
        }

        [TestCase(null)]
        [TestCase("")]
        public void AnEmptyNicknameShowsThePlaceholder(string nickname) =>
            Assert.AreSame("Player", PlayerNames.Shown(nickname, "Player"));

        [Test]
        public void ANicknameShowsItself()
        {
            const string nickname = "Ayla";

            Assert.AreSame(nickname, PlayerNames.Shown(nickname, "Player"));
        }
    }
}
