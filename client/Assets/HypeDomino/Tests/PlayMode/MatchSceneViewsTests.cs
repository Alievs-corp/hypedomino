using System;
using System.Collections.Generic;
using Google.Protobuf;
using HypeDomino.Game;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Emoji;
using HypeDomino.UI.Graphics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Match;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Rarity;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class MatchSceneViewsTests
    {
        const ulong ServerNowMs = 1_000_000;
        const string FoxStatic = "cosmetic/avatar/fox/static";
        const string NeonStatic = "cosmetic/table/neon/static";
        const string Laugh = "emoji.laugh";
        const string LaughStatic = "cosmetic/emoji/laugh/static";
        const string Legendary = "legendary";
        const float Tolerance = 0.001f;
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[
{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""},
{""id"":""table.neon"",""type"":""table"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/table/neon/static""},
{""id"":""frame.gold"",""type"":""frame"",""rarityId"":""legendary"",""availability"":""live"",""staticAddress"":""cosmetic/frame/gold/static""},
{""id"":""emoji.laugh"",""type"":""emoji"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/emoji/laugh/static""}]}";

        sealed class ServerFrames : IFrameTransport
        {
            readonly Queue<byte[]> pending = new();

            public void Send(ReadOnlySpan<byte> frame)
            {
            }

            public bool TryReceive(out ReadOnlySpan<byte> frame)
            {
                var has = pending.Count > 0;
                frame = has ? pending.Dequeue() : null;
                return has;
            }

            public void Deliver(Envelope.PayloadOneofCase payload, IMessage message) =>
                pending.Enqueue(new FrameWriter(Session.MaxSentFrameBytes).Write(new FrameHeader(Session.ProtocolVersion, 0, 0, 1), payload, message).ToArray());
        }

        GameObject root;
        MatchLayoutView layout;
        BoardTilesView boardTiles;
        MatchHudViewTests.HudParts hud;
        MatchSceneViews views;
        SeatEmojiLayer emojiLayer;
        AvatarView[] seatViews;
        RarityHalo[] halos;
        ActiveRarityTable rarities;
        FakeArtSource<Sprite> stills;
        ServerFrames frames;
        Connection connection;
        AppFlow flow;
        MatchHost host;
        ActiveCosmeticManifest manifest;
        GameStrings strings;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Match", typeof(RectTransform), typeof(Canvas));
            ((RectTransform)root.transform).sizeDelta = new Vector2(1920f, 1080f);
            layout = NewChild<MatchLayoutView>("Layout");
            ((RectTransform)layout.transform).sizeDelta = new Vector2(1920f, 1080f);
            var tiles = NewChild<BoardTilesView>("Board");
            hud = MatchHudViewTests.NewHud(root.transform, 1);
            tiles.Configure(hud.Faces, Array.Empty<Image>(), null, Array.Empty<Image>());
            boardTiles = tiles;
            var seats = new AvatarView[TableSides.Count];
            halos = new RarityHalo[TableSides.Count];
            for (var side = 0; side < seats.Length; side++)
            {
                seats[side] = NewChild<AvatarView>("Seat " + side);
                halos[side] = NewHalo(seats[side].transform);
                seats[side].Configure(NewArt("Avatar", seats[side].transform), NewArt("Frame", seats[side].transform), halos[side], null);
                seats[side].transform.localPosition = new Vector3(100f * side, 50f, 0f);
            }
            seatViews = seats;
            var loading = NewChild<LoadingScreenView>("Loading");
            loading.Configure(Array.Empty<LoadingCardView>(), NewImage("Total", loading.transform), NewTable(loading.transform));
            rarities = new ActiveRarityTable(RarityTable.LoadBuiltin());
            views = NewChild<MatchSceneViews>("Views");
            emojiLayer = NewChild<SeatEmojiLayer>("Emoji");
            var popup = new GameObject("Popup", typeof(RectTransform), typeof(Image)).AddComponent<EmojiPopup>();
            popup.transform.SetParent(root.transform, false);
            views.Configure(layout, NewScorePanel(), tiles, seats, NewTable(root.transform), loading, hud.View, emojiLayer, popup, NewSprite(), NewSprite());
            stills = new FakeArtSource<Sprite>();
            frames = new ServerFrames();
            connection = new Connection(() => throw new InvalidOperationException(), () => new ClientHello(), () => 0d, new System.Random(1));
            flow = new AppFlow();
            connection.Session.HelloReceived += flow.Hello;
            connection.Session.ViewReceived += FollowView;
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var parsed));
            manifest = new ActiveCosmeticManifest(parsed);
            host = new MatchHost(flow, connection, manifest);
            strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(GameLanguage.English, () => Array.Empty<string>()), false);
            connection.Session.Hello(frames, new ClientHello());
            frames.Deliver(Envelope.PayloadOneofCase.ServerHello, new ServerHello { Status = HelloStatus.Accepted, SessionNonce = 1, ServerTimeUnixMs = ServerNowMs });
            connection.Session.Poll();
        }

        [TearDown]
        public void Destroy()
        {
            views.Detach();
            Object.DestroyImmediate(root);
            Object.DestroyImmediate(hud.Faces);
        }

        void FollowView(MatchSnapshot view) => flow.View(view.State);

        void Attach() => views.Attach(host, connection, manifest, rarities, strings, stills, new FakeArtSource<LoopArt>(), new GraphicsQuality(DeviceTier.Mid, null, 0, false), new MotionPreference(null, false));

        T NewChild<T>(string name) where T : Component
        {
            var child = new GameObject(name, typeof(RectTransform)).AddComponent<T>();
            child.transform.SetParent(root.transform, false);
            return child;
        }

        static Sprite NewSprite() => Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), Vector2.zero);

        static RarityHalo NewHalo(Transform parent)
        {
            var halo = new GameObject("Halo", typeof(RectTransform), typeof(Image)).AddComponent<RarityHalo>();
            halo.transform.SetParent(parent, false);
            halo.Configure(NewSprite(), NewSprite());
            halo.GetComponent<Image>().enabled = false;
            return halo;
        }

        static Image NewImage(string name, Transform parent)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).GetComponent<Image>();
            image.transform.SetParent(parent, false);
            image.type = Image.Type.Filled;
            return image;
        }

        static TMP_Text NewText(string name, Transform parent)
        {
            var text = new GameObject(name, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            text.transform.SetParent(parent, false);
            return text;
        }

        static CosmeticImage NewArt(string name, Transform parent)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<AnimatedImage>().gameObject.AddComponent<CosmeticImage>();
            image.transform.SetParent(parent, false);
            return image;
        }

        static TableView NewTable(Transform parent)
        {
            var built = new GameObject("Table", typeof(RectTransform)).AddComponent<TableView>();
            built.transform.SetParent(parent, false);
            var layer = new GameObject("Layer", typeof(RectTransform), typeof(Image)).AddComponent<AnimatedImage>();
            layer.transform.SetParent(built.transform, false);
            built.Configure(NewImage("Flat", built.transform), new[] { layer }, NewSprite());
            return built;
        }

        ScorePanelView NewScorePanel()
        {
            var panel = NewChild<ScorePanelView>("Score");
            panel.Configure(NewTeam("Mine", panel.transform), NewTeam("Theirs", panel.transform), NewText("Target", panel.transform));
            return panel;
        }

        static TeamScoreView NewTeam(string name, Transform parent)
        {
            var team = new GameObject(name, typeof(RectTransform)).AddComponent<TeamScoreView>();
            team.transform.SetParent(parent, false);
            var badge = new GameObject("Lock", typeof(RectTransform));
            badge.transform.SetParent(team.transform, false);
            team.Configure(NewText("Score", team.transform), NewImage("Gem", team.transform), NewImage("Meter", team.transform), badge);
            return team;
        }

        static MatchView View(MatchState state, Mode mode)
        {
            var two = mode == Mode._1V1;
            var view = new MatchView
            {
                State = state,
                Mode = mode,
                Seat = two ? Seat.X : Seat.South,
                Current = two ? Seat.X : Seat.South,
                Table = "table.neon",
                TurnDeadlineUnixMs = ServerNowMs + 10_000,
                TurnSeconds = 15,
                TurnWarningSeconds = 5,
            };
            view.Roster.Add(new RosterEntry { Seat = two ? Seat.Y : Seat.North, Avatar = "avatar.fox", Frame = "frame.gold" });
            view.Hand.Add(new Hypedomino.V1.Tile { Low = 2, High = 5 });
            return view;
        }

        void Receive(MatchView view)
        {
            frames.Deliver(Envelope.PayloadOneofCase.MatchView, view);
            connection.Session.Poll();
        }

        [Test]
        public void TheViewsAreNotReadyUntilAttached()
        {
            Assert.IsFalse(views.Ready);

            Attach();

            Assert.IsTrue(views.Ready);
            Assert.IsNotNull(views.Table);
            Assert.IsNotNull(views.LoadingScreen);
            Assert.IsNotNull(views.Stills);
            Assert.IsNotNull(views.Emoji);
        }

        [Test]
        public void AttachedViewsLetTheHostBuildTheMatchAndShowTheTableAndSeats()
        {
            Attach();

            Receive(View(MatchState.Playing, Mode._1V1));

            Assert.IsTrue(host.Live);
            CollectionAssert.Contains(stills.Loads, NeonStatic);
            CollectionAssert.Contains(stills.Loads, FoxStatic);
        }

        [Test]
        public void TheLayoutTakesTheSeatCountOfTheMode()
        {
            Attach();

            Receive(View(MatchState.Playing, Mode._1V1));
            Assert.AreEqual(2, layout.SeatCount);

            Receive(View(MatchState.Playing, Mode._2V2));
            Assert.AreEqual(4, layout.SeatCount);
        }

        [Test]
        public void TheHudShowsTheHandOfTheHostedMatch()
        {
            Attach();

            Receive(View(MatchState.Playing, Mode._2V2));
            hud.View.Refresh();

            Assert.IsTrue(hud.Tiles[0].gameObject.activeSelf);
            Assert.AreSame(hud.Faces.Face(new Domino(2, 5)), hud.Tiles[0].image.sprite);
        }

        [TestCase(DeviceTier.Low, TokenEffectLevel.LowMaxTweens)]
        [TestCase(DeviceTier.Mid, TokenEffectLevel.MediumMaxTweens)]
        public void AttachingGivesTheBoardATilePlaceLayerAtTheEffectLevel(DeviceTier tier, int capacity)
        {
            views.Attach(host, connection, manifest, rarities, strings, stills, new FakeArtSource<LoopArt>(), new GraphicsQuality(tier, null, 0, false), new MotionPreference(null, false));

            Assert.IsNotNull(views.TilePlacing);
            Assert.AreSame(boardTiles.GetComponent<TilePlaceLayer>(), views.TilePlacing);
            Assert.AreEqual(capacity, views.TilePlacing.Capacity);
        }

        [Test]
        public void ReattachingKeepsOneTilePlaceLayer()
        {
            Attach();
            var first = views.TilePlacing;

            views.Detach();
            Attach();

            Assert.AreSame(first, views.TilePlacing);
            Assert.AreEqual(1, boardTiles.GetComponents<TilePlaceLayer>().Length);
        }

        [Test]
        public void DetachingTearsTheMatchDown()
        {
            Attach();
            Receive(View(MatchState.Playing, Mode._1V1));

            views.Detach();
            Receive(View(MatchState.Playing, Mode._1V1));

            Assert.IsFalse(host.Live);
            Assert.IsFalse(views.Ready);
        }

        static InventoryReply WornLaugh(int emojiSlots) => new() { loadout = new LoadoutReply { emojis = new[] { Laugh } }, emoji_slots = emojiSlots };

        [Test]
        public void TheEmojiWornBeforeAttachingArePrefetchedWhenTheMatchBegins()
        {
            views.Wear(WornLaugh(6));
            Attach();

            Receive(View(MatchState.Playing, Mode._2V2));

            CollectionAssert.Contains(stills.Loads, LaughStatic);
        }

        [Test]
        public void AnInventoryWithMoreSlotsArrivingAfterAttachReachesTheNextMatch()
        {
            Attach();
            var before = views.Emoji;

            views.Wear(WornLaugh(6));
            Receive(View(MatchState.Playing, Mode._2V2));

            Assert.AreNotSame(before, views.Emoji);
            Assert.AreSame(views.Emoji, host.Root.Emoji);
            CollectionAssert.Contains(stills.Loads, LaughStatic);
        }

        [Test]
        public void NoInventoryLeavesTheWheelEmpty()
        {
            Attach();

            Receive(View(MatchState.Playing, Mode._2V2));

            CollectionAssert.DoesNotContain(stills.Loads, LaughStatic);
        }

        [Test]
        public void AnEmojiShownAtASeatPopsOverThatSeat()
        {
            views.Wear(WornLaugh(6));
            Attach();
            Receive(View(MatchState.Playing, Mode._2V2));
            stills.Finish(LaughStatic, NewSprite());

            frames.Deliver(Envelope.PayloadOneofCase.EmojiShown, new EmojiShown { Seat = Seat.West, ItemId = Laugh });
            connection.Session.Poll();

            Assert.AreEqual(1, emojiLayer.LiveCount);
            var shown = 0;
            for (var side = 0; side < seatViews.Length; side++)
            {
                var popup = emojiLayer.LiveAt(side);
                if (popup == null) continue;
                shown++;
                Assert.AreEqual(seatViews[side].transform.position, popup.transform.position);
                Assert.IsNotNull(popup.Shown);
            }
            Assert.AreEqual(1, shown);
        }

        static RarityTable Delivered(string fill, float glowAlpha) => RarityTable.From(new RarityTableFile
        {
            diamondMinSize = 72,
            flowPeriodMs = 2000,
            tiers = new[]
            {
                new RarityTierEntry { id = Legendary, ordinal = 1, nameKey = "rarity.legendary.name", fill = fill, on = "#000000", text = "#FFFFFF", pips = 1, glowAlpha = glowAlpha, glowScale = 1.25f },
            },
        });

        RarityHalo ShownHalo()
        {
            RarityHalo found = null;
            foreach (var halo in halos)
            {
                if (!halo.IsShown) continue;
                Assert.IsNull(found);
                found = halo;
            }
            Assert.IsNotNull(found);
            return found;
        }

        static void AssertTint(Color fill, float glowAlpha, RarityHalo halo)
        {
            var tint = halo.GetComponent<Image>().color;
            Assert.AreEqual(fill.r, tint.r, Tolerance);
            Assert.AreEqual(fill.g, tint.g, Tolerance);
            Assert.AreEqual(fill.b, tint.b, Tolerance);
            Assert.AreEqual(glowAlpha, tint.a, Tolerance);
        }

        [Test]
        public void WithoutADeliveredTableTheSeatHaloUsesTheShippedDefaults()
        {
            Assert.IsTrue(RarityTable.LoadBuiltin().TryGet(Legendary, out var shipped));
            Attach();

            Receive(View(MatchState.Playing, Mode._2V2));

            AssertTint(shipped.Fill, shipped.GlowAlpha, ShownHalo());
        }

        [Test]
        public void ADeliveredTableColoursTheSeatHalo()
        {
            rarities.Apply(Delivered("#00FF00", 0.6f));
            Attach();

            Receive(View(MatchState.Playing, Mode._2V2));

            var halo = ShownHalo();
            AssertTint(Color.green, 0.6f, halo);
            Assert.AreEqual(1.25f, halo.transform.localScale.x, Tolerance);
        }

        [Test]
        public void ATableDeliveredAfterAttachReachesTheSeatsAtTheNextAttach()
        {
            Assert.IsTrue(RarityTable.LoadBuiltin().TryGet(Legendary, out var shipped));
            Attach();
            rarities.Apply(Delivered("#00FF00", 0.6f));

            Receive(View(MatchState.Playing, Mode._2V2));
            AssertTint(shipped.Fill, shipped.GlowAlpha, ShownHalo());

            views.Detach();
            Attach();
            Receive(View(MatchState.Playing, Mode._2V2));
            AssertTint(Color.green, 0.6f, ShownHalo());
        }

        [Test]
        public void AttachingNeedsARarityTable()
        {
            Assert.Throws<ArgumentNullException>(() => views.Attach(host, connection, manifest, null, strings, stills, new FakeArtSource<LoopArt>(), new GraphicsQuality(DeviceTier.Mid, null, 0, false), new MotionPreference(null, false)));

            Assert.IsFalse(views.Ready);
        }
    }
}
