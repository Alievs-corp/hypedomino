using System;
using System.Collections.Generic;
using HypeDomino.UI.Localization;

namespace HypeDomino.Game
{
    public static class CacheSizeText
    {
        public const string BelowOneMegabyteKey = "settings.cache.size_below_one_megabyte";
        public const string MegabytesKey = "settings.cache.size";
        public const string GigabytesKey = "settings.cache.size_gigabytes";
        public const string MegabytesArgument = "megabytes";
        public const string GigabytesArgument = "gigabytes";
        const decimal BytesPerMegabyte = 1_000_000m;
        const decimal BytesPerGigabyte = 1_000_000_000m;
        const decimal MegabytesPerGigabyte = 1_000m;
        const int ShownDecimals = 1;
        const decimal OneDecimalZero = 0.0m;

        public static string Of(GameStrings strings, long bytes)
        {
            if (bytes < BytesPerMegabyte) return strings.Text(BelowOneMegabyteKey);
            var megabytes = OneDecimal(bytes / BytesPerMegabyte);
            if (megabytes < MegabytesPerGigabyte) return strings.Text(MegabytesKey, Argument(MegabytesArgument, megabytes));
            return strings.Text(GigabytesKey, Argument(GigabytesArgument, OneDecimal(bytes / BytesPerGigabyte)));
        }

        static decimal OneDecimal(decimal amount) =>
            decimal.Round(amount, ShownDecimals, MidpointRounding.AwayFromZero) + OneDecimalZero;

        static Dictionary<string, object> Argument(string name, decimal amount) =>
            new Dictionary<string, object> { [name] = amount };
    }
}
