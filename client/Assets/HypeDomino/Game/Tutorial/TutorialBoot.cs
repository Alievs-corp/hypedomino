using UnityEngine;

namespace HypeDomino.Game.Tutorial
{
    public static class TutorialBoot
    {
        public static TutorialOffer Offer { get; } = new();

        [RuntimeInitializeOnLoadMethod(RuntimeInitializeLoadType.AfterSceneLoad)]
        static void Start() => Offer.Follow(AppBoot.Flow);
    }
}
