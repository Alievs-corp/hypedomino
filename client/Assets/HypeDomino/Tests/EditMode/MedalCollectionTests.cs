using HypeDomino.UI.Medals;
using NUnit.Framework;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MedalCollectionTests
    {
        const string CollectionJson = @"{""version"":7,""slots"":3,""medals"":[
{""id"":""medal.first_win"",""revision"":1,""nameKey"":""medal.first_win.name"",""descriptionKey"":""medal.first_win.description"",""rarityId"":""common"",""staticAddress"":""medal/first_win/static"",""sources"":[{""type"":""mission"",""ref"":""mission.first_win""}],""availability"":""live""},
{""id"":""medal.hundred_wins"",""revision"":2,""nameKey"":""medal.hundred_wins.name"",""descriptionKey"":""medal.hundred_wins.description"",""rarityId"":""epic"",""staticAddress"":""medal/hundred_wins/static"",""sources"":[{""type"":""mission"",""ref"":""mission.hundred_wins""}],""availability"":""live""},
{""id"":""medal.moon_event"",""revision"":1,""nameKey"":""medal.moon_event.name"",""descriptionKey"":""medal.moon_event.description"",""rarityId"":""rare"",""staticAddress"":""medal/moon_event/static"",""sources"":[{""type"":""mission"",""ref"":""mission.moon_event""}],""availability"":""retired""},
{""id"":""medal.season_top"",""revision"":1,""nameKey"":""medal.season_top.name"",""descriptionKey"":""medal.season_top.description"",""rarityId"":""legendary"",""staticAddress"":""medal/season_top/static"",""sources"":[{""type"":""rank_reward"",""ref"":""rank.s1.top""}],""availability"":""live""},
{""id"":""medal.domino_ten"",""revision"":1,""nameKey"":""medal.domino_ten.name"",""descriptionKey"":""medal.domino_ten.description"",""rarityId"":""rare"",""staticAddress"":""medal/domino_ten/static"",""sources"":[{""type"":""mission"",""ref"":""mission.domino_ten""}],""availability"":""live""}],
""owned"":[
{""medal"":""medal.first_win"",""earned_at"":""2026-11-01T10:00:00Z""},
{""medal"":""medal.moon_event"",""earned_at"":""2026-11-02T10:00:00Z""},
{""medal"":""medal.domino_ten"",""earned_at"":""2026-11-03T10:00:00Z""},
{""medal"":""medal.switched_off"",""earned_at"":""2026-11-04T10:00:00Z""}],
""showcase"":[""medal.moon_event"",""medal.switched_off"",""medal.first_win""]}";

        static MedalCollection Collection(string json = CollectionJson)
        {
            var collection = new MedalCollection();
            collection.Show(MedalCollectionReply.FromJson(json));
            return collection;
        }

        static string[] Showcase(MedalCollection collection)
        {
            var ids = new string[collection.ShowcaseCount];
            for (var at = 0; at < ids.Length; at++) ids[at] = collection.ShowcaseAt(at).id;
            return ids;
        }

        [Test]
        public void TheHeaderCountsEarnedOverShownMedals()
        {
            var collection = Collection();

            Assert.AreEqual(5, collection.ShownCount);
            Assert.AreEqual(3, collection.EarnedCount);
            Assert.AreEqual(3, collection.Slots);
        }

        [Test]
        public void AnEarnedMedalCarriesTheServerDateEarned()
        {
            Assert.IsTrue(Collection().TryGet("medal.moon_event", out var entry));

            Assert.IsTrue(entry.Earned);
            Assert.AreEqual("2026-11-02T10:00:00Z", entry.EarnedAt);
            Assert.AreEqual("medal.moon_event.name", entry.Medal.nameKey);
            Assert.AreEqual("medal.moon_event.description", entry.Medal.descriptionKey);
            Assert.AreEqual("rare", entry.Medal.rarityId);
            Assert.AreEqual("medal/moon_event/static", entry.Medal.staticAddress);
        }

        [Test]
        public void ALockedMedalIsShownWithItsDescriptionAndNoDate()
        {
            Assert.IsTrue(Collection().TryGet("medal.hundred_wins", out var entry));

            Assert.IsFalse(entry.Earned);
            Assert.IsNull(entry.EarnedAt);
            Assert.AreEqual("medal.hundred_wins.description", entry.Medal.descriptionKey);
        }

        [Test]
        public void AMedalTheServerDidNotListIsNotShownEvenWhenOwned()
        {
            var collection = Collection();

            Assert.IsFalse(collection.TryGet("medal.switched_off", out _));
            Assert.IsFalse(collection.TryGet("medal.unknown", out _));
            Assert.IsFalse(collection.TryGet(null, out _));
        }

        [Test]
        public void TheShowcaseKeepsTheServerOrderOfShownEarnedMedals()
        {
            Assert.AreEqual(new[] { "medal.moon_event", "medal.first_win" }, Showcase(Collection()));
        }

        [Test]
        public void AddingAnEarnedMedalAsksTheServerWithItAppended()
        {
            var collection = Collection();

            Assert.IsTrue(collection.TryAddToShowcase("medal.domino_ten", out var request));

            Assert.AreEqual(new[] { "medal.moon_event", "medal.first_win", "medal.domino_ten" }, request.medals);
            Assert.AreEqual(new[] { "medal.moon_event", "medal.first_win" }, Showcase(collection));
        }

        [Test]
        public void ALockedShowcasedOrUnknownMedalIsNeverAsked()
        {
            var collection = Collection();

            Assert.IsFalse(collection.TryAddToShowcase("medal.hundred_wins", out var locked));
            Assert.IsFalse(collection.TryAddToShowcase("medal.first_win", out var twice));
            Assert.IsFalse(collection.TryAddToShowcase("medal.switched_off", out var hidden));
            Assert.IsFalse(collection.TryAddToShowcase(null, out var none));
            Assert.IsNull(locked);
            Assert.IsNull(twice);
            Assert.IsNull(hidden);
            Assert.IsNull(none);
        }

        [Test]
        public void AFullShowcaseTakesNoMoreMedals()
        {
            var twoSlots = Collection(CollectionJson.Replace(@"""slots"":3", @"""slots"":2"));
            var noSlots = Collection(CollectionJson.Replace(@"""slots"":3", @"""slots"":0"));

            Assert.AreEqual(2, twoSlots.ShowcaseCount);
            Assert.IsFalse(twoSlots.TryAddToShowcase("medal.domino_ten", out _));
            Assert.IsFalse(noSlots.TryAddToShowcase("medal.domino_ten", out _));

            twoSlots.Apply(MedalShowcaseReply.FromJson(@"{""showcase"":[""medal.first_win""]}"));

            Assert.IsTrue(twoSlots.TryAddToShowcase("medal.domino_ten", out _));
        }

        [Test]
        public void RemovingAMedalAsksTheServerWithoutIt()
        {
            var collection = Collection();

            Assert.IsTrue(collection.TryRemoveFromShowcase("medal.moon_event", out var request));
            Assert.AreEqual(new[] { "medal.first_win" }, request.medals);
            Assert.IsFalse(collection.TryRemoveFromShowcase("medal.domino_ten", out _));
        }

        [Test]
        public void MovingAMedalAsksTheServerForTheNewOrder()
        {
            var collection = Collection();
            collection.Apply(MedalShowcaseReply.FromJson(@"{""showcase"":[""medal.moon_event"",""medal.first_win"",""medal.domino_ten""]}"));

            Assert.IsTrue(collection.TryMoveInShowcase(2, 0, out var request));
            Assert.AreEqual(new[] { "medal.domino_ten", "medal.moon_event", "medal.first_win" }, request.medals);
            Assert.IsFalse(collection.TryMoveInShowcase(1, 1, out _));
            Assert.IsFalse(collection.TryMoveInShowcase(0, 3, out _));
            Assert.IsFalse(collection.TryMoveInShowcase(-1, 0, out _));
        }

        [Test]
        public void TheShowcaseFollowsTheServerReply()
        {
            var collection = Collection();

            collection.Apply(MedalShowcaseReply.FromJson(@"{""showcase"":[""medal.first_win"",""medal.hundred_wins"",""medal.first_win""]}"));
            Assert.AreEqual(new[] { "medal.first_win" }, Showcase(collection));

            collection.Apply(MedalShowcaseReply.FromJson(@"{""error"":""medal.not_owned""}"));
            Assert.AreEqual(new[] { "medal.first_win" }, Showcase(collection));
        }

        [Test]
        public void TheShowcaseIntentIsTheServerRequestShape()
        {
            Assert.IsTrue(Collection().TryAddToShowcase("medal.domino_ten", out var request));

            Assert.AreEqual(@"{""medals"":[""medal.moon_event"",""medal.first_win"",""medal.domino_ten""]}", request.ToJson());
            Assert.AreEqual("/medals", MedalCollection.CollectionPath);
            Assert.AreEqual("/medals/showcase", MedalCollection.ShowcasePath);
        }

        [Test]
        public void AnEmptyReplyShowsNothing()
        {
            var collection = new MedalCollection();
            collection.Show(null);

            Assert.AreEqual(0, collection.ShownCount);
            Assert.AreEqual(0, collection.EarnedCount);
            Assert.AreEqual(0, collection.Slots);
            Assert.AreEqual(0, collection.ShowcaseCount);
        }

        [Test]
        public void ReadingTheCollectionAllocatesNothing()
        {
            var collection = Collection();
            TestDelegate read = () =>
            {
                collection.TryGet("medal.first_win", out _);
                collection.TryGet("medal.hundred_wins", out _);
                for (var at = 0; at < collection.ShowcaseCount; at++) collection.ShowcaseAt(at);
                collection.IsShowcased("medal.domino_ten");
            };
            read();

            Assert.That(read, Allocates.Nothing);
        }
    }
}
