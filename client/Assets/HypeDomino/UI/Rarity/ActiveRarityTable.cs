namespace HypeDomino.UI.Rarity
{
    public sealed class ActiveRarityTable
    {
        public ActiveRarityTable(RarityTable builtin) => Current = builtin;

        public RarityTable Current { get; private set; }

        public void Apply(RarityTable delivered)
        {
            if (delivered != null) Current = delivered;
        }
    }
}
