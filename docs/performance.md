# Performance budget and measurement method

Budgets come from `CLAUDE.md` and are not changed here; this file fixes how each is measured and when it passes. A budget without a passing measurement is not met. Tool and harness work is tracked in #192 (load bot, replay harness); device certification is #191; battery measurement is #164. Decisions that affect these paths are recorded in `docs/adr/`.

## 1. Device matrix

Initial reference matrix. #191 certifies on it; changing a model needs a PR to this table.

| Tier | Android reference | SoC / GPU / RAM / OS | iOS reference | Role |
|---|---|---|---|---|
| Low | Samsung Galaxy A14 4G | Helio G80, Mali-G52, 4 GB, Android 13 | iPhone XR (A12, 3 GB, iOS 17) | Playable at 30 fps minimum, reduced effects |
| Mid | Samsung Galaxy A34 5G | Dimensity 1080, Mali-G68, 6 GB, Android 14 | iPhone 12 (A14, 4 GB, iOS 17) | Gating tier: 60 fps and all start, draw call, GC, size, battery gates |
| High | Samsung Galaxy S24 | Snapdragon 8 Gen 3, Adreno 750, 8 GB, Android 14 | iPhone 15 Pro (A17 Pro, 8 GB, 120 Hz ProMotion) | 120 fps option |

Class rule for devices outside the matrix: Low = under 4 GB RAM or GPU older than Mali-G57 / Adreno 610 class; Mid = 4 to 8 GB with Mali-G68 / Adreno 619 class or better; High = 8 GB or more, flagship SoC, 120 Hz panel. The effects quality level (admin-tunable) is chosen per tier; Low runs the reduced-effects level.

## 2. Test conditions (all device measurements)

- Development build with the Unity Profiler attached, deep profiling off, for Profiler-based gates. Release build (IL2CPP, ARM64, no Profiler) for cold start, fps and battery gates.
- Device between 100% and 30% battery, not charging, Wi-Fi to the test server, brightness fixed at 50%, no other foreground apps, idle-cool (surface under 35 C) before each run, battery saver / Low Power Mode off.
- `Application.targetFrameRate` 60 (120 for the high option), vsync on. Player Settings "Frame Timing Stats" enabled so `FrameTimingManager` reports per-frame CPU and GPU time.
- With vsync on, the presented interval is quantised to the refresh period (16.7 ms at 60 Hz), so smoothness is judged by two numbers: work time (`max(cpuFrameTime, gpuFrameTime)` from `FrameTimingManager.GetLatestTimings`) shows headroom, and missed vsyncs (presented interval longer than 1.5 refresh periods) show visible stutter.
- Scripted match replay: a fixed-seed recorded 2v2 match of 3 minutes, server-fed, effects at the tier's default level, worst-case effect moments included. Played 3 times back to back; all 3 runs must pass and the worst is reported. Harness: #192.

## 3. Client budgets

