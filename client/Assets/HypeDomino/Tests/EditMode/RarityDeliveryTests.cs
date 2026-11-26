using HypeDomino.Game;
using HypeDomino.UI.Medals;
using HypeDomino.UI.Rarity;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class RarityDeliveryTests
    {
        ActiveRarityTable active;
        RarityDelivery delivery;

        [SetUp]
        public void Create()
        {
            active = new ActiveRarityTable(RarityTable.LoadBuiltin());
            delivery = new RarityDelivery(active);
        }

        static RarityTierConfig Tier(string id, uint ordinal, string fill, uint pips) => new()
        {
            Id = id,
            Ordinal = ordinal,
            NameKey = $"rarity.{id}.name",
            Fill = fill,
            On = "#0B0605",
            Text = "#FFFFFF",
            Pips = pips,
            GlowAlpha = 0.35f,
            GlowScale = 1.15f,
            Flow = id == "rare",
        };

        static RarityTableConfig Swapped() => new()
        {
            DiamondMinSize = 96,
            FlowPeriodMs = 3000,
            Tiers = { Tier("common", 2, "#00FF00", 1), Tier("rare", 1, "#FF0000", 2) },
        };

        static ServerHello Hello(HelloStatus status, RarityTableConfig table) => new() { Status = status, RarityTable = table };

        static string[] GridOrder(RarityTable table)
        {
            var collection = new MedalCollection(table);
            collection.Show(MedalCollectionReply.FromJson(@"{""slots"":3,""medals"":[
{""id"":""medal.rare"",""rarityId"":""rare"",""group"":""achievement""},
{""id"":""medal.common"",""rarityId"":""common"",""group"":""achievement""}]}"));
            var ids = new string[collection.ShownCount];
            for (var at = 0; at < ids.Length; at++) ids[at] = collection.ShownAt(at).id;
            return ids;
        }

        [Test]
        public void TheShippedDefaultsHoldUntilATableIsDelivered()
        {
            Assert.AreEqual(5, active.Current.Count);
            Assert.AreEqual(72, active.Current.DiamondMinSize);
            Assert.AreEqual(2.4f, active.Current.FlowPeriodSeconds, 1e-6f);
            Assert.IsTrue(active.Current.TryGet("mythic", out var mythic));
            Assert.AreEqual(5, mythic.Ordinal);
            Assert.IsTrue(mythic.Flow);
        }

        [Test]
        public void AnAcceptedHelloReplacesTheTableWithTheDeliveredOne()
        {
            delivery.Read(Hello(HelloStatus.Accepted, Swapped()));

            Assert.AreEqual(2, active.Current.Count);
            Assert.AreEqual(96, active.Current.DiamondMinSize);
            Assert.AreEqual(3f, active.Current.FlowPeriodSeconds, 1e-6f);
            Assert.IsTrue(active.Current.TryGet("rare", out var rare));
            Assert.AreEqual(1, rare.Ordinal);
            Assert.AreEqual(Color.red, rare.Fill);
            Assert.AreEqual(0.35f, rare.GlowAlpha, 1e-6f);
            Assert.AreEqual(1.15f, rare.GlowScale, 1e-6f);
            Assert.IsTrue(rare.Flow);
            Assert.IsFalse(active.Current.TryGet("mythic", out _));
        }

        [Test]
        public void AHelloWithoutATableKeepsTheCurrentOne()
        {
            var builtin = active.Current;
            delivery.Read(Hello(HelloStatus.Accepted, null));
            Assert.AreSame(builtin, active.Current);

            delivery.Read(Hello(HelloStatus.Accepted, Swapped()));
            var delivered = active.Current;
            delivery.Read(Hello(HelloStatus.Accepted, null));
            Assert.AreSame(delivered, active.Current);
        }

        [Test]
        public void AHelloThatIsNotAcceptedChangesNothing()
        {
            var builtin = active.Current;
            delivery.Read(Hello(HelloStatus.UpdateRequired, Swapped()));
            delivery.Read(Hello(HelloStatus.Maintenance, Swapped()));
            Assert.AreSame(builtin, active.Current);
        }

        [Test]
        public void AnInvalidRarityTierConfigIsLeftOutLikeAnyOther()
        {
            var table = Swapped();
            table.Tiers.Add(Tier("epic", 3, "purple", 3));
            delivery.Read(Hello(HelloStatus.Accepted, table));

            Assert.AreEqual(2, active.Current.Count);
            Assert.IsFalse(active.Current.TryGet("epic", out _));
        }

        [Test]
        public void TheMedalGridOrdersByTheDeliveredOrdinals()
        {
            CollectionAssert.AreEqual(new[] { "medal.rare", "medal.common" }, GridOrder(active.Current));

            delivery.Read(Hello(HelloStatus.Accepted, Swapped()));

            CollectionAssert.AreEqual(new[] { "medal.common", "medal.rare" }, GridOrder(active.Current));
        }

        [Test]
        public void ApplyingNoTableKeepsTheCurrentOne()
        {
            var builtin = active.Current;
            active.Apply(null);
            Assert.AreSame(builtin, active.Current);
        }
    }
}
