using HypeDomino.Game.Privacy;
using HypeDomino.UI.Audio;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Settings;
using HypeDomino.UI.Tokens;
using UnityEngine;

namespace HypeDomino.Game.Screens
{
    public static class SettingsScreens
    {
        public const string NavKey = "ui.plain.nav.settings";
        public const string LanguageKey = "ui.plain.settings.language";
        public const string ReduceMotionKey = "ui.plain.settings.reduce_motion";
        public const string HapticsKey = "ui.plain.settings.haptics";
        public const string AutoFrameRateKey = "ui.plain.settings.auto_frame_rate";
        public const string FrameRateKey = "ui.plain.settings.frame_rate";
        public const string BatterySaverKey = "ui.plain.settings.battery_saver";
        public const string ClearCacheKey = "ui.plain.settings.clear_cache";
        public const string ExportDataKey = "ui.plain.settings.export_data";
        public const string DeleteAccountKey = "ui.plain.settings.delete_account";
        public const string DeleteWarningKey = "ui.plain.settings.delete_warning";
        public const string DeleteWithGoogleKey = "ui.plain.settings.delete_with_google";
        public const string DeleteWithAppleKey = "ui.plain.settings.delete_with_apple";
        public const string KeepAccountKey = "ui.plain.settings.keep_account";

        public static readonly string[] VolumeKeys =
        {
            "ui.plain.settings.music",
            "ui.plain.settings.sfx",
            "ui.plain.settings.ui",
        };

        public static ScreenView Settings(ScreenHost host, SettingsModel model)
        {
            var screen = host.PushedScreen("Settings", out var column);
            var list = PlainUi.Scroll(column);
            var languages = PlainUi.Dropdown(Setting(list, host, LanguageKey));
            var language = languages.gameObject.AddComponent<LanguageRow>();
            language.Configure(languages);
            var volumes = new VolumeSliderRow[VolumeKeys.Length];
            for (var group = 0; group < volumes.Length; group++) volumes[group] = Volume(list, host, (AudioGroup)group);
            var reduceMotion = Toggle(list, host, ReduceMotionKey);
            var haptics = Toggle(list, host, HapticsKey);
            var autoFrameRate = Toggle(list, host, AutoFrameRateKey);
            var frameRate = PlainUi.Dropdown(Setting(list, host, FrameRateKey));
            var batterySaver = Toggle(list, host, BatterySaverKey);
            var cacheSize = PlainUi.Text(list, TokenTextStyle.Body);
            var clearCache = PlainUi.Button(list, ClearCacheKey, host.Strings);
            var message = PlainUi.Text(list, TokenTextStyle.Body);
            var promoCode = PlainUi.Button(list, PromoScreens.NavKey, host.Strings);
            var view = screen.gameObject.AddComponent<SettingsScreenView>();
            view.Configure(language, volumes, reduceMotion, haptics, autoFrameRate, frameRate, batterySaver, cacheSize, clearCache, promoCode, message);
            view.Bind(model, host.Strings, () => host.Open(host.Promo));
            Account(screen, list, host);
            Legal(screen, list, host);
            return screen;
        }

        static void Legal(ScreenView screen, Transform list, ScreenHost host)
        {
            var policy = PlainUi.Button(list, FirstLaunchScreens.PolicyKey, host.Strings);
            var terms = PlainUi.Button(list, FirstLaunchScreens.TermsKey, host.Strings);
            screen.gameObject.AddComponent<LegalLinksView>().Configure(policy, terms);
        }

        static void Account(ScreenView screen, Transform list, ScreenHost host)
        {
            var export = PlainUi.Button(list, ExportDataKey, host.Strings);
            var delete = PlainUi.Button(list, DeleteAccountKey, host.Strings);
            var warning = PlainUi.Column(list);
            PlainUi.Text(warning.transform, TokenTextStyle.Body, DeleteWarningKey, host.Strings);
            var google = PlainUi.Button(warning.transform, DeleteWithGoogleKey, host.Strings);
            var apple = PlainUi.Button(warning.transform, DeleteWithAppleKey, host.Strings);
            var keep = PlainUi.Button(warning.transform, KeepAccountKey, host.Strings);
            warning.SetActive(false);
            var pending = PlainUi.Mark(list);
            pending.SetActive(false);
            var message = PlainUi.Text(list, TokenTextStyle.Body);
            screen.gameObject.AddComponent<AccountSettingsView>().Configure(export, delete, warning, google, apple, keep, pending, message);
        }

        static Transform Setting(Transform list, ScreenHost host, string labelKey)
        {
            var row = PlainUi.Row(list).transform;
            PlainUi.Text(row, TokenTextStyle.Label, labelKey, host.Strings);
            return row;
        }

        static VolumeSliderRow Volume(Transform list, ScreenHost host, AudioGroup group)
        {
            var slider = PlainUi.Slider(Setting(list, host, VolumeKeys[(int)group]));
            var row = slider.gameObject.AddComponent<VolumeSliderRow>();
            row.Configure(slider, group);
            return row;
        }

        static SettingToggleRow Toggle(Transform list, ScreenHost host, string labelKey)
        {
            var toggle = PlainUi.Toggle(Setting(list, host, labelKey));
            var row = toggle.gameObject.AddComponent<SettingToggleRow>();
            row.Configure(toggle);
            return row;
        }
    }
}
