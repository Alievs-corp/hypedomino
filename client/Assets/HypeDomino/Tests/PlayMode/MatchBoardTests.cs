using System.IO;
using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Effects;
using HypeDomino.UI.Match;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class MatchBoardTests
    {
        const float PlaceMs = TokenDurationMs.Base + TokenDurationMs.Micro;
        const float Tolerance = 0.01f;
        static readonly Vector2 BoardSize = new Vector2(1200f, 600f);

        readonly MatchSnapshot snapshot = new();
        GameObject root;
        BoardTilesView tiles;
        TilePlaceLayer placing;
        Transform[] origins;
        BoardChain chain;
        BoardPlacements placements;
        MatchBoard board;
        float nowMs;
        int tilePlaces;
        MatchSounds otherSounds;

        [SetUp]
        public void Build()
        {
            root = new GameObject("Match", typeof(RectTransform), typeof(Canvas));
            var boardRect = new GameObject("Board", typeof(RectTransform)).GetComponent<RectTransform>();
            boardRect.SetParent(root.transform, false);
            boardRect.sizeDelta = BoardSize;
            var faces = new Sprite[TileFaces.SetSize];
            for (var i = 0; i < faces.Length; i++) faces[i] = Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), new Vector2(0.5f, 0.5f));
            var set = ScriptableObject.CreateInstance<TileFaceSet>();
            set.SetFaces(faces);
            var shadows = NewImages("Shadow", boardRect);
            var images = NewImages("Tile", boardRect);
            tiles = boardRect.gameObject.AddComponent<BoardTilesView>();
            tiles.Configure(set, images, Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), new Vector2(0.5f, 0.5f)), shadows);
            placing = boardRect.gameObject.AddComponent<TilePlaceLayer>();
            var moved = new Transform[TileFaces.SetSize];
            var cast = new Transform[TileFaces.SetSize];
            for (var slot = 0; slot < moved.Length; slot++)
            {
                moved[slot] = tiles.TileAt(slot);
                cast[slot] = tiles.ShadowAt(slot);
            }
            placing.Build(moved, cast);
            origins = new Transform[TableSides.Count];
            for (var side = 0; side < origins.Length; side++)
            {
                origins[side] = new GameObject("Seat" + side, typeof(RectTransform)).transform;
                origins[side].SetParent(root.transform, false);
                origins[side].localPosition = new Vector3(-500f + 300f * side, -400f + 100f * side, 0f);
            }
            chain = new BoardChain();
            placements = new BoardPlacements();
            nowMs = 0f;
            tilePlaces = 0;
            otherSounds = MatchSounds.None;
            board = new MatchBoard(tiles, placing, origins, null, () => nowMs, Hear);
        }

        [TearDown]
        public void DestroyBoard()
        {
            board.Dispose();
            Object.DestroyImmediate(root);
        }

        void Hear(MatchSounds sounds)
        {
            if (sounds == MatchSounds.TilePlace) tilePlaces++;
            else otherSounds |= sounds;
        }

        static Image[] NewImages(string name, Transform parent)
        {
            var created = new Image[TileFaces.SetSize];
            for (var i = 0; i < created.Length; i++)
            {
                created[i] = new GameObject(name + i, typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                created[i].transform.SetParent(parent, false);
            }
            return created;
        }

        static OrientedTile O(uint first, uint second) => new() { First = first, Second = second };

        static MatchView View(uint south, uint west, uint north, uint east, params OrientedTile[] line)
        {
            var view = new MatchView { State = MatchState.Playing, Mode = Mode._2V2, Seat = Seat.South, HandsPlayed = 1, Board = new BoardView() };
            view.Board.Line.Add(line);
            view.Seats.Add(new SeatView { Seat = Seat.North, HandSize = north });
            view.Seats.Add(new SeatView { Seat = Seat.East, HandSize = east });
            view.Seats.Add(new SeatView { Seat = Seat.South, HandSize = south });
            view.Seats.Add(new SeatView { Seat = Seat.West, HandSize = west });
            return view;
        }

        void Show(MatchView view) => Show(view.ToByteArray());

        void Show(byte[] view)
        {
            Assert.IsTrue(snapshot.TryRead(view));
            chain.Arrange(snapshot, BoardSize, float.MaxValue);
            placements.Read(snapshot);
            board.Show(chain.Layout, placements);
        }

        void Tick(float atMs)
        {
            nowMs = atMs;
            placing.Tick(atMs);
        }

        static int Slot(uint first, uint second) => TileFaces.Index(new Domino((byte)first, (byte)second));

        Vector3 OriginOnBoard(TableSide side) => tiles.transform.InverseTransformPoint(origins[(int)side].position);

        [Test]
        public void TheFirstViewSnapsEveryTileAndPlaysNothing()
        {
            Show(View(6, 6, 6, 6, O(2, 3), O(3, 6)));

            Assert.AreEqual(0, placing.MovingCount);
            Assert.IsFalse(placing.IsSettling(Slot(2, 3)));
            Tick(PlaceMs * 2f);
            Assert.AreEqual(0, tilePlaces);
        }

        [Test]
        public void ANewTileTravelsFromItsSeatAndSoundsWhenItSettlesNotWhenItsViewArrives()
        {
            Show(View(6, 6, 6, 6, O(2, 3)));

            Show(View(6, 5, 6, 6, O(2, 3), O(3, 6)));

            var tile = tiles.TileAt(Slot(3, 6));
            Assert.Less((tile.localPosition - OriginOnBoard(TableSide.Left)).magnitude, Tolerance);
            Assert.AreEqual(0, tilePlaces, "the view arrived");
            Tick(TokenDurationMs.Base);
            Assert.AreEqual(0, tilePlaces, "landed, still settling");
            Tick(PlaceMs);
            Assert.AreEqual(1, tilePlaces);
            Assert.Less(((Vector2)tile.localPosition - tile.anchoredPosition).magnitude, Tolerance);
            Tick(PlaceMs * 3f);
            Assert.AreEqual(1, tilePlaces);
            Assert.AreEqual(MatchSounds.None, otherSounds);
        }

        [Test]
        public void EverySeatsPlacementSoundsOnceWhenItsTileSettles()
        {
            Show(View(6, 6, 6, 6, O(3, 3)));
            var at = 0f;

            Show(View(5, 6, 6, 6, O(3, 3), O(3, 4)));
            Tick(at += PlaceMs);
            Show(View(5, 5, 6, 6, O(1, 3), O(3, 3), O(3, 4)));
            Tick(at += PlaceMs);
            Show(View(5, 5, 5, 6, O(1, 3), O(3, 3), O(3, 4), O(4, 4)));
            Tick(at += PlaceMs);
            Show(View(5, 5, 5, 5, O(0, 1), O(1, 3), O(3, 3), O(3, 4), O(4, 4)));
            Tick(at += PlaceMs);

            Assert.AreEqual(4, tilePlaces);
        }

        [Test]
        public void TheLocalPlayersTileTravelsFromTheBottomSeat()
        {
            Show(View(6, 6, 6, 6, O(3, 3)));

            Show(View(5, 6, 6, 6, O(3, 3), O(3, 4)));

            Assert.Less((tiles.TileAt(Slot(3, 4)).localPosition - OriginOnBoard(TableSide.Bottom)).magnitude, Tolerance);
        }

        [Test]
        public void AReconnectSnapsTheTilesAndPlaysNothing()
        {
            Show(View(6, 6, 6, 6, O(2, 3)));
            Show(View(6, 5, 6, 6, O(2, 3), O(3, 6)));

            placements.StartConnection();
            Show(View(6, 5, 5, 6, O(2, 3), O(3, 6), O(6, 6)));
            Tick(PlaceMs * 2f);

            Assert.AreEqual(0, tilePlaces);
            Assert.AreEqual(0, placing.MovingCount);
            var snapped = tiles.TileAt(Slot(6, 6));
            Assert.Less(((Vector2)snapped.localPosition - snapped.anchoredPosition).magnitude, Tolerance);
        }

        [Test]
        public void WithReduceMotionTheTileStillTravelsInFastAndSoundsWhenItLands()
        {
            placing.Configure(EffectLevel.Medium, true);
            Show(View(6, 6, 6, 6, O(2, 3)));

            Show(View(6, 6, 5, 6, O(2, 3), O(3, 6)));

            Assert.Less((tiles.TileAt(Slot(3, 6)).localPosition - OriginOnBoard(TableSide.Top)).magnitude, Tolerance);
            Tick(TokenDurationMs.Fast - 1f);
            Assert.AreEqual(0, tilePlaces);
            Tick(TokenDurationMs.Fast);
            Assert.AreEqual(1, tilePlaces);
            Assert.AreEqual(Vector3.one, tiles.TileAt(Slot(3, 6)).localScale);
        }

        [Test]
        public void ATileStillMovingLandsOnItsNewSlotWhenTheNextPlacementLaysTheBoardOutAgain()
        {
            Show(View(6, 6, 6, 6, O(2, 3)));
            Show(View(6, 5, 6, 6, O(2, 3), O(3, 6)));
            Tick(TokenDurationMs.Base * 0.5f);

            Show(View(6, 5, 5, 6, O(1, 2), O(2, 3), O(3, 6)));
            Tick(PlaceMs * 2f);

            var first = tiles.TileAt(Slot(3, 6));
            var second = tiles.TileAt(Slot(1, 2));
            Assert.Less(((Vector2)first.localPosition - first.anchoredPosition).magnitude, Tolerance);
            Assert.Less(((Vector2)second.localPosition - second.anchoredPosition).magnitude, Tolerance);
            Assert.AreEqual(2, tilePlaces);
        }

        [Test]
        public void ANewHandSnapsTheBoardWithoutSoundingTheTilesStillMoving()
        {
            Show(View(6, 6, 6, 6, O(2, 3)));
            Show(View(6, 5, 6, 6, O(2, 3), O(3, 6)));

            var dealt = View(7, 7, 7, 7);
            dealt.HandsPlayed = 2;
            Show(dealt);
            Tick(PlaceMs * 2f);

            Assert.AreEqual(0, tilePlaces);
            Assert.AreEqual(0, placing.MovingCount);
        }

        [Test]
        public void DisposingStopsTheSound()
        {
            Show(View(6, 6, 6, 6, O(2, 3)));
            Show(View(6, 5, 6, 6, O(2, 3), O(3, 6)));

            board.Dispose();
            Tick(PlaceMs);

            Assert.AreEqual(0, tilePlaces);
        }

        EffectPlayer NewEffectPlayer()
        {
            var layer = new GameObject("Effects", typeof(RectTransform));
            layer.transform.SetParent(root.transform, false);
            var player = layer.AddComponent<EffectPlayer>();
            var library = EffectLibrary.Parse(File.ReadAllText(Path.Combine(Application.dataPath, "HypeDomino", "UI", "Effects", "effects.json")));
            var sprites = new Sprite[library.SpriteCount];
            for (var sprite = 0; sprite < sprites.Length; sprite++)
            {
                sprites[sprite] = Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), new Vector2(0.5f, 0.5f));
                sprites[sprite].name = library.SpriteName(sprite);
            }
            var shared = Canvas.GetDefaultCanvasMaterial();
            player.Build(library, sprites, shared, shared, EffectLevel.High, false);
            return player;
        }

        ScorePopupView NewPopups()
        {
            var view = new GameObject("Popups", typeof(RectTransform)).AddComponent<ScorePopupView>();
            view.transform.SetParent(root.transform, false);
            var texts = new TMP_Text[TokenEffect.ScorePopupPooled];
            for (var at = 0; at < texts.Length; at++)
            {
                texts[at] = new GameObject("Popup " + at, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
                texts[at].transform.SetParent(view.transform, false);
            }
            view.Configure(texts);
            return view;
        }

        RectTransform NewAnchor(string name)
        {
            var anchor = new GameObject(name, typeof(RectTransform)).GetComponent<RectTransform>();
            anchor.SetParent(root.transform, false);
            return anchor;
        }

        [Test]
        public void TheSceneEffectsPlayTheDustOfAPlacementAndReadEveryViewAfterTheBoard()
        {
            var player = NewEffectPlayer();
            placing.Attach(player);
            var effects = new MatchEffects(player, placing, (RectTransform)tiles.transform, NewAnchor("Banner"), NewAnchor("Mine"), NewAnchor("Theirs"), NewPopups());
            board.Dispose();
            board = new MatchBoard(tiles, placing, origins, effects, () => nowMs, Hear);
            board.StartConnection();
            var first = View(6, 6, 6, 6, O(2, 3));
            Show(first);
            board.ShowEffects(snapshot);

            var scored = View(5, 6, 6, 6, O(2, 3), O(3, 6));
            scored.Scores = new TeamPoints { TeamA = 10 };
            Show(scored);
            board.ShowEffects(snapshot);

            Assert.AreEqual(1, player.PoolOf(TokenEffectKind.TilePlace).LiveCount);
            Assert.AreEqual(1, player.PoolOf(TokenEffectKind.ScoreCascade).LiveCount);
        }

        [Test]
        public void PlacingATileAndLettingItSettleAllocatesNothing()
        {
            var before = View(6, 6, 6, 6, O(2, 3)).ToByteArray();
            var after = View(6, 5, 6, 6, O(2, 3), O(3, 6)).ToByteArray();
            TestDelegate move = () =>
            {
                Show(before);
                Show(after);
                Tick(nowMs + 16f);
                Tick(nowMs + PlaceMs);
            };
            move();
            move();

            Assert.That(move, Allocates.Nothing);
            Assert.AreEqual(3, tilePlaces);
        }
    }
}
