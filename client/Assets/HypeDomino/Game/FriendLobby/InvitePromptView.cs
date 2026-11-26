using System.Collections.Generic;
using HypeDomino.Net;
using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class InvitePromptView : MonoBehaviour
    {
        public const string FromKey = "ui.plain.lobby.invite_from";
        public const string UnnamedKey = "ui.plain.lobby.invite_received";
        public const string NicknameArgument = "nickname";

        [SerializeField] GameObject panel;
        [SerializeField] TMP_Text message;
        [SerializeField] Button accept;
        [SerializeField] Button decline;

        readonly Dictionary<string, object> messageArgs = new(1);
        InvitePrompt prompt;
        LobbyCommands commands;
        Session session;
        GameStrings strings;

        public void Configure(GameObject panelRoot, TMP_Text messageLine, Button acceptButton, Button declineButton)
        {
            Unbind();
            panel = panelRoot;
            message = messageLine;
            accept = acceptButton;
            decline = declineButton;
        }

        public void Bind(InvitePrompt invitePrompt, LobbyCommands lobbyCommands, Session source, GameStrings texts)
        {
            Unbind();
            prompt = invitePrompt;
            commands = lobbyCommands;
            session = source;
            strings = texts;
            accept.onClick.AddListener(OnAccept);
            decline.onClick.AddListener(OnDecline);
            prompt.Changed += Redraw;
            strings.Changed += Redraw;
            Redraw();
        }

        void Start()
        {
            if (prompt == null && FriendLobbyBoot.Link != null)
                Bind(FriendLobbyBoot.Link.Incoming, NetBoot.Lobby, NetBoot.Connection.Session, LanguageBoot.Strings);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (prompt == null) return;
            accept.onClick.RemoveListener(OnAccept);
            decline.onClick.RemoveListener(OnDecline);
            prompt.Changed -= Redraw;
            strings.Changed -= Redraw;
            prompt = null;
            commands = null;
            session = null;
            strings = null;
        }

        void OnAccept() => prompt.TryAccept(session, commands);

        void OnDecline() => prompt.Decline();

        void Redraw()
        {
            panel.SetActive(prompt.Shown);
            if (!prompt.Shown) return;
            accept.interactable = prompt.CanAccept;
            message.text = Message();
        }

        string Message()
        {
            if (string.IsNullOrEmpty(prompt.Nickname)) return strings.Text(UnnamedKey);
            messageArgs[NicknameArgument] = prompt.Nickname;
            return strings.Text(FromKey, messageArgs);
        }
    }
}
