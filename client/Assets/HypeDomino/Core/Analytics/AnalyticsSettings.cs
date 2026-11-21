using System;

namespace HypeDomino.Core.Analytics
{
    public readonly struct AnalyticsSettings
    {
        public const int MaxBatchSize = 400;

        public static readonly AnalyticsSettings Draft = new AnalyticsSettings(50, 500, 60);

        public readonly int BatchSize;
        public readonly int QueueLimit;
        public readonly double FlushSeconds;

        public AnalyticsSettings(int batchSize, int queueLimit, double flushSeconds)
        {
            if (batchSize < 1 || batchSize > MaxBatchSize) throw new ArgumentOutOfRangeException(nameof(batchSize));
            if (queueLimit < batchSize) throw new ArgumentOutOfRangeException(nameof(queueLimit));
            if (!(flushSeconds > 0)) throw new ArgumentOutOfRangeException(nameof(flushSeconds));
            BatchSize = batchSize;
            QueueLimit = queueLimit;
            FlushSeconds = flushSeconds;
        }
    }
}
