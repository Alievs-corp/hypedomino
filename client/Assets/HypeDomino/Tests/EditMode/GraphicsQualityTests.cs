using HypeDomino.UI.Graphics;
using HypeDomino.UI.Motion;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class GraphicsQualityTests
    {
        [TestCase(3650, "Mali-G52 MC2", 60, DeviceTier.Low)]
        [TestCase(1850, "Mali-G57 MC2", 60, DeviceTier.Low)]
        [TestCase(2800, "Apple A12 GPU", 60, DeviceTier.Low)]
        [TestCase(5600, "Adreno (TM) 506", 60, DeviceTier.Low)]
        [TestCase(5600, "Mali-G72 MP3", 60, DeviceTier.Low)]
        [TestCase(5600, "PowerVR Rogue GE8320", 60, DeviceTier.Low)]
        [TestCase(5500, "Mali-G68 MC4", 120, DeviceTier.Mid)]
        [TestCase(3700, "Apple A14 GPU", 60, DeviceTier.Mid)]
        [TestCase(7400, "Adreno (TM) 710", 120, DeviceTier.Mid)]
        [TestCase(5600, "Adreno (TM) 610", 60, DeviceTier.Mid)]
        [TestCase(5600, "Adreno (TM) 750", 120, DeviceTier.Mid)]
        [TestCase(7400, "Adreno (TM) 750", 60, DeviceTier.Mid)]
        [TestCase(7400, "", 120, DeviceTier.Mid)]
        [TestCase(7400, "Adreno (TM) 750", 120, DeviceTier.High)]
        [TestCase(7400, "Samsung Xclipse 940", 120, DeviceTier.High)]
        [TestCase(7400, "Mali-G710 MC10", 120, DeviceTier.High)]
        [TestCase(7400, "Immortalis-G720 MC12", 120, DeviceTier.High)]
        [TestCase(7500, "Apple A17 Pro GPU", 120, DeviceTier.High)]
        public void TheClassRulePutsEachDeviceInItsTier(int memoryMb, string gpuName, int maxRefreshHz, DeviceTier tier)
        {
            Assert.AreEqual(tier, DeviceClass.TierOf(new DeviceFacts(memoryMb, gpuName, maxRefreshHz)));
        }

        [TestCase(DeviceTier.Low, GraphicsLevel.Low, GraphicsLevel.Low)]
        [TestCase(DeviceTier.Mid, GraphicsLevel.Medium, GraphicsLevel.Medium)]
        [TestCase(DeviceTier.High, GraphicsLevel.High, GraphicsLevel.Ultra)]
        public void AutoPicksTheTierDefaultBelowTheHighest(DeviceTier tier, GraphicsLevel auto, GraphicsLevel highest)
        {
            var quality = new GraphicsQuality(tier, null);
            Assert.IsNull(quality.Choice);
            Assert.AreEqual(auto, quality.Level);
            Assert.AreEqual(highest, quality.Highest);
        }

        [Test]
        public void ALevelAboveTheHighestCannotBeChosen()
        {
            var quality = new GraphicsQuality(DeviceTier.Mid, null);
            var changes = 0;
            quality.Changed += () => changes++;

            Assert.IsFalse(quality.Choose(GraphicsLevel.High));
            Assert.IsFalse(quality.Choose(GraphicsLevel.Ultra));
            Assert.AreEqual(GraphicsLevel.Medium, quality.Level);
            Assert.AreEqual(0, changes);
        }

        [Test]
        public void ALowerLevelAppliesOnceAndAutoComesBack()
        {
            var quality = new GraphicsQuality(DeviceTier.High, null);
            var changes = 0;
            quality.Changed += () => changes++;

            Assert.IsTrue(quality.Choose(GraphicsLevel.Low));
            Assert.IsTrue(quality.Choose(GraphicsLevel.Low));
            Assert.AreEqual(GraphicsLevel.Low, quality.Level);
            Assert.AreEqual(1, changes);

            Assert.IsTrue(quality.Choose(null));
            Assert.AreEqual(GraphicsLevel.High, quality.Level);
            Assert.AreEqual(2, changes);
        }

        [Test]
        public void ASavedChoiceAboveTheHighestCountsAsTheHighest()
        {
            var quality = new GraphicsQuality(DeviceTier.Mid, GraphicsLevel.Ultra);
            Assert.AreEqual(GraphicsLevel.Medium, quality.Choice);
            Assert.AreEqual(GraphicsLevel.Medium, quality.Level);
        }

        [TestCase(GraphicsLevel.Low, EffectLevel.Low)]
        [TestCase(GraphicsLevel.Medium, EffectLevel.Medium)]
        [TestCase(GraphicsLevel.High, EffectLevel.High)]
        [TestCase(GraphicsLevel.Ultra, EffectLevel.High)]
        public void EachLevelRunsItsEffectLevel(GraphicsLevel level, EffectLevel effects)
        {
            Assert.AreEqual(effects, new GraphicsQuality(DeviceTier.High, level).Effects);
        }

        [TestCase(GraphicsLevel.Low, 1080, 0.6667f)]
        [TestCase(GraphicsLevel.Low, 720, 1f)]
        [TestCase(GraphicsLevel.Medium, 1440, 0.75f)]
        [TestCase(GraphicsLevel.High, 1440, 0.75f)]
        [TestCase(GraphicsLevel.Ultra, 1440, 1f)]
        [TestCase(GraphicsLevel.Ultra, 0, 1f)]
        public void TheCameraRendersNoLargerThanTheLevelCap(GraphicsLevel level, int nativeShortSide, float renderScale)
        {
            Assert.AreEqual(renderScale, new GraphicsQuality(DeviceTier.High, level).RenderScale(nativeShortSide), 0.0001f);
        }
    }
}
