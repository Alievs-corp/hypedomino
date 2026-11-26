using System.Collections.Generic;
using HypeDomino.UI.Board;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using NUnit.Framework;
using Unity.Profiling;
using UnityEngine;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class TilePlaceLayerTests
    {
        const float PlaceMs = TokenDurationMs.Base + TokenDurationMs.Micro;
        const int MarkedTicks = 3;
        const int RecordedSamplesCap = 64;

        static readonly Vector3 Hand = new Vector3(0f, -500f, 0f);
        static readonly Vector3 Slot = new Vector3(200f, 40f, 0f);

        GameObject root;
        TilePlaceLayer layer;
        Transform[] tiles;
        Transform[] shadows;
        List<int> settled;
        int settledCount;

        [SetUp]
        public void BuildLayer()
        {
            root = new GameObject("Board");
            tiles = new Transform[4];
            shadows = new Transform[4];
            for (var i = 0; i < tiles.Length; i++)
            {
                shadows[i] = new GameObject("Shadow" + i).transform;
                shadows[i].SetParent(root.transform, false);
                tiles[i] = new GameObject("Tile" + i).transform;
                tiles[i].SetParent(root.transform, false);
            }
            layer = root.AddComponent<TilePlaceLayer>();
            layer.Build(tiles, shadows);
            settled = new List<int>();
            settledCount = 0;
        }

        void Record(int tile) => settled.Add(tile);

        void Count(int tile) => settledCount++;

        [TearDown]
        public void DestroyLayer() => Object.Destroy(root);

        [Test]
        public void APlacedTileTravelsTurnsAndSettlesOnItsSlot()
        {
            layer.Place(2, Hand, 0f, Slot, 90f, 0f);

            Assert.AreEqual(Hand, tiles[2].localPosition);
            layer.Tick(TokenDurationMs.Base);
            Assert.AreEqual(Slot, tiles[2].localPosition);
            Assert.AreEqual(90f, tiles[2].localEulerAngles.z, 1e-3f);
            Assert.AreEqual(TokenMotion.PressScale, tiles[2].localScale.x, 1e-3f);
            layer.Tick(PlaceMs);
            Assert.AreEqual(Vector3.one, tiles[2].localScale);
            Assert.AreEqual(0, layer.MovingCount);
        }

        [Test]
        public void ReduceMotionLandsInFastWithoutTheSettle()
        {
            layer.Configure(EffectLevel.Medium, true);
            layer.Place(1, Hand, 0f, Slot, -90f, 0f);

            layer.Tick(TokenDurationMs.Fast);

            Assert.AreEqual(Slot, tiles[1].localPosition);
            Assert.AreEqual(Vector3.one, tiles[1].localScale);
            Assert.AreEqual(0, layer.MovingCount);
        }

        [TestCase(EffectLevel.Low, TokenEffectLevel.LowMaxTweens)]
        [TestCase(EffectLevel.Medium, TokenEffectLevel.MediumMaxTweens)]
        [TestCase(EffectLevel.High, TokenEffectLevel.HighMaxTweens)]
        public void ThePoolFollowsTheEffectLevel(EffectLevel level, int capacity)
        {
            layer.Configure(level, false);
            Assert.AreEqual(capacity, layer.Capacity);
        }

        [Test]
        public void ChangingTheEffectLevelLandsTheMovingTiles()
        {
            layer.Place(0, Hand, 0f, Slot, 90f, 0f);

            layer.Configure(EffectLevel.Low, false);

            Assert.AreEqual(Slot, tiles[0].localPosition);
            Assert.AreEqual(0, layer.MovingCount);
        }

        [Test]
        public void TheShadowTravelsUnderItsTileAtTheRestingElevation()
        {
            layer.Place(2, Hand, 0f, Slot, 90f, 0f);

            Assert.AreEqual(Hand + Vector3.down * TokenElevation.E1OffsetY, shadows[2].localPosition);
            layer.Tick(TokenDurationMs.Base * 0.5f);
            Assert.AreEqual(tiles[2].localPosition + Vector3.down * TokenElevation.E1OffsetY, shadows[2].localPosition);
            Assert.AreEqual(tiles[2].localRotation, shadows[2].localRotation);
            layer.Tick(PlaceMs);
            Assert.AreEqual(Slot + Vector3.down * TokenElevation.E1OffsetY, shadows[2].localPosition);
            Assert.AreEqual(Vector3.one, shadows[2].localScale);
        }

        [Test]
        public void ATileSettlesOnceWhenItLandsAndNotWhenItIsPlaced()
        {
            layer.Settled += Record;

            layer.Place(1, Hand, 0f, Slot, 90f, 0f);
            layer.Tick(0f);
            Assert.IsEmpty(settled);
            Assert.IsTrue(layer.IsSettling(1));
            layer.Tick(TokenDurationMs.Base);
            Assert.IsEmpty(settled, "still settling into place");
            layer.Tick(PlaceMs);
            CollectionAssert.AreEqual(new[] { 1 }, settled);
            Assert.IsFalse(layer.IsSettling(1));
            layer.Tick(PlaceMs * 2f);
            CollectionAssert.AreEqual(new[] { 1 }, settled);
        }

        [Test]
        public void WithReduceMotionATileSettlesWhenItsFastTravelEnds()
        {
            layer.Configure(EffectLevel.Medium, true);
            layer.Settled += Record;

            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Tick(TokenDurationMs.Fast - 1f);
            Assert.IsEmpty(settled);
            layer.Tick(TokenDurationMs.Fast);

            CollectionAssert.AreEqual(new[] { 0 }, settled);
        }

        [Test]
        public void PlacementsThatArriveTogetherSettleOneAfterAnother()
        {
            layer.Settled += Record;

            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Place(1, Hand, 0f, Slot, 90f, 0f);
            layer.Tick(PlaceMs);
            CollectionAssert.AreEqual(new[] { 0 }, settled);
            Assert.AreEqual(Hand, tiles[1].localPosition, "the second waits where it was shown");
            layer.Tick(PlaceMs * 2f);

            CollectionAssert.AreEqual(new[] { 0, 1 }, settled);
        }

        [Test]
        public void APlacementPastThePresentationLagLandsEveryTileAndEachSettlesOnce()
        {
            layer.Settled += Record;

            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Place(1, Hand, 0f, Slot, 90f, 0f);
            layer.Place(2, Hand, 0f, Slot, 90f, 0f);
            Assert.AreEqual(Slot, tiles[0].localPosition);
            Assert.AreEqual(Slot, tiles[2].localPosition);
            layer.Tick(0f);

            CollectionAssert.AreEquivalent(new[] { 0, 1, 2 }, settled);
            layer.Tick(PlaceMs * 3f);
            Assert.AreEqual(3, settled.Count);
        }

        [Test]
        public void AnEffectLevelChangeLandsTheMovingTilesAndSettlesThem()
        {
            layer.Settled += Record;
            layer.Place(3, Hand, 0f, Slot, 90f, 0f);

            layer.Configure(EffectLevel.High, false);
            layer.Tick(1f);

            CollectionAssert.AreEqual(new[] { 3 }, settled);
        }

        [Test]
        public void LandingSnapsEveryTileWithoutASettle()
        {
            layer.Settled += Record;
            layer.Place(0, Hand, 0f, Slot, 90f, 0f);

            layer.Land();
            layer.Tick(PlaceMs);

            Assert.AreEqual(Slot, tiles[0].localPosition);
            Assert.AreEqual(Vector3.one, tiles[0].localScale);
            Assert.IsEmpty(settled);
            Assert.AreEqual(0, layer.MovingCount);
        }

        [Test]
        public void AMovingTileFollowsItsSlotWhenTheBoardIsLaidOutAgain()
        {
            var moved = new Vector3(-80f, 10f, 0f);
            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Place(1, Hand, 0f, Slot, 90f, 0f);

            layer.Follow(0, moved, 180f, TokenDurationMs.Base * 0.5f);
            layer.Follow(1, moved, 180f, TokenDurationMs.Base * 0.5f);

            Assert.AreEqual(Hand, tiles[1].localPosition, "a waiting tile stays where it was shown");
            layer.Tick(PlaceMs);
            Assert.AreEqual(moved, tiles[0].localPosition);
            Assert.AreEqual(180f, tiles[0].localEulerAngles.z, 1e-3f);
            layer.Tick(PlaceMs * 2f);
            Assert.AreEqual(moved, tiles[1].localPosition);
        }

        [Test]
        public void FollowingASettledTileChangesNothing()
        {
            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Tick(PlaceMs);

            layer.Follow(0, Hand, 0f, PlaceMs);

            Assert.AreEqual(Slot, tiles[0].localPosition);
        }

        [Test]
        public void PlacingAndTickingDoesNotAllocate()
        {
            var now = 0f;
            layer.Settled += Count;
            layer.Place(0, Hand, 0f, Slot, 90f, now);
            layer.Tick(now += PlaceMs);
            Assert.That(() =>
            {
                for (var turn = 0; turn < 200; turn++)
                {
                    layer.Place(turn % tiles.Length, Hand, 0f, Slot, 90f, now);
                    layer.Follow(turn % tiles.Length, Slot, 90f, now);
                    layer.Tick(now += 16f);
                    layer.Tick(now += PlaceMs);
                }
            }, Allocates.Nothing);
        }

        [Test]
        public void EveryTickIsOneSampleOfTheTilePlaceEffectMarker()
        {
            layer.Place(0, Hand, 0f, Slot, 90f, 0f);
            layer.Tick(0f);
            var recorder = new ProfilerRecorder(TilePlaceLayer.ProfilerMarkerName, RecordedSamplesCap, ProfilerRecorderOptions.CollectOnlyOnCurrentThread);
            try
            {
                Assert.IsTrue(recorder.Valid);
                recorder.Start();
                for (var tick = 1; tick <= MarkedTicks; tick++) layer.Tick(tick * PlaceMs / MarkedTicks);
                recorder.Stop();
                Assert.AreEqual(MarkedTicks, recorder.Count);
            }
            finally
            {
                recorder.Dispose();
            }
        }
    }
}
