using System;
using HypeDomino.Content;
using HypeDomino.Game;
using HypeDomino.UI.Graphics;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class AppBootTests
    {
        const int ChosenRate = 120;
        const int MenuRate = 60;

        sealed class IdleCatalogSource : ICatalogSource
        {
            public string StoredVersion => null;

            public string StoredBaseUrl => null;

            public void LoadStored(Action<bool> done) { }

            public void Download(string url, Action<bool> done) { }

            public void Commit(string version, string baseUrl, Action<bool> done) { }
        }

        static FrameRatePolicy FrameRate() =>
            new FrameRatePolicy(new GraphicsQuality(DeviceTier.High, GraphicsLevel.High, 0, false), new[] { 60, 120 }, ChosenRate, false);

        static CatalogUpdater Catalog() => new CatalogUpdater(new IdleCatalogSource(), "Android");

        [TestCase(AppState.Boot, false)]
        [TestCase(AppState.Login, false)]
        [TestCase(AppState.Lobby, false)]
        [TestCase(AppState.Match, true)]
        [TestCase(AppState.UpdateRequired, false)]
        [TestCase(AppState.Maintenance, false)]
        public void OnlyTheMatchStateCountsAsInMatch(AppState state, bool inMatch)
        {
            var frameRate = FrameRate();
            var catalog = Catalog();
            AppBoot.FollowMatch(state, frameRate, catalog);
            Assert.AreEqual(inMatch, frameRate.InMatch);
            Assert.AreEqual(inMatch, catalog.InMatch);
        }

        [Test]
        public void TheChosenRateAndTheCatalogHoldLastOnlyWhileInTheMatch()
        {
            var frameRate = FrameRate();
            var catalog = Catalog();
            AppBoot.FollowMatch(AppState.Match, frameRate, catalog);
            Assert.AreEqual(ChosenRate, frameRate.Target);
            AppBoot.FollowMatch(AppState.Lobby, frameRate, catalog);
            Assert.AreEqual(MenuRate, frameRate.Target);
            Assert.IsFalse(catalog.InMatch);
        }

        [Test]
        public void TheEditorHasNoCatalogAndStillFollowsTheMatch()
        {
            var frameRate = FrameRate();
            AppBoot.FollowMatch(AppState.Match, frameRate, null);
            Assert.IsTrue(frameRate.InMatch);
        }
    }
}
