using System.Collections.Generic;
using HypeDomino.Net;
using HypeDomino.UI.Cosmetics;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public sealed class MatchPrefetch
    {
        const int ItemsPerSeat = 3;

        readonly CosmeticArtCache<Sprite> stills;
        readonly List<string> held = new(MatchSnapshot.MaxSeats * ItemsPerSeat + 1);
        int ticket;

        public MatchPrefetch(CosmeticArtCache<Sprite> stills) => this.stills = stills;

        public int Count => held.Count;

        public int Settled { get; private set; }

        public float Loaded => held.Count == 0 ? 1f : Settled / (float)held.Count;

        public bool Done => Settled == held.Count;

        public string this[int index] => held[index];

        public void Start(MatchSnapshot view)
        {
            Release();
            Hold(CosmeticAddresses.Table, view.Table);
            var roster = view.Roster;
            for (var at = 0; at < roster.Count; at++)
            {
                var entry = roster[at];
                Hold(CosmeticAddresses.Avatar, entry.Avatar);
                Hold(CosmeticAddresses.Frame, entry.Frame);
                Hold(CosmeticAddresses.Tile, entry.Tile);
            }
            var current = ticket;
            for (var at = 0; at < held.Count; at++) stills.Acquire(held[at], _ => OnSettled(current));
        }

        public void Release()
        {
            ticket++;
            for (var at = 0; at < held.Count; at++) stills.Release(held[at]);
            held.Clear();
            Settled = 0;
        }

        void Hold(string type, string itemId)
        {
            var address = CosmeticAddresses.Static(type, itemId);
            if (address != null && !held.Contains(address)) held.Add(address);
        }

        void OnSettled(int from)
        {
            if (from == ticket) Settled++;
        }
    }
}
