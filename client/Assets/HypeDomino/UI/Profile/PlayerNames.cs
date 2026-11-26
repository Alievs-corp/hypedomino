using HypeDomino.UI.Localization;

namespace HypeDomino.UI.Profile
{
    public static class PlayerNames
    {
        public const string PlaceholderKey = "player.name.placeholder";

        public static string Placeholder(GameStrings strings) => strings.Text(PlaceholderKey);

        public static string Shown(string nickname, string placeholder) => string.IsNullOrEmpty(nickname) ? placeholder : nickname;
    }
}
