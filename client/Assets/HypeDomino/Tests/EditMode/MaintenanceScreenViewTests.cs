using HypeDomino.Game;
using Hypedomino.V1;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MaintenanceScreenViewTests
    {
        const ulong ServerNowMs = 1_800_000_000_000;

        GameObject root;
        GameObject panel;
        Button retry;
        MaintenanceScreen screen;
        int dials;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(MaintenanceScreenViewTests), typeof(RectTransform));
            panel = new GameObject("Panel", typeof(RectTransform));
            panel.transform.SetParent(root.transform, false);
            retry = new GameObject("Retry", typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            retry.transform.SetParent(panel.transform, false);
            dials = 0;
            screen = new MaintenanceScreen(() => dials++, () => 100);
            var view = root.AddComponent<MaintenanceScreenView>();
            view.Configure(panel, retry);
            view.Bind(screen);
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        static ServerHello Hello(HelloStatus status) => new() { Status = status, ServerTimeUnixMs = ServerNowMs, MaintenanceEndsUnixMs = ServerNowMs + 60_000 };

        [Test]
        public void ThePanelIsHiddenUntilAMaintenanceHello()
        {
            Assert.IsFalse(panel.activeSelf);

            screen.Read(Hello(HelloStatus.Maintenance));

            Assert.IsTrue(panel.activeSelf);
            Assert.IsTrue(retry.interactable);
        }

        [Test]
        public void RetryDialsOnceAndLocksUntilTheServerAnswers()
        {
            screen.Read(Hello(HelloStatus.Maintenance));

            retry.onClick.Invoke();
            retry.onClick.Invoke();

            Assert.AreEqual(1, dials);
            Assert.IsFalse(retry.interactable);

            screen.Read(Hello(HelloStatus.Maintenance));

            Assert.IsTrue(panel.activeSelf);
            Assert.IsTrue(retry.interactable);
        }

        [Test]
        public void AnAcceptedHelloHidesThePanel()
        {
            screen.Read(Hello(HelloStatus.Maintenance));

            screen.Read(Hello(HelloStatus.Accepted));

            Assert.IsFalse(panel.activeSelf);
        }
    }
}
