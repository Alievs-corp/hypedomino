using System;
using HypeDomino.Game;
using HypeDomino.Game.Tutorial;
using HypeDomino.UI.Localization;
using NUnit.Framework;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using Object = UnityEngine.Object;

namespace HypeDomino.Tests.EditMode
{
    public sealed class TutorialOfferViewTests
    {
        const int Unset = -1;

        GameObject root;
        GameObject panel;
        TMP_Text title;
        TMP_Text body;
        Button play;
        TMP_Text playLabel;
        Button skip;
        TMP_Text skipLabel;
        LanguageChoice choice;
        TutorialOffer offer;
        int keptFinished;

        [SetUp]
        public void Build()
        {
            keptFinished = PlayerPrefs.GetInt(TutorialProgressStore.FinishedKey, Unset);
            PlayerPrefs.DeleteKey(TutorialProgressStore.FinishedKey);
            root = new GameObject(nameof(TutorialOfferViewTests), typeof(RectTransform));
            panel = new GameObject("Panel", typeof(RectTransform));
            panel.transform.SetParent(root.transform, false);
            title = NewText("Title");
            body = NewText("Body");
            play = NewButton("Play");
            playLabel = NewText("Play label");
            skip = NewButton("Skip");
            skipLabel = NewText("Skip label");
            choice = LanguageChoice.Load(GameLanguage.English, () => Array.Empty<string>());
            offer = new TutorialOffer();
            var view = root.AddComponent<TutorialOfferView>();
            view.Configure(panel, title, body, play, playLabel, skip, skipLabel);
            view.Bind(offer, new GameStrings(BuiltinStrings.Load(), choice, false));
        }

        [TearDown]
        public void Destroy()
        {
            Object.DestroyImmediate(root);
            if (keptFinished == Unset) PlayerPrefs.DeleteKey(TutorialProgressStore.FinishedKey);
            else PlayerPrefs.SetInt(TutorialProgressStore.FinishedKey, keptFinished);
            PlayerPrefs.Save();
        }

        TMP_Text NewText(string name)
        {
            var text = new GameObject(name, typeof(RectTransform)).AddComponent<TextMeshProUGUI>();
            text.transform.SetParent(panel.transform, false);
            return text;
        }

        Button NewButton(string name)
        {
            var button = new GameObject(name, typeof(RectTransform), typeof(Image)).AddComponent<Button>();
            button.transform.SetParent(panel.transform, false);
            return button;
        }

        static string Builtin(string key, string language) => BuiltinStrings.Load().Value(key, language);

        [Test]
        public void ThePanelIsHiddenUntilTheOfferAsks()
        {
            Assert.IsFalse(panel.activeSelf);

            offer.Read(AppState.Lobby);

            Assert.IsTrue(panel.activeSelf);
            Assert.AreEqual(Builtin(TutorialOffer.Title, "en"), title.text);
            Assert.AreEqual(Builtin(TutorialOffer.Body, "en"), body.text);
            Assert.AreEqual(Builtin(TutorialOffer.Play, "en"), playLabel.text);
            Assert.AreEqual(Builtin(TutorialOffer.Skip, "en"), skipLabel.text);
        }

        [Test]
        public void PlayStartsTheRunOnceAndHidesThePanel()
        {
            offer.Read(AppState.Lobby);

            play.onClick.Invoke();
            play.onClick.Invoke();

            Assert.IsFalse(panel.activeSelf);
            Assert.AreEqual(TutorialState.Running, offer.Run.State);
            Assert.AreEqual(0, offer.Run.StepIndex);
            Assert.IsTrue(TutorialProgressStore.ShouldOffer);
        }

        [Test]
        public void SkipSkipsTheRunAndHidesThePanel()
        {
            offer.Read(AppState.Lobby);

            skip.onClick.Invoke();

            Assert.IsFalse(panel.activeSelf);
            Assert.AreEqual(TutorialState.Skipped, offer.Run.State);
            Assert.IsFalse(TutorialProgressStore.ShouldOffer);
        }

        [Test]
        public void LeavingTheLobbyHidesAnUnansweredPrompt()
        {
            offer.Read(AppState.Lobby);

            offer.Read(AppState.Room);

            Assert.IsFalse(panel.activeSelf);
            Assert.IsTrue(TutorialProgressStore.ShouldOffer);
        }

        [Test]
        public void TheShownTextFollowsALanguageChange()
        {
            offer.Read(AppState.Lobby);

            Assert.IsTrue(choice.Choose("az"));

            Assert.AreEqual(Builtin(TutorialOffer.Title, "az"), title.text);
            Assert.AreEqual(Builtin(TutorialOffer.Body, "az"), body.text);
            Assert.AreEqual(Builtin(TutorialOffer.Play, "az"), playLabel.text);
            Assert.AreEqual(Builtin(TutorialOffer.Skip, "az"), skipLabel.text);
        }
    }
}
