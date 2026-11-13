using HypeDomino.Content;
using HypeDomino.UI.Graphics;
using UnityEngine;

namespace HypeDomino.Game
{
    public static class AppBoot
    {
        public static AppFlow Flow { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.BeforeSceneLoad)]
        static void Start()
        {
            Flow = new AppFlow();
            Flow.Changed += OnChanged;
        }

        public static void FollowMatch(AppState state, FrameRatePolicy frameRate, CatalogUpdater catalog)
        {
            var inMatch = state == AppState.Match;
            frameRate.SetInMatch(inMatch);
            catalog?.SetInMatch(inMatch);
        }

        static void OnChanged(AppState state)
        {
            FollowMatch(state, GraphicsBoot.FrameRate, ContentBoot.Catalog);
            Debug.Log($"App state={state}");
        }
    }
}
