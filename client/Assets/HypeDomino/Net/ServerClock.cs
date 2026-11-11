namespace HypeDomino.Net
{
    public sealed class ServerClock
    {
        double offsetMs;

        public bool Synced { get; private set; }

        public void Sync(ulong serverUnixMs, double sentSeconds, double receivedSeconds)
        {
            offsetMs = serverUnixMs - (sentSeconds + receivedSeconds) * 500;
            Synced = true;
        }

        public double UnixMs(double nowSeconds) => nowSeconds * 1000 + offsetMs;

        public double SecondsUntil(ulong serverUnixMs, double nowSeconds) => (serverUnixMs - UnixMs(nowSeconds)) / 1000;
    }
}
