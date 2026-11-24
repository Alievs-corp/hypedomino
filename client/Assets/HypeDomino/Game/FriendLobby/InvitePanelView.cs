using System;
using HypeDomino.Game.Friends;
using HypeDomino.UI.Lists;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class InvitePanelView : MonoBehaviour, IRowBinder
    {
        [SerializeField] GameObject panel;
        [SerializeField] RecyclingList friends;
        [SerializeField] FriendInviteRow friendRow;
        [SerializeField] TMP_Text code;
        [SerializeField] Button close;

        Action<Friend> invite;
        FriendLobbyLink link;
        bool built;

        public void Configure(GameObject panelRoot, RecyclingList friendList, FriendInviteRow friendRowTemplate, TMP_Text codeText, Button closeButton)
        {
            Unbind();
            panel = panelRoot;
            friends = friendList;
            friendRow = friendRowTemplate;
            code = codeText;
            close = closeButton;
        }

        public void Bind(FriendLobbyLink lobby)
        {
            Unbind();
            link = lobby;
            invite ??= OnInvite;
            close.onClick.AddListener(Hide);
            link.Invites.Changed += Redraw;
            BuildList();
            panel.SetActive(false);
        }

        public void Show()
        {
            panel.SetActive(true);
            Redraw();
        }

        public void Hide() => panel.SetActive(false);

        void IRowBinder.Bind(RectTransform row, int index)
        {
            var friend = link.Invites.Friends[index];
            row.GetComponent<FriendInviteRow>().Show(friend, link.Invites.StateOf(friend), invite);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (link == null) return;
            close.onClick.RemoveListener(Hide);
            link.Invites.Changed -= Redraw;
            link = null;
        }

        void BuildList()
        {
            if (built) return;
            var template = (RectTransform)friendRow.transform;
            friends.Build(template, template.rect.height, TokenSpacing.Sm);
            built = true;
        }

        void OnInvite(Friend friend) => link.TryInvite(friend);

        void Redraw()
        {
            if (!panel.activeSelf) return;
            code.text = link.Invites.Code;
            friends.Show(link.Invites.Friends.Count, this);
        }
    }
}