| Budget | Target | Tier / environment | Method | Sample and duration | Pass / fail |
|---|---|---|---|---|---|
| Stable 60 fps | 60 fps | Mid (Low: section 4) | Replay harness records per frame `FrameTimingManager` CPU and GPU time and the presented interval (`Time.unscaledDeltaTime`), exported as CSV and analysed in Profile Analyzer. Cross-check presented frames with `adb shell dumpsys SurfaceFlinger --latency <layer>` or a Perfetto frame timeline | 3 replays x 3 min (about 10,800 frames each) | All hold: work time p50 <= 11.0 ms and p95 <= 14.0 ms; missed vsyncs (interval > 25.0 ms) <= 1% of frames; no interval over 100 ms |
| 120 fps option | 120 fps | High, 120 Hz panel, option enabled | Same as above, target 120 | 3 replays x 3 min | Work time p50 <= 5.5 ms and p95 <= 7.0 ms; missed vsyncs (interval > 12.5 ms) <= 2%; no interval over 50 ms |
| Cold start | < 3 s | Mid, release build | Command below | 10 runs, force-stop before each, first run after install discarded | Median time to interactive lobby < 3000 ms and worst run < 3500 ms. Low and High are recorded, not gated |
| Draw calls | < 150 | Mid, match scene | `ProfilerRecorder` on `ProfilerCategory.Render` "Batches Count" and "SetPass Calls Count" per frame; inspect the worst frame in Frame Debugger (Window > Analysis > Frame Debugger) | 1 full replay, worst-case effect frame included | Maximum Batches over all frames < 150 (maximum, not average) |
| GC allocation | 0 B/frame | Mid, match scene, development build | `ProfilerRecorder.StartNew(ProfilerCategory.Memory, "GC Allocated In Frame")` per frame; `GC.CollectionCount(0)` before and after; Profiler Memory module and Memory Profiler snapshot to find the source of any allocation | First deal to result screen, 3 replays | Every frame 0 B and `GC.CollectionCount(0)` delta 0. One non-zero frame fails |
| Install size | < 150 MB | Android AAB and iOS IPA, release | Unity Build Report (`Library/LastBuild.buildreport`, Build Report Inspector) for the content breakdown. Android: `bundletool build-apks --bundle=app.aab --output=app.apks --device-spec=<tier>.json` then `bundletool get-size total --apks=app.apks`. iOS: App Thinning Size Report from the Xcode archive | Every release candidate | Android: max download size over the low, mid and high device specs < 150 MB. iOS: compressed size for the mid and high models < 150 MB. Content fetched on demand through Addressables is excluded from this number but listed in the report |

Cold start (Android, run 10 times; `<activity>` is the launcher activity):

```bash
adb shell am force-stop <package>
adb shell am start -W -n <package>/<activity>
```

`am start -W` prints `TotalTime` (process start to first frame). The gate is time to interactive lobby: the client's startup marker in `adb logcat -s Unity` minus the launch time. iOS: Xcode Instruments "App Launch" template, same 10-run rule.

Leak check, reported but not gated until #191 sets a number: Memory Profiler snapshots after replay 1 and replay 3; managed and native totals must not grow.

## 4. Low tier target

| Target | Method | Pass / fail |
|---|---|---|
| Playable at 30 fps minimum with reduced effects | Same replay on the Low reference devices, `targetFrameRate` 30, reduced-effects level | Work time p95 <= 28.0 ms; missed vsyncs (interval > 50.0 ms) <= 1%; no interval over 100 ms. Draw call and GC gates still apply. Move, draw, pass and timer fully functional |

## 5. Battery

Initial target, to be recalibrated after the #164 measurement.

| Target | Tier | Method | Pass / fail |
|---|---|---|---|
| 30-minute match session drains at most 8% battery | Mid reference devices, 60 fps, default effects, Wi-Fi | Android: `adb shell dumpsys batterystats --reset`, play 30 min of back-to-back matches (replay harness in live mode), `adb bugreport`, open in Battery Historian; read battery level delta, app estimated power use and end temperature. iOS: Instruments "Energy Log" over the same 30 min plus battery percentage delta | 3 sessions; median delta <= 8% and worst <= 10%. Until #164 reports, a miss is logged, not blocking |

## 6. Server budgets

Run on the reference node spec (CPU, RAM, region) fixed in #192 and recorded with the results.

