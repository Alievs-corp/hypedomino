using System;
using HypeDomino.Core.Diagnostics;
using HypeDomino.UI.Motion;
using Unity.Profiling;
using UnityEngine;
using UnityEngine.UI;

namespace HypeDomino.Bench
{
    public sealed class AnimationBench : MonoBehaviour
    {
        const int TargetFrameRate = 60;
        const int WindowFrames = 1200;
        const int WindowsPerPhase = 2;
        const int Cycles = 3;
        const float SettleSeconds = 5f;
        const float LoopFramesPerSecond = 12f;
        const double MissedVsyncFactor = 1.5;
        const int PanelCount = 40;
        const int ColumnCount = 2;
        const float AvatarSize = 300f;

        static readonly string[] Phases = { "still", "nested", "shared" };

        [SerializeField] Texture2D[] atlases;
        [SerializeField] int cellSize;
        [SerializeField] int framesPerAvatar;

        readonly FrameTiming[] latestTiming = new FrameTiming[1];
        FrameWindow window;
        ProfilerRecorder gcAllocatedInFrame;
        ProfilerRecorder batchesCount;
        ProfilerRecorder setPassCount;
        ProfilerRecorder usedMemory;
        Transform screenRoot;
        Transform nestedRoot;
        AnimatedImage[] avatars;
        LoopArt[] art;
        AnimationBudget budget;
        int cycle;
        int phase;
        int windowsDone;
        float settledAt;
        bool skipReportFrame;
        bool finished;

        public void Configure(Texture2D[] avatarAtlases, int cell, int frames)
        {
            atlases = avatarAtlases;
            cellSize = cell;
            framesPerAvatar = frames;
        }

        void Awake()
        {
            Application.targetFrameRate = TargetFrameRate;
            Screen.sleepTimeout = SleepTimeout.NeverSleep;
            window = new FrameWindow(WindowFrames, MissedVsyncFactor / TargetFrameRate);
            budget = AnimationBudget.For(EffectLevel.Medium);
        }

        void OnEnable()
        {
            gcAllocatedInFrame = ProfilerRecorder.StartNew(ProfilerCategory.Memory, "GC Allocated In Frame");
            batchesCount = ProfilerRecorder.StartNew(ProfilerCategory.Render, "Batches Count");
            setPassCount = ProfilerRecorder.StartNew(ProfilerCategory.Render, "SetPass Calls Count");
            usedMemory = ProfilerRecorder.StartNew(ProfilerCategory.Memory, "System Used Memory");
        }

        void OnDisable()
        {
            gcAllocatedInFrame.Dispose();
            batchesCount.Dispose();
            setPassCount.Dispose();
            usedMemory.Dispose();
        }

        void Start()
        {
            BuildScreen();
            BeginPhase();
        }

        void Update()
        {
            FrameTimingManager.CaptureFrameTimings();
            if (finished || Time.unscaledTime < settledAt) return;
            if (FrameTimingManager.GetLatestTimings(1, latestTiming) == 0) return;

            var gcBytes = skipReportFrame ? 0 : LastValue(gcAllocatedInFrame);
            skipReportFrame = false;
            var work = Math.Max(latestTiming[0].cpuFrameTime, latestTiming[0].gpuFrameTime);
            window.Add(work, Time.unscaledDeltaTime, gcBytes, LastValue(batchesCount));
            if (!window.IsFull) return;

            Report(window.Close());
            skipReportFrame = true;
            windowsDone++;
            if (windowsDone == WindowsPerPhase) NextPhase();
        }

        void Report(FrameWindowReport report)
        {
            Debug.Log($"AnimationBench cycle={cycle + 1} phase={Phases[phase]} window={windowsDone + 1} frames={report.Frames} " +
                      $"workP50={report.WorkP50Ms:F2} workP95={report.WorkP95Ms:F2} missed={report.MissedVsyncPercent:F2} " +
                      $"gcFrames={report.AllocatingFrames} gcMax={report.MaxGcBytes} batches={report.MaxBatches} " +
                      $"setPass={LastValue(setPassCount)} usedMemoryMb={LastValue(usedMemory) / (1024 * 1024)} active={budget.Active} " +
                      $"size={Screen.width}x{Screen.height} refresh={Screen.currentResolution.refreshRateRatio.value:F0}");
        }

