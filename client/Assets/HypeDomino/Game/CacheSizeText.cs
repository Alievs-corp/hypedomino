using System;
using System.Collections.Generic;
using HypeDomino.UI.Localization;

namespace HypeDomino.Game
{
    public static class CacheSizeText
    {
        public const string Key = "settings.cache.size";
        public const string MegabytesArgument = "megabytes";
        const decimal BytesPerMegabyte = 1_000_000m;
        const int ShownDecimals = 1;
        const decimal OneDecimalZero = 0.0m;

        public static decimal Megabytes(long bytes) =>
            decimal.Round(bytes / BytesPerMegabyte, ShownDecimals, MidpointRounding.AwayFromZero) + OneDecimalZero;

        public static string Of(GameStrings strings, long bytes) =>
            strings.Text(Key, new Dictionary<string, object> { [MegabytesArgument] = Megabytes(bytes) });
    }
}
