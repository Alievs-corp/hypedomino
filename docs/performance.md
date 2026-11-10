# Performance budget and measurement method

Budgets come from `CLAUDE.md` and are not changed here; this file fixes how each is measured and when it passes. A budget without a passing measurement is not met. Tool and harness work is tracked in #192 (load bot, replay harness); device certification is #191; battery measurement is #164. Decisions that affect these paths are recorded in `docs/adr/`.

## 1. Device matrix

Initial reference matrix. #191 certifies on it; changing a model needs a PR to this table.

| Tier | Android reference | SoC / GPU / RAM / OS | iOS reference | Role |
|---|---|---|---|---|
| Low | Samsung Galaxy A14 4G | Helio G80, Mali-G52, 4 GB, Android 13 | iPhone XR (A12, 3 GB, iOS 17) | Playable at 30 fps minimum, reduced effects |
| Mid | Samsung Galaxy A34 5G | Dimensity 1080, Mali-G68, 6 GB, Android 14 | iPhone 12 (A14, 4 GB, iOS 17) | Gating tier: 60 fps and all start, draw call, GC, size, battery gates |
| High | Samsung Galaxy S24 | Snapdragon 8 Gen 3, Adreno 750, 8 GB, Android 14 | iPhone 15 Pro (A17 Pro, 8 GB, 120 Hz ProMotion) | 120 fps option |

Class rule for devices outside the matrix: Low = under 4 GB RAM or GPU older than Mali-G57 / Adreno 610 class; Mid = 4 to 8 GB with Mali-G68 / Adreno 619 class or better; High = 8 GB or more, flagship SoC, 120 Hz panel. The tier decides the graphics level (1.1); Low runs the reduced-effects level.

### 1.1 Graphics levels

