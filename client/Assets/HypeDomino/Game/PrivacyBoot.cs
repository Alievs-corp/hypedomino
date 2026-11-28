using HypeDomino.Game.Privacy;
using HypeDomino.Game.Push;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using Hypedomino.V1;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class PrivacyBoot
    {
        static AuthSession auth;

        public static PolicyVersions Versions { get; private set; }

        public static ConsentGates Gates { get; private set; }

        public static ConsentStartup Consents { get; private set; }

        public static FirstLaunchConsents FirstLaunch { get; private set; }

        public static PrivacySettings Settings { get; private set; }

        public static PushSettings Push { get; private set; }

        public static void Start(ApiClient api, AuthSession session, Session net)
        {
            auth = session;
            Versions = new PolicyVersions();
            Gates = new ConsentGates(AnalyticsBoot.Queue, CrashReportingBoot.Reports, PrivacySettings.RandomAnalyticsId);
            session.AccountSignedIn += Gates.SignedIn;
            Consents = new ConsentStartup(api, Gates);
            FirstLaunch = new FirstLaunchConsents(CurrentPolicy, Application.version);
            Settings = new PrivacySettings(api, AnalyticsBoot.Queue, CrashReportingBoot.Reports, PrivacySettings.RandomAnalyticsId, CurrentPolicy, Application.version);
            Push = new PushSettings(api, CurrentPolicy, Application.version);
            net.HelloReceived += SignedIn;
        }

        public static bool Follow(ServerHello hello, PolicyVersions versions, ConsentStartup consents, string accountId)
        {
            versions.Read(hello);
            return hello.Status == HelloStatus.Accepted && consents.SignedIn(accountId);
        }

        static string CurrentPolicy() => Versions.PrivacyPolicy;

        static void SignedIn(ServerHello hello) => Follow(hello, Versions, Consents, auth.AccountId);
    }
}
