using Google.Protobuf;
using HypeDomino.Game.Match;
using HypeDomino.Net;
using HypeDomino.UI.Match;
using HypeDomino.UI.Tokens;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Tests.PlayMode
{
    public sealed class ScorePanelViewTests
    {
        GameObject root;
        ScorePanelView panel;
        TeamParts mine;
        TeamParts theirs;
        TMP_Text target;
        ScoreBoard board;

        sealed class TeamParts
        {
            public TMP_Text Score;
            public Image Gem;
            public Image Meter;
            public GameObject Badge;
            public RectTransform Mark;
            public TeamScoreView View;
        }

        [SetUp]
        public void BuildPanel()
        {
            root = new GameObject("Screen", typeof(RectTransform), typeof(Canvas));
            mine = NewTeam("Mine");
            theirs = NewTeam("Theirs");
            target = NewText("Target", root.transform);
            panel = new GameObject("Score panel", typeof(RectTransform)).AddComponent<ScorePanelView>();
            panel.transform.SetParent(root.transform, false);
            panel.Configure(mine.View, theirs.View, target);
            board = new ScoreBoard();
        }

        [TearDown]
        public void DestroyPanel() => Object.Destroy(root);

        TeamParts NewTeam(string name)
        {
            var view = new GameObject(name, typeof(RectTransform)).AddComponent<TeamScoreView>();
            view.transform.SetParent(root.transform, false);
            var meter = NewImage("Meter", view.transform);
            meter.type = Image.Type.Filled;
            meter.fillMethod = Image.FillMethod.Horizontal;
            var mark = new GameObject("Lock mark", typeof(RectTransform)).GetComponent<RectTransform>();
            mark.SetParent(meter.transform, false);
            var badge = new GameObject("Lock", typeof(RectTransform));
            badge.transform.SetParent(view.transform, false);
            var parts = new TeamParts { Score = NewText("Score", view.transform), Gem = NewImage("Gem", view.transform), Meter = meter, Badge = badge, Mark = mark, View = view };
            view.Configure(parts.Score, parts.Gem, parts.Meter, parts.Badge, parts.Mark);
            return parts;
        }

        static TMP_Text NewText(string name, Transform parent)
        {
            var text = new GameObject(name, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            text.transform.SetParent(parent, false);
            return text;
        }

        static Image NewImage(string name, Transform parent)
        {
            var image = new GameObject(name, typeof(RectTransform), typeof(Image)).GetComponent<Image>();
            image.transform.SetParent(parent, false);
            return image;
        }

        static MatchSnapshot Snapshot(Seat seat, uint teamA, uint teamB, uint targetScore, params Team[] locked)
        {
            var view = new MatchView
            {
                State = MatchState.Playing,
                Seat = seat,
                Mode = Mode._2V2,
                Scores = new TeamPoints { TeamA = teamA, TeamB = teamB },
                TargetScore = targetScore,
                HandPointsLockThreshold = 300,
            };
            view.LockedTeams.Add(locked);
            var snapshot = new MatchSnapshot();
            Assert.IsTrue(snapshot.TryRead(view.ToByteArray()));
            return snapshot;
        }

        void Show(MatchSnapshot snapshot)
        {
            board.Read(snapshot);
            ScorePanel.Show(panel, board);
        }

        [Test]
        public void TheLocalTeamComesFirstInItsColourWithTheServerScores()
        {
            Show(Snapshot(Seat.East, 100, 310, 365, Team.B));

            Assert.AreEqual("310", mine.Score.text);
            Assert.AreEqual((Color)TokenColor.RoleTeamBFill, mine.Gem.color);
            Assert.AreEqual((Color)TokenColor.RoleTeamBFill, mine.Meter.color);
            Assert.AreEqual(310f / 365, mine.Meter.fillAmount, 1e-6f);
            Assert.IsTrue(mine.Badge.activeSelf);

            Assert.AreEqual("100", theirs.Score.text);
            Assert.AreEqual((Color)TokenColor.RoleTeamAFill, theirs.Gem.color);
            Assert.AreEqual((Color)TokenColor.RoleTeamAFill, theirs.Meter.color);
            Assert.AreEqual(100f / 365, theirs.Meter.fillAmount, 1e-6f);
            Assert.IsFalse(theirs.Badge.activeSelf);

            Assert.IsTrue(target.enabled);
            Assert.AreEqual("365", target.text);
            Assert.AreEqual((Color)TokenColor.RoleScoreText, mine.Score.color);
            Assert.AreEqual((Color)TokenColor.RoleScoreText, target.color);
        }

        [Test]
        public void TheLockBadgeFollowsTheServerList()
        {
            Show(Snapshot(Seat.North, 305, 0, 365, Team.A));
            Assert.IsTrue(mine.Badge.activeSelf);

            Show(Snapshot(Seat.North, 305, 0, 365));
            Assert.IsFalse(mine.Badge.activeSelf);
        }

        [Test]
        public void TheLockMarkSitsAtTheServerThresholdAlongTheMeterFill()
        {
            Show(Snapshot(Seat.South, 50, 20, 365));

            Assert.IsTrue(mine.Mark.gameObject.activeSelf);
            Assert.AreEqual(300f / 365, mine.Mark.anchorMin.x, 1e-6f);
            Assert.AreEqual(300f / 365, mine.Mark.anchorMax.x, 1e-6f);
            Assert.AreEqual(0f, mine.Mark.anchorMin.y);
            Assert.AreEqual(1f, mine.Mark.anchorMax.y);
            Assert.AreEqual(300f / 365, theirs.Mark.anchorMin.x, 1e-6f);
        }

        [Test]
        public void TheLockMarkFollowsAVerticalMeterFromTheTop()
        {
            mine.Meter.fillMethod = Image.FillMethod.Vertical;
            mine.Meter.fillOrigin = (int)Image.OriginVertical.Top;

            Show(Snapshot(Seat.South, 50, 20, 365));

            Assert.AreEqual(1f - 300f / 365, mine.Mark.anchorMin.y, 1e-6f);
            Assert.AreEqual(0f, mine.Mark.anchorMin.x);
            Assert.AreEqual(1f, mine.Mark.anchorMax.x);
        }

        [Test]
        public void AMissingTargetHidesTheLockMark()
        {
            Show(Snapshot(Seat.South, 50, 20, 365));
            Show(Snapshot(Seat.South, 50, 20, 0));

            Assert.IsFalse(mine.Mark.gameObject.activeSelf);
            Assert.IsFalse(theirs.Mark.gameObject.activeSelf);
        }

        [Test]
        public void AMissingTargetHidesItAndLeavesTheMetersEmpty()
        {
            Show(Snapshot(Seat.South, 50, 20, 365));
            Show(Snapshot(Seat.South, 55, 20, 0));

            Assert.IsFalse(target.enabled);
            Assert.AreEqual(0f, mine.Meter.fillAmount);
            Assert.AreEqual("55", mine.Score.text);
        }

        [Test]
        public void AViewWithUnchangedScoresAllocatesNothing()
        {
            var first = Snapshot(Seat.West, 305, 120, 365);
            var second = Snapshot(Seat.West, 305, 120, 365, Team.A, Team.B);
            Show(first);
            Show(second);
            Show(first);

            Assert.That(() =>
            {
                Show(second);
                Show(first);
            }, Allocates.Nothing);
        }
    }
}
