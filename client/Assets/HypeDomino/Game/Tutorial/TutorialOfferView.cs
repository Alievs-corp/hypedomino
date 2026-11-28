using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Game.Tutorial
{
    public sealed class TutorialOfferView : MonoBehaviour
    {
        [SerializeField] GameObject panel;
        [SerializeField] TMP_Text title;
        [SerializeField] TMP_Text body;
        [SerializeField] Button play;
        [SerializeField] TMP_Text playLabel;
        [SerializeField] Button skip;
        [SerializeField] TMP_Text skipLabel;

        TutorialOffer offer;
        GameStrings strings;

        public void Configure(GameObject offerPanel, TMP_Text titleText, TMP_Text bodyText, Button playButton, TMP_Text playText, Button skipButton, TMP_Text skipText)
        {
            Unbind();
            panel = offerPanel;
            title = titleText;
            body = bodyText;
            play = playButton;
            playLabel = playText;
            skip = skipButton;
            skipLabel = skipText;
        }

        public void Bind(TutorialOffer tutorialOffer, GameStrings texts)
        {
            Unbind();
            offer = tutorialOffer;
            strings = texts;
            offer.Changed += Refresh;
            strings.Changed += Refresh;
            play.onClick.AddListener(Play);
            skip.onClick.AddListener(Skip);
            Refresh();
        }

        void Start()
        {
            if (offer == null) Bind(TutorialBoot.Offer, LanguageBoot.Strings);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (offer == null) return;
            offer.Changed -= Refresh;
            strings.Changed -= Refresh;
            play.onClick.RemoveListener(Play);
            skip.onClick.RemoveListener(Skip);
            offer = null;
            strings = null;
        }

        void Play() => offer.Accept();

        void Skip() => offer.Decline();

        void Refresh()
        {
            panel.SetActive(offer.Asking);
            if (!offer.Asking) return;
            title.text = strings.Text(TutorialOffer.Title);
            body.text = strings.Text(TutorialOffer.Body);
            playLabel.text = strings.Text(TutorialOffer.Play);
            skipLabel.text = strings.Text(TutorialOffer.Skip);
        }
    }
}
