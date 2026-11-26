using Hypedomino.V1;

namespace HypeDomino.Game.Privacy
{
    public sealed class PolicyVersions
    {
        public string PrivacyPolicy { get; private set; } = string.Empty;

        public string Terms { get; private set; } = string.Empty;

        public void Read(ServerHello hello)
        {
            var versions = hello.PrivacyVersions;
            if (hello.Status != HelloStatus.Accepted || versions == null) return;
            PrivacyPolicy = versions.PrivacyPolicyVersion;
            Terms = versions.TermsVersion;
        }
    }
}
