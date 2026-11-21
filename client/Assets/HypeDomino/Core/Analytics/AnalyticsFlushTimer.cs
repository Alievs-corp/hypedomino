using System;

namespace HypeDomino.Core.Analytics
{
    public sealed class AnalyticsFlushTimer
    {
        readonly AnalyticsQueue queue;
        readonly double intervalSeconds;
        double dueAt;
        bool started;

        public AnalyticsFlushTimer(AnalyticsQueue queue, double intervalSeconds)
        {
            if (!(intervalSeconds > 0)) throw new ArgumentOutOfRangeException(nameof(intervalSeconds));
            this.queue = queue ?? throw new ArgumentNullException(nameof(queue));
            this.intervalSeconds = intervalSeconds;
        }

        public void Tick(double nowSeconds)
        {
            if (!started)
            {
                started = true;
                dueAt = nowSeconds + intervalSeconds;
                return;
            }
            if (nowSeconds < dueAt) return;
            dueAt = nowSeconds + intervalSeconds;
            queue.Flush();
        }

        public void AppPaused() => queue.Flush();
    }
}
