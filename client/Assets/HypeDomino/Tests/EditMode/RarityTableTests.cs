using System.Collections.Generic;
using System.IO;
using HypeDomino.UI.Rarity;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class RarityTableTests
    {
        const string DefaultsPath = "Assets/HypeDomino/UI/Rarity/rarity-defaults.json";
        static readonly string[] TierIds = { "common", "rare", "epic", "legendary", "mythic" };

        static RarityTable Defaults() => RarityTable.Parse(File.ReadAllText(DefaultsPath));

        [Test]
        public void TheBuiltInDefaultsHoldEveryTierInOrder()
        {
            var table = Defaults();
            Assert.AreEqual(TierIds.Length, table.Count);
            for (var at = 0; at < TierIds.Length; at++)
            {
                Assert.IsTrue(table.TryGet(TierIds[at], out var tier));
                Assert.AreEqual(at + 1, tier.Ordinal);
                Assert.AreEqual($"rarity.{TierIds[at]}.name", tier.NameKey);
            }
        }

        [Test]
        public void EveryTierHasItsOwnPipPatternSoColourIsNeverTheOnlyCue()
        {
            var table = Defaults();
            var patterns = new HashSet<int>();
            foreach (var id in TierIds)
            {
                table.TryGet(id, out var tier);
                Assert.IsTrue(patterns.Add(tier.Pips), id);
            }
        }

        [Test]
        public void OnlyTheTopTierFlows()
        {
            var table = Defaults();
            foreach (var id in TierIds)
            {
                table.TryGet(id, out var tier);
                Assert.AreEqual(id == "mythic", tier.Flow, id);
            }
        }

        [Test]
        public void HexValuesBecomeOpaqueColours()
        {
            var table = RarityTable.Parse(@"{""tiers"":[{""id"":""a"",""fill"":""#FF0000"",""on"":""#00FF00"",""text"":""#0000FF"",""pips"":1}]}");
            table.TryGet("a", out var tier);
            Assert.AreEqual(Color.red, tier.Fill);
            Assert.AreEqual(Color.green, tier.On);
            Assert.AreEqual(Color.blue, tier.Text);
        }

        [Test]
        public void AnUnknownOrEmptyIdHasNoTier()
        {
            var table = Defaults();
            Assert.IsFalse(table.TryGet("unknown", out _));
            Assert.IsFalse(table.TryGet("", out _));
            Assert.IsFalse(table.TryGet(null, out _));
        }

        [TestCase(@"{""id"":""a"",""fill"":""nope"",""on"":""#000000"",""text"":""#000000"",""pips"":1}")]
        [TestCase(@"{""id"":""a"",""fill"":""#000000"",""on"":""#000000"",""text"":""#000000"",""pips"":0}")]
        [TestCase(@"{""id"":"""",""fill"":""#000000"",""on"":""#000000"",""text"":""#000000"",""pips"":1}")]
        public void AnInvalidTierIsLeftOutSoItDrawsNoMark(string entry)
        {
            Assert.AreEqual(0, RarityTable.Parse(@"{""tiers"":[" + entry + "]}").Count);
        }

        [Test]
        public void TheFlowPeriodIsInSeconds()
        {
            Assert.AreEqual(2.4f, RarityTable.Parse(@"{""flowPeriodMs"":2400}").FlowPeriodSeconds, 0.0001f);
        }
    }
}
