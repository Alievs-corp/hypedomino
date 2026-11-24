using System;
using HypeDomino.UI.Cosmetics;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Inventory
{
    public sealed class CollectionCellSlot : MonoBehaviour
    {
        [SerializeField] CosmeticImage icon;
        [SerializeField] TMP_Text itemName;
        [SerializeField] GameObject lockedLook;
        [SerializeField] GameObject equippedLook;
        [SerializeField] Button select;

        Action<int> picked;
        int shownIndex;
        bool listening;

        public void Configure(CosmeticImage iconImage, TMP_Text nameText, GameObject lockedRoot, GameObject equippedRoot, Button selectButton)
        {
            StopListening();
            icon = iconImage;
            itemName = nameText;
            lockedLook = lockedRoot;
            equippedLook = equippedRoot;
            select = selectButton;
        }

        public void Show(CosmeticStage stage, CollectionEntry entry, string nameLabel, int index, Action<int> onPick)
        {
            if (!gameObject.activeSelf) gameObject.SetActive(true);
            shownIndex = index;
            picked = onPick;
            if (!listening)
            {
                select.onClick.AddListener(OnPick);
                listening = true;
            }
            icon.Show(stage, entry.Item.Art, null);
            itemName.text = nameLabel;
            ShowLooks(lockedLook, equippedLook, entry);
        }

        public static void ShowLooks(GameObject locked, GameObject equipped, CollectionEntry entry)
        {
            if (locked.activeSelf == entry.Held) locked.SetActive(!entry.Held);
            if (equipped.activeSelf != entry.Equipped) equipped.SetActive(entry.Equipped);
        }

        public void Hide()
        {
            if (gameObject.activeSelf) gameObject.SetActive(false);
        }

        void OnPick() => picked?.Invoke(shownIndex);

        void StopListening()
        {
            if (!listening) return;
            select.onClick.RemoveListener(OnPick);
            listening = false;
        }

        void OnDestroy() => StopListening();
    }
}
