using System;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Privacy
{
    public sealed class PrivacyNoticeView : MonoBehaviour
    {
        [SerializeField] Button policy;
        [SerializeField] Button terms;
        [SerializeField] Button proceed;

        string policyUrl;
        string termsUrl;
        Action<string> open;
        Action continued;

        public void Configure(Button policyButton, Button termsButton, Button continueButton)
        {
            Unbind();
            policy = policyButton;
            terms = termsButton;
            proceed = continueButton;
            ShowLinks(null, null);
        }

        public void Bind(string privacyPolicyUrl, string termsOfUseUrl, Action<string> openLink, Action done)
        {
            Unbind();
            open = openLink;
            continued = done;
            ShowLinks(privacyPolicyUrl, termsOfUseUrl);
            policy.onClick.AddListener(OpenPolicy);
            terms.onClick.AddListener(OpenTerms);
            proceed.onClick.AddListener(Continue);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (continued == null) return;
            policy.onClick.RemoveListener(OpenPolicy);
            terms.onClick.RemoveListener(OpenTerms);
            proceed.onClick.RemoveListener(Continue);
            open = null;
            continued = null;
        }

        void ShowLinks(string policyLink, string termsLink)
        {
            policyUrl = policyLink;
            termsUrl = termsLink;
            policy.gameObject.SetActive(!string.IsNullOrEmpty(policyUrl));
            terms.gameObject.SetActive(!string.IsNullOrEmpty(termsUrl));
        }

        void OpenPolicy() => open?.Invoke(policyUrl);

        void OpenTerms() => open?.Invoke(termsUrl);

        void Continue()
        {
            if (continued == null) return;
            var next = continued;
            Unbind();
            next();
        }
    }
}
