using HypeDomino.Game.Lobby;
using HypeDomino.Game.Training;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class LobbyTrainingCardViewTests
    {
        sealed class BotGame : ITrainingGame
        {
            public int Starts;

            public bool TryStart()
            {
                Starts++;
                return true;
            }
        }

        GameObject root;
        Button start;
        MatchQueue queue;
        BotGame game;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(LobbyTrainingCardViewTests), typeof(RectTransform));
            start = new GameObject("Start", typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            start.transform.SetParent(root.transform, false);
            queue = new MatchQueue();
            var play = new LobbyPlay(queue, (_, _) => { });
            game = new BotGame();
            var view = root.AddComponent<LobbyTrainingCardView>();
            view.Configure(start);
            view.Bind(new LobbyTraining(queue, play, game), play);
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        [Test]
        public void TappingTheReadyCardOpensTheBotGame()
        {
            Assert.IsTrue(start.interactable);

            start.onClick.Invoke();

            Assert.AreEqual(1, game.Starts);
        }

        [Test]
        public void ARunningSearchLocksTheCardUntilItEnds()
        {
            queue.Read(new QueueState { Status = QueueStatus.Searching, Mode = Mode._1V1, Queue = QueueKind.Ranked });

            Assert.IsFalse(start.interactable);
            start.onClick.Invoke();
            Assert.AreEqual(0, game.Starts);

            queue.Read(new QueueState { Status = QueueStatus.Left });

            Assert.IsTrue(start.interactable);
        }
    }
}
