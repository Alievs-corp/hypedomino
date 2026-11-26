using System;
using HypeDomino.Game.Profile;
using HypeDomino.Net;
using HypeDomino.Net.Api;

namespace HypeDomino.Game.FriendLobby
{
    public sealed class InvitePrompt
    {
        readonly RoomInvites invites;
        readonly ProfileLoader loader;
        readonly Action<ProfileLoadResult> onLoaded;
        string named;
        string loading;

        public InvitePrompt(RoomInvites invites, ApiClient api)
        {
            this.invites = invites ?? throw new ArgumentNullException(nameof(invites));
            loader = new ProfileLoader(api ?? throw new ArgumentNullException(nameof(api)));
            onLoaded = Loaded;
            invites.Changed += Read;
        }

        public bool Shown => invites.Pending.Count > 0;
        public RoomInvite Current => Shown ? invites.Pending[invites.Pending.Count - 1] : default;
        public string Nickname { get; private set; }
        public bool CanAccept => Shown && invites.CanAccept;

        public event Action Changed;

        public bool TryAccept(Session session, LobbyCommands commands) => Shown && invites.TryAccept(session, commands, Current.Code);

        public void Decline()
        {
            if (Shown) invites.Decline(Current.Code);
        }

        void Read()
        {
            var from = Current.From;
            if (from != named)
            {
                named = from;
                Nickname = null;
                LoadName();
            }
            Changed?.Invoke();
        }

        void LoadName()
        {
            if (named == null || loading != null) return;
            loading = named;
            if (!loader.LoadPublic(loading, onLoaded)) loading = null;
        }

        void Loaded(ProfileLoadResult result)
        {
            var account = loading;
            loading = null;
            if (account != named)
            {
                LoadName();
                return;
            }
            if (result.Profile == null) return;
            Nickname = result.Profile.Nickname;
            Changed?.Invoke();
        }
    }
}
