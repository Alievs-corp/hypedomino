using System;
using HypeDomino.Net;
using HypeDomino.UI.Cosmetics;
using UnityEngine;

namespace HypeDomino.Game.Match
{
    public sealed class MatchRoot
    {
        bool loadingWasOpen;

        public MatchRoot(IMatchTable table, IMatchHud hud, ILoadingScreen loadingScreen, CosmeticArtCache<Sprite> stills, ActiveCosmeticManifest manifest, MatchEmoji emoji)
        {
            Screen = new MatchScreen(table, hud, AudioBoot.PlayMatch);
            Loading = new MatchLoading(loadingScreen, new MatchPrefetch(stills), manifest);
            Emoji = emoji ?? throw new ArgumentNullException(nameof(emoji));
        }

        public MatchScreen Screen { get; }
        public MatchLoading Loading { get; }
        public MatchEmoji Emoji { get; }

        public void Bind(Session session)
        {
            Loading.Bind(session);
            Screen.Bind(session);
            Emoji.Bind(session);
        }

        public void Unbind()
        {
            Loading.Unbind();
            Screen.Unbind();
            Emoji.Unbind();
        }

        public void Tick(Connection connection, double nowSeconds, bool reduceMotion, bool sceneReady)
        {
            Loading.Tick(sceneReady);
            Screen.Tick(connection, nowSeconds, reduceMotion);
            var open = Loading.Open;
            if (loadingWasOpen && !open) Loading.ReleaseArt();
            loadingWasOpen = open;
        }
    }
}
