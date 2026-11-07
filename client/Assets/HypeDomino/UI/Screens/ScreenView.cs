using HypeDomino.UI.Tokens;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.UI.Screens
{
    [RequireComponent(typeof(Canvas), typeof(CanvasScaler), typeof(GraphicRaycaster))]
    public sealed class ScreenView : MonoBehaviour
    {
        const float ReferenceShortSide = 1080f;

        [SerializeField] CanvasGroup content;
        [SerializeField] Image scrim;

        Canvas canvas;
        GraphicRaycaster raycaster;
        ScreenFade body;
        ScreenFade shade;

        public bool IsModal => scrim != null;

        public bool IsShown => canvas.enabled;

        public bool TakesInput => raycaster.enabled;

        public float Visibility => body.Value;

        public int SortingOrder => canvas.sortingOrder;

        public void Configure(CanvasGroup screenContent, Image modalScrim)
        {
            content = screenContent;
            scrim = modalScrim;
            PaintScrim();
            Apply(false);
        }

        void Awake()
        {
            canvas = GetComponent<Canvas>();
            canvas.renderMode = RenderMode.ScreenSpaceOverlay;
            canvas.enabled = false;
            raycaster = GetComponent<GraphicRaycaster>();
            raycaster.enabled = false;
            var scaler = GetComponent<CanvasScaler>();
            scaler.uiScaleMode = CanvasScaler.ScaleMode.ScaleWithScreenSize;
            scaler.referenceResolution = new Vector2(ReferenceShortSide, ReferenceShortSide);
            scaler.screenMatchMode = CanvasScaler.ScreenMatchMode.Expand;
            PaintScrim();
            Apply(false);
        }

        void PaintScrim()
        {
            if (scrim != null) scrim.color = TokenColor.Scrim;
        }

        internal void Show(int sortingOrder)
        {
            canvas.sortingOrder = sortingOrder;
            canvas.enabled = true;
        }

        internal void Hide() => canvas.enabled = false;

        internal void TakeInput(bool on) => raycaster.enabled = on;

        internal void Enter()
        {
            body.Retarget(1f, IsModal ? TokenDurationMs.Moderate : TokenDurationMs.Slow, TokenEasing.Enter);
            if (IsModal) shade.Retarget(1f, TokenDurationMs.Base, TokenEasing.Standard);
        }

        internal void Exit()
        {
            body.Retarget(0f, TokenDurationMs.Base, TokenEasing.Exit);
            if (IsModal) shade.Retarget(0f, TokenDurationMs.Base, TokenEasing.Standard);
        }

        internal bool Advance(float deltaMs, bool reduceMotion)
        {
            var bodyDone = body.Advance(deltaMs);
            var shadeDone = shade.Advance(deltaMs);
            Apply(reduceMotion);
            return bodyDone && shadeDone;
        }

        void Apply(bool reduceMotion)
        {
            if (content != null)
            {
                content.alpha = body.Value;
                var scale = reduceMotion ? 1f : Mathf.LerpUnclamped(TokenMotion.PanelEnterScale, 1f, body.Value);
                content.transform.localScale = new Vector3(scale, scale, 1f);
            }
            if (scrim != null) scrim.canvasRenderer.SetAlpha(TokenOpacity.Scrim * shade.Value);
        }
    }
}
