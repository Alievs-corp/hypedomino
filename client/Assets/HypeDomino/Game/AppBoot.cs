using HypeDomino.Content;
using HypeDomino.Net;
using HypeDomino.UI.Graphics;
using Hypedomino.V1;
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
            new StartupMarker(Flow, () => Time.realtimeSinceStartupAsDouble, line => Debug.Log(line));
        }

        public static void Listen(Session session, AppFlow flow, CatalogUpdater catalog)
        {
            session.HelloReceived += hello => Hello(hello, flow, catalog);
            session.ViewReceived += view => flow.View(view.State);
        }

        public static void FollowMatch(AppState state, FrameRatePolicy frameRate, CatalogUpdater catalog)
        {
            var inMatch = state == AppState.Match;
            frameRate.SetInMatch(inMatch);
            catalog?.SetInMatch(inMatch);
        }

        static void Hello(ServerHello hello, AppFlow flow, CatalogUpdater catalog)
        {
            flow.Hello(hello);
            if (hello.Status == HelloStatus.Accepted) catalog?.Announce(hello.CatalogVersion, hello.ContentBaseUrl);
        }

        static void OnChanged(AppState state)
        {
            FollowMatch(state, GraphicsBoot.FrameRate, ContentBoot.Catalog);
            Debug.Log($"App state={state}");
        }
    }
}
