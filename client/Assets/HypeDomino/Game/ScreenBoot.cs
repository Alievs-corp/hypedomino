using HypeDomino.UI.Screens;
using UnityEngine;
using UnityEngine.EventSystems;

namespace HypeDomino.Game
{
    public static class ScreenBoot
    {
        public static ScreenHost Host { get; private set; }

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.AfterSceneLoad)]
        static void Start()
        {
            var root = new GameObject(nameof(ScreenBoot));
            Object.DontDestroyOnLoad(root);
            if (EventSystem.current == null) root.AddComponent<EventSystem>().gameObject.AddComponent<StandaloneInputModule>();
            var stack = root.AddComponent<ScreenStack>();
            var motion = MotionBoot.Motion;
            stack.ReduceMotion = motion.ReduceMotion;
            motion.Changed += () => stack.ReduceMotion = motion.ReduceMotion;
            Host = new ScreenHost(stack);
            Host.Follow(AppBoot.Flow.State);
            AppBoot.Flow.Changed += Host.Follow;
        }
    }
}
