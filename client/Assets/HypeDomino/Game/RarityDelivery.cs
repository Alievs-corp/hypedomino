using HypeDomino.UI.Rarity;
using Hypedomino.V1;

namespace HypeDomino.Game
{
    public sealed class RarityDelivery
    {
        readonly ActiveRarityTable active;

        public RarityDelivery(ActiveRarityTable active) => this.active = active;

        public void Read(ServerHello hello)
        {
            var delivered = hello.RarityTable;
            if (hello.Status != HelloStatus.Accepted || delivered == null) return;
            active.Apply(RarityTable.From(FileOf(delivered)));
        }

        static RarityTableFile FileOf(RarityTableConfig delivered)
        {
            var tiers = new RarityTierEntry[delivered.Tiers.Count];
            for (var at = 0; at < tiers.Length; at++) tiers[at] = EntryOf(delivered.Tiers[at]);
            return new RarityTableFile { diamondMinSize = (int)delivered.DiamondMinSize, flowPeriodMs = (int)delivered.FlowPeriodMs, tiers = tiers };
        }

        static RarityTierEntry EntryOf(RarityTierConfig tier) => new()
        {
            id = tier.Id,
            ordinal = (int)tier.Ordinal,
            nameKey = tier.NameKey,
            fill = tier.Fill,
            on = tier.On,
            text = tier.Text,
            pips = (int)tier.Pips,
            glowAlpha = tier.GlowAlpha,
            glowScale = tier.GlowScale,
            flow = tier.Flow,
        };
    }
}