The player picks Auto (the default) or one of four graphics levels in settings (#157). A level sets the effect level (`docs/ui-motion.md` section 7) and the resolution the camera renders at. Values are admin config (`effects.graphics` in `docs/design-tokens.json`), with the defaults below.

| Graphics level | Effect level | Camera render short side, at most |
|---|---|---|
| Low | low | 720 px |
| Medium | medium | 1080 px |
| High | high | 1080 px |
| Ultra | high | 2160 px |

The render scale is `min(1, cap / native short side)` and is set on the one URP asset (`renderScale`); there is one URP asset for every level. UI on screen-space overlay canvases always renders at the native resolution, so text stays sharp. Ultra differs from High only on a panel whose short side is over 1080 px. The frame rate is #158 and does not change the level.

| Device tier | Auto picks | Highest the player may pick |
|---|---|---|
| Low | Low | Low |
| Mid | Medium | Medium |
| High | High | Ultra |

The player may pick any level from Low up to the highest; a higher one is shown locked (`docs/ui-motion.md` section 7). A saved choice above the highest, for example after the class rule changes, counts as the highest.

**Device tier.** The client applies the class rule above at every start, from `SystemInfo.systemMemorySize`, `SystemInfo.graphicsDeviceName` and the highest refresh rate in `Screen.resolutions` (`effects.graphics.deviceClass`):

| Tier | Rule |
|---|---|
| Low | Memory under 3200 MB (a 4 GB device reports less than 4096 MB), or the GPU name matches a low pattern |
| High | Memory 7000 MB or more, the GPU name matches a high pattern, and a refresh rate of 120 Hz or more |
| Mid | Every other device, including a GPU between the Low and Mid classes; the probe below lowers it if it cannot carry Medium |

| Patterns | Default (regular expressions on the GPU name) |
|---|---|
| Low | `^Mali-(T\d+\|G31\|G51\|G52\|G71\|G72)\b`, `^Adreno \(TM\) ([1-5]\d\d\|60\d)\b`, `^PowerVR Rogue` |
| High | `^Adreno \(TM\) (7[3-9]\d\|8\d\d)\b`, `^Mali-G7[1-2]\d\b`, `^Immortalis`, `^Samsung Xclipse`, `^Apple (A1[5-9]\|A[2-9]\d\|M\d)\b` |

#191 checks the rule on every certified device; a device in the wrong tier is fixed by a PR to the patterns or thresholds.

**Probe.** The class rule is a guess; the first matches check it. While the choice is Auto, a match measures work time (section 2) with `FrameTimingManager` over 1800 frames after 300 warm-up frames (`probe`). If work time p95 is over the `workTimeP95Ms` of the level's effect level, Auto drops one level for the next match, the dropped level becomes the highest the player may pick, and the probe runs again at the new level. Once a window passes, the probe stops. A new app version clears the drops and runs the probe again. The probe never raises a level, never drops below Low, and makes no change on a device without frame timing data.

**Thermal.** When the device reports heat, the level in use drops without changing the saved choice. Android 10 and later report it through the `PowerManager` thermal status listener; Android 7.1 to 9 give no signal and never drop. iOS reports `ProcessInfo.thermalState`.

| Status | Android | iOS | Level in use |
|---|---|---|---|
| Normal | none, light | nominal, fair | The chosen level |
| Moderate | moderate | serious | One level lower, not below Low |
| Severe | severe, critical, emergency, shutdown | critical | Low |

A hotter status applies at once. A cooler one applies only after it has held for 60 s (`thermal.recoverAfterSeconds`), and any hotter reading in between restarts the wait, so the level does not flicker at a boundary. While the status is not normal, the probe does not run and a window that ends hot is ignored. The client watches the status on Android only for now; iOS is in section 9.

### 1.2 Frame rate

The player picks Auto (the default) or a frame rate in settings (#157). Values are admin config (`effects.graphics.frameRate` in `docs/design-tokens.json`, and `targetFps` of each effect level), with the defaults below.

| Rule | Default |
|---|---|
| Rates offered | 30, 60, 90, 120 |
| A rate is offered when | It is at most the `targetFps` of the effect level in use (low 30, medium 60, high 120), and a refresh rate the display reports in `Screen.resolutions` is a whole multiple of it, so frames are paced evenly |
| Auto | The highest offered rate up to 60. A 120 option is opt-in |
| Menus (lobby, shop, profile, settings) | At most 60. The chosen rate applies in a match only |
| Battery saver | 30 everywhere, while the in-game toggle or the system saver (Android `PowerManager.isPowerSaveMode`) is on. The graphics level does not change |

A saved rate above the highest one offered counts as the highest. The level in use decides the cap, so a thermal drop (1.1) lowers the rate too, and it comes back with the level.

The client sets `QualitySettings.vSyncCount` to 0 and `Application.targetFrameRate` to the rate in use, on every platform. On Android, Optimized Frame Pacing (Swappy, on in Player Settings) paces frames and asks the display for a matching refresh rate. The match scene tells the client when a match starts and ends. The system saver is read at start and each time the app regains focus, because the quick settings shade takes focus; it is not polled during a match, so a JNI call never allocates in a match frame. iOS Low Power Mode needs a native plugin (section 9).

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
| Canvas rebuild | p95 <= 1.5 ms per frame | Mid, development build, UI screens | Canvas bench (`client/README.md` "Canvas bench"): `ProfilerRecorder` on the `PostLateUpdate.PlayerUpdateCanvases` marker per frame over the `static`, `scroll`, `rebind` and `pushpop` phases, read with `adb logcat -s Unity` | 3 cycles x 4 phases x 1200 frames; cycle 1 is warm-up and not gated | Every phase of cycles 2 and 3: p95 <= 1.5 ms, and the `static` phase p95 <= 0.3 ms; `scroll` and `rebind` allocate 0 B. First run (Galaxy A36, Mid, 2026-11-10): static 0.27, scroll 0.98, rebind 1.38, pushpop 1.25 ms p95, pass |
| Install size | < 150 MB | Android AAB and iOS IPA, release | Unity Build Report (`Library/LastBuild.buildreport`, Build Report Inspector) for the content breakdown. Android: `bundletool build-apks --bundle=app.aab --output=app.apks --device-spec=<tier>.json` then `bundletool get-size total --apks=app.apks`. iOS: App Thinning Size Report from the Xcode archive | Every release candidate | Android: max download size over the low, mid and high device specs (`tools/install-size/`) < 150 MB, where 1 MB is 1,000,000 bytes; the `Release build` workflow measures it and fails over the budget (`docs/content-delivery.md` 6). iOS: compressed size for the mid and high models < 150 MB. Content fetched on demand through Addressables is excluded from this number but listed in the report |

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

Run on the reference node spec fixed in #192 and recorded with the results.

### 6.1 Reference run

- Node under test: Hetzner Cloud CCX23 (4 dedicated vCPU, 16 GB RAM), location Falkenstein (`fsn1`), Ubuntu 24.04, one server process and nothing else. Dedicated vCPU keeps a neighbour's load out of the numbers, so runs can be compared.
- Bot host: a second CCX23 in the same location, joined to the node by a Hetzner private network; the load bot never runs on the node. The debug port (`HD_DEBUG_ADDR`) listens on the private address only.
- Peak target: the issue's "expected first-days peak ×3" has no number yet. It is set from the CB-3 size after the closed beta (`docs/closed-beta.md` CE-14); until then the run proves the 10,000-room node target of 6.2.

Steps:

1. Build on any machine: `cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o loadbot ./cmd/loadbot`, then copy `loadbot` and `tools/loadtest/*.sh` to both hosts.
2. Node: `./node.sh <node private address> node-run`. It raises the file limit, records the machine (`node.txt`), samples CPU every 10 s (`vmstat.log`) and runs `loadbot serve` with the same `HD_*` variables as the server.
3. Bot host: `./bot.sh <node private address> bot-run`. Defaults are those of 6.2: 10,000 rooms, 50% 2v2, 5 min ramp, 30 min hold; extra `loadbot run` flags are appended and override them. It samples `/metrics` every 60 s, keeps the bot report (`report.json`, `report.txt`), the final metrics and a heap profile, and exits non-zero when the bot gates fail.
4. Copy `node-run/vmstat.log` into `bot-run`, then `./check.sh bot-run 5 30`. It checks heap drift between hold minute 10 and minute 30 (live heap after GC, `go_heap_live_bytes`) and average CPU during the hold, and prints the server histogram p99. Exit code 1 means a gate failed.
5. Record in #192: commit, `node.txt`, `report.txt` and the `check.sh` output. A soak is the same with `-hold 12h` and `./check.sh bot-run 5 720`.

Reconnect storm: the same run with `./bot.sh <node private address> storm-run -storm 15m` drops every live connection at once at hold minute 10, and each seat dials again at once and resumes with its acked sequence. The report adds the dropped seats, the seats back in their match and the slowest return; the 1-minute windows around the storm show its cost on the move round-trip. The usual gates still decide the exit code. There is no recovery target yet: the first reference storm run is recorded in #192 and the owner sets the target from it.

TLS run: production terminates TLS on the Hetzner load balancer (ADR 0010, ADR 0011), so the TLS run puts one in front of the node and the bot connects through it; the node binary never serves TLS. 10,000 rooms at 50% 2v2 hold 30,000 WebSocket connections (5,000 × 2 + 5,000 × 4), so the run uses an LB31 in `fsn1`; check its connection limit against the report's peak connections before the run. Setup: attach the LB to the node's private network, add the node as a target by its private IP, add an `https` service on 443 to destination port 8080 with a Hetzner managed certificate for a subdomain of the project domain (for example `load.<domain>`) and an HTTP health check on `/readyz`. Then run `./bot.sh <node private address> tls-run -target wss://load.<domain>/ws`: the extra `-target` overrides the default, while `/metrics` and the heap profile are still read from the node's private address. Compare its report with the plain run of the same size; the difference is the cost of the TLS handshake and the load balancer hop, and the storm variant (`-storm 15m` added) shows the handshake cost of 30,000 reconnects at once. Delete the LB and its certificate after the run.

Database and Redis run: a move touches neither PostgreSQL nor Redis, and a connection verifies its token in process, so their load comes per match: the match row and its participants when the room opens, the result, replay and seat results when it ends, and one route claim and one release in Redis per seat. The node runs `node.sh` with `HD_DATABASE_URL` (migrated with `cmd/migrate up`) and `HD_REDIS_URL` plus `HD_NODE_ID` exported, and `loadbot serve` makes those writes for every room. `check.sh` then prints the p99 of `match_begin_seconds`, `match_finish_seconds`, `route_claim_seconds` and `route_release_seconds` and the `hd_store_failures` count; the matches finished in the bot report over the run time give the write rate the node needs. The usual gates still decide the exit code, so the run shows whether the writes cost the move round-trip. Record the database and Redis hosts with the results. There is no write latency target yet: the owner sets it from the first run. Rating, reward and outbox writes after a match are not wired to rooms yet and are not in this run.

### 6.2 Budgets

| Budget | Target | Environment | Method | Sample and duration | Pass / fail |
|---|---|---|---|---|---|
| Move round-trip | p99 < 100 ms | Load bot in the same cloud region as the node, on a different host; production-like TLS and WebSocket | Load bot (`server/cmd/loadbot`, #192) timestamps each move intent send and the resulting broadcast receive. The server also exports a Prometheus histogram `move_roundtrip_seconds` (with the node gauges `go_heap_live_bytes`, `go_goroutines`, `hd_rooms` and `hd_connections` beside it) for its in-process share: an accepted move intent from its receive by the gateway to the resulting broadcast queued for every seat, so the socket writes are left to the bot measurement | 30 min steady state at the 10k-room load, at least 100,000 moves, first 2 min discarded as warm-up | Bot-measured p99 < 100 ms overall and in every 1-minute window; server histogram p99 reported alongside |
| Concurrent rooms | 10,000 per node | Reference node | Load bot ramps to 10,000 active rooms (1v1 and 2v2 mix, moves at human cadence inside the admin-config turn time), then holds. Observe with pprof, `runtime/metrics` and node CPU / RAM | Ramp, then 30 min hold | Holds 10,000 rooms within the move budget; error and disconnect rate under 0.1%; average CPU under 80%; heap flat within 5% between minute 10 and minute 30 |
| Hot-path cost | No regression | Developer machine and CI | `go test -run='^$' -bench=. -benchmem -count=10 ./internal/rules/... > new.txt`, then `benchstat old.txt new.txt` | 10 runs per side | No benchmark slower than 5% at `p < 0.05`; allocs/op not higher |

Profiling:

```bash
go test -run='^$' -bench=<Name> -cpuprofile cpu.out -memprofile mem.out ./internal/rules
go tool pprof -http=:0 cpu.out
go tool pprof -http=:0 "http://<node>:<debug-port>/debug/pprof/profile?seconds=30"
go tool pprof -http=:0 "http://<node>:<debug-port>/debug/pprof/heap"
```

The debug port is `HD_DEBUG_ADDR`; it is never exposed publicly.

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
- The load bot exists (`server/cmd/loadbot`, #53, `server/README.md` "Load test"); its bot-measured round-trip, error rate and heap drift are the server gates above. The `move_roundtrip_seconds` histogram is served at `/metrics` and pprof at `/debug/pprof/` on the debug listener (`HD_DEBUG_ADDR`, off when empty, `server/README.md`). The replay harness does not exist yet (#192); until it does, client gates are checked by hand with the Profiler and the commands above.
- Thermal status on iOS (1.1): reading `ProcessInfo.thermalState` needs a native plugin, added when the first iOS build exists (#159). Until then an iOS device never drops for heat. iOS Low Power Mode (1.2, `ProcessInfo.isLowPowerModeEnabled`) needs the same plugin; until then only the in-game battery saver toggle applies on iOS.
