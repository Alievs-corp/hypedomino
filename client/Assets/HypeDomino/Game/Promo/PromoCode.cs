using System.Text;

namespace HypeDomino.Game.Promo
{
    public static class PromoCode
    {
        public const string Alphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789";
        public const int MinLength = 4;
        public const int MaxLength = 32;

        public static bool TryNormalize(string typed, out string code)
        {
            code = null;
            if (typed == null)
            {
                return false;
            }
            var kept = new StringBuilder(typed.Length);
            foreach (var character in typed)
            {
                if (IsSeparator(character))
                {
                    continue;
                }
                var upper = char.ToUpperInvariant(character);
                if (Alphabet.IndexOf(upper) < 0)
                {
                    return false;
                }
                kept.Append(upper);
            }
            if (kept.Length < MinLength || kept.Length > MaxLength)
            {
                return false;
            }
            code = kept.ToString();
            return true;
        }

        static bool IsSeparator(char character) => character == ' ' || character == '-' || character == '_';
    }
}
