using System;
using System.Collections.Generic;
using UnityEngine;

namespace HypeDomino.UI.Localization
{
    public sealed class BuiltinStrings
    {
        [Serializable]
        sealed class Entry
        {
            public string key;
            public string en;
            public string az;
            public string tr;
            public string ru;
        }

        [Serializable]
        sealed class File
        {
            public Entry[] entries;
        }

        static readonly Dictionary<string, string[]> Chains = new()
        {
            ["az"] = new[] { "az", "tr", "en" },
            ["tr"] = new[] { "tr", "en" },
            ["ru"] = new[] { "ru", "en" },
            ["en"] = new[] { "en" },
        };

        readonly Dictionary<string, Entry> entries = new();

        BuiltinStrings(File file)
        {
            foreach (var entry in file.entries ?? Array.Empty<Entry>())
            {
                entries[entry.key] = entry;
            }
        }

        public static BuiltinStrings Parse(string json) => new(JsonUtility.FromJson<File>(json));

        public bool TryGet(string key, string language, out string text)
        {
            text = null;
            if (!entries.TryGetValue(key, out var entry) || !Chains.TryGetValue(language, out var chain))
            {
                return false;
            }
            foreach (var step in chain)
            {
                text = ValueIn(entry, step);
                if (!string.IsNullOrEmpty(text))
                {
                    return true;
                }
            }
            return false;
        }

        static string ValueIn(Entry entry, string language) => language switch
        {
            "az" => entry.az,
            "tr" => entry.tr,
            "ru" => entry.ru,
            _ => entry.en,
        };
    }
}
