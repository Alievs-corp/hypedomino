using System;
using System.Collections.Generic;
using HypeDomino.Content;
using HypeDomino.Game.Inventory;
using HypeDomino.Net;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using TMPro;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.UI;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class FriendRoomView : MonoBehaviour
    {
        [SerializeField] GameObject lobbyPanel;
        [SerializeField] GameObject teamsPanel;
        [SerializeField] RoomMemberCell[] members;
        [SerializeField] RoomMemberCell[] teamASeats;
        [SerializeField] RoomMemberCell[] teamBSeats;
        [SerializeField] RoomMemberCell[] waiting;
        [SerializeField] TMP_Text code;
        [SerializeField] TMP_Text message;
        [SerializeField] Button leave;
        [SerializeField] Button invite;
        [SerializeField] InvitePanelView invitePanel;
        [SerializeField] Button startOneVsOne;
        [SerializeField] Button startTwoVsTwo;
        [SerializeField] Button startMatch;
        [SerializeField] Button back;
        [SerializeField] CanvasGroup emojiRoot;
        [SerializeField] Button wheel;
        [SerializeField] GameObject wheelPanel;
        [SerializeField] Button[] wheelSlots;
        [SerializeField] CosmeticImage[] wheelArt;
        [SerializeField] Button slotTemplate;

        UnityAction[] teamATaps;
        UnityAction[] teamBTaps;
        UnityAction[] waitingTaps;
        UnityAction[] slotTaps;
        Action<MemberEmoji> onEmoji;
        Action<InventoryLoadResult> onInventory;
        string[] emojiAccounts;
        long[] emojiEndsAtMs;
        bool shownCalm;
        FriendRoom room;
        FriendLobbyLink link;
        LobbyCommands commands;
        Session session;
        CosmeticStage stage;
        ActiveCosmeticManifest manifest;
        GameStrings strings;
        string selected;

        public void Configure(GameObject lobbyRoot, GameObject teamsRoot, RoomMemberCell[] memberCells, TMP_Text codeText, TMP_Text messageLine,
            Button leaveButton, Button inviteButton, InvitePanelView invites)
        {
            Unbind();
            lobbyPanel = lobbyRoot;
            teamsPanel = teamsRoot;
            members = memberCells;
            code = codeText;
            message = messageLine;
            leave = leaveButton;
            invite = inviteButton;
            invitePanel = invites;
        }

        public void ConfigureStart(Button oneVsOneButton, Button twoVsTwoButton)
        {
            Unbind();
            startOneVsOne = oneVsOneButton;
            startTwoVsTwo = twoVsTwoButton;
        }

        public void ConfigureTeams(RoomMemberCell[] teamACells, RoomMemberCell[] teamBCells, RoomMemberCell[] waitingCells, Button startMatchButton, Button backButton)
        {
            Unbind();
            teamASeats = teamACells;
            teamBSeats = teamBCells;
            waiting = waitingCells;
            startMatch = startMatchButton;
            back = backButton;
            teamATaps = null;
            teamBTaps = null;
            waitingTaps = null;
        }

        public void ConfigureEmoji(CanvasGroup root, Button wheelButton, GameObject panel, Button[] slotButtons, CosmeticImage[] slotArt, Button slotButtonTemplate)
        {
            Unbind();
            emojiRoot = root;
            wheel = wheelButton;
            wheelPanel = panel;
            wheelSlots = slotButtons;
            wheelArt = slotArt;
            slotTemplate = slotButtonTemplate;
            slotTaps = null;
        }

        public void Bind(FriendRoom friendRoom, FriendLobbyLink lobby, LobbyCommands lobbyCommands, Session source, CosmeticStage cosmetics, ActiveCosmeticManifest cosmeticManifest, GameStrings texts)
        {
            Unbind();
            room = friendRoom;
            link = lobby;
            commands = lobbyCommands;
            session = source;
            stage = cosmetics;
            manifest = cosmeticManifest;
            strings = texts;
            selected = null;
            teamATaps ??= Taps(teamASeats.Length, OnTeamASeat);
            teamBTaps ??= Taps(teamBSeats.Length, OnTeamBSeat);
            waitingTaps ??= Taps(waiting.Length, OnWaiting);
            slotTaps ??= Taps(wheelSlots.Length, OnWheelSlot);
            onEmoji ??= OnEmojiShown;
            onInventory ??= OnInventoryLoaded;
            emojiAccounts = new string[members.Length];
            emojiEndsAtMs = new long[members.Length];
            Listen(teamASeats, teamATaps);
            Listen(teamBSeats, teamBTaps);
            Listen(waiting, waitingTaps);
            for (var at = 0; at < wheelSlots.Length; at++) wheelSlots[at].onClick.AddListener(slotTaps[at]);
            wheel.onClick.AddListener(OnWheel);
            link.EmojiShown += onEmoji;
            link.InventoryLoaded += onInventory;
            HideEmoji();
            shownCalm = false;
            emojiRoot.alpha = 1f;
            leave.onClick.AddListener(OnLeave);
            invite.onClick.AddListener(invitePanel.Show);
            startOneVsOne.onClick.AddListener(OnStartOneVsOne);
            startTwoVsTwo.onClick.AddListener(OnStartTwoVsTwo);
            startMatch.onClick.AddListener(OnStartMatch);
            back.onClick.AddListener(OnBack);
            link.Members.Changed += Redraw;
            strings.Changed += Redraw;
            manifest.Changed += Redraw;
            invitePanel.Bind(link);
            Redraw();
        }

        void Start()
        {
            if (room == null && FriendLobbyBoot.Link != null)
                Bind(NetBoot.Room, FriendLobbyBoot.Link, NetBoot.Lobby, NetBoot.Connection.Session, StillsStage(), ContentBoot.Cosmetics, LanguageBoot.Strings);
        }

        static CosmeticStage StillsStage() =>
            new(new AddressablesArtSource<Sprite>(), new StillsOnlyArtSource(), null, null, null, GraphicsBoot.Quality.Effects, MotionBoot.Motion.ReduceMotion);

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (room == null) return;
            Forget(teamASeats, teamATaps);
            Forget(teamBSeats, teamBTaps);
            Forget(waiting, waitingTaps);
            for (var at = 0; at < wheelSlots.Length; at++) wheelSlots[at].onClick.RemoveListener(slotTaps[at]);
            wheel.onClick.RemoveListener(OnWheel);
            link.EmojiShown -= onEmoji;
            link.InventoryLoaded -= onInventory;
            leave.onClick.RemoveListener(OnLeave);
            invite.onClick.RemoveListener(invitePanel.Show);
            startOneVsOne.onClick.RemoveListener(OnStartOneVsOne);
            startTwoVsTwo.onClick.RemoveListener(OnStartTwoVsTwo);
            startMatch.onClick.RemoveListener(OnStartMatch);
            back.onClick.RemoveListener(OnBack);
            link.Members.Changed -= Redraw;
            strings.Changed -= Redraw;
            manifest.Changed -= Redraw;
            room = null;
            link = null;
            commands = null;
            session = null;
        }

        static UnityAction[] Taps(int count, Action<int> pick)
        {
            var taps = new UnityAction[count];
            for (var at = 0; at < count; at++) taps[at] = Tap(pick, at);
            return taps;
        }

        static UnityAction Tap(Action<int> pick, int index) => () => pick(index);

        static void Listen(RoomMemberCell[] cells, UnityAction[] taps)
        {
            for (var at = 0; at < cells.Length; at++) cells[at].Tap.onClick.AddListener(taps[at]);
        }

        static void Forget(RoomMemberCell[] cells, UnityAction[] taps)
        {
            for (var at = 0; at < cells.Length; at++) cells[at].Tap.onClick.RemoveListener(taps[at]);
        }

        void OnLeave() => commands.TryLeave(session);

        void OnStartOneVsOne() => ChooseMode(Mode._1V1);

        void OnStartTwoVsTwo() => ChooseMode(Mode._2V2);

        void ChooseMode(Mode mode)
        {
            if (room.CanChooseMode) commands.TryChooseMode(session, mode);
        }

        void OnStartMatch()
        {
            if (room.CanStart) commands.TryStart(session);
        }

        void OnBack()
        {
            if (room.CanGoBack) commands.TryBack(session);
        }

        void OnTeamASeat(int index) => OnSeat(room.TeamA, index);

        void OnTeamBSeat(int index) => OnSeat(room.TeamB, index);

        void OnSeat(IReadOnlyList<RoomSeat> seats, int index)
        {
            if (index >= seats.Count) return;
            var seat = seats[index];
            if (!room.IsLeader)
            {
                if (room.CanRequestSeat(seat.Seat)) commands.TryRequestSeat(session, seat.Seat);
                return;
            }
            if (selected != null)
            {
                commands.TryPlace(session, selected, seat.Seat);
                selected = null;
            }
            else if (!seat.Free)
            {
                selected = seat.Account;
            }
            Redraw();
        }

        void OnWaiting(int index)
        {
            if (!room.IsLeader || index >= room.Waiting.Count) return;
            selected = room.Waiting[index] == selected ? null : room.Waiting[index];
            Redraw();
        }

        void Redraw()
        {
            if (room == null) return;
            if (selected != null && link.Members.IndexOf(selected) < 0) selected = null;
            lobbyPanel.SetActive(room.InRoom && room.Phase == LobbyPhase.Lobby);
            teamsPanel.SetActive(room.InRoom && room.Phase == LobbyPhase.Teams);
            if (!room.CanInvite) invitePanel.Hide();
            code.text = room.Code;
            invite.interactable = room.CanInvite;
            startOneVsOne.gameObject.SetActive(room.CanChooseMode);
            startTwoVsTwo.gameObject.SetActive(room.CanChooseMode);
            startMatch.gameObject.SetActive(room.IsLeader);
            startMatch.interactable = room.CanStart;
            back.gameObject.SetActive(room.CanGoBack);
            ShowMembers();
            ShowSeats(teamASeats, room.TeamA);
            ShowSeats(teamBSeats, room.TeamB);
            ShowWaiting();
            RedrawWheel();
            var key = room.RefusalKey;
            message.enabled = key != null;
            message.text = key == null ? string.Empty : strings.Text(key);
        }

        public void Tick()
        {
            if (room == null) return;
            var nowMs = link.NowMs;
            ShowCalm(link.Wheel != null && link.Wheel.Calm(nowMs));
            for (var at = 0; at < emojiAccounts.Length; at++)
            {
                var account = emojiAccounts[at];
                if (account == null) continue;
                if (nowMs < emojiEndsAtMs[at] && lobbyPanel.activeSelf && link.Members.IndexOf(account) == at) continue;
                members[at].HideEmoji();
                emojiAccounts[at] = null;
            }
        }

        void Update() => Tick();

        void OnWheel()
        {
            link.Wheel?.Toggle();
            RedrawWheel();
        }

        void OnWheelSlot(int index)
        {
            link.TrySend(index);
            RedrawWheel();
        }

        void OnInventoryLoaded(InventoryLoadResult result) => RedrawWheel();

        void OnEmojiShown(MemberEmoji shown)
        {
            if (!lobbyPanel.activeSelf) return;
            var at = link.Members.IndexOf(shown.Account);
            if (at < 0 || at >= members.Length || !TryEmojiArt(shown.ItemId, out var art)) return;
            members[at].ShowEmoji(stage, art);
            emojiAccounts[at] = shown.Account;
            emojiEndsAtMs[at] = link.NowMs + TokenEffect.EmojiDurationMs;
        }

        void RedrawWheel()
        {
            if (link.Wheel != null) GrowWheel(link.Wheel.Capacity);
            var usable = link.Emoji != null && link.Emoji.CanSend;
            if (!usable) link.Wheel?.Close();
            emojiRoot.gameObject.SetActive(usable);
            wheelPanel.SetActive(usable && link.Wheel.IsOpen);
            if (!wheelPanel.activeSelf) return;
            for (var at = 0; at < wheelSlots.Length; at++)
            {
                if (at < link.Wheel.Count && TryEmojiArt(link.Wheel.Slot(at), out var art))
                {
                    wheelSlots[at].gameObject.SetActive(true);
                    wheelArt[at].Show(stage, art, null);
                }
                else
                {
                    wheelSlots[at].gameObject.SetActive(false);
                }
            }
        }

        void GrowWheel(int capacity)
        {
            var built = wheelSlots.Length;
            if (slotTemplate == null || capacity <= built) return;
            Array.Resize(ref wheelSlots, capacity);
            Array.Resize(ref wheelArt, capacity);
            Array.Resize(ref slotTaps, capacity);
            for (var at = built; at < capacity; at++)
            {
                wheelSlots[at] = Instantiate(slotTemplate, wheelPanel.transform);
                wheelArt[at] = wheelSlots[at].GetComponentInChildren<CosmeticImage>(true);
                slotTaps[at] = Tap(OnWheelSlot, at);
                wheelSlots[at].onClick.AddListener(slotTaps[at]);
            }
        }

        bool TryEmojiArt(string itemId, out CosmeticArt art)
        {
            art = default;
            if (string.IsNullOrEmpty(itemId) || !manifest.Current.TryGet(itemId, out var item)) return false;
            if (item.Type != CosmeticAddresses.Emoji || !item.IsRendered) return false;
            art = item.Art;
            return true;
        }

        void ShowCalm(bool calm)
        {
            if (calm == shownCalm) return;
            shownCalm = calm;
            emojiRoot.alpha = calm ? TokenOpacity.Disabled : 1f;
        }

        void HideEmoji()
        {
            for (var at = 0; at < emojiAccounts.Length; at++)
            {
                members[at].HideEmoji();
                emojiAccounts[at] = null;
            }
        }

        void ShowMembers()
        {
            var all = link.Members.All;
            for (var at = 0; at < members.Length; at++)
            {
                if (at < all.Count) members[at].Show(all[at], stage, manifest.Current, false);
                else members[at].Hide();
            }
        }

        void ShowSeats(RoomMemberCell[] cells, IReadOnlyList<RoomSeat> seats)
        {
            for (var at = 0; at < cells.Length; at++)
            {
                if (at >= seats.Count) cells[at].Hide();
                else if (seats[at].Free) cells[at].ShowFree();
                else ShowAccount(cells[at], seats[at].Account);
            }
        }

        void ShowWaiting()
        {
            for (var at = 0; at < waiting.Length; at++)
            {
                if (at < room.Waiting.Count) ShowAccount(waiting[at], room.Waiting[at]);
                else waiting[at].Hide();
            }
        }

        void ShowAccount(RoomMemberCell cell, string account)
        {
            var index = link.Members.IndexOf(account);
            if (index < 0)
            {
                cell.ShowFree();
                return;
            }
            cell.Show(link.Members.All[index], stage, manifest.Current, account == selected);
        }
    }
}
