using System;
using System.Collections.Generic;
using HypeDomino.Game;
using HypeDomino.UI.Localization;
using Hypedomino.V1;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class MaintenanceScreenViewTests
    {
        const ulong ServerNowMs = 1_800_000_000_000;
        const ulong EndsMs = ServerNowMs + 60_000;

        GameObject root;
        GameObject panel;
        TMP_Text ends;
        Button retry;
        LanguageChoice choice;
        GameStrings strings;
        MaintenanceScreen screen;
        int dials;

        [SetUp]
        public void Build()
        {
            root = new GameObject(nameof(MaintenanceScreenViewTests), typeof(RectTransform));
            panel = new GameObject("Panel", typeof(RectTransform));
            panel.transform.SetParent(root.transform, false);
            ends = new GameObject("Ends", typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            ends.transform.SetParent(panel.transform, false);
            retry = new GameObject("Retry", typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            retry.transform.SetParent(panel.transform, false);
            dials = 0;
            choice = LanguageChoice.Load(GameLanguage.English, () => Array.Empty<string>());
            strings = new GameStrings(BuiltinStrings.Load(), choice, false);
            screen = new MaintenanceScreen(() => dials++, () => 100);
            var view = root.AddComponent<MaintenanceScreenView>();
            view.Configure(panel, ends, retry);
            view.Bind(screen, strings);
        }

        [TearDown]
        public void Destroy() => Object.DestroyImmediate(root);

        static ServerHello Hello(HelloStatus status) => new() { Status = status, ServerTimeUnixMs = ServerNowMs, MaintenanceEndsUnixMs = EndsMs };

        static ServerHello WithoutEnd() => new() { Status = HelloStatus.Maintenance, ServerTimeUnixMs = ServerNowMs };

        static string Expected(GameStrings texts, ulong endsMs) =>
            texts.Text(MaintenanceScreenView.EndsKey, new Dictionary<string, object> { [MaintenanceScreenView.EndArgument] = MaintenanceScreenView.EndText(endsMs, texts.Language) });

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

        [Test]
        public void TheEndTimeIsWrittenInUtc()
        {
            Assert.AreEqual("1/15/2027, 08:01", MaintenanceScreenView.EndText(EndsMs, GameLanguage.English));
            Assert.AreEqual("15.01.2027 08:01", MaintenanceScreenView.EndText(EndsMs, "az"));
        }

        [Test]
        public void TheEndTimeTheServerSentIsShownInUtc()
        {
            screen.Read(Hello(HelloStatus.Maintenance));

            Assert.IsTrue(ends.gameObject.activeSelf);
            Assert.AreEqual("Back at 1/15/2027, 08:01 UTC.", ends.text);
        }

        [Test]
        public void AHelloWithoutAnEndHidesTheEndTime()
        {
            screen.Read(Hello(HelloStatus.Maintenance));

            screen.Read(WithoutEnd());

            Assert.IsTrue(panel.activeSelf);
            Assert.IsFalse(ends.gameObject.activeSelf);
        }

        [Test]
        public void AMovedEndTimeIsShown()
        {
            const ulong movedMs = EndsMs + 3_600_000;
            screen.Read(Hello(HelloStatus.Maintenance));

            screen.Read(new ServerHello { Status = HelloStatus.Maintenance, ServerTimeUnixMs = ServerNowMs, MaintenanceEndsUnixMs = movedMs });

            Assert.AreEqual(Expected(strings, movedMs), ends.text);
        }

        [Test]
        public void AChosenLanguageRewritesTheEndTime()
        {
            screen.Read(Hello(HelloStatus.Maintenance));

            choice.Choose("az");

            Assert.AreEqual(Expected(strings, EndsMs), ends.text);
            StringAssert.Contains(MaintenanceScreenView.EndText(EndsMs, "az"), ends.text);
        }
    }
}
