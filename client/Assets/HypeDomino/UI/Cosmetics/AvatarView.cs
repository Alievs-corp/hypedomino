using HypeDomino.UI.Rarity;
using UnityEngine;

namespace HypeDomino.UI.Cosmetics
{
    public sealed class AvatarView : MonoBehaviour
    {
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

        void ShowMarks()
        {
            if (halo != null) halo.Show(stage.Rarities, markRarityId, stage.Level, stage.ReduceMotion, Time.unscaledTime);
            if (diamond != null) diamond.Show(stage.Rarities, markRarityId);
        }
    }
}
