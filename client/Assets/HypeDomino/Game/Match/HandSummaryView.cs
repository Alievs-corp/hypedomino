using System;
using HypeDomino.UI.Board;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Match
{
    public sealed class HandSummaryView : MonoBehaviour
    {
        [SerializeField] TileFaceSet faceSet;
        [SerializeField] GameObject domino;
        [SerializeField] GameObject block;
        [SerializeField] TMP_Text pipsA;
        [SerializeField] TMP_Text pipsB;
        [SerializeField] TMP_Text countedPips;
        [SerializeField] TMP_Text handPoints;
        [SerializeField] TMP_Text carriedIn;
        [SerializeField] TMP_Text carryOver;
        [SerializeField] GameObject lockBadge;
        [SerializeField] GameObject[] seatRows = Array.Empty<GameObject>();
        [SerializeField] Image[] tiles = Array.Empty<Image>();

        public void Configure(TileFaceSet faces, GameObject dominoMark, GameObject blockMark, TMP_Text teamAPips, TMP_Text teamBPips, TMP_Text counted,
            TMP_Text points, TMP_Text carried, TMP_Text kept, GameObject badge, GameObject[] rows, Image[] rowTiles)
        {
            if (rows.Length == 0 || rowTiles.Length % rows.Length != 0) throw new ArgumentException($"{rowTiles.Length} tiles for {rows.Length} rows", nameof(rowTiles));
            faceSet = faces;
            domino = dominoMark;
            block = blockMark;
            pipsA = teamAPips;
            pipsB = teamBPips;
            countedPips = counted;
            handPoints = points;
            carriedIn = carried;
            carryOver = kept;
            lockBadge = badge;
            seatRows = rows;
            tiles = rowTiles;
        }

        public void Show(HandSummary summary)
        {
            SetActive(gameObject, summary != null);
            if (summary == null) return;
            SetActive(domino, summary.Ending == HandEnding.Domino);
            SetActive(block, summary.Ending == HandEnding.Block);
            ShowPoints(pipsA, summary.PipsA, true);
            ShowPoints(pipsB, summary.PipsB, true);
            var awarded = summary.HasWinner && !summary.Locked;
            ShowPoints(countedPips, summary.CountedPips, summary.HasWinner);
            ShowPoints(handPoints, summary.HandPoints, awarded);
            ShowPoints(carriedIn, summary.CarriedIn, awarded && summary.CarriedIn > 0);
            ShowPoints(carryOver, summary.CarryOverKept, !summary.HasWinner);
            SetActive(lockBadge, summary.HasWinner && summary.Locked);
            ShowLeft(summary);
        }

        void ShowLeft(HandSummary summary)
        {
            var seats = summary.LeftSeats;
            var perRow = tiles.Length / seatRows.Length;
            for (var row = 0; row < seatRows.Length; row++)
            {
                var shown = row < seats.Length;
                SetActive(seatRows[row], shown);
                var left = shown ? summary.LeftTiles(seats[row]) : default;
                for (var slot = 0; slot < perRow; slot++)
                {
                    var image = tiles[row * perRow + slot];
                    var tileShown = slot < left.Length;
                    SetActive(image.gameObject, tileShown);
                    if (tileShown) image.sprite = faceSet.Face(new Domino((byte)left[slot].Low, (byte)left[slot].High));
                }
            }
        }

        static void ShowPoints(TMP_Text text, uint points, bool shown)
        {
            SetActive(text.gameObject, shown);
            if (!shown) return;
            text.color = TokenColor.RoleScoreText;
            text.SetText("{0}", points);
        }

        static void SetActive(GameObject target, bool active)
        {
            if (target.activeSelf != active) target.SetActive(active);
        }
    }
}
