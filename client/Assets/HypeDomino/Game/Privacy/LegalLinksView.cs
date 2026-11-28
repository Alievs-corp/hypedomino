using System;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Privacy
{
    public sealed class LegalLinksView : MonoBehaviour
    {
        [SerializeField] Button policy;
        [SerializeField] Button terms;

        PolicyVersions versions;
        Action<string> open;

        public void Configure(Button policyButton, Button termsButton)
        {
            Unbind();
            policy = policyButton;
            terms = termsButton;
            Show();
        }

        public void Bind(PolicyVersions announced, Action<string> openLink)
        {
            Unbind();
            versions = announced;
            open = openLink;
            versions.Changed += Show;
            policy.onClick.AddListener(OpenPolicy);
            terms.onClick.AddListener(OpenTerms);
            Show();
        }

        void Start()
        {
            if (versions == null && PrivacyBoot.Versions != null) Bind(PrivacyBoot.Versions, Application.OpenURL);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (versions == null) return;
            versions.Changed -= Show;
            policy.onClick.RemoveListener(OpenPolicy);
            terms.onClick.RemoveListener(OpenTerms);
            versions = null;
            open = null;
        }

        void Show()
        {
            policy.gameObject.SetActive(!string.IsNullOrEmpty(versions?.PrivacyPolicyUrl));
            terms.gameObject.SetActive(!string.IsNullOrEmpty(versions?.TermsOfUseUrl));
        }

        void OpenPolicy() => Open(versions?.PrivacyPolicyUrl);

        void OpenTerms() => Open(versions?.TermsOfUseUrl);

        void Open(string url)
        {
            if (!string.IsNullOrEmpty(url)) open?.Invoke(url);
        }
    }
}
