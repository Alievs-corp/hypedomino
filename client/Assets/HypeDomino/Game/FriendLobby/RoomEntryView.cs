using HypeDomino.Net;
using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class RoomEntryView : MonoBehaviour
    {
        [SerializeField] Button create;
        [SerializeField] TMP_InputField code;
        [SerializeField] Button join;
        [SerializeField] TMP_Text message;

        FriendRoom room;
        RoomMembers members;
        LobbyCommands commands;
        Session session;
        GameStrings strings;
        string typedRefusal;

        public void Configure(Button createButton, TMP_InputField codeField, Button joinButton, TMP_Text messageLine)
        {
            Unbind();
            create = createButton;
            code = codeField;
            join = joinButton;
            message = messageLine;
        }

        public void Bind(FriendRoom friendRoom, RoomMembers roomMembers, LobbyCommands lobbyCommands, Session source, GameStrings texts)
        {
            Unbind();
            room = friendRoom;
            members = roomMembers;
            commands = lobbyCommands;
            session = source;
            strings = texts;
            typedRefusal = null;
            create.onClick.AddListener(OnCreate);
            join.onClick.AddListener(OnJoin);
            members.Changed += OnRoomRead;
            strings.Changed += Show;
            Show();
        }

        void Start()
        {
            if (room == null && FriendLobbyBoot.Link != null)
                Bind(NetBoot.Room, FriendLobbyBoot.Link.Members, NetBoot.Lobby, NetBoot.Connection.Session, LanguageBoot.Strings);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (room == null) return;
            create.onClick.RemoveListener(OnCreate);
            join.onClick.RemoveListener(OnJoin);
            members.Changed -= OnRoomRead;
            strings.Changed -= Show;
            room = null;
            members = null;
            commands = null;
            session = null;
            strings = null;
        }

        void OnCreate()
        {
            typedRefusal = null;
            commands.TryCreate(session);
            Show();
        }

        void OnJoin()
        {
            typedRefusal = RoomCode.TryNormalize(code.text, out _) ? null : LobbyRefusalKeys.UnknownCode;
            if (typedRefusal == null) commands.TryJoin(session, code.text);
            Show();
        }

        void OnRoomRead()
        {
            typedRefusal = null;
            Show();
        }

        void Show()
        {
            var key = typedRefusal ?? (room.InRoom ? null : room.RefusalKey);
            message.enabled = key != null;
            message.text = key == null ? string.Empty : strings.Text(key);
        }
    }
}
