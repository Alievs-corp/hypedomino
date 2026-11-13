using System;
using System.Collections.Generic;
using UnityEngine;

namespace HypeDomino.UI.Cosmetics
{
    public sealed class CosmeticItem
    {
        public CosmeticItem(string id, string type, string rarityId, string nameKey, string availability, bool isDefault, CosmeticArt art, TableLook table)
        {
            Id = id;
            Type = type;
            RarityId = rarityId;
            NameKey = nameKey;
            Availability = availability;
            IsDefault = isDefault;
            Art = art;
            Table = table;
        }

        public string Id { get; }

        public string Type { get; }

        public string RarityId { get; }

        public string NameKey { get; }

        public string Availability { get; }

        public bool IsDefault { get; }

        public bool IsLive => Availability == CosmeticManifest.Live;

        public CosmeticArt Art { get; }

        public TableLook Table { get; }
    }

    public sealed class CosmeticManifest
    {
        public const int KnownSchemaVersion = 1;
        public const string Live = "live";
        public const string Retired = "retired";
        const string DefaultSource = "default";

        static readonly string[] KnownTypes = { CosmeticAddresses.Avatar, CosmeticAddresses.Frame, CosmeticAddresses.Table, CosmeticAddresses.Tile, CosmeticAddresses.Emoji };

        [Serializable]
        sealed class SourceEntry
        {
            public string type;
        }

        [Serializable]
        sealed class LayerEntry
        {
            public string name;
            public string address;
            public bool animated;
        }

        [Serializable]
        sealed class ItemEntry
        {
            public string id;
            public string type;
            public string rarityId;
            public string nameKey;
            public string availability;
            public bool animated;
            public string staticAddress;
            public string animationAddress;
            public LayerEntry[] layers;
            public SourceEntry[] sources;
        }

        [Serializable]
        sealed class ManifestFile
        {
            public int schemaVersion;
            public int manifestVersion;
            public ItemEntry[] items;
        }

        readonly List<CosmeticItem> items = new();
        readonly Dictionary<string, CosmeticItem> byId = new();

        CosmeticManifest(int manifestVersion) => ManifestVersion = manifestVersion;

        public int ManifestVersion { get; }

        public int Count => items.Count;

        public CosmeticItem this[int index] => items[index];

        public bool TryGet(string id, out CosmeticItem item)
        {
            item = null;
            return !string.IsNullOrEmpty(id) && byId.TryGetValue(id, out item);
        }

        public CosmeticItem DefaultOf(string type)
        {
            foreach (var item in items)
                if (item.Type == type && item.IsDefault && item.IsLive) return item;
            return null;
        }

        public static bool TryParse(string json, out CosmeticManifest manifest)
        {
            manifest = null;
            ManifestFile file;
            try
            {
                file = JsonUtility.FromJson<ManifestFile>(json);
            }
            catch (ArgumentException)
            {
                return false;
            }
            if (file == null || file.schemaVersion != KnownSchemaVersion || file.items == null || file.items.Length == 0) return false;
            manifest = new CosmeticManifest(file.manifestVersion);
            foreach (var entry in file.items) manifest.Add(entry);
            return true;
        }

        void Add(ItemEntry entry)
        {
            if (string.IsNullOrEmpty(entry.id) || string.IsNullOrEmpty(entry.staticAddress)) return;
            if (Array.IndexOf(KnownTypes, entry.type) < 0 || byId.ContainsKey(entry.id)) return;
            var item = new CosmeticItem(entry.id, entry.type, entry.rarityId, entry.nameKey, entry.availability, IsDefault(entry), Art(entry), Table(entry));
            items.Add(item);
            byId.Add(item.Id, item);
        }

        static CosmeticArt Art(ItemEntry entry) =>
            new(entry.staticAddress, entry.animated && entry.type != CosmeticAddresses.Table ? entry.animationAddress : null);

        static TableLook Table(ItemEntry entry)
        {
            if (entry.type != CosmeticAddresses.Table || !entry.animated || entry.layers == null) return new TableLook(entry.staticAddress, null);
            var layers = new TableLayer[entry.layers.Length];
            for (var at = 0; at < layers.Length; at++)
            {
                var layer = entry.layers[at];
                layers[at] = new TableLayer(layer.name, layer.address, layer.animated);
            }
            return new TableLook(entry.staticAddress, layers);
        }

        static bool IsDefault(ItemEntry entry)
        {
            if (entry.sources == null) return false;
            foreach (var source in entry.sources)
                if (source != null && source.type == DefaultSource) return true;
            return false;
        }
    }
}
