using System;
using HypeDomino.Game.Friends;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class FriendInviteRow : MonoBehaviour
    {
        [SerializeField] TMP_Text nickname;
        [SerializeField] Button invite;
        [SerializeField] GameObject[] stateMarks;

        Friend friend;
        Action<Friend> invited;
        bool listening;

        public void Configure(TMP_Text nicknameText, Button inviteButton, GameObject[] stateMarkers)
        {
            StopListening();
            nickname = nicknameText;
            invite = inviteButton;
            stateMarks = stateMarkers;
        }

        public void Show(Friend shown, FriendInviteState state, Action<Friend> onInvite)
        {
            friend = shown;
            invited = onInvite;
            if (!listening)
            {
                invite.onClick.AddListener(OnInvite);
                listening = true;
            }
            nickname.text = shown.Nickname;
            invite.interactable = state == FriendInviteState.Invitable;
            for (var at = 0; at < stateMarks.Length; at++) stateMarks[at].SetActive(at == (int)state);
        }

        void OnInvite() => invited?.Invoke(friend);

        void StopListening()
        {
            if (!listening) return;
            invite.onClick.RemoveListener(OnInvite);
            listening = false;
        }

        void OnDestroy() => StopListening();
    }
}