| Budget | Target | Environment | Method | Sample and duration | Pass / fail |
|---|---|---|---|---|---|
| Move round-trip | p99 < 100 ms | Load bot in the same cloud region as the node, on a different host; production-like TLS and WebSocket | Load bot (`server/cmd/loadbot`, #192) timestamps each move intent send and the resulting broadcast receive. The server also exports a Prometheus histogram `move_roundtrip_seconds` (receive to last broadcast write) for its in-process share | 30 min steady state at the 10k-room load, at least 100,000 moves, first 2 min discarded as warm-up | Bot-measured p99 < 100 ms overall and in every 1-minute window; server histogram p99 reported alongside |
| Concurrent rooms | 10,000 per node | Reference node | Load bot ramps to 10,000 active rooms (1v1 and 2v2 mix, moves at human cadence inside the admin-config turn time), then holds. Observe with pprof, `runtime/metrics` and node CPU / RAM | Ramp, then 30 min hold | Holds 10,000 rooms within the move budget; error and disconnect rate under 0.1%; average CPU under 80%; heap flat within 5% between minute 10 and minute 30 |
| Hot-path cost | No regression | Developer machine and CI | `go test -run='^$' -bench=. -benchmem -count=10 ./internal/rules/... > new.txt`, then `benchstat old.txt new.txt` | 10 runs per side | No benchmark slower than 5% at `p < 0.05`; allocs/op not higher |

Profiling:

```bash
go test -run='^$' -bench=<Name> -cpuprofile cpu.out -memprofile mem.out ./internal/rules
go tool pprof -http=:0 cpu.out
go tool pprof -http=:0 "http://<node>:<debug-port>/debug/pprof/profile?seconds=30"
go tool pprof -http=:0 "http://<node>:<debug-port>/debug/pprof/heap"
```

The debug port is never exposed publicly.

## 7. Tools

| Tool | Used for |
|---|---|
| Unity Profiler (development build, deep profiling off), Profile Analyzer | Frame time, GC, batches, before/after capture comparison |
| Frame Debugger | Draw call composition |
| Memory Profiler | Snapshots, leak and allocation source |
| Android GPU Inspector, Perfetto (`adb shell perfetto`, ui.perfetto.dev) | GPU and frame timeline, CPU scheduling, thermal |
| Xcode Instruments (Time Profiler, Metal System Trace, Energy Log, App Launch) | iOS frame, GPU, energy, start |
| `adb shell am start -W`, `adb logcat` | Cold start |
| Unity Build Report, `bundletool`, App Thinning Size Report | Install size |
| Battery Historian, `dumpsys batterystats` | Android battery |
| Go benchmarks, `benchstat`, `pprof`, `runtime/metrics` | Server CPU, allocation, regression |
| Load bot (`server/cmd/loadbot`, #192), Prometheus histograms | Server latency and capacity |

## 8. Pull request rule

A performance path is any change to: client `Update` / `LateUpdate` / `FixedUpdate` loops, input handling, animation and effects, the match scene (prefabs, materials, shaders, atlases, UI), pooling, the server room loop and scheduler, protocol encode / decode and connection handling, and `server/internal/rules` code that runs per move.

Such a PR includes before and after numbers for every budget it can affect, each with device (model and tier), build type, method (the row in section 3 or 6), sample size and duration, and a pass / fail verdict. If a budget cannot be affected, the PR says which and why. A regression against a budget is fixed before merge, or the exception is recorded in `docs/adr/`. The checklist item "Changes on a performance path include a measurement" in `.github/pull_request_template.md` is ticked only when this section is satisfied.

Paste-ready block for the PR body under Verification:

```markdown
### Performance measurement
- Path touched:
- Device / tier / build:
- Method (docs/performance.md row):
- Sample (runs x duration):

| Metric | Before | After | Budget | Result |
|---|---|---|---|---|
| Work time p50 / p95 (ms) | | | 11.0 / 14.0 | |
| Missed vsyncs (%) | | | <= 1 | |
| Max batches | | | < 150 | |
| GC alloc per frame (B) | | | 0 | |
| Cold start median (ms) | | | < 3000 | |
| Install size (MB) | | | < 150 | |
| Move round-trip p99 (ms) | | | < 100 | |
| Benchmark delta (benchstat) | | | no regression | |

Not affected and why:
```

## 9. Open items

- Thresholds are initial; #191 (certification) and #164 (battery) may tighten them by PR to this file.
- Skinned 3D characters (ADR 0009) are part of the match scene: their skinning, draw calls and memory are inside the client budgets above and are first measured in #93.
- The replay harness, load bot and metrics endpoint do not exist yet (#192). Until they do, gates are checked by hand with the Profiler and the commands above.
