using System;
using Hypedomino.V1;

namespace HypeDomino.Game.Privacy
{
    public sealed class PolicyVersions
    {
        public event Action Changed;

        public string PrivacyPolicy { get; private set; } = string.Empty;

        public string Terms { get; private set; } = string.Empty;

        public string PrivacyPolicyUrl { get; private set; } = string.Empty;

        public string TermsOfUseUrl { get; private set; } = string.Empty;

        public void Read(ServerHello hello)
        {
            var versions = hello.PrivacyVersions;
            if (hello.Status != HelloStatus.Accepted || versions == null) return;
            PrivacyPolicy = versions.PrivacyPolicyVersion;
            Terms = versions.TermsVersion;
            if (PrivacyPolicyUrl == versions.PrivacyPolicyUrl && TermsOfUseUrl == versions.TermsOfUseUrl) return;
            PrivacyPolicyUrl = versions.PrivacyPolicyUrl;
            TermsOfUseUrl = versions.TermsOfUseUrl;
            Changed?.Invoke();
        }
    }
}
