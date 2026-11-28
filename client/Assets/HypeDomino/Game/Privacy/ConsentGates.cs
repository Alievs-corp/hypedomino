using System;
using HypeDomino.Core.Analytics;
using HypeDomino.Core.Diagnostics;

namespace HypeDomino.Game.Privacy
{
    public sealed class ConsentGates
    {
        readonly AnalyticsQueue analytics;
        readonly ErrorReports diagnostics;
        readonly Func<string> newAnalyticsId;
        string analyticsAccount;

        public ConsentGates(AnalyticsQueue analytics, ErrorReports diagnostics, Func<string> newAnalyticsId)
        {
            this.analytics = analytics ?? throw new ArgumentNullException(nameof(analytics));
            this.diagnostics = diagnostics ?? throw new ArgumentNullException(nameof(diagnostics));
            this.newAnalyticsId = newAnalyticsId ?? throw new ArgumentNullException(nameof(newAnalyticsId));
        }

        public void Apply(ConsentChoicesReply choices)
        {
            if (choices == null) throw new ArgumentNullException(nameof(choices));
            Analytics(choices.analytics);
            Diagnostics(choices.diagnostics);
        }

        public void SignedIn(string accountId)
        {
            if (string.IsNullOrEmpty(accountId) || accountId == analyticsAccount) return;
            if (analyticsAccount != null) analytics.Withdraw();
            analyticsAccount = accountId;
        }

        public void Withdraw(bool analyticsOn, bool diagnosticsOn)
        {
            if (!analyticsOn) analytics.Withdraw();
            if (!diagnosticsOn) diagnostics.Withdraw();
        }

        void Analytics(bool on)
        {
            if (!on)
            {
                analytics.Withdraw();
                return;
            }
            if (!analytics.Consented) analytics.Consent(newAnalyticsId());
        }

        void Diagnostics(bool on)
        {
            if (on)
            {
                diagnostics.Consent();
                return;
            }
            diagnostics.Withdraw();
        }
    }
}
