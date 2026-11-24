using System.Collections.Generic;
using HypeDomino.Game.Shop;
using HypeDomino.Net;
using HypeDomino.UI.Localization;
using TMPro;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.UI;

namespace HypeDomino.Game.Lobby
{
    public sealed class LobbyScreenView : MonoBehaviour
    {
        const long NotShown = -1;

        [SerializeField] Button[] playCards;
        [SerializeField] GameObject[] playCardsBusy;
        [SerializeField] Button cancel;
        [SerializeField] TMP_Text countdown;
        [SerializeField] TMP_Text message;
        [SerializeField] TMP_Text coins;
        [SerializeField] TMP_Text crystals;
        [SerializeField] TMP_Text level;
        [SerializeField] Button retryHeader;

        readonly Dictionary<string, object> durationArgs = new();
        UnityAction[] cardPicks;
        LobbyScreen screen;
        MatchQueue queue;
        LobbyHeaderLoader headerLoader;
        Session session;
        ServerClock clock;
        GameStrings strings;
        long shownCountdownSecond = NotShown;

        public void Configure(Button[] playCardButtons, GameObject[] playCardBusyMarks, Button cancelButton, TMP_Text countdownText, TMP_Text messageLine,
            TMP_Text coinsText, TMP_Text crystalsText, TMP_Text levelText, Button retryHeaderButton)
        {
            Unbind();
            playCards = playCardButtons;
            playCardsBusy = playCardBusyMarks;
            cancel = cancelButton;
            countdown = countdownText;
            message = messageLine;
            coins = coinsText;
            crystals = crystalsText;
            level = levelText;
            retryHeader = retryHeaderButton;
            cardPicks = null;
        }

        public void Bind(LobbyScreen lobby, MatchQueue matchQueue, LobbyHeaderLoader loader, Session netSession, ServerClock serverClock, GameStrings texts)
        {
            Unbind();
            screen = lobby;
            queue = matchQueue;
            headerLoader = loader;
            session = netSession;
            clock = serverClock;
            strings = texts;
            cardPicks ??= CardPicks(playCards.Length);
            for (var i = 0; i < playCards.Length; i++) playCards[i].onClick.AddListener(cardPicks[i]);
            cancel.onClick.AddListener(OnCancel);
            retryHeader.onClick.AddListener(OnRetryHeader);
            strings.Changed += screen.ShowAll;
            screen.ShowAll();
        }

        public void Tick(double nowSeconds)
        {
            var changes = screen.Tick(clock, nowSeconds);
            if ((changes & LobbyChanges.Header) != 0) ShowHeader();
            if ((changes & LobbyChanges.Play) != 0) ShowPlay();
            if ((changes & LobbyChanges.Countdown) != 0) ShowCountdown();
            if (retryHeader.interactable == headerLoader.Busy) retryHeader.interactable = !headerLoader.Busy;
        }

        void Start()
        {
            if (screen == null && NetBoot.LobbyScreen != null)
                Bind(NetBoot.LobbyScreen, NetBoot.Queue, NetBoot.LobbyHeaderLoader, NetBoot.Connection.Session, NetBoot.Connection.Clock, LanguageBoot.Strings);
        }

        void Update()
        {
            if (screen != null) Tick(Time.realtimeSinceStartupAsDouble);
        }

        void OnDestroy() => Unbind();

        void Unbind()
        {
            if (screen == null) return;
            for (var i = 0; i < playCards.Length; i++) playCards[i].onClick.RemoveListener(cardPicks[i]);
            cancel.onClick.RemoveListener(OnCancel);
            retryHeader.onClick.RemoveListener(OnRetryHeader);
            strings.Changed -= screen.ShowAll;
            screen = null;
            queue = null;
            headerLoader = null;
            session = null;
            clock = null;
            strings = null;
        }

        UnityAction[] CardPicks(int count)
        {
            var picks = new UnityAction[count];
            for (var i = 0; i < count; i++)
            {
                var card = i;
                picks[i] = () => OnPlay(card);
            }
            return picks;
        }

        void OnPlay(int card) => screen.Play.Play(session, card);

        void OnCancel() => screen.Play.Cancel(session);

        void OnRetryHeader()
        {
            if (headerLoader.Load(null)) retryHeader.interactable = false;
        }

        void ShowHeader()
        {
            var header = screen.Header;
            coins.enabled = header.HasBalances;
            crystals.enabled = header.HasBalances;
            level.enabled = header.HasLevel;
            retryHeader.gameObject.SetActive(!header.HasBalances || !header.HasLevel);
            if (header.HasBalances)
            {
                coins.text = MessageFormat.Number(header.Balances.Coins.Amount, strings.Language);
                crystals.text = MessageFormat.Number(header.Balances.Crystals.Amount, strings.Language);
            }
            if (header.HasLevel) level.text = MessageFormat.Number(header.Level, strings.Language);
        }

        void ShowPlay()
        {
            var play = screen.Play;
            for (var card = 0; card < playCards.Length; card++)
            {
                var state = play.StateOf(card);
                playCards[card].interactable = state == PlayCardState.Ready;
                playCardsBusy[card].SetActive(state == PlayCardState.Starting || state == PlayCardState.Searching);
            }
            cancel.gameObject.SetActive(play.CanCancel);
            var key = queue.RefusalKey;
            message.enabled = key != null;
            message.text = key == null ? string.Empty : strings.Text(key);
        }

        void ShowCountdown()
        {
            shownCountdownSecond = NotShown;
            ShopScreenView.ShowTimeLeft(countdown, screen.Countdown != QueueCountdown.None, screen.CountdownSeconds, ref shownCountdownSecond, strings, durationArgs);
        }
    }
}
