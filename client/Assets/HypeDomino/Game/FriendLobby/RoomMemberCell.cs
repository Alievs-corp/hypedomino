using HypeDomino.UI.Cosmetics;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class RoomMemberCell : MonoBehaviour
    {
        [SerializeField] AvatarView avatar;
        [SerializeField] TMP_Text nickname;
        [SerializeField] GameObject leaderMark;
        [SerializeField] GameObject selectedMark;
        [SerializeField] Button tap;
        [SerializeField] CosmeticImage emoji;

        public Button Tap => tap;

        public void ConfigureEmoji(CosmeticImage emojiArt) => emoji = emojiArt;

        public void ShowEmoji(CosmeticStage stage, CosmeticArt art)
        {
            emoji.gameObject.SetActive(true);
            emoji.Show(stage, art, null);
        }

        public void HideEmoji()
        {
            if (emoji.gameObject.activeSelf) emoji.gameObject.SetActive(false);
        }

        public void Configure(AvatarView avatarView, TMP_Text nicknameText, GameObject leaderMarker, GameObject selectedMarker, Button tapButton)
        {
            avatar = avatarView;
            nickname = nicknameText;
            leaderMark = leaderMarker;
            selectedMark = selectedMarker;
            tap = tapButton;
        }

        public void Show(RoomMember member, CosmeticStage stage, CosmeticManifest manifest, bool selected)
        {
            gameObject.SetActive(true);
            avatar.gameObject.SetActive(member.ProfileLoaded);
            if (member.ProfileLoaded) avatar.Show(stage, member.Profile.Look(manifest));
            nickname.enabled = member.ProfileLoaded;
            nickname.text = member.Nickname ?? string.Empty;
            leaderMark.SetActive(member.IsLeader);
            selectedMark.SetActive(selected);
        }

        public void ShowFree()
        {
            gameObject.SetActive(true);
            avatar.gameObject.SetActive(false);
            nickname.enabled = false;
            leaderMark.SetActive(false);
            selectedMark.SetActive(false);
        }

        public void Hide() => gameObject.SetActive(false);
    }
}
