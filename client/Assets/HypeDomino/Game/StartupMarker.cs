using System;

namespace HypeDomino.Game
{
    public sealed class StartupMarker
    {
        const double MillisecondsPerSecond = 1000d;

        readonly AppFlow flow;
        readonly Func<double> secondsSinceStartup;
        readonly Action<string> log;

        public StartupMarker(AppFlow flow, Func<double> secondsSinceStartup, Action<string> log)
        {
            this.flow = flow;
            this.secondsSinceStartup = secondsSinceStartup;
            this.log = log;
            flow.Changed += OnFirstScreen;
        }

        void OnFirstScreen(AppState state)
        {
            flow.Changed -= OnFirstScreen;
            log($"Startup state={state} ms={(long)(secondsSinceStartup() * MillisecondsPerSecond)}");
        }
    }
}
