using System;
using HypeDomino.Core.Analytics;
using HypeDomino.Core.Diagnostics;
using HypeDomino.Net.Api;

namespace HypeDomino.Game.Privacy
{
    public sealed class PrivacySettings
    {
        public const string Path = "/consents";

        readonly ApiClient api;
        readonly AnalyticsQueue analytics;
        readonly ErrorReports diagnostics;
        readonly Func<string> newAnalyticsId;
        readonly string policyVersion;
        readonly string appVersion;

        public PrivacySettings(ApiClient api, AnalyticsQueue analytics, ErrorReports diagnostics, Func<string> newAnalyticsId, string policyVersion, string appVersion)
        {
            if (string.IsNullOrEmpty(policyVersion)) throw new ArgumentException("policy version required", nameof(policyVersion));
            if (string.IsNullOrEmpty(appVersion)) throw new ArgumentException("app version required", nameof(appVersion));
            this.api = api;
            this.analytics = analytics ?? throw new ArgumentNullException(nameof(analytics));
            this.diagnostics = diagnostics ?? throw new ArgumentNullException(nameof(diagnostics));
            this.newAnalyticsId = newAnalyticsId ?? throw new ArgumentNullException(nameof(newAnalyticsId));
            this.policyVersion = policyVersion;
            this.appVersion = appVersion;
        }

        public event Action Changed;

        public bool Busy { get; private set; }

        public bool Loaded { get; private set; }

        public bool Analytics { get; private set; }

        public bool Diagnostics { get; private set; }

        public static string RandomAnalyticsId() => Guid.NewGuid().ToString("D");

        public bool Load(Action<PrivacySettingsResult> done)
        {
            if (Busy) return false;
            Busy = true;
            api.Get(Path, reply => Answer(reply, done));
            return true;
        }

        public bool SetAnalytics(bool on, Action<PrivacySettingsResult> done) => Save(on, Diagnostics, done);

        public bool SetDiagnostics(bool on, Action<PrivacySettingsResult> done) => Save(Analytics, on, done);

        bool Save(bool analyticsOn, bool diagnosticsOn, Action<PrivacySettingsResult> done)
        {
            if (Busy || !Loaded) return false;
            Busy = true;
            if (!analyticsOn) analytics.Withdraw();
            if (!diagnosticsOn) diagnostics.Withdraw();
            var request = new ConsentChangeRequest { analytics = analyticsOn, diagnostics = diagnosticsOn, policyVersion = policyVersion, appVersion = appVersion };
            api.Put(Path, request.ToJson(), reply => Answer(reply, done));
            return true;
        }

        void Answer(ApiReply reply, Action<PrivacySettingsResult> done)
        {
            Busy = false;
            var result = PrivacySettingsResult.From(reply);
            if (result.Outcome == PrivacySettingsOutcome.Done) Apply(result.Choices);
            done(result);
        }

        void Apply(ConsentChoicesReply choices)
        {
            var changed = !Loaded || choices.analytics != Analytics || choices.diagnostics != Diagnostics;
            Loaded = true;
            Analytics = choices.analytics;
            Diagnostics = choices.diagnostics;
            GateAnalytics();
            GateDiagnostics();
            if (changed) Changed?.Invoke();
        }

        void GateAnalytics()
        {
            if (!Analytics)
            {
                analytics.Withdraw();
                return;
            }
            if (!analytics.Consented) analytics.Consent(newAnalyticsId());
        }

        void GateDiagnostics()
        {
            if (Diagnostics)
            {
                diagnostics.Consent();
                return;
            }
            diagnostics.Withdraw();
        }
    }
}
