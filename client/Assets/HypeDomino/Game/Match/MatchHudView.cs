using System;
using HypeDomino.Net;
using HypeDomino.UI.Board;
using HypeDomino.UI.Effects;
using HypeDomino.UI.Localization;
using HypeDomino.UI.Match;
using HypeDomino.UI.Motion;
using HypeDomino.UI.Tokens;
using TMPro;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.UI;

namespace HypeDomino.Game.Match
{
    public sealed class MatchHudView : MonoBehaviour, IEffectSettingsTarget
    {
        const float MillisecondsPerSecond = 1000f;

        static readonly Color RingColor = TokenColor.GoldDefault;
        static readonly Color RingWarningColor = TokenColor.RoleWarningFill;

        [SerializeField] TileFaceSet faceSet;
        [SerializeField] Button[] handTiles = Array.Empty<Button>();
        [SerializeField] GameObject[] legalMarks = Array.Empty<GameObject>();
        [SerializeField] Button draw;
        [SerializeField] Button pass;
        [SerializeField] Button endTurn;
        [SerializeField] Image turnRing;
        [SerializeField] TMP_Text network;
        [SerializeField] TMP_Text afkWarning;
        [SerializeField] Button takeBack;
        [SerializeField] TMP_Text takeBackText;
        [SerializeField] TMP_Text[] seatBadges = new TMP_Text[TableSides.Count];
        [SerializeField] CanvasGroup[] passLabels = new CanvasGroup[TableSides.Count];
        [SerializeField] RectTransform dealOrigin;
        [SerializeField] RectTransform drawOrigin;
        [SerializeField] HandSummaryView summary;

        UnityAction[] tileTaps = Array.Empty<UnityAction>();
        MatchHost host;
        Connection connection;
        GameStrings strings;
        TweenPool arrivalPool = TweenPool.For(EffectLevel.Medium);
        TransformTweenSink arrivalSink;
        Vector3[] restPositions = Array.Empty<Vector3>();
        EffectLevel level = EffectLevel.Medium;
        bool reduceMotion;
        int handVersion = -1;
        int arrivalsVersion = -1;
        int noticesVersion = -1;
        int passVersion = -1;
        int summaryVersion = -1;

        public int MovingCount => arrivalPool.Count;

        MatchScreen LiveScreen => host?.Root?.Screen;

        public void Configure(TileFaceSet faces, Button[] tiles, GameObject[] marks, Button drawButton, Button passButton, Button endTurnButton, Image ring,
            TMP_Text networkText, TMP_Text afkText, Button takeBackButton, TMP_Text takeBackLabel, TMP_Text[] badges, CanvasGroup[] passGroups,
            RectTransform dealFrom, RectTransform drawFrom, HandSummaryView summaryView = null)
        {
            Unbind();
            faceSet = faces;
            handTiles = tiles ?? Array.Empty<Button>();
            legalMarks = marks ?? Array.Empty<GameObject>();
            if (legalMarks.Length != handTiles.Length) throw new ArgumentException($"{legalMarks.Length} marks for {handTiles.Length} tiles", nameof(marks));
            draw = drawButton;
            pass = passButton;
            endTurn = endTurnButton;
            turnRing = ring;
            network = networkText;
            afkWarning = afkText;
            takeBack = takeBackButton;
            takeBackText = takeBackLabel;
            seatBadges = badges;
            passLabels = passGroups;
            dealOrigin = dealFrom;
            drawOrigin = drawFrom;
            summary = summaryView;
        }

        public void Configure(EffectLevel effectLevel, bool reduce)
        {
            reduceMotion = reduce;
            if (effectLevel == level) return;
            level = effectLevel;
            arrivalPool.FinishAll(ref arrivalSink);
            arrivalPool = TweenPool.For(level);
        }

        public void Bind(MatchHost matchHost, Connection matchConnection, GameStrings texts)
        {
            Unbind();
            host = matchHost ?? throw new ArgumentNullException(nameof(matchHost));
            connection = matchConnection ?? throw new ArgumentNullException(nameof(matchConnection));
            strings = texts ?? throw new ArgumentNullException(nameof(texts));
            strings.Changed += RedrawTexts;
            tileTaps = new UnityAction[handTiles.Length];
            var moved = new Transform[handTiles.Length];
            restPositions = new Vector3[handTiles.Length];
            for (var index = 0; index < handTiles.Length; index++)
            {
                moved[index] = handTiles[index].transform;
                var slot = index;
                tileTaps[index] = () => LiveScreen?.TapTile(slot);
                handTiles[index].onClick.AddListener(tileTaps[index]);
            }
            draw.onClick.AddListener(Draw);
            pass.onClick.AddListener(Pass);
            endTurn.onClick.AddListener(EndTurn);
            takeBack.onClick.AddListener(TakeBack);
            arrivalSink = new TransformTweenSink(moved);
            arrivalsVersion = host.Hud.ArrivalsVersion;
            handVersion = noticesVersion = passVersion = summaryVersion = -1;
            Refresh();
        }

        public void Unbind()
        {
            if (host == null) return;
            arrivalPool.FinishAll(ref arrivalSink);
            strings.Changed -= RedrawTexts;
            for (var index = 0; index < tileTaps.Length; index++) handTiles[index].onClick.RemoveListener(tileTaps[index]);
            tileTaps = Array.Empty<UnityAction>();
            draw.onClick.RemoveListener(Draw);
            pass.onClick.RemoveListener(Pass);
            endTurn.onClick.RemoveListener(EndTurn);
            takeBack.onClick.RemoveListener(TakeBack);
            host = null;
            connection = null;
            strings = null;
        }

