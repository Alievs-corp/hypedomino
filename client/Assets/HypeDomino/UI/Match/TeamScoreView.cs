using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Match
{
    public sealed class TeamScoreView : MonoBehaviour
    {
        [SerializeField] TMP_Text score;
        [SerializeField] Graphic gem;
        [SerializeField] Image meter;
        [SerializeField] GameObject lockBadge;
        [SerializeField] RectTransform lockMark;

        bool hasScore;
        uint shownScore;
        float shownMark = -1f;

        public void Configure(TMP_Text scoreText, Graphic teamGem, Image meterFill, GameObject badge, RectTransform mark = null)
        {
            score = scoreText;
            gem = teamGem;
            meter = meterFill;
            lockBadge = badge;
            lockMark = mark;
            hasScore = false;
            shownMark = -1f;
        }

        public void ShowLockMark(float at)
        {
            if (lockMark == null || at == shownMark) return;
            shownMark = at;
            var visible = at > 0f;
            if (lockMark.gameObject.activeSelf != visible) lockMark.gameObject.SetActive(visible);
            if (!visible) return;
            var along = meter.fillOrigin == 0 ? at : 1f - at;
            var vertical = meter.fillMethod == Image.FillMethod.Vertical;
            lockMark.anchorMin = vertical ? new Vector2(0f, along) : new Vector2(along, 0f);
            lockMark.anchorMax = vertical ? new Vector2(1f, along) : new Vector2(along, 1f);
        }

        public void Show(Color32 team, uint points, float fill, bool locked)
        {
            gem.color = team;
            meter.color = team;
            meter.fillAmount = fill;
            score.color = TokenColor.RoleScoreText;
            if (lockBadge.activeSelf != locked) lockBadge.SetActive(locked);
            if (hasScore && shownScore == points) return;
            hasScore = true;
            shownScore = points;
            score.SetText("{0}", points);
        }
    }
}
