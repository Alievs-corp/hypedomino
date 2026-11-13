using System.IO;
using HypeDomino.UI.Cosmetics;
using NUnit.Framework;
using UnityEngine;

namespace HypeDomino.Tests.EditMode
{
    public sealed class CosmeticManifestTests
    {
        const string BuiltinPath = "Assets/HypeDomino/UI/Cosmetics/cosmetics-builtin.json";

        static string RepoFile(params string[] parts)
        {
            var path = Path.Combine(Application.dataPath, "..", "..");
            foreach (var part in parts) path = Path.Combine(path, part);
            return File.ReadAllText(Path.GetFullPath(path));
        }

        static CosmeticManifest Parse(string json)
        {
            Assert.IsTrue(CosmeticManifest.TryParse(json, out var manifest));
            return manifest;
        }

        static string Manifest(string items, int schemaVersion = 1) =>
            "{\"schemaVersion\":" + schemaVersion + ",\"manifestVersion\":3,\"items\":[" + items + "]}";

        static string Item(string id, string type, string extra = "", string rarityId = "common", string availability = "live") =>
            "{\"id\":\"" + id + "\",\"type\":\"" + type + "\",\"rarityId\":\"" + rarityId + "\",\"availability\":\"" + availability + "\",\"staticAddress\":\"cosmetic/" + type + "/x/static\"" + extra + "}";

        [Test]
        public void TheBuiltInManifestIsTheServersBuiltInCopy()
        {
            Assert.AreEqual(RepoFile("server", "internal", "cosmetics", "builtin.json"), File.ReadAllText(BuiltinPath));
        }

        [TestCase(CosmeticAddresses.Avatar, "avatar.starter")]
        [TestCase(CosmeticAddresses.Frame, "frame.starter")]
        [TestCase(CosmeticAddresses.Table, "table.starter")]
        [TestCase(CosmeticAddresses.Tile, "tile.starter")]
        [TestCase(CosmeticAddresses.Emoji, "emoji.smile")]
        public void TheBuiltInManifestHasALiveDefaultOfEveryType(string type, string id)
        {
            var manifest = Parse(File.ReadAllText(BuiltinPath));

            Assert.AreEqual(id, manifest.DefaultOf(type).Id);
            Assert.IsFalse(manifest.DefaultOf(type).Art.IsAnimated);
        }

        [Test]
        public void TheExampleCatalogGivesEachItemItsArt()
        {
            var manifest = Parse(RepoFile("docs", "cosmetics", "examples", "catalog.json"));

            Assert.AreEqual(7, manifest.ManifestVersion);
            Assert.AreEqual(16, manifest.Count);
            Assert.IsTrue(manifest.TryGet("avatar.comet_fox", out var fox));
            Assert.AreEqual("cosmetic/avatar/comet_fox/static", fox.Art.StaticAddress);
            Assert.AreEqual("cosmetic/avatar/comet_fox/anim", fox.Art.AnimationAddress);
            Assert.IsTrue(manifest.TryGet("avatar.moon_owl", out var owl));
            Assert.IsFalse(owl.Art.IsAnimated);
        }

        [Test]
        public void AnAnimatedTableGivesItsLayersBackToFront()
        {
            var manifest = Parse(RepoFile("docs", "cosmetics", "examples", "catalog.json"));

            Assert.IsTrue(manifest.TryGet("table.night_market", out var table));
            var look = table.Table;
            Assert.AreEqual("cosmetic/table/night_market/static", look.StaticAddress);
            Assert.AreEqual(new[] { "sky", "stalls", "lanterns" }, new[] { look.Layers[0].Name, look.Layers[1].Name, look.Layers[2].Name });
            Assert.AreEqual("cosmetic/table/night_market/layer/stalls", look.Layers[1].Address);
            Assert.AreEqual(2, look.AnimatedLayerCount);
            Assert.IsFalse(table.Art.IsAnimated);
        }

        [Test]
        public void AStaticTableHasNoLayers()
        {
            var manifest = Parse(File.ReadAllText(BuiltinPath));

            Assert.IsTrue(manifest.TryGet("table.starter", out var table));
            Assert.IsFalse(table.Table.IsAnimated);
            Assert.AreEqual("cosmetic/table/starter/static", table.Table.StaticAddress);
        }

        [Test]
        public void AnUnknownSchemaVersionIsRefused()
        {
            Assert.IsFalse(CosmeticManifest.TryParse(Manifest(Item("avatar.a", "avatar"), 2), out var manifest));
            Assert.IsNull(manifest);
        }

        [TestCase("")]
        [TestCase("{not json")]
        [TestCase("{\"schemaVersion\":1}")]
        [TestCase("{\"schemaVersion\":1,\"items\":[]}")]
        public void AnUnreadableManifestIsRefused(string json)
        {
            Assert.IsFalse(CosmeticManifest.TryParse(json, out _));
        }

        [Test]
        public void UnknownTypesAndFieldsAreSkipped()
        {
            var manifest = Parse(Manifest(Item("hat.top", "hat") + "," + Item("avatar.a", "avatar", ",\"sparkle\":true")));

            Assert.AreEqual(1, manifest.Count);
            Assert.IsTrue(manifest.TryGet("avatar.a", out _));
            Assert.IsFalse(manifest.TryGet("hat.top", out _));
        }

        [Test]
        public void AnItemWithoutAStaticOrIdIsSkipped()
        {
            var manifest = Parse(Manifest("{\"id\":\"avatar.a\",\"type\":\"avatar\"},{\"type\":\"avatar\",\"staticAddress\":\"cosmetic/avatar/b/static\"}"));

            Assert.AreEqual(0, manifest.Count);
        }

        [Test]
        public void TheFirstOfTwoItemsWithOneIdWins()
        {
            var manifest = Parse(Manifest(Item("avatar.a", "avatar", rarityId: "rare") + "," + Item("avatar.a", "avatar")));

            Assert.AreEqual(1, manifest.Count);
            Assert.IsTrue(manifest.TryGet("avatar.a", out var item));
            Assert.AreEqual("rare", item.RarityId);
        }

        [Test]
        public void AnAnimatedItemWithoutItsLoopShowsItsStatic()
        {
            var manifest = Parse(Manifest(Item("avatar.a", "avatar", ",\"animated\":true")));

            Assert.IsTrue(manifest.TryGet("avatar.a", out var item));
            Assert.IsFalse(item.Art.IsAnimated);
        }

        [Test]
        public void TheDefaultIsTheFirstLiveOne()
        {
            var retired = Item("avatar.old", "avatar", ",\"sources\":[{\"type\":\"default\"}]", availability: "retired");
            var bought = Item("avatar.shop", "avatar", ",\"sources\":[{\"type\":\"shop\"}]");
            var starter = Item("avatar.starter", "avatar", ",\"sources\":[{\"type\":\"default\"}]");
            var manifest = Parse(Manifest(retired + "," + bought + "," + starter));

            Assert.AreEqual("avatar.starter", manifest.DefaultOf(CosmeticAddresses.Avatar).Id);
            Assert.IsNull(manifest.DefaultOf(CosmeticAddresses.Frame));
        }

        [TestCase(null)]
        [TestCase("")]
        [TestCase("avatar.missing")]
        public void AnUnknownIdIsNotFound(string id)
        {
            var manifest = Parse(File.ReadAllText(BuiltinPath));

            Assert.IsFalse(manifest.TryGet(id, out var item));
            Assert.IsNull(item);
        }
    }
}