        public void Refresh() => Refresh(Time.unscaledTime * MillisecondsPerSecond);

        public void Refresh(float nowMs)
        {
            if (host == null) return;
            var hud = host.Hud;
            if (hud.HandVersion != handVersion)
            {
                handVersion = hud.HandVersion;
                ShowHand(hud.Hand);
            }
            if (hud.ArrivalsVersion != arrivalsVersion)
            {
                arrivalsVersion = hud.ArrivalsVersion;
                ShowArrivals(hud.Arrivals, nowMs);
            }
            arrivalPool.Step(nowMs, ref arrivalSink);
            ShowTimer(hud.Timer);
            if (hud.NoticesVersion != noticesVersion)
            {
                noticesVersion = hud.NoticesVersion;
                ShowNotices(hud.Notices);
            }
            if (hud.PassVersion != passVersion)
            {
                passVersion = hud.PassVersion;
                ShowPasses(hud);
            }
            if (summary != null && hud.SummaryVersion != summaryVersion)
            {
                summaryVersion = hud.SummaryVersion;
                summary.Show(hud.Summary);
            }
        }

        void LateUpdate() => Refresh();

        void OnDestroy() => Unbind();

        void RedrawTexts() => noticesVersion = -1;

        void Draw() => LiveScreen?.Draw();

        void Pass() => LiveScreen?.Pass();

        void EndTurn() => LiveScreen?.EndTurn();

        void TakeBack() => LiveScreen?.TapTakeBack(connection);

        void ShowHand(HandChoice hand)
        {
            var count = hand?.HandCount ?? 0;
            for (var index = 0; index < handTiles.Length; index++)
            {
                var shown = index < count;
                SetActive(handTiles[index].gameObject, shown);
                var selectable = shown && hand.IsSelectable(index);
                SetActive(legalMarks[index], selectable);
                handTiles[index].interactable = selectable;
                if (!shown) continue;
                var tile = hand.TileAt(index);
                handTiles[index].image.sprite = faceSet.Face(new Domino((byte)tile.Low, (byte)tile.High));
            }
            SetActive(draw.gameObject, hand != null && hand.CanDraw);
            SetActive(pass.gameObject, hand != null && hand.CanPass);
            SetActive(endTurn.gameObject, hand != null && hand.CanEndTurn);
        }

        void ShowArrivals(HandArrivals arrivals, float nowMs)
        {
            if (arrivals == null)
            {
                arrivalPool.FinishAll(ref arrivalSink);
                return;
            }
            var origin = arrivals.Kind == HandArrival.Deal ? dealOrigin : drawOrigin;
            if (origin == null) return;
            var arrived = arrivals.ArrivedTiles;
            for (var order = 0; order < arrived.Length; order++)
            {
                var index = arrived[order];
                if (index >= handTiles.Length) continue;
                var tile = handTiles[index].transform;
                if (!arrivalPool.IsMoving(index, TweenChannel.Position)) restPositions[index] = tile.localPosition;
                var from = tile.parent.InverseTransformPoint(origin.position);
                arrivalSink.Apply(index, TweenChannel.Position, from);
                TileTravel.Arrive(arrivalPool, ref arrivalSink, index, order, from, restPositions[index], nowMs, reduceMotion);
            }
        }

        void ShowTimer(TurnTimer timer)
        {
            var running = timer != null && timer.Running;
            SetActive(turnRing.gameObject, running);
            if (!running) return;
            turnRing.fillAmount = timer.Fill;
            turnRing.color = timer.IsWarning ? RingWarningColor : RingColor;
        }

        void ShowNotices(MatchNotices notices)
        {
            ShowKey(network, notices == null ? null : ConnectionNoticeKeys.Of(notices.Network));
            ShowKey(afkWarning, notices != null && notices.ShowsAfkWarning ? ConnectionNoticeKeys.OneTimeoutFromBot : null);
            var banner = notices != null && notices.ShowsTakeBack;
            SetActive(takeBack.gameObject, banner);
            if (banner) takeBackText.text = strings.Text(ConnectionNoticeKeys.TakeBack);
            for (var side = 0; side < seatBadges.Length; side++)
                ShowKey(seatBadges[side], notices == null ? null : ConnectionNoticeKeys.Of(notices.Seats.At((TableSide)side)));
        }

        void ShowPasses(MatchHud hud)
        {
            for (var side = 0; side < passLabels.Length; side++)
            {
                var label = passLabels[side];
                if (label == null) continue;
                var shown = hud.TryGetPass((TableSide)side, out var notice);
                SetActive(label.gameObject, shown);
                if (!shown) continue;
                label.alpha = notice.Alpha;
                label.transform.localScale = new Vector3(notice.Scale, notice.Scale, 1f);
            }
        }

        void ShowKey(TMP_Text text, string key)
        {
            if (text == null) return;
            text.enabled = key != null;
            if (key != null) text.text = strings.Text(key);
        }

        static void SetActive(GameObject target, bool active)
        {
            if (target.activeSelf != active) target.SetActive(active);
        }
    }
}
