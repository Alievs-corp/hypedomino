using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;
using UnityEngine;

namespace HypeDomino.UI.Rarity
{
    public sealed class RarityTier
    {
        public RarityTier(string id, int ordinal, string nameKey, Color fill, Color on, Color text, int pips, float glowAlpha, float glowScale, bool flow)
        {
            Id = id;
            Ordinal = ordinal;
            NameKey = nameKey;
            Fill = fill;
            On = on;
            Text = text;
            Pips = pips;
            GlowAlpha = glowAlpha;
            GlowScale = glowScale;
            Flow = flow;
        }

        public string Id { get; }
        public int Ordinal { get; }
        public string NameKey { get; }
        public Color Fill { get; }
        public Color On { get; }
        public Color Text { get; }
        public int Pips { get; }
        public float GlowAlpha { get; }
        public float GlowScale { get; }
        public bool Flow { get; }
    }

    [Serializable]
    public sealed class RarityTierEntry
    {
        public string id;
        public int ordinal;
        public string nameKey;
        public string fill;
        public string on;
        public string text;
        public int pips;
        public float glowAlpha;
        public float glowScale;
        public bool flow;
    }

    [Serializable]
    public sealed class RarityTableFile
    {
        public int diamondMinSize;
        public int flowPeriodMs;
        public RarityTierEntry[] tiers;
    }

    public sealed class RarityTable
    {
        public const string BuiltinResource = "rarity-defaults";

        const float MillisecondsPerSecond = 1000f;
        const int MaxPips = 6;
        static readonly Regex TierId = new(@"^[a-z][a-z0-9_]*\z", RegexOptions.CultureInvariant);
        static readonly Regex Hex = new(@"^#[0-9A-Fa-f]{6}\z", RegexOptions.CultureInvariant);

        readonly Dictionary<string, RarityTier> tiers = new();

        RarityTable(RarityTableFile file)
        {
            DiamondMinSize = file.diamondMinSize;
            FlowPeriodSeconds = file.flowPeriodMs / MillisecondsPerSecond;
            var valid = new List<RarityTierEntry>();
            foreach (var entry in file.tiers ?? Array.Empty<RarityTierEntry>())
            {
                if (IsValid(entry)) valid.Add(entry);
            }
            foreach (var entry in valid)
            {
                if (IsUnique(valid, entry)) Add(entry);
            }
        }

        public int DiamondMinSize { get; }

        public float FlowPeriodSeconds { get; }

        public int Count => tiers.Count;

        public static RarityTable Parse(string json) => new(JsonUtility.FromJson<RarityTableFile>(json));

        public static RarityTable From(RarityTableFile file) => new(file ?? new RarityTableFile());

        public static RarityTable LoadBuiltin()
        {
            var asset = Resources.Load<TextAsset>(BuiltinResource);
            if (asset == null)
            {
                Debug.LogError($"Rarity builtin missing resource={BuiltinResource}");
                return From(null);
            }
            var table = Parse(asset.text);
            Resources.UnloadAsset(asset);
            return table;
        }

        public bool TryGet(string id, out RarityTier tier)
        {
            tier = null;
            return !string.IsNullOrEmpty(id) && tiers.TryGetValue(id, out tier);
        }

        static bool IsValid(RarityTierEntry entry) =>
            entry != null
            && entry.id != null && TierId.IsMatch(entry.id)
            && entry.ordinal >= 1
            && entry.nameKey == $"rarity.{entry.id}.name"
            && entry.pips >= 1 && entry.pips <= MaxPips
            && IsHex(entry.fill) && IsHex(entry.on) && IsHex(entry.text);

        static bool IsHex(string value) => value != null && Hex.IsMatch(value);

        static bool IsUnique(List<RarityTierEntry> entries, RarityTierEntry entry)
        {
            foreach (var other in entries)
            {
                if (other == entry) continue;
                if (other.id == entry.id || other.ordinal == entry.ordinal || other.pips == entry.pips) return false;
            }
            return true;
        }

        void Add(RarityTierEntry entry)
        {
            ColorUtility.TryParseHtmlString(entry.fill, out var fill);
            ColorUtility.TryParseHtmlString(entry.on, out var on);
            ColorUtility.TryParseHtmlString(entry.text, out var text);
            tiers[entry.id] = new RarityTier(entry.id, entry.ordinal, entry.nameKey, fill, on, text, entry.pips, entry.glowAlpha, entry.glowScale, entry.flow);
        }
    }
}
