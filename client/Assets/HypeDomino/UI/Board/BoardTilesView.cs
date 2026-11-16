using System;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Board
{
    public sealed class BoardTilesView : MonoBehaviour
    {
        [SerializeField] TileFaceSet faceSet;
        [SerializeField] Image[] tiles = Array.Empty<Image>();

        public int Count { get; private set; }

        public int Capacity => tiles.Length;

        public void Configure(TileFaceSet faces, Image[] tileImages)
        {
            faceSet = faces;
            tiles = tileImages ?? Array.Empty<Image>();
            for (var i = 0; i < tiles.Length; i++) tiles[i].raycastTarget = false;
        }

        public void Show(ChainLayout layout)
        {
            Count = Mathf.Min(layout.Count, tiles.Length);
            for (var i = 0; i < tiles.Length; i++)
            {
                var shown = i < Count;
                tiles[i].enabled = shown;
                if (shown) Place(tiles[i], layout[i], layout.OnBoard(layout[i].Center), layout.Scale);
            }
        }

        void Place(Image image, PlacedDomino tile, Vector2 position, float shortSide)
        {
            image.sprite = faceSet.Face(new Domino(tile.Near, tile.Far));
            var rect = image.rectTransform;
            rect.anchoredPosition = position;
            rect.sizeDelta = new Vector2(shortSide, shortSide * 2f);
            rect.localRotation = Quaternion.Euler(0f, 0f, TileFaces.Angle(tile));
        }
    }
}