        void NextPhase()
        {
            windowsDone = 0;
            phase++;
            if (phase == Phases.Length)
            {
                phase = 0;
                cycle++;
            }
            if (cycle == Cycles)
            {
                finished = true;
                Debug.Log("AnimationBench finished");
                return;
            }
            BeginPhase();
        }

        void BeginPhase()
        {
            var parent = Phases[phase] == "nested" ? nestedRoot : screenRoot;
            var level = Phases[phase] == "still" ? EffectLevel.Low : EffectLevel.Medium;
            for (var at = 0; at < avatars.Length; at++)
            {
                avatars[at].transform.SetParent(parent, false);
                avatars[at].Show(art[at], level, false, budget, Time.unscaledTime);
            }
            settledAt = Time.unscaledTime + SettleSeconds;
        }

        void BuildScreen()
        {
            new GameObject("Camera", typeof(Camera)).GetComponent<Camera>().backgroundColor = Color.black;
            var canvas = new GameObject("Screen", typeof(RectTransform), typeof(Canvas), typeof(CanvasScaler)).GetComponent<Canvas>();
            canvas.renderMode = RenderMode.ScreenSpaceOverlay;
            var scaler = canvas.GetComponent<CanvasScaler>();
            scaler.uiScaleMode = CanvasScaler.ScaleMode.ScaleWithScreenSize;
            scaler.referenceResolution = new Vector2(1080f, 2340f);
            screenRoot = canvas.transform;
            AddStaticPanels();
            nestedRoot = new GameObject("Animated", typeof(RectTransform), typeof(Canvas)).transform;
            nestedRoot.SetParent(screenRoot, false);
            Stretch((RectTransform)nestedRoot);
            BuildAvatars();
        }

        void AddStaticPanels()
        {
            var white = Sprite.Create(Texture2D.whiteTexture, new Rect(0, 0, 4, 4), Vector2.zero);
            for (var at = 0; at < PanelCount; at++)
            {
                var image = new GameObject("Panel", typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                image.sprite = white;
                image.raycastTarget = false;
                image.color = new Color(0.1f, 0.1f, 0.2f + 0.01f * at, 0.5f);
                image.transform.SetParent(screenRoot, false);
                Place(image.rectTransform, at % 4, at / 4, 4, 10, 240f);
            }
        }

        void BuildAvatars()
        {
            avatars = new AnimatedImage[atlases.Length];
            art = new LoopArt[atlases.Length];
            for (var at = 0; at < atlases.Length; at++)
            {
                art[at] = ArtOf(atlases[at]);
                var image = new GameObject("Avatar", typeof(RectTransform), typeof(Image)).GetComponent<Image>();
                image.raycastTarget = false;
                image.transform.SetParent(screenRoot, false);
                Place(image.rectTransform, at % ColumnCount, at / ColumnCount, ColumnCount, atlases.Length / ColumnCount, AvatarSize);
                avatars[at] = image.gameObject.AddComponent<AnimatedImage>();
            }
        }

        LoopArt ArtOf(Texture2D atlas)
        {
            var frames = new Sprite[framesPerAvatar];
            var columns = atlas.width / cellSize;
            for (var at = 0; at < frames.Length; at++)
            {
                var cell = new Rect(at % columns * cellSize, at / columns * cellSize, cellSize, cellSize);
                frames[at] = Sprite.Create(atlas, cell, new Vector2(0.5f, 0.5f), 100f, 0, SpriteMeshType.FullRect);
            }
            return new LoopArt(frames[0], frames, LoopFramesPerSecond);
        }

        static void Place(RectTransform rect, int column, int row, int columns, int rows, float size)
        {
            var anchor = new Vector2((column + 0.5f) / columns, (row + 0.5f) / rows);
            rect.anchorMin = anchor;
            rect.anchorMax = anchor;
            rect.sizeDelta = new Vector2(size, size);
        }

        static void Stretch(RectTransform rect)
        {
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.offsetMin = Vector2.zero;
            rect.offsetMax = Vector2.zero;
        }

        static long LastValue(ProfilerRecorder recorder) => recorder.Valid ? recorder.LastValue : 0;
    }
}
