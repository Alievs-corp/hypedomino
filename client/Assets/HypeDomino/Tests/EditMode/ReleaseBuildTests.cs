using HypeDomino.Content;
using HypeDomino.Editor;
using NUnit.Framework;
using UnityEditor.AddressableAssets;
using UnityEditor.AddressableAssets.Settings;

namespace HypeDomino.Tests.EditMode
{
    public sealed class ReleaseBuildTests
    {
        [TestCase("2026.11.08-1")]
        [TestCase("v7_hotfix")]
        [TestCase("a")]
        [TestCase("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")]
        public void ACatalogVersionTheServerAcceptsIsAccepted(string version)
        {
            Assert.IsTrue(ReleaseBuild.IsCatalogVersion(version));
        }

        [TestCase(null)]
        [TestCase("")]
        [TestCase(".hidden")]
        [TestCase("-1")]
        [TestCase("a/b")]
        [TestCase("a b")]
        [TestCase("a?b")]
        [TestCase("1\n")]
        [TestCase("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")]
        public void ACatalogVersionTheServerRefusesIsRefused(string version)
        {
            Assert.IsFalse(ReleaseBuild.IsCatalogVersion(version));
        }

        [Test]
        public void TheBuiltCatalogHasTheNameTheClientDownloads()
        {
            Assert.AreEqual("catalog_2026.11.08-1.bin", ContentUrls.CatalogFile("2026.11.08-1"));
            StringAssert.EndsWith("/Android/" + ContentUrls.CatalogFile("2026.11.08-1"), ContentUrls.Catalog("https://content.example.com", "Android", "2026.11.08-1"));
        }

        [Test]
        public void RemoteContentIsBuiltIntoTheFolderTheWorkflowUploads()
        {
            var settings = AddressableAssetSettingsDefaultObject.Settings;
            Assert.AreEqual(AddressableAssetSettings.kRemoteBuildPathValue, settings.profileSettings.GetValueByName(settings.activeProfileId, AddressableAssetSettings.kRemoteBuildPath));
            Assert.AreEqual(ReleaseBuild.AndroidRemoteFolder, AddressableAssetSettings.kRemoteBuildPathValue.Replace("[BuildTarget]", "Android"));
        }
    }
}
