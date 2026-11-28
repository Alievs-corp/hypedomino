using HypeDomino.Game.Analytics;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class ScreenViewIdsTests
    {
        [TestCase("Lobby", "lobby")]
        [TestCase("Settings", "settings")]
        [TestCase("UpdateRequired", "update_required")]
        [TestCase("PrivacyNotice", "privacy_notice")]
        public void TheIdIsTheScreenNameInSnakeCase(string screenName, string id)
        {
            Assert.AreEqual(id, ScreenViewIds.Of(screenName));
        }
    }
}
