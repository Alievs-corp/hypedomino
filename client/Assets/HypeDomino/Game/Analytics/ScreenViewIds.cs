using System.Text;

namespace HypeDomino.Game.Analytics
{
    public static class ScreenViewIds
    {
        const char WordSeparator = '_';

        public static string Of(string screenName)
        {
            var id = new StringBuilder(screenName.Length * 2);
            for (var at = 0; at < screenName.Length; at++)
            {
                var letter = screenName[at];
                if (char.IsUpper(letter) && at > 0) id.Append(WordSeparator);
                id.Append(char.ToLowerInvariant(letter));
            }
            return id.ToString();
        }
    }
}
