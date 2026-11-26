using System;
using System.Collections.Generic;
using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game
{
    public sealed class MaintenanceScreenView : MonoBehaviour
    {
        public const string EndsKey = "ui.plain.maintenance.ends";
        public const string EndArgument = "end";

        [SerializeField] GameObject panel;
        [SerializeField] TMP_Text ends;
        [SerializeField] Button retry;

        readonly Dictionary<string, object> endsArgs = new();
        MaintenanceScreen screen;
        GameStrings strings;

        public static string EndText(ulong endsUnixMs, string language) =>
            MessageFormat.Timestamp(DateTimeOffset.FromUnixTimeMilliseconds((long)endsUnixMs), TimeZoneInfo.Utc, language);

        public void Configure(GameObject screenPanel, TMP_Text endsText, Button retryButton)
        {
            Unbind();
            panel = screenPanel;
            ends = endsText;
            retry = retryButton;
        }

        public void Bind(MaintenanceScreen maintenance, GameStrings texts)
        {
            Unbind();
            screen = maintenance;
            strings = texts;
            screen.Changed += Refresh;
            strings.Changed += Refresh;
            retry.onClick.AddListener(Retry);
            Refresh();
        }

        void Start()
        {
            if (screen == null && NetBoot.Maintenance != null) Bind(NetBoot.Maintenance, LanguageBoot.Strings);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (screen == null) return;
            screen.Changed -= Refresh;
            strings.Changed -= Refresh;
            retry.onClick.RemoveListener(Retry);
            screen = null;
            strings = null;
        }

        void Retry() => screen.Retry();

        void Refresh()
        {
            panel.SetActive(screen.Active);
            retry.interactable = !screen.Checking;
            ShowEnd();
        }

        void ShowEnd()
        {
            ends.gameObject.SetActive(screen.HasEnd);
            if (!screen.HasEnd) return;
            endsArgs[EndArgument] = EndText(screen.EndsUnixMs, strings.Language);
            ends.text = strings.Text(EndsKey, endsArgs);
        }
    }
}
