using HypeDomino.Game.FriendLobby;
using HypeDomino.Game.Inventory;
using HypeDomino.Net;
using HypeDomino.Net.Api;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class FriendLobbyBoot
    {
        public static FriendLobbyLink Link { get; private set; }

        public static void Start(Session session, FriendRoom room, ApiClient api)
        {
            Link = new FriendLobbyLink(room, api, ClockMs);
            Link.InventoryLoaded += InventoryLoaded;
            Link.Follow(session);
        }

        static long ClockMs() => (long)(Time.realtimeSinceStartupAsDouble * 1000d);

        static void InventoryLoaded(InventoryLoadResult result)
        {
            if (result.Outcome != InventoryLoadOutcome.Loaded) Debug.LogWarning($"Friend lobby inventory outcome={result.Outcome}");
        }
    }
}
