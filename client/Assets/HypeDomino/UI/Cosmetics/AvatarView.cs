using HypeDomino.UI.Rarity;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Cosmetics
{
    [RequireComponent(typeof(RectTransform))]
    public sealed class AvatarView : MonoBehaviour
    {
        static readonly Vector2 Centre = new Vector2(0.5f, 0.5f);

        [SerializeField] CosmeticImage avatar;
        [SerializeField] CosmeticImage frame;
        [SerializeField] RarityHalo halo;
        [SerializeField] RarityDiamond diamond;

        CosmeticStage stage;
        string markRarityId;

        public CosmeticImage Avatar => avatar;

        public CosmeticImage Frame => frame;

        public void Configure(CosmeticImage avatarImage, CosmeticImage frameImage, RarityHalo haloMark, RarityDiamond diamondMark)
        {
            avatar = avatarImage;
            frame = frameImage;
            halo = haloMark;
            diamond = diamondMark;
            FitArt();
        }

        public void Show(CosmeticStage onStage, AvatarLook look)
        {
            stage = onStage;
            markRarityId = look.MarkRarityId;
            avatar.Show(stage, look.Avatar, stage.BuiltInAvatar);
            frame.Show(stage, look.Frame, stage.BuiltInFrame);
            ShowMarks();
        }

        public void RefreshEffects()
        {
            if (stage == null) return;
            avatar.Refresh();
            frame.Refresh();
            ShowMarks();
        }

        void OnEnable() => FitArt();

        void OnRectTransformDimensionsChange() => FitArt();

        void FitArt()
        {
            var box = ((RectTransform)transform).rect;
            var side = Mathf.Max(0f, Mathf.Min(box.width, box.height));
            FitSquare(avatar, side);
            FitSquare(frame, side);
            FitSquare(halo, side);
        }

        static void FitSquare(Component art, float side)
        {
            if (art == null) return;
            var rect = (RectTransform)art.transform;
            rect.anchorMin = Centre;
            rect.anchorMax = Centre;
            rect.pivot = Centre;
            rect.anchoredPosition = Vector2.zero;
            rect.sizeDelta = new Vector2(side, side);
            if (art.TryGetComponent<Image>(out var image)) image.preserveAspect = true;
        }

        void ShowMarks()
        {
            if (halo != null) halo.Show(stage.Rarities, markRarityId, stage.Level, stage.ReduceMotion, Time.unscaledTime);
            if (diamond != null) diamond.Show(stage.Rarities, markRarityId);
        }
    }
}
