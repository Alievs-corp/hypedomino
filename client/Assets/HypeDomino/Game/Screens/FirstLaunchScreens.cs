using HypeDomino.Game.Privacy;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;

namespace HypeDomino.Game.Screens
{
    public static class FirstLaunchScreens
    {
        public const string NoticeKey = "ui.plain.privacy.notice";
        public const string PolicyKey = "ui.plain.privacy.policy";
        public const string TermsKey = "ui.plain.privacy.terms";
        public const string ContinueKey = "ui.plain.privacy.continue";

        public static ScreenView PrivacyNotice(ScreenHost host)
        {
            var screen = host.Screen(nameof(PrivacyNotice), out var column);
            PlainUi.Text(column, TokenTextStyle.Body, NoticeKey, host.Strings);
            var policy = PlainUi.Button(column, PolicyKey, host.Strings);
            var terms = PlainUi.Button(column, TermsKey, host.Strings);
            var proceed = PlainUi.Button(column, ContinueKey, host.Strings);
            screen.gameObject.AddComponent<PrivacyNoticeView>().Configure(policy, terms, proceed);
            return screen;
        }
    }
}
