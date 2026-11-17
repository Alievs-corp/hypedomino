using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Match
{
    public sealed class LoadingScreenView : MonoBehaviour
    {
        public const int Rows = 2;
        public const int CardsPerRow = 2;

        [SerializeField] LoadingCardView[] cards;
        [SerializeField] Image total;

        public int CardCount => cards.Length;

        public void Configure(LoadingCardView[] rowMajorCards, Image totalFill)
        {
            cards = rowMajorCards;
            total = totalFill;
        }

        public LoadingCardView Card(int row, int place) => cards[row * CardsPerRow + place];

        public LoadingCardView Card(int index) => cards[index];

        public void SetOpen(bool open)
        {
            if (gameObject.activeSelf != open) gameObject.SetActive(open);
        }

        public void ShowTotal(float fill)
        {
            total.color = TokenColor.RolePrimaryFill;
            total.fillAmount = fill;
        }
    }
}
