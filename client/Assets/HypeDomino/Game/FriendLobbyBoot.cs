using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Friends;
using HypeDomino.Game.Inventory;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using HypeDomino.Net.Auth;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class FriendLobbyBoot
    {
        public static FriendLobbyLink Link { get; private set; }

        public static void Start(Session session, FriendRoom room, LobbyCommands commands, ApiClient api, AuthSession auth)
        {
            Link = new FriendLobbyLink(room, commands, api, ClockMs);
            Link.InventoryLoaded += InventoryLoaded;
            Link.FriendsLoaded += FriendsLoaded;
            auth.AccountSignedIn += Link.SignedIn;
            Link.Follow(session);
        }

        public static void Tick() => Link?.Tick();

        public static void SignedOut() => Link?.SignedOut();

        static long ClockMs() => (long)(Time.realtimeSinceStartupAsDouble * 1000d);

        static void InventoryLoaded(InventoryLoadResult result)
        {
            if (result.Outcome != InventoryLoadOutcome.Loaded) Debug.LogWarning($"Friend lobby inventory outcome={result.Outcome}");
        }

        static void FriendsLoaded(FriendListLoadOutcome outcome)
        {
            if (outcome != FriendListLoadOutcome.Loaded) Debug.LogWarning($"Friend lobby friends outcome={outcome}");
        }
    }
}
