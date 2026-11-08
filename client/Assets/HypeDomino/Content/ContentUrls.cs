using System;

namespace HypeDomino.Content
{
    public static class ContentUrls
    {
        public const string RemotePrefix = "hd-content:/";
        const string AddressablesFolder = "/addressables/";

        public static string Catalog(string baseUrl, string platform, string version) =>
            baseUrl + AddressablesFolder + platform + "/" + CatalogFile(version);

        public static string CatalogFile(string version) => "catalog_" + version + ".bin";

        public static string Resolve(string internalId, string baseUrl)
        {
            if (string.IsNullOrEmpty(baseUrl) || !internalId.StartsWith(RemotePrefix, StringComparison.Ordinal)) return internalId;
            return baseUrl + AddressablesFolder + internalId.Substring(RemotePrefix.Length);
        }
    }
}
