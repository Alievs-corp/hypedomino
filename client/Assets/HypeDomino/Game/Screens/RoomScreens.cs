using System;
using HypeDomino.Game.FriendLobby;
using HypeDomino.Net;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Screens;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Screens
{
    public static class RoomScreens
    {
        public const string CreateKey = "ui.plain.lobby.create_room";
        public const string JoinKey = "ui.plain.lobby.join_room";
        public const string CodeKey = "ui.plain.room.code";
        public const string LeaveKey = "ui.plain.room.leave";
        public const string InviteKey = "ui.plain.room.invite";
        public const string SoloKey = "ui.plain.room.solo";
        public const string TeamsKey = "ui.plain.room.teams";
        public const string EmojiKey = "ui.plain.room.emoji";
        public const string TeamAKey = "ui.plain.room.team_a";
        public const string TeamBKey = "ui.plain.room.team_b";
        public const string WaitingKey = "ui.plain.room.waiting";
        public const string StartMatchKey = "ui.plain.room.start_match";
        public const string InviteFriendKey = "ui.plain.room.invite_friend";
        public const string AcceptInviteKey = "ui.plain.lobby.accept_invite";
        public const string DeclineInviteKey = "ui.plain.lobby.decline_invite";

        public static readonly string[] FriendStateKeys =
        {
            "ui.plain.room.friend.invitable",
            "ui.plain.room.friend.invited",
            "ui.plain.room.friend.already_in_room",
            "ui.plain.room.friend.in_match",
            "ui.plain.room.friend.offline",
            "ui.plain.room.friend.room_closed",
        };

        public static readonly int TeamSeats = SeatOrder.Size(Mode._2V2) / 2;
        public static readonly int WaitingSeats = FriendRoom.Capacity - SeatOrder.Size(Mode._1V1);

        public static ScreenView Room(ScreenHost host)
        {
            var screen = host.Screen(nameof(AppState.Room), out var column);
            var strings = host.Strings;
            PlainUi.Text(column, TokenTextStyle.Caption, CodeKey, strings);
            var code = PlainUi.Text(column, TokenTextStyle.Counter);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            var leave = PlainUi.Button(column, LeaveKey, strings);
            var invite = PlainUi.Button(column, InviteKey, strings);
            var invites = Invites(screen, column, strings);
            var lobby = PlainUi.Column(column);
            var teams = PlainUi.Column(column);
            var view = screen.gameObject.AddComponent<FriendRoomView>();
            view.Configure(lobby, teams, Members(lobby.transform), code, message, leave, invite, invites);
            LobbyPanel(view, screen.transform, lobby.transform, strings);
            TeamsPanel(view, teams.transform, strings);
            teams.SetActive(false);
            return screen;
        }

        public static void Entry(ScreenView screen, Transform column, GameStrings strings)
        {
            var create = PlainUi.Button(column, CreateKey, strings);
            PlainUi.Text(column, TokenTextStyle.Caption, CodeKey, strings);
            var code = PlainUi.Input(column);
            var join = PlainUi.Button(column, JoinKey, strings);
            var message = PlainUi.Text(column, TokenTextStyle.Body);
            screen.gameObject.AddComponent<RoomEntryView>().Configure(create, code, join, message);
        }

        public static void Invitation(ScreenView screen, Transform column, GameStrings strings)
        {
            var panel = PlainUi.Column(column);
            var at = panel.transform;
            var message = PlainUi.Text(at, TokenTextStyle.Body);
            var answers = PlainUi.Row(at).transform;
            var accept = PlainUi.Button(answers, AcceptInviteKey, strings);
            var decline = PlainUi.Button(answers, DeclineInviteKey, strings);
            panel.SetActive(false);
            screen.gameObject.AddComponent<InvitePromptView>().Configure(panel, message, accept, decline);
        }

        static void LobbyPanel(FriendRoomView view, Transform screen, Transform lobby, GameStrings strings)
        {
            var modes = PlainUi.Row(lobby).transform;
            view.ConfigureStart(PlainUi.Button(modes, SoloKey, strings), PlainUi.Button(modes, TeamsKey, strings));
            var emoji = PlainUi.Column(lobby).AddComponent<CanvasGroup>();
            var wheel = PlainUi.Button(emoji.transform, EmojiKey, strings);
            var slots = PlainUi.Row(emoji.transform);
            view.ConfigureEmoji(emoji, wheel, slots, Array.Empty<Button>(), Array.Empty<CosmeticImage>(), WheelSlot(screen));
        }

        static Button WheelSlot(Transform parent)
        {
            var slot = PlainUi.Button(parent);
            var art = (RectTransform)PlainUi.Icon(slot.transform).transform;
            art.anchorMin = Vector2.zero;
            art.anchorMax = Vector2.one;
            art.offsetMin = Vector2.zero;
            art.offsetMax = Vector2.zero;
            slot.gameObject.SetActive(false);
            return slot;
        }

        static void TeamsPanel(FriendRoomView view, Transform teams, GameStrings strings)
        {
            var teamA = Seats(teams, TeamAKey, TeamSeats, strings);
            var teamB = Seats(teams, TeamBKey, TeamSeats, strings);
            var waiting = Seats(teams, WaitingKey, WaitingSeats, strings);
            var start = PlainUi.Button(teams, StartMatchKey, strings);
            var back = PlainUi.Button(teams, ScreenHost.BackKey, strings);
            view.ConfigureTeams(teamA, teamB, waiting, start, back);
        }

        static RoomMemberCell[] Members(Transform lobby)
        {
            var cells = new RoomMemberCell[FriendRoom.Capacity];
            for (var at = 0; at < cells.Length; at++)
            {
                cells[at] = Cell(lobby);
                var emoji = Art(cells[at].transform, "Emoji");
                emoji.gameObject.SetActive(false);
                cells[at].ConfigureEmoji(emoji);
            }
            return cells;
        }

        static RoomMemberCell[] Seats(Transform parent, string titleKey, int count, GameStrings strings)
        {
            PlainUi.Text(parent, TokenTextStyle.Label, titleKey, strings);
            var cells = new RoomMemberCell[count];
            for (var at = 0; at < count; at++) cells[at] = Cell(parent);
            return cells;
        }

        static RoomMemberCell Cell(Transform parent)
        {
            var row = PlainUi.Row(parent);
            var fill = row.AddComponent<Image>();
            fill.color = TokenColor.SurfaceDefault;
            var tap = row.AddComponent<Button>();
            tap.targetGraphic = fill;
            var at = row.transform;
            var avatar = Avatar(at);
            var nickname = PlainUi.Text(at, TokenTextStyle.Label);
            var leader = PlainUi.Mark(at);
            var selected = PlainUi.Mark(at);
            selected.GetComponent<Image>().color = TokenColor.RolePrimaryFill;
            var cell = row.AddComponent<RoomMemberCell>();
            cell.Configure(avatar, nickname, leader, selected, tap);
            return cell;
        }

        static AvatarView Avatar(Transform parent)
        {
            var holder = new GameObject("Avatar", typeof(RectTransform), typeof(LayoutElement));
            holder.transform.SetParent(parent, false);
            holder.GetComponent<LayoutElement>().preferredWidth = TokenSize.TouchTargetMin;
            var avatar = holder.AddComponent<AvatarView>();
            avatar.Configure(Art(holder.transform, "Art"), Art(holder.transform, "Frame"), null, null);
            return avatar;
        }

        static CosmeticImage Art(Transform parent, string name)
        {
            var art = new GameObject(name, typeof(RectTransform), typeof(Image), typeof(AnimatedImage), typeof(CosmeticImage));
            art.transform.SetParent(parent, false);
            var image = art.GetComponent<Image>();
            image.raycastTarget = false;
            image.enabled = false;
            return art.GetComponent<CosmeticImage>();
        }

        static InvitePanelView Invites(ScreenView screen, Transform column, GameStrings strings)
        {
            var panel = PlainUi.Column(column);
            var at = panel.transform;
            PlainUi.Text(at, TokenTextStyle.Caption, CodeKey, strings);
            var code = PlainUi.Text(at, TokenTextStyle.Counter);
            var friends = PlainUi.List(at);
            var close = PlainUi.Button(at, ShopScreens.CloseKey, strings);
            panel.SetActive(false);
            var view = screen.gameObject.AddComponent<InvitePanelView>();
            view.Configure(panel, friends, FriendRow(screen.transform, strings), code, close);
            return view;
        }

        static FriendInviteRow FriendRow(Transform parent, GameStrings strings)
        {
            var row = PlainUi.Row(parent);
            var at = row.transform;
            var nickname = PlainUi.Text(at, TokenTextStyle.Label);
            var marks = new GameObject[FriendStateKeys.Length];
            for (var state = 0; state < marks.Length; state++) marks[state] = PlainUi.Text(at, TokenTextStyle.Caption, FriendStateKeys[state], strings).gameObject;
            var invite = PlainUi.Button(at, InviteFriendKey, strings);
            var friendRow = row.AddComponent<FriendInviteRow>();
            friendRow.Configure(nickname, invite, marks);
            row.SetActive(false);
            return friendRow;
        }
    }
}
