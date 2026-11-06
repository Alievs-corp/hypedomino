using System;
using System.Collections.Generic;
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

    public sealed class RarityTable
    {
        const float MillisecondsPerSecond = 1000f;

        [Serializable]
        sealed class TierEntry
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
        sealed class TableFile
        {
            public int diamondMinSize;
            public int flowPeriodMs;
            public TierEntry[] tiers;
        }

        readonly Dictionary<string, RarityTier> tiers = new();

        RarityTable(TableFile file)
        {
            DiamondMinSize = file.diamondMinSize;
            FlowPeriodSeconds = file.flowPeriodMs / MillisecondsPerSecond;
            foreach (var entry in file.tiers ?? Array.Empty<TierEntry>()) Add(entry);
        }

        public int DiamondMinSize { get; }

        public float FlowPeriodSeconds { get; }

        public int Count => tiers.Count;

        public static RarityTable Parse(string json) => new(JsonUtility.FromJson<TableFile>(json));

        public bool TryGet(string id, out RarityTier tier)
        {
            tier = null;
            return !string.IsNullOrEmpty(id) && tiers.TryGetValue(id, out tier);
        }

        void Add(TierEntry entry)
        {
            if (string.IsNullOrEmpty(entry.id) || entry.pips < 1) return;
            if (!ColorUtility.TryParseHtmlString(entry.fill, out var fill)) return;
            if (!ColorUtility.TryParseHtmlString(entry.on, out var on)) return;
            if (!ColorUtility.TryParseHtmlString(entry.text, out var text)) return;
            tiers[entry.id] = new RarityTier(entry.id, entry.ordinal, entry.nameKey, fill, on, text, entry.pips, entry.glowAlpha, entry.glowScale, entry.flow);
        }
    }
}
