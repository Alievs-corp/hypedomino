using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Profile;
using TMPro;
using UnityEngine;

namespace HypeDomino.Game.Profile
{
    public sealed class ProfileSummaryView : MonoBehaviour
    {
        [SerializeField] AvatarView avatar;
        [SerializeField] TMP_Text nickname;
        [SerializeField] TMP_Text level;
        [SerializeField] TMP_Text[] statValues;

        public void Configure(AvatarView avatarView, TMP_Text nicknameText, TMP_Text levelText, TMP_Text[] statValueTexts)
        {
            avatar = avatarView;
            nickname = nicknameText;
            level = levelText;
            statValues = statValueTexts;
        }

        public void Show(PlayerProfile profile, CosmeticStage stage, CosmeticManifest manifest, GameStrings strings)
        {
            var shown = profile != null;
            avatar.gameObject.SetActive(shown);
            nickname.enabled = shown;
            level.enabled = shown;
            for (var at = 0; at < statValues.Length; at++) statValues[at].enabled = false;
            if (!shown) return;
            avatar.Show(stage, profile.Look(manifest));
            nickname.text = profile.Nickname;
            level.text = MessageFormat.Number(profile.Level, strings.Language);
            for (var at = 0; at < profile.StatCount; at++) ShowStat(profile.Stat(at), strings.Language);
        }

        void ShowStat(ProfileStatValue stat, string language)
        {
            var at = (int)stat.Stat;
            if (at >= statValues.Length) return;
            statValues[at].enabled = true;
            statValues[at].text = MessageFormat.Number(stat.Value, language);
        }
    }
}
