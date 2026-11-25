namespace HypeDomino.Content
{
    public sealed class NoCacheStore : ICacheStore
    {
        public long OccupiedBytes => 0;

        public bool ClearAtStart { get; set; }

        public void ApplyLimits(long limitBytes, int unusedForSeconds)
        {
        }

        public bool TryClear() => true;
    }
}
