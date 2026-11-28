using System;

namespace HypeDomino.Game.Inventory
{
    public sealed class InventoryRetry
    {
        static readonly long[] BackoffMs = { 2000, 4000, 8000 };

        readonly InventoryLoader loader;
        readonly Func<long> clockMs;
        readonly Action<InventoryLoadResult> done;
        readonly Action<InventoryLoadResult> onAnswer;
        int failures;
        long retryAtMs;
        bool retryPending;
        bool answerWanted;
        bool loadWhenAnswered;

        public InventoryRetry(InventoryLoader loader, Func<long> clockMs, Action<InventoryLoadResult> done)
        {
            this.loader = loader ?? throw new ArgumentNullException(nameof(loader));
            this.clockMs = clockMs ?? throw new ArgumentNullException(nameof(clockMs));
            this.done = done ?? throw new ArgumentNullException(nameof(done));
            onAnswer = Answered;
        }

        public void Load()
        {
            failures = 0;
            retryPending = false;
            if (loader.Busy)
            {
                loadWhenAnswered = !answerWanted;
                return;
            }
            Attempt();
        }

        public void Tick()
        {
            if (!retryPending || clockMs() < retryAtMs) return;
            retryPending = false;
            Attempt();
        }

        public void Cancel()
        {
            failures = 0;
            retryPending = false;
            answerWanted = false;
            loadWhenAnswered = false;
        }

        void Attempt()
        {
            answerWanted = true;
            loader.Load(onAnswer);
        }

        void Answered(InventoryLoadResult result)
        {
            if (!answerWanted)
            {
                LoadIfAsked();
                return;
            }
            answerWanted = false;
            if (result.Transient && failures < BackoffMs.Length)
            {
                retryAtMs = clockMs() + Math.Max(BackoffMs[failures++], result.RetryAfterMs);
                retryPending = true;
                return;
            }
            failures = 0;
            done(result);
        }

        void LoadIfAsked()
        {
            if (!loadWhenAnswered) return;
            loadWhenAnswered = false;
            Attempt();
        }
    }
}
