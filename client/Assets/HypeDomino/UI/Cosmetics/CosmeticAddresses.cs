namespace HypeDomino.UI.Cosmetics
{
    public static class CosmeticAddresses
    {
        public const string Avatar = "avatar";
        public const string Frame = "frame";
        public const string Tile = "tile";
        public const string Table = "table";

        const string Root = "cosmetic/";
        const string StaticTail = "/static";

        public static string Static(string type, string itemId)
        {
            if (string.IsNullOrEmpty(itemId) || itemId.Length <= type.Length + 1) return null;
            if (!itemId.StartsWith(type, System.StringComparison.Ordinal) || itemId[type.Length] != '.') return null;
            var slug = itemId.Substring(type.Length + 1);
            if (slug.IndexOf('.') >= 0 || slug.IndexOf('/') >= 0) return null;
            return Root + type + "/" + slug + StaticTail;
        }
    }
}
