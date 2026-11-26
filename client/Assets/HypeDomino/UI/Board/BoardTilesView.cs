using System;
using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Board
{
    public sealed class BoardTilesView : MonoBehaviour
    {
        [SerializeField] TileFaceSet faceSet;
        [SerializeField] Image[] tiles = Array.Empty<Image>();
        [SerializeField] Sprite shadowSprite;
        [SerializeField] Image[] shadows = Array.Empty<Image>();

        bool[] laid = Array.Empty<bool>();

        public int Count { get; private set; }

        public int Capacity => tiles.Length;

        public RectTransform TileAt(int slot) => tiles[slot].rectTransform;

        public RectTransform ShadowAt(int slot) => shadows[slot].rectTransform;

        public static int SlotOf(PlacedDomino tile) => TileFaces.Index(new Domino(tile.Near, tile.Far));

        public void Configure(TileFaceSet faces, Image[] tileImages, Sprite shadow, Image[] shadowImages)
        {
            faceSet = faces;
            tiles = tileImages ?? Array.Empty<Image>();
            shadowSprite = shadow;
            shadows = shadowImages ?? Array.Empty<Image>();
            if (shadows.Length != tiles.Length) throw new ArgumentException($"{shadows.Length} shadows for {tiles.Length} tiles", nameof(shadowImages));
            laid = new bool[tiles.Length];
            Color restingShadow = TokenColor.Shadow;
            restingShadow.a = TokenElevation.E1Alpha;
            for (var i = 0; i < tiles.Length; i++)
            {
                tiles[i].raycastTarget = false;
                shadows[i].raycastTarget = false;
                shadows[i].sprite = shadowSprite;
                shadows[i].color = restingShadow;
            }
        }

        public void Show(ChainLayout layout)
        {
            if (laid.Length != tiles.Length) laid = new bool[tiles.Length];
            Array.Clear(laid, 0, laid.Length);
            Count = 0;
            for (var i = 0; i < layout.Count; i++)
            {
                var slot = SlotOf(layout[i]);
                if (slot >= tiles.Length) continue;
                laid[slot] = true;
                Count++;
                Place(slot, layout[i], layout.OnBoard(layout[i].Center), layout.Scale);
            }
            for (var slot = 0; slot < tiles.Length; slot++)
            {
                Enable(tiles[slot], laid[slot]);
                Enable(shadows[slot], laid[slot]);
            }
        }

        static void Enable(Behaviour shown, bool enabled)
        {
            if (shown.enabled != enabled) shown.enabled = enabled;
        }

        void Place(int slot, PlacedDomino tile, Vector2 position, float shortSide)
        {
            var face = faceSet.Face(new Domino(tile.Near, tile.Far));
            var rotation = Quaternion.Euler(0f, 0f, TileFaces.Angle(tile));
            var image = tiles[slot];
            image.sprite = face;
            var rect = image.rectTransform;
            rect.anchoredPosition = position;
            rect.sizeDelta = new Vector2(shortSide, shortSide * 2f);
            rect.localRotation = rotation;
            var shadow = shadows[slot].rectTransform;
            shadow.anchoredPosition = new Vector2(position.x, position.y - TokenElevation.E1OffsetY);
            shadow.sizeDelta = shadowSprite.rect.size * (shortSide / face.rect.width);
            shadow.localRotation = rotation;
        }
    }
}
