using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Ranking
{
    public sealed class SeasonHistoryRowSlot : MonoBehaviour
    {
        [SerializeField] TMP_Text seasonName;
        [SerializeField] Image badge;
        [SerializeField] TMP_Text lp;
        [SerializeField] TMP_Text position;
        [SerializeField] TMP_Text reward;

        public void Configure(TMP_Text seasonNameText, Image badgeImage, TMP_Text lpText, TMP_Text positionText, TMP_Text rewardText)
        {
            seasonName = seasonNameText;
            badge = badgeImage;
            lp = lpText;
            position = positionText;
            reward = rewardText;
        }

        public void Show(string seasonLabel, RankBadge rankBadge, Sprite[] tierBadges, string lpLabel, string positionLabel, string rewardLabel)
        {
            seasonName.text = seasonLabel;
            RankBadgeArt.Show(badge, tierBadges, rankBadge);
            lp.text = lpLabel;
            position.text = positionLabel;
            reward.text = rewardLabel;
        }
    }
}
