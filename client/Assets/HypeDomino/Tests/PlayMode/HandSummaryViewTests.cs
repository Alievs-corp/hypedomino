using System;
using Google.Protobuf;
using HypeDomino.Game;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Cosmetics;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class HandSummaryViewTests
    {
        const int Rows = 4;
        const int TilesPerRow = 3;
        const string Catalog = @"{""schemaVersion"":1,""manifestVersion"":1,""items"":[{""id"":""avatar.fox"",""type"":""avatar"",""rarityId"":""common"",""availability"":""live"",""staticAddress"":""cosmetic/avatar/fox/static""}]}";

        GameObject root;
        MatchHudViewTests.HudParts hud;
        HandSummaryView view;
        GameObject domino;
        GameObject block;
        TMP_Text pipsA;
        TMP_Text pipsB;
        TMP_Text counted;
        TMP_Text handPoints;
        TMP_Text carriedIn;
        TMP_Text carryOver;
        GameObject lockBadge;
        GameObject[] rows;
        Image[] tiles;
        HandSummary summary;
        readonly MatchSnapshot snapshot = new();

        [SetUp]
        public void Build()
        {
            root = new GameObject("Match", typeof(RectTransform), typeof(Canvas));
            summary = new HandSummary();
            hud = MatchHudViewTests.NewHud(root.transform, 1);
            view = new GameObject("Summary", typeof(RectTransform)).AddComponent<HandSummaryView>();
            view.transform.SetParent(root.transform, false);
            var parent = view.transform;
            domino = NewObject("Domino", parent);
            block = NewObject("Block", parent);
            pipsA = NewText("Pips A", parent);
            pipsB = NewText("Pips B", parent);
            counted = NewText("Counted", parent);
            handPoints = NewText("Hand points", parent);
            carriedIn = NewText("Carried in", parent);
            carryOver = NewText("Carry over", parent);
            lockBadge = NewObject("Lock", parent);
            rows = new GameObject[Rows];
            tiles = new Image[Rows * TilesPerRow];
            for (var row = 0; row < Rows; row++)
            {
                rows[row] = NewObject("Row " + row, parent);
                for (var slot = 0; slot < TilesPerRow; slot++)
                {
                    var image = new GameObject("Tile " + slot, typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                    image.transform.SetParent(rows[row].transform, false);
                    tiles[row * TilesPerRow + slot] = image;
                }
            }
            view.Configure(hud.Faces, domino, block, pipsA, pipsB, counted, handPoints, carriedIn, carryOver, lockBadge, rows, tiles);
        }

        [TearDown]
        public void Destroy()
        {
            Object.DestroyImmediate(root);
            Object.DestroyImmediate(hud.Faces);
        }

        static GameObject NewObject(string name, Transform parent)
        {
            var child = new GameObject(name, typeof(RectTransform));
            child.transform.SetParent(parent, false);
            return child;
        }

        static TMP_Text NewText(string name, Transform parent)
        {
            var text = new GameObject(name, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            text.transform.SetParent(parent, false);
            return text;
        }

        static Hypedomino.V1.Tile T(uint low, uint high) => new() { Low = low, High = high };

        static MatchView Ended(uint hands, HandResult result)
        {
            var view = new MatchView { State = MatchState.Playing, Seat = Seat.South, Mode = Mode._2V2, HandsPlayed = hands, LastResult = result };
            result.Left.Add(new SeatTiles { Seat = Seat.North, Tiles = { T(6, 6) } });
            result.Left.Add(new SeatTiles { Seat = Seat.East });
            result.Left.Add(new SeatTiles { Seat = Seat.South, Tiles = { T(1, 2), T(0, 3) } });
            return view;
        }

        void Show(MatchView ended)
        {
            Assert.IsTrue(snapshot.TryRead(ended.ToByteArray()));
            Assert.IsTrue(summary.Read(snapshot));
            view.Show(summary);
        }

        Sprite FaceOf(byte first, byte second) => hud.Faces.Face(new Domino(first, second));

        [Test]
        public void ADominoShowsThePipsTheRoundedHandPointsAndTheCarryOverAdded()
        {
            Show(Ended(1, new HandResult
            {
                Ending = HandEnding.Domino,
                Winner = Team.B,
                Pips = new TeamPoints { TeamA = 13, TeamB = 6 },
                Awarded = new TeamPoints { TeamB = 50 },
                CarriedIn = 35,
            }));

            Assert.IsTrue(view.gameObject.activeSelf);
            Assert.IsTrue(domino.activeSelf);
            Assert.IsFalse(block.activeSelf);
            Assert.AreEqual("13", pipsA.text);
            Assert.AreEqual("6", pipsB.text);
            Assert.AreEqual("13", counted.text);
            Assert.AreEqual("15", handPoints.text);
            Assert.IsTrue(carriedIn.gameObject.activeSelf);
            Assert.AreEqual("35", carriedIn.text);
            Assert.IsFalse(carryOver.gameObject.activeSelf);
            Assert.IsFalse(lockBadge.activeSelf);
            Assert.AreEqual((Color)TokenColor.RoleScoreText, handPoints.color);
        }

        [Test]
        public void TheTilesLeftShowPerSeatRow()
        {
            Show(Ended(1, new HandResult { Ending = HandEnding.Domino, Winner = Team.A, Awarded = new TeamPoints { TeamA = 10 } }));

            CollectionAssert.AreEqual(new[] { true, true, true, false }, Array.ConvertAll(rows, row => row.activeSelf));
            Assert.IsTrue(tiles[0].gameObject.activeSelf);
            Assert.AreSame(FaceOf(6, 6), tiles[0].sprite);
            Assert.IsFalse(tiles[1].gameObject.activeSelf);
            Assert.IsFalse(tiles[TilesPerRow].gameObject.activeSelf);
            Assert.AreSame(FaceOf(1, 2), tiles[2 * TilesPerRow].sprite);
            Assert.AreSame(FaceOf(0, 3), tiles[2 * TilesPerRow + 1].sprite);
            Assert.IsFalse(tiles[2 * TilesPerRow + 2].gameObject.activeSelf);
        }

        [Test]
        public void ALockedWinnerShowsTheLockInsteadOfHandPoints()
        {
            Show(Ended(1, new HandResult
            {
                Ending = HandEnding.Block,
                Winner = Team.A,
                Pips = new TeamPoints { TeamA = 4, TeamB = 12 },
                HandPointsLocked = true,
            }));

            Assert.IsTrue(block.activeSelf);
            Assert.IsFalse(domino.activeSelf);
            Assert.IsTrue(lockBadge.activeSelf);
            Assert.IsFalse(handPoints.gameObject.activeSelf);
            Assert.IsFalse(carriedIn.gameObject.activeSelf);
            Assert.IsTrue(counted.gameObject.activeSelf);
        }

        [Test]
        public void ABlockWithEqualTotalsShowsTheCarryOverKept()
        {
            Show(Ended(1, new HandResult
            {
                Ending = HandEnding.Block,
                Pips = new TeamPoints { TeamA = 9, TeamB = 9 },
                CarryOver = 20,
            }));

            Assert.IsTrue(carryOver.gameObject.activeSelf);
            Assert.AreEqual("20", carryOver.text);
            Assert.IsFalse(counted.gameObject.activeSelf);
            Assert.IsFalse(handPoints.gameObject.activeSelf);
            Assert.IsFalse(lockBadge.activeSelf);
        }

        [Test]
        public void TheHudOpensTheSummaryFromTheHostAndClosesItWhenCleared()
        {
            var connection = new Connection(() => throw new InvalidOperationException(), () => new ClientHello(), () => 0d, new System.Random(1));
            Assert.IsTrue(CosmeticManifest.TryParse(Catalog, out var manifest));
            var host = new MatchHost(new AppFlow(), connection, new ActiveCosmeticManifest(manifest));
            var strings = new GameStrings(BuiltinStrings.Load(), LanguageChoice.Load(GameLanguage.English, () => Array.Empty<string>()), false);
            hud.View.Configure(hud.Faces, hud.Tiles, hud.Marks, hud.Draw, hud.Pass, hud.EndTurn, hud.Ring, hud.Network, hud.Afk, hud.TakeBack, hud.TakeBackText,
                hud.Badges, hud.PassLabels, view);
            hud.View.Bind(host, connection, strings);
            Assert.IsFalse(view.gameObject.activeSelf);

            Assert.IsTrue(snapshot.TryRead(Ended(1, new HandResult { Ending = HandEnding.Domino, Winner = Team.A, Awarded = new TeamPoints { TeamA = 10 } }).ToByteArray()));
            Assert.IsTrue(summary.Read(snapshot));
            host.Hud.ShowSummary(summary);
            hud.View.Refresh();
            Assert.IsTrue(view.gameObject.activeSelf);

            host.Hud.Clear();
            hud.View.Refresh();
            Assert.IsFalse(view.gameObject.activeSelf);
            hud.View.Unbind();
        }
    }
}
