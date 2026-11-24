using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;

namespace HypeDomino.UI.Screens
{
    public sealed class PlainLabel : MonoBehaviour
    {
        TMP_Text text;
        GameStrings strings;

        public string Key { get; private set; }

        public void Configure(TMP_Text labelText, string labelKey, GameStrings gameStrings)
        {
            Unbind();
            text = labelText;
            Key = labelKey;
            strings = gameStrings;
            strings.Changed += Show;
            Show();
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            strings.Changed -= Show;
            strings = null;
        }

        void Show() => text.text = strings.Text(Key);
    }
}
