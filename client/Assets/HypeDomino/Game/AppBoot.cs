using System;
using System.Collections.Generic;
using HypeDomino.Content;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Localization;
using Hypedomino.V1;
using UnityEngine;
using AnnouncedBundle = Hypedomino.V1.StringBundle;

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

        public static void Listen(Session session, AppFlow flow, CatalogUpdater catalog, GameStrings strings, StringBundleUpdater bundles, Action<HapticCue> playHaptic)
        {
            var cues = new HapticCues();
            session.HelloReceived += hello =>
            {
                cues.StartConnection();
                Hello(hello, flow, catalog, strings, bundles);
            };
            session.ViewReceived += view =>
            {
                flow.View(view.State);
                playHaptic(cues.Read(view));
            };
        }

        public static void FollowMatch(AppState state, FrameRatePolicy frameRate, CatalogUpdater catalog)
        {
            var inMatch = state == AppState.Match;
            frameRate.SetInMatch(inMatch);
            catalog?.SetInMatch(inMatch);
        }

        public static List<AnnouncedBundle> FollowStrings(ServerHello hello, GameStrings strings)
        {
            var chains = hello.FallbackChains;
            for (var i = 0; i < chains.Count; i++)
                if (!strings.SetChain(chains[i].Language, chains[i].Languages)) Debug.LogWarning($"Strings chain refused language={chains[i].Language}");
            var bundles = hello.StringBundles;
            var missing = new List<AnnouncedBundle>(bundles.Count);
            for (var i = 0; i < bundles.Count; i++)
                if (IsWanted(bundles[i].Language, strings.Language) && !strings.Holds(bundles[i].Language, bundles[i].Hash)) missing.Add(bundles[i]);
            return missing;
        }

        static bool IsWanted(string language, string selected) => language == selected || language == GameLanguage.English;

        static void Hello(ServerHello hello, AppFlow flow, CatalogUpdater catalog, GameStrings strings, StringBundleUpdater bundles)
        {
            flow.Hello(hello);
            if (hello.Status != HelloStatus.Accepted) return;
            catalog?.Announce(hello.CatalogVersion, hello.ContentBaseUrl);
            var missing = FollowStrings(hello, strings);
            for (var i = 0; i < missing.Count; i++)
            {
                Debug.Log($"Strings fetch language={missing[i].Language} version={missing[i].Version}");
                bundles?.Fetch(hello.ContentBaseUrl, missing[i].Language, (int)missing[i].Version, missing[i].Hash);
            }
        }

        static void OnChanged(AppState state)
        {
            FollowMatch(state, GraphicsBoot.FrameRate, ContentBoot.Catalog);
            Debug.Log($"App state={state}");
        }
    }
}
