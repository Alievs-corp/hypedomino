using System;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.UI;

namespace HypeDomino.Game.Inventory
{
    public sealed class CollectionScreenView : MonoBehaviour, IRowBinder
    {
        const int NoSelection = -1;

        [SerializeField] Button[] filters;
        [SerializeField] string[] filterTypes;
        [SerializeField] RecyclingList grid;
        [SerializeField] CollectionRowSlot rowTemplate;
        [SerializeField] TMP_Text message;
        [SerializeField] GameObject preview;
        [SerializeField] CosmeticImage previewArt;
        [SerializeField] TMP_Text previewName;
        [SerializeField] GameObject previewLocked;
        [SerializeField] GameObject previewEquipped;
        [SerializeField] Button equip;
        [SerializeField] Button closePreview;

        Action<InventoryLoadResult> loaded;
        Action<EquipResult> equipped;
        Action<int> pick;
        UnityAction[] filterPicks;
        InventoryLoader loader;
        LoadoutEquipper equipper;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        CosmeticCollection collection;
        InventoryReply inventory;
        string selectedId;
        string loadMessageKey;
        string equipMessageKey;
        int filterAt;
        bool built;

        public void Configure(Button[] filterButtons, string[] filterItemTypes, RecyclingList rows, CollectionRowSlot template, TMP_Text messageLine)
        {
            Unbind();
            filters = filterButtons;
            filterTypes = filterItemTypes;
            grid = rows;
            rowTemplate = template;
            message = messageLine;
            filterPicks = null;
        }

        public void ConfigurePreview(GameObject panel, CosmeticImage art, TMP_Text title, GameObject lockedRoot, GameObject equippedRoot, Button equipButton, Button closeButton)
        {
            Unbind();
            preview = panel;
            previewArt = art;
            previewName = title;
            previewLocked = lockedRoot;
            previewEquipped = equippedRoot;
            equip = equipButton;
            closePreview = closeButton;
        }

        public void Bind(InventoryLoader inventoryLoader, LoadoutEquipper loadoutEquipper, CosmeticStage cosmetics, ActiveCosmeticManifest cosmeticManifest, GameStrings texts)
        {
            Unbind();
            loader = inventoryLoader;
            equipper = loadoutEquipper;
            stage = cosmetics;
            manifest = cosmeticManifest;
            strings = texts;
            inventory = null;
            selectedId = null;
            loadMessageKey = null;
            equipMessageKey = null;
            filterAt = 0;
            loaded ??= OnLoaded;
            equipped ??= OnEquipped;
            pick ??= OnPick;
            filterPicks ??= FilterPicks(filters.Length);
            for (var i = 0; i < filters.Length; i++) filters[i].onClick.AddListener(filterPicks[i]);
            equip.onClick.AddListener(OnEquip);
            closePreview.onClick.AddListener(OnClosePreview);
            strings.Changed += Redraw;
            manifest.Changed += OnManifestChanged;
            BuildGrid();
            Rebuild();
            loader.Load(loaded);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (strings == null) return;
            for (var i = 0; i < filters.Length; i++) filters[i].onClick.RemoveListener(filterPicks[i]);
            equip.onClick.RemoveListener(OnEquip);
            closePreview.onClick.RemoveListener(OnClosePreview);
            strings.Changed -= Redraw;
            manifest.Changed -= OnManifestChanged;
            strings = null;
            manifest = null;
            loader = null;
            equipper = null;
        }

        UnityAction[] FilterPicks(int count)
        {
            var picks = new UnityAction[count];
            for (var i = 0; i < count; i++)
            {
                var index = i;
                picks[i] = () => OnFilter(index);
            }
            return picks;
        }

        void BuildGrid()
        {
            if (built) return;
            var template = (RectTransform)rowTemplate.transform;
            grid.Build(template, template.rect.height, TokenSpacing.Sm);
            built = true;
        }

        void Rebuild()
        {
            collection = new CosmeticCollection(manifest.Current);
            collection.Show(inventory);
            collection.Filter(filterAt < filterTypes.Length ? filterTypes[filterAt] : null, null);
            Redraw();
        }

        void OnManifestChanged() => Rebuild();

        void OnLoaded(InventoryLoadResult result)
        {
            if (strings == null) return;
            loadMessageKey = result.MessageKey;
            if (result.Outcome == InventoryLoadOutcome.Loaded)
            {
                inventory = result.Inventory;
                collection.Show(inventory);
            }
            Redraw();
        }

        void OnFilter(int index)
        {
            filterAt = index;
            selectedId = null;
            equipMessageKey = null;
            collection.Filter(filterTypes[index], null);
            Redraw();
        }

        void OnPick(int index)
        {
            selectedId = collection[index].Item.Id;
            equipMessageKey = null;
            ShowPreview();
            ShowMessage();
        }

        void OnClosePreview()
        {
            selectedId = null;
            ShowPreview();
        }

        void OnEquip()
        {
            if (!collection.TryEquip(SelectedIndex, out var request)) return;
            if (!equipper.Equip(request, equipped)) return;
            equipMessageKey = null;
            ShowPreview();
            ShowMessage();
        }

        void OnEquipped(EquipResult result)
        {
            if (strings == null) return;
            equipMessageKey = result.MessageKey;
            if (result.Outcome == EquipOutcome.Equipped) collection.Apply(result.Equipped);
            if (result.InventoryStale) loader.Load(loaded);
            Redraw();
        }

        int SelectedIndex
        {
            get
            {
                if (selectedId == null) return NoSelection;
                for (var at = 0; at < collection.Count; at++)
                    if (collection[at].Item.Id == selectedId) return at;
                return NoSelection;
            }
        }

        void Redraw()
        {
            ShowFilters();
            var columns = rowTemplate.Columns;
            grid.Show((ShownCount() + columns - 1) / columns, this);
            ShowPreview();
            ShowMessage();
        }

        int ShownCount() => inventory == null ? 0 : collection.Count;

        void ShowFilters()
        {
            for (var i = 0; i < filters.Length; i++) filters[i].interactable = i != filterAt;
        }

        public void Bind(RectTransform row, int index)
        {
            var slot = row.GetComponent<CollectionRowSlot>();
            var count = ShownCount();
            for (var column = 0; column < slot.Columns; column++)
            {
                var at = index * slot.Columns + column;
                var cell = slot.CellAt(column);
                if (at >= count)
                {
                    cell.Hide();
                    continue;
                }
                var entry = collection[at];
                cell.Show(stage, entry, strings.Text(entry.Item.NameKey), at, pick);
            }
        }

        void ShowPreview()
        {
            var index = inventory == null ? NoSelection : SelectedIndex;
            preview.SetActive(index != NoSelection);
            if (index == NoSelection) return;
            var entry = collection[index];
            previewArt.Show(stage, entry.Item.Art, null);
            previewName.text = strings.Text(entry.Item.NameKey);
            CollectionCellSlot.ShowLooks(previewLocked, previewEquipped, entry);
            equip.interactable = !equipper.Busy && collection.TryEquip(index, out _);
        }

        void ShowMessage()
        {
            var key = equipMessageKey ?? loadMessageKey;
            message.enabled = key != null;
            message.text = key == null ? string.Empty : strings.Text(key);
        }
    }
}
