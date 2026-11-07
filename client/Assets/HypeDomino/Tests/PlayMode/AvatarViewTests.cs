using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Rarity;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class AvatarViewTests
    {
        const float FramesPerSecond = 10f;
        const string AvatarStatic = "cosmetic/avatar/a/static";
        const string AvatarAnim = "cosmetic/avatar/a/anim";
        const string OtherStatic = "cosmetic/avatar/b/static";
        const string FrameStatic = "cosmetic/frame/a/static";
        const string FrameAnim = "cosmetic/frame/a/anim";
        const string Rarities = @"{""diamondMinSize"":72,""flowPeriodMs"":2000,""tiers"":[
{""id"":""top"",""fill"":""#0000FF"",""on"":""#000000"",""text"":""#FFFFFF"",""pips"":1,""glowAlpha"":0.75,""glowScale"":1.5,""flow"":true}]}";

        GameObject root;
        FakeArtSource<Sprite> stills;
        FakeArtSource<LoopArt> loops;
        Sprite builtInAvatar;
        Sprite builtInFrame;
        Sprite still;
        Sprite frameOne;
        LoopArt loop;
        RarityTable table;

        static readonly CosmeticArt AnimatedAvatar = new CosmeticArt(AvatarStatic, AvatarAnim);
        static readonly CosmeticArt StaticFrame = new CosmeticArt(FrameStatic, null);

        [SetUp]
        public void Build()
        {
            root = new GameObject("Screen", typeof(RectTransform), typeof(Canvas));
            stills = new FakeArtSource<Sprite>();
            loops = new FakeArtSource<LoopArt>();
            builtInAvatar = NewSprite();
            builtInFrame = NewSprite();
            still = NewSprite();
            frameOne = NewSprite();
            loop = new LoopArt(still, new[] { frameOne, NewSprite() }, FramesPerSecond);
            table = RarityTable.Parse(Rarities);
        }

        [TearDown]
        public void DestroyScreen() => Object.Destroy(root);

        static Sprite NewSprite() => Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), Vector2.zero);

        CosmeticStage NewStage(EffectLevel level, bool reduceMotion = false) =>
            new CosmeticStage(stills, loops, table, builtInAvatar, builtInFrame, level, reduceMotion);

        CosmeticImage NewImage(string name)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<AnimatedImage>().gameObject.AddComponent<CosmeticImage>();
            image.transform.SetParent(root.transform, false);
            return image;
        }

        AvatarView NewView(RarityHalo halo = null)
        {
            var view = new GameObject("Avatar", typeof(RectTransform)).AddComponent<AvatarView>();
            view.transform.SetParent(root.transform, false);
            view.Configure(NewImage("Art"), NewImage("Frame"), halo, null);
            return view;
        }

        static AvatarLook Look(CosmeticArt avatar) => new AvatarLook(avatar, StaticFrame, "top");

        [Test]
        public void TheBuiltInStaticShowsUntilTheItemsStaticHasLoaded()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.Medium), Look(AnimatedAvatar));
            Assert.AreSame(builtInAvatar, view.Avatar.Shown);
            Assert.AreSame(builtInFrame, view.Frame.Shown);

            stills.Finish(AvatarStatic, still);
            Assert.AreSame(still, view.Avatar.Shown);
        }

        [Test]
        public void TheStaticLoadsFirstAndTheLoopOnlyAfterIt()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.Medium), Look(AnimatedAvatar));
            Assert.IsEmpty(loops.Loads);

            stills.Finish(AvatarStatic, still);
            Assert.AreEqual(new[] { AvatarAnim }, loops.Loads);
            Assert.IsFalse(view.Avatar.IsPlaying);

            loops.Finish(AvatarAnim, loop);
            Assert.IsTrue(view.Avatar.IsPlaying);
            Assert.AreSame(frameOne, view.Avatar.Shown);
        }

        [Test]
        public void OnTheLowLevelEveryAvatarIsStaticAndNoLoopIsLoaded()
        {
            var view = NewView();
            var look = new AvatarLook(AnimatedAvatar, new CosmeticArt(FrameStatic, FrameAnim), "top");
            view.Show(NewStage(EffectLevel.Low), look);
            stills.Finish(AvatarStatic, still);
            stills.Finish(FrameStatic, still);

            Assert.IsFalse(view.Avatar.IsPlaying);
            Assert.IsFalse(view.Frame.IsPlaying);
            Assert.AreSame(still, view.Avatar.Shown);
            Assert.AreSame(still, view.Frame.Shown);
            Assert.IsEmpty(loops.Loads);
        }

        [Test]
        public void ReduceMotionKeepsTheAvatarStatic()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.High, true), Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            Assert.IsFalse(view.Avatar.IsPlaying);
            Assert.IsEmpty(loops.Loads);
        }

        [Test]
        public void AStaticItemNeverAsksForALoop()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.High), Look(new CosmeticArt(AvatarStatic, null)));
            stills.Finish(AvatarStatic, still);
            Assert.IsEmpty(loops.Loads);
        }

        [Test]
        public void TurningGraphicsDownStopsTheLoopFreesItAndItsSlot()
        {
            var stage = NewStage(EffectLevel.Medium);
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            loops.Finish(AvatarAnim, loop);
            var budget = stage.Budget;
            Assert.AreEqual(1, budget.Active);

            stage.SetEffects(EffectLevel.Low, false);
            view.RefreshEffects();

            Assert.IsFalse(view.Avatar.IsPlaying);
            Assert.AreSame(still, view.Avatar.Shown);
            Assert.AreEqual(0, budget.Active);
            Assert.AreEqual(new[] { AvatarAnim }, loops.Releases);
        }

        [Test]
        public void TurningGraphicsUpStartsTheLoop()
        {
            var stage = NewStage(EffectLevel.Low);
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);

            stage.SetEffects(EffectLevel.High, false);
            view.RefreshEffects();
            loops.Finish(AvatarAnim, loop);

            Assert.IsTrue(view.Avatar.IsPlaying);
        }

        [Test]
        public void AFailedStaticKeepsTheBuiltInAndAsksForNoLoop()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.Medium), Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, null);
            Assert.AreSame(builtInAvatar, view.Avatar.Shown);
            Assert.IsEmpty(loops.Loads);
        }

        [Test]
        public void AFailedLoopStaysOnTheStatic()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.Medium), Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            loops.Finish(AvatarAnim, null);
            Assert.IsFalse(view.Avatar.IsPlaying);
            Assert.AreSame(still, view.Avatar.Shown);
        }

        [Test]
        public void ArtThatArrivesAfterTheRowWasReusedIsIgnoredAndFreed()
        {
            var stage = NewStage(EffectLevel.Medium);
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            view.Show(stage, Look(new CosmeticArt(OtherStatic, null)));

            stills.Finish(AvatarStatic, still);

            Assert.AreSame(builtInAvatar, view.Avatar.Shown);
            Assert.AreEqual(new[] { AvatarStatic }, stills.Releases);
        }

        [Test]
        public void ShowingTheSameItemAgainDoesNotReloadIt()
        {
            var stage = NewStage(EffectLevel.Medium);
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            view.Show(stage, Look(AnimatedAvatar));
            Assert.AreEqual(1, stills.Loads.FindAll(address => address == AvatarStatic).Count);
            Assert.IsEmpty(stills.Releases);
        }

        [Test]
        public void TheLoopIsNotLoadedWhenTheScreensBudgetIsSpent()
        {
            var stage = NewStage(EffectLevel.Medium);
            while (stage.Budget.TryAcquire()) { }
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            Assert.IsEmpty(loops.Loads);
            Assert.AreSame(still, view.Avatar.Shown);
        }

        [Test]
        public void ARowHiddenAndShownAgainPlaysAgain()
        {
            var stage = NewStage(EffectLevel.Medium);
            var view = NewView();
            view.Show(stage, Look(AnimatedAvatar));
            stills.Finish(AvatarStatic, still);
            loops.Finish(AvatarAnim, loop);

            view.Avatar.gameObject.SetActive(false);
            Assert.AreEqual(0, stage.Budget.Active);
            view.Avatar.gameObject.SetActive(true);

            Assert.IsTrue(view.Avatar.IsPlaying);
            Assert.AreEqual(1, stage.Budget.Active);
        }

        [Test]
        public void TheHaloFollowsTheLooksRarityAndTheEffectLevel()
        {
            var image = new GameObject("Halo", typeof(RectTransform), typeof(Image)).AddComponent<RarityHalo>();
            image.transform.SetParent(root.transform, false);
            image.Configure(NewSprite(), NewSprite());
            var stage = NewStage(EffectLevel.Medium);
            var view = NewView(image);

            view.Show(stage, Look(AnimatedAvatar));
            Assert.IsTrue(image.IsRotating);

            stage.SetEffects(EffectLevel.Low, false);
            view.RefreshEffects();
            Assert.IsTrue(image.IsShown);
            Assert.IsFalse(image.IsRotating);
        }

        [Test]
        public void AnAbsentArtShowsTheBuiltInWithoutLoading()
        {
            var view = NewView();
            view.Show(NewStage(EffectLevel.Medium), Look(default));
            Assert.AreSame(builtInAvatar, view.Avatar.Shown);
            Assert.AreEqual(new[] { FrameStatic }, stills.Loads);
        }
    }
}
