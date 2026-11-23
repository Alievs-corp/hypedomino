using System;
using HypeDomino.UI.Settings;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Privacy
{
    public sealed class FirstLaunchConsentView : MonoBehaviour
    {
        [SerializeField] SettingToggleRow analytics;
        [SerializeField] SettingToggleRow diagnostics;
        [SerializeField] Button proceed;

        FirstLaunchConsents choices;
        ConsentStartup startup;
        Action continued;

        public void Configure(SettingToggleRow analyticsRow, SettingToggleRow diagnosticsRow, Button continueButton)
        {
            Unbind();
            analytics = analyticsRow;
            diagnostics = diagnosticsRow;
            proceed = continueButton;
        }

        public void Bind(FirstLaunchConsents held, ConsentStartup consents, Action done)
        {
            Unbind();
            choices = held;
            startup = consents;
            continued = done;
            analytics.Bind(false, OnAnalytics);
            diagnostics.Bind(false, OnDiagnostics);
            proceed.onClick.AddListener(Continue);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (choices == null) return;
            proceed.onClick.RemoveListener(Continue);
            choices = null;
            startup = null;
            continued = null;
        }

        void OnAnalytics(bool on) => choices?.Choose(ConsentPurpose.Analytics, on);

        void OnDiagnostics(bool on) => choices?.Choose(ConsentPurpose.Diagnostics, on);

        void Continue()
        {
            startup.Hold(choices);
            continued();
        }
    }
}
