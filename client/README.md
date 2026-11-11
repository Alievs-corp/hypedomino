# client

Unity 6 LTS game client (URP Universal Renderer with 2D sprites and real-time 3D characters, ADR 0009; C#, Addressables). It renders state and sends player intents only; the server decides every rule, score and reward.

## Editor

Unity `6000.3.27f1` (Unity 6.3 LTS), pinned in `ProjectSettings/ProjectVersion.txt`. Install it with Unity Hub together with Android Build Support (OpenJDK, Android SDK & NDK Tools). An upgrade is its own PR with a measurement (ADR 0001).

## Layout

| Path | Assembly | References |
|---|---|---|
| `Assets/HypeDomino/Core` | `HypeDomino.Core` | none |
| `Assets/HypeDomino/Net` | `HypeDomino.Net` | Core, Proto |
| `Assets/HypeDomino/UI` | `HypeDomino.UI` | Core, `UnityEngine.UI` |
| `Assets/HypeDomino/Cosmetics` | `HypeDomino.Cosmetics` | Core |
| `Assets/HypeDomino/Game` | `HypeDomino.Game` | Core, Net, Proto, UI, Cosmetics |
| `Assets/HypeDomino/Generated` | `HypeDomino.Proto` | `Google.Protobuf.dll` only; output of `npm run proto:gen`, never edited |
| `Assets/HypeDomino/Tests/EditMode` | `HypeDomino.Tests.EditMode` | Editor only; references Core, UI, Game, Proto and Content |
| `Assets/HypeDomino/Tests/PlayMode` | `HypeDomino.Tests.PlayMode` | References Core, UI and `UnityEngine.UI` |
| `Assets/HypeDomino/Settings` | | URP asset, Universal Renderer, texture preset |
| `Assets/HypeDomino/Scenes` | | `Boot.unity`, the first build scene |
| `Assets/Plugins/Google.Protobuf` | | Google.Protobuf 3.36.2 (`netstandard2.0`), the runtime matching the `protocolbuffers/csharp:v36.2` plugin in `proto/buf.gen.yaml`, and its dependency System.Runtime.CompilerServices.Unsafe 4.5.3 (`netstandard2.0`, assembly 4.0.4.1), which the player profile lacks; `System.Memory` comes from the .NET Standard 2.1 profile |

References only point down this table; Core depends on nothing. `Assets/link.xml` keeps the protobuf assemblies from high managed stripping, because message descriptors are read by reflection.

## Settings

The `ProjectSettingsTests` EditMode tests guard the scripting, Android architecture, texture compression, rendering, quality level and profiling rows.

| Area | Setting | Why |
|---|---|---|
| Scripting | IL2CPP, high managed stripping, engine code stripping, .NET Standard 2.1 on Android and iOS | ADR 0001; install size |
| Android | ARM64 only; Vulkan, then OpenGL ES 3 | Install size; no 32-bit devices in the matrix (`docs/performance.md` 1) |
| iOS | Metal | Only API on target devices |
| Rendering | URP Universal Renderer, Forward path; no HDR, MSAA, depth or opaque texture, post-processing, shadows or additional lights | ADR 0009, `docs/ui-motion.md` 4 |
| Quality levels | No per-level pipeline; every level uses the default URP asset; `vSyncCount` 0 | One renderer; graphics levels set only its render scale (`docs/performance.md` 1.1); `Application.targetFrameRate` alone sets the rate (1.2) |
| Profiling | Frame Timing Stats on | `FrameTimingManager` measurements, `docs/performance.md` 2 |
| Color space | Linear | Lit 3D characters, ADR 0009 |

## Textures and atlases

- Android texture compression is ASTC only; the build subtarget is ASTC (ADR 0002).
- `Settings/SpriteTextureASTC.preset` is the default preset for every new texture: Sprite, no mipmaps, ASTC 6x6 on Android and iOS. ASTC 6x6 is the base-colour block size of ADR 0009. Per-class block and maximum sizes are open item O-5 in `docs/asset-workflow.md` (#93, #135); a class that needs other values adds a preset with a folder filter above this one in the Preset Manager.
- Sprite Atlas V2 is the packer. One atlas per cosmetic item inside its Addressables group and effect sprites only in the shared effects atlas (`docs/cosmetics.md` 8.2, `docs/art-direction.md` 6).

## Measuring the empty build

`Scenes/Boot.unity` holds an orthographic camera and `FrameTimingSampler`. The `BootSceneSmokeTests` PlayMode smoke test checks that it is the first build scene, loads it, runs 60 frames and fails on any error or exception logged meanwhile. The sampler measures at the frame rate `GraphicsBoot` set (`docs/performance.md` 1.2; the highest refresh rate when none is set) and logs every 1200 frames:

```
FrameTiming target=<Hz> frames=1200 workP50=<ms>ms workP95=<ms>ms missedVsync=<percent>% gcFrames=<n> gcMax=<bytes>B gcCollections=<n> batchesMax=<n>
```

Work time is `max(cpuMainThreadFrameTime, cpuRenderThreadFrameTime, gpuFrameTime)` (`FrameWindow.WorkMs`), shared with `GraphicsProbe` and the benches; a missed vsync is a presented interval longer than 1.5 refresh periods (`docs/performance.md` 2). Read it with `adb logcat -s Unity`.

`gcFrames` counts frames whose `ProfilerRecorder` "GC Allocated In Frame" was above 0, `gcMax` is the largest of them, `gcCollections` is the `GC.CollectionCount(0)` delta and `batchesMax` the largest "Batches Count" (`docs/performance.md` 3). A counter the player does not expose prints `n/a`; the Profiler counters need a development build. The frame that writes the log line allocates that line, so its GC value is not counted. `FrameWindow` holds the window and the `FrameWindowTests` EditMode tests cover it.

In a development build `FrameStatsOverlay`, on the same object as the sampler in `Scenes/Boot.unity`, shows the sampler's live window of the last 60 frames in the top-left of the safe area: work time p95 in ms, missed vsyncs in percent, frames that allocated (in the danger colour when above 0) and the most batches. It refreshes once per window, takes its numbers from string tables filled at start, so showing them allocates nothing (`FrameStatsOverlayTests`), and removes itself in a release build, so store screenshots never show it.

Size: the `Release build` workflow below; the Build Report (`Library/LastBuild.buildreport`) of a local release build gives the content breakdown.

## App flow

`Game/AppFlow` holds which part of the app is shown: `Boot`, `Login`, `Lobby`, `Match`, `UpdateRequired` or `Maintenance`. It is plain C#, fed by server messages, and raises `Changed` once per new state.

- `SignInNeeded` (no session or `UNAUTHENTICATED`) goes to `Login`, except while an update is required.
- An accepted `ServerHello` goes to `Match` when `in_match` is set (R-58), otherwise to `Lobby`. `UPDATE_REQUIRED` and `MAINTENANCE` go to their screens (ADR 0004). An unknown status changes nothing.
- A `MatchView` that is not finished takes the lobby into the match, so a match found by the queue or a friend room starts the same way.
- `LeaveMatch` returns to the lobby only after the match is finished.

The `AppFlowTests` EditMode tests cover every transition. Scenes, DI and the network wiring that feeds it are not built yet.

## Network frames

`Net` reads and writes the `Envelope` frames of `proto/README.md` without the generated `Envelope` parser, which allocates a new payload message on every parse. `FrameReader.TryRead` walks the frame by hand and returns a `Frame` (a `ref struct`): the header (`protocol_version`, `sequence`, `acked_sequence`, `session_nonce`), the payload case and the payload bytes as a span over the frame. The caller merges those bytes into a message it keeps (`message.MergeFrom(frame.PayloadBytes)`). A field it does not know, including a payload number newer than this client, is skipped, so the payload reads as `None`; a truncated or malformed frame returns `false`. `FrameWriter` writes a header and one intent into its own buffer and returns a span over it, valid until the next write; an intent kept and changed in place is sent without allocating. The `FrameCodecTests` EditMode tests read every golden frame in `proto/testdata/frames` like the generated parser does, write every client golden frame byte for byte, refuse malformed frames, and check that reading a `MatchView` frame and writing a reused intent allocate nothing.

`MatchSnapshot.TryRead` decodes a `MatchView` payload by hand, including each seat's `disconnected` and `consecutive_timeouts`, into arrays it keeps, because the generated parser allocates every tile, seat and legal move. Lists come back as `ReadOnlySpan`s of structs (`TileFace`, `SeatState`, `OrientedTileFace`, `OpenEndValue`, `LegalMoveValue`), sized by the set: 28 tiles (`docs/rules.md` R-1), 4 seats, 4 open ends, 112 legal moves; more is refused. A field absent from the frame reads as its default, and `HasSpinner`, `HasForcedOpening` and `HasLastResult` tell an absent message from a zero one. The roster and table follow `proto/README.md`: a view without them keeps the ones already read, and they are the only parts that allocate, once per connection. A malformed view returns `false` and leaves the snapshot partly written, so the caller drops it. The `MatchSnapshotTests` EditMode tests compare every golden `MatchView` and one that sets every field with the generated parser, and check roster keeping, shrinking lists, the size limit, unknown fields, truncation and that a view without a roster allocates nothing.

`Session` speaks the protocol over an `IFrameTransport`, which sends one frame and hands back received frames on the main thread. `Hello` takes the transport of the current connection and sends the `ClientHello` with the last acknowledged sequence. An accepted `ServerHello` gives the `session_nonce` that `Send` stamps on every later intent. `PlaceTileIntent`, `PassIntent`, `DrawIntent` and `EndTurnIntent` take the next sequence; loading, emoji, quick message and queue intents carry none. The counter never falls behind the server's `acked_sequence`, so a resumed session goes on after the last accepted intent instead of being dropped as stale. `Poll` reads every waiting frame: a `MatchView` goes into the `MatchSnapshot` the session keeps, and `ServerHello`, `Error`, `QueueState`, `EmojiShown` and `QuickMessageShown` are raised as events with the generated message. A frame that does not read, carries a newer `protocol_version` or a payload only a client sends breaks the session: it stops reading and refuses intents until the next hello. Room frames arrive without `protocol_version`, so a missing one is read. The `SessionTests` EditMode tests send the golden hellos byte for byte, number intents, resume after an acked sequence, route every server payload, break on bad frames, and check that polling a `MatchView` without a roster and sending a reused turn intent allocate nothing. Every allocation test runs its step once before measuring, so first-call setup is not counted.

`WebSocketTransport` is the `IFrameTransport` over `System.Net.WebSockets.ClientWebSocket`. `Connect` dials once on a background receive thread; a send thread starts when the socket opens. Both rings are preallocated: `Send` copies the frame into one of 8 send slots of `Session.MaxSentFrameBytes`, so frames sent while connecting go out in order once the socket opens, and the receive thread fills one of 8 receive slots of 16 KiB. `TryReceive` returns the oldest frame as a span over its slot, valid until the next call; a full receive ring stops reading the socket until the main thread polls again. Neither call allocates on the main thread. A close from the server, a text frame, a frame over 16 KiB, a full send queue, a failed connect or `Dispose` moves `State` to `Closed` with `CloseStatus` and `CloseDescription`; frames received before the close can still be read. The background threads still allocate inside `ClientWebSocket`, a few times per frame. The `WebSocketTransportTests` EditMode tests run it against a loopback WebSocket server written over `TcpListener`.

`Connection` keeps one `Session` alive across sockets; the app calls `Tick` once per frame. `Start` dials, sends the hello from the app's `ClientHello` provider and waits in `Connecting`; an accepted hello moves it to `Online`. A socket that closes, or a session that breaks, is dropped after the frames already received are read, and the next socket is dialled on this schedule: at once, then after 0.5, 1, 2 and 4 s, then every 5 s, each delay spread by ±20 % so a server restart does not bring every player back in the same instant. The server keeps a dropped seat for `HD_RECONNECT_GRACE` (60 s), about 14 attempts. An accepted hello starts the schedule again, so a close after `MATCH_FOUND` on another node is redialled at once. The resume hello carries the last acked sequence. `UPDATE_REQUIRED`, `MAINTENANCE`, `UNAUTHENTICATED`, `INTEGRITY_FAILED` and `UNSUPPORTED_PROTOCOL` stop the connection: dialling again cannot fix them, and the app shows the matching screen and calls `Start` when it can go on.

`ServerClock` maps the local monotonic clock (`Time.realtimeSinceStartupAsDouble`, not the device clock the player can change) to server time from the hello's `server_time_unix_ms`, taken as the middle of the hello round trip; the server's sign-in work inside that round trip makes it a little late, which is well under a second and fine for turn and loading deadlines. `RoundTrip` times each turn intent from `Send` to the first frame whose `acked_sequence` covers it, as read by `Tick`, so a sample includes up to one frame of polling; a rejected intent is not timed. `SmoothedMs` follows the samples with a gain of 1/8, the smoothing TCP uses for its round-trip estimate (RFC 6298). The `ConnectionTests` EditMode tests drive it with a fake socket and clock: the schedule and its jitter, reset on resume, the resume hello, frames read before a close, every stopping answer, clock mapping, round-trip samples and that a quiet `Tick` allocates nothing.

`ConnectionNotices` decides what the connection UX of #87 shows, from the `Connection` and the latest `MatchSnapshot`. `Network` is `Reconnecting` while a session that was online waits or dials again (the first dial is the boot screen's, not a notice), and `Weak` while online when the smoothed round trip is over 300 ms or a turn intent has waited more than 2 s for its ack (thresholds chosen by the owner). `Seat` is `Dropped` for a seat whose player lost the connection but still holds the seat, and `BotPlaying` for a human seat the bot plays after the AFK mark or the reconnect grace; a bot from the roster has no notice. `OneTimeoutFromBot` is true when the player's own next timeout would hand the seat to the bot (R-51), from the seat's `consecutive_timeouts` and the view's `afk_consecutive_timeouts`. The screens that show them come with the match scene. The `ConnectionNoticesTests` and `ConnectionTests` EditMode tests cover every notice and both thresholds.

## Shop and wallet

`UI/Shop` turns server replies into what the shop shows, without computing anything (`docs/economy.md` 3.4, 6.1). `WalletReply`, `ShopReply` and `PurchaseReply` read `GET /wallet`, `GET /shop` and `POST /shop/buy` with `JsonUtility`. `Balances` keeps coins and crystals apart and takes the new balance of a purchase's `currency` from the reply. `ChestCard` copies a `chest` offer's pool, owned items, odds (`[{"rarity","percent"}]`), opens left to the pity, price and refusal reason as sent. `ChestReveal` holds only the granted item, its rarity key and the opens left to the pity, so a reveal can show nothing but the result: no reels, wheels, slot sounds or near-miss effects (`docs/localization.md` section 6). Names are keys: `currency.coin.name`, `currency.crystal.name`, `rarity.<id>.name`.

`UI/Localization/builtin-strings.json` is the built-in bundle (`docs/localization.md` 2.2) for keys the client needs before a published bundle arrives, starting with the currency names of `docs/economy.md` OQ-13. `BuiltinStrings` gives the built-in value of a key in one language. The `ShopTests` EditMode tests cover these types.

## Localization

`UI/Localization` is the client side of `docs/localization.md` (section 2.4).
- **`GameLanguage`:** picks the language.
- **`GameStrings`:** looks a key up along the fallback chain over downloaded and built-in bundles, and swaps a staged bundle in outside a match.
- **`StringBundle`:** reads a published bundle after its hash check.
- **`MessageFormat` and `PluralRules`:** render placeholders, plurals and selects.

The `LocalizationTests` EditMode tests cover them. Format a string when a screen opens or a value changes, never per frame.

## Sign-in

`Net/Auth/AuthSession` is the client side of `POST /auth/nonce`, `/auth/signin`, `/auth/refresh` and `/auth/logout` (`server/internal/auth`). It is plain C# behind three seams:
- `IIdentityProvider` is the native Google or Apple sign-in. It gets the nonce to put in the ID token: the server nonce for Google, its lowercase SHA-256 hex for Apple. It answers a token, `Cancelled` or `Failed`.
- `IAuthTransport` posts JSON and answers the status and body; status 0 means no response.
- `ISecureStore` is the platform secure storage (Keystore, Keychain; `docs/privacy.md` P-1). It holds only the refresh token (`auth.refreshToken`) and a random device id made once (`auth.deviceId`), never a hardware identifier.

`SignIn` gets a nonce, asks the provider for an ID token and sends it with the raw nonce and the device id. `Resume` is the silent sign-in at later starts: it refreshes the stored token with the same device id. The access token, the request key and the account id stay in memory only.

Outcomes:
- `SignedIn`, with `NewAccount` set on the first sign-in so the profile setup can follow.
- `Cancelled` when the player closes the account picker, `ProviderFailed` when the SDK fails.
- `Offline`, `Unavailable` (5xx or an unreadable reply) and `RateLimited` (429). A refresh that fails this way keeps the stored token.
- `Rejected` for a refused sign-in, `Banned`, `Deleted` and `DeletionPending` with `DeleteAt`. A refresh refused this way forgets the stored token; a refused refresh token answers `SignInRequired`.
- `Busy` while another call runs, `Unsupported` when the platform has no such provider.

`SignOut` forgets the session at once and revokes the refresh token on the server. The `AuthSessionTests` EditMode tests cover these flows. The native Google Sign-In and Sign in with Apple plugins, the Keystore and Keychain store and the HTTP transport are not built yet.

## Graphics levels

`UI/Graphics` holds the device class rule and `GraphicsQuality`: Auto or a chosen level, the highest level the device tier allows, its effect level and render scale (`docs/performance.md` 1.1). Values come from `TokenGraphics`. `Game/GraphicsBoot` runs before the first scene: it classes the device, reads the saved choice, sets the render scale on the URP asset outside the editor and logs one `Graphics tier=... level=...` line. `Game/GraphicsProbe` is the first-match probe: the match scene adds it, and while the choice is Auto and the probe is not settled it measures work time and hands the p95 to `GraphicsBoot.SettleProbe`, which saves the drops with the app version and logs one `GraphicsProbe ...` line. `Game/ThermalWatch` starts with it on Android 10 and later: a `PowerManager` thermal status listener feeds `GraphicsQuality.UpdateHeat` every frame, and each level change logs one `Graphics level=... heat=...` line. `FrameRatePolicy` holds the frame rate (`docs/performance.md` 1.2): the rates offered for the level in use and the display, Auto or the chosen rate, the menu cap and the battery saver. `GraphicsBoot` sets `Application.targetFrameRate` to its `Target` and logs one `Graphics targetFps=...` line per change; the match scene calls `FrameRate.SetInMatch`, and `Game/PowerSaveWatch` reads the Android system saver at start and on focus. The `GraphicsQualityTests` and `FrameRatePolicyTests` EditMode tests cover the class rule on the matrix devices, the choice rules and the frame rate rules.

## Avatars

`UI/Cosmetics` holds `AvatarView` and the rules for choosing static or animated art (`docs/cosmetics.md` 8.3). It takes art through `ICosmeticArtSource`, which has no Addressables implementation until #68.

## Rarity marks

`UI/Rarity` draws the rarity diamond and halo from a `RarityTable` (`docs/rarity.md` 2.2.1). `rarity-defaults.json` there is generated by `npm run tokens:gen`; never edit it by hand.

## Design tokens

`UI/Tokens/DesignTokens.cs` is generated from `docs/design-tokens.json` by `npm run tokens:gen`; edit the JSON and regenerate, never the file. It holds `TokenColor` (`Color32`), `TokenOpacity`, `TokenSpacing`, `TokenRadius` and `TokenSize` (UI units on the 1080-unit short side), `TokenTextSize` (the font size of each type style), `TokenDurationMs`, `TokenEasing` (cubic-bezier points as `Vector4`), `TokenMotion`, `TokenEffectLevel` (per effect level: `MaxAnimatedCosmetics`, `WorkTimeP95Ms`, `MaxTweens`) and `TokenEffect` (each effect's `DurationMs` and `Pooled` count, the largest over the levels). `UI/Motion/Easing` evaluates a `TokenEasing` curve. Rarity colours are not constants: they arrive with the admin-published rarity config (`docs/admin.md`). `npm run test:tools` fails when the file is stale, and `RuntimeCodeRulesTests` fails on `new Color(...)`, `new Color32(...)` or a named `Color` such as `Color.white` in runtime code outside it.

## Screens

`UI/Screens` holds the screen stack of `docs/ui-motion.md` 5 "Screens". A `ScreenView` is one screen: its own overlay `Canvas` scaled so the short side is 1080 units, a `CanvasGroup` content root that fades and scales, and for a modal a full-screen scrim `Image` set by `Configure`. `Configure` moves the content and the scrim into a nested `Body` canvas, which also holds the screen's `GraphicRaycaster`. Hiding disables the outer `Canvas` only, so the graphics stay registered with the body and a push or pop allocates nothing; a graphic placed straight under the outer canvas would be registered again on every show (#505). `ScreenStack` pushes and pops them, hands input to the top screen at once (the `GraphicRaycaster` of every other screen is off), hides a covered screen by disabling its `Canvas` only, and pops on the Android back key. Transitions are `ScreenFade` values driven by unscaled time, so an interrupted one reverses from where it is. `ScreenStack.ReduceMotion` drops the scale. The `ScreenStackTests` PlayMode tests cover push, pop, reversal both ways, modals, the root screen, reduce motion and the canvas scale.

## Lists

`UI/Lists` holds the recycling list of `docs/ui-motion.md` 5 "Lists". Put `RecyclingList` next to a `ScrollRect`, call `Build(rowTemplate, rowHeight, gap)` once when the screen builds (height and gap are `TokenSize` and `TokenSpacing` values), then `Show(count, binder)`. `Build` creates the rows that cover the viewport plus a spare row above and below, makes the content its own sub-canvas with a `GraphicRaycaster`, and pins rows to the top. `Show` sets the content height and binds the visible rows. An `IRowBinder` fills a row for an item index. Scrolling rebinds a row only when it starts to show another item, and `Refresh(index)` rebinds one item if it is on screen. `RecyclingLayout` is the row arithmetic, covered by the `RecyclingLayoutTests` EditMode tests. The `RecyclingListTests` PlayMode tests cover row count, placement, hiding past the end, one rebind per scrolled row, the canvas, and scrolling allocating 0 B.

## Board layout

`UI/Board/ChainLayout` places the chain of `docs/ui-motion.md` 5 "Board". Create it once with the set size as capacity and call `Arrange(line, anchorIndex, top, bottom, boardSize, maxScale)` when the board changes: `anchorIndex` is the spinner's index in the line, or the opening tile's while there is no spinner. Each `PlacedDomino` gives the tile's center and heading in tile units, its near and far pips and whether it lies across. `Scale` turns tile units into board units and `Focus` is where the camera centers. It allocates nothing after construction. The `ChainLayoutTests` EditMode tests cover the arms, doubles, folding and a full set on every board shape.

## Tweens

`UI/Motion/TweenPool` runs the tweens of `docs/ui-motion.md` 6 rule 3. Create it once per scene with `TweenPool.For(level)`, which takes `TokenEffectLevel.*MaxTweens`. `Start(tween, ref sink)` and `Step(nowMs, ref sink)` write values through an `ITweenSink` struct, so neither allocates: a `Tween` is a value with a target index, a channel (position, rotation, scale, opacity), its from and to values, start, duration, a `TokenEasing` curve, and whether it is cosmetic. When the pool is full, the oldest cosmetic tween lands on its end value. `FinishAll` is the jump to the final state of rule 5. A tween does nothing to its target before its start time. `UI/Board/TileTravel.Start` runs `tilePlace`: position and the shorter turn in `base` with `standard`, then the settle from `pressScale` in `micro` with `overshoot` (300 ms, the `tilePlace` duration); with reduce motion `fast` and no settle. `TileTravel.Arrive` is the deal and draw of 6 "Hand motion": tile n starts n × `micro` later. `UI/Match/PassNotice` gives the scale and opacity of the pass label at a time, like `EmojiPop`. The `TweenPoolTests` EditMode tests cover the curve, eviction, replacement, finishing, tile travel, settle, deal spacing, draw and 0 B per frame; `PassNoticeTests` cover the pass label.

## Canvas bench

`Bench/CanvasBench` measures what UI costs per frame on a device (#69). It builds a lobby screen (40 panels with labels) and a shop screen with a `RecyclingList` of 1000 rows (background, icon, name, price) on the `ScreenStack`, then runs four phases of 1200 frames, three times. `static` changes nothing. `scroll` scrolls the list at 1500 units per second. `rebind` changes the price text of every visible row each frame. `pushpop` pushes or pops the shop every 0.6 s. After each phase it logs `CanvasBench ...` with work time p50 and p95, missed vsyncs, GC frames, batches, and p50, p95 and the maximum in ms of every canvas profiler marker the player exposes (`Canvas.*`, `CanvasUpdate.*`, `UGUI.*`, `PostLateUpdate.PlayerUpdateCanvases`). Build it with `gh workflow run animation-bench.yml --ref <branch> -f bench=canvas` (artifact `canvas-bench-apk`, a development build, package `com.hypedomino.canvasbench`) and read `adb logcat -s Unity`.

The budget it checks is the "Canvas rebuild" row of `docs/performance.md` section 3: `PostLateUpdate.PlayerUpdateCanvases` p95 at most 1.5 ms in every phase, 0.3 ms in `static`.

`RecyclingList.Build` reads the viewport height, so it runs once the screen's canvas is shown and scaled, a frame after the push.

## Content

`Content` holds the client side of `docs/content-delivery.md` on Addressables 2.10.3, the version Unity 6000.3 ships. `CatalogUpdater` decides when a remote catalog is downloaded, committed, retried (2 s, 8 s, 30 s) or held back until a match ends; it is plain C# behind `ICatalogSource` and covered by the `CatalogUpdaterTests` EditMode tests. `AddressablesCatalogSource` downloads the catalog to `persistentDataPath/content`, keeps it there as `catalog_<version>.bin`, loads it with `Addressables.LoadContentCatalogAsync`, keeps only the last remote locator and stores its version and root in `PlayerPrefs` (`content.catalogVersion`, `content.baseUrl`). Every successful load runs `Addressables.CleanBundleCache`, which reads the loaded catalogs from their files and removes cached bundles none of them references (CD-24). Remote bundles are built with the load path prefix `hd-content:/`, and `ContentUrls.Resolve` rewrites it to `<content_base_url>/addressables/` through `Addressables.InternalIdTransformFunc`. `AddressablesArtSource<T>` is the `ICosmeticArtSource<T>` for `CosmeticArtCache`: one handle per address, released on the cache's last release, and a failed load is released at once. `ContentBoot` starts all of this before the first scene in a player build; in the editor Addressables reads from the AssetDatabase and no catalog is downloaded. The network layer calls `ContentBoot.Catalog.Announce(catalog_version, content_base_url)` after an accepted `ServerHello` and `SetInMatch` around a match.

`ContentCache` is the disk cache of `docs/content-delivery.md` 7: `Start` caps Unity's AssetBundle cache at 1,000,000,000 bytes and 30 days unused and runs a clear left pending from the last session, `OccupiedBytes` is the size to show, and `Clear` answers `Cleared`, `AtNextStart` (a bundle is in use) or `NotInMatch`. It is plain C# behind `ICacheStore`, covered by `ContentCacheTests`; `UnityCacheStore` is the `Caching` and `PlayerPrefs` (`content.clearAtStart`) side. The settings screen (#157) reads `ContentBoot.Cache`.

`Editor/ContentSetup` (menu HypeDomino > Apply Content Setup, also run by Apply Project Setup) writes the Addressables settings under `Assets/AddressableAssetsData`: the remote load path `hd-content:/[BuildTarget]`, a binary remote catalog that is never checked at start (the client loads catalogs itself), content built with every player build, the install-time group `cosmetic-builtin` as the default group, and the group template `Cosmetic Item` for remote item groups: one bundle per entry, a content hash in the file name, 2 retries and a 20 s timeout without progress.

## Editor methods in CI

Without a local Unity license, an editor method that creates or changes assets runs in CI: `gh workflow run unity-editor.yml --ref <branch> -f method=HypeDomino.Editor.<Class>.<Method>`. The workflow runs the method in batch mode and uploads everything it changed under `client/` as `editor-changes.patch` (artifact `editor-changes`); `gh run download <run id> -n editor-changes` and `git apply editor-changes.patch` bring it into the branch. The method runs its work through `EditorMethod.Run`, as `ContentSetup.ApplyAndExit` does: it exits the editor and writes `Logs/editor-method-succeeded` on success. The builder step always reports a failed build, because no player is built, so the run fails only when that file is missing.

## Release build

`Editor/ReleaseBuild.BuildAndroid` builds the release AAB for one catalog version (`docs/content-delivery.md` 6): it sets the Addressables player version to `-catalogVersion <version>`, builds the scenes in Build Settings into `-customBuildPath` (default `Builds/hypedomino.aab`) together with the content, and fails when `ServerData/Android/catalog_<version>.bin` is missing. It restores the player version and the AAB switch afterwards. In CI: `gh workflow run release-build.yml -f catalog_version=<version>`. The workflow measures the download size with `bundletool` on the low, mid and high device specs in `tools/install-size/specs`, fails at 150 MB or more (`docs/performance.md` 3) and keeps the AAB, the size report and the remote folder as artifacts. `ReleaseBuildTests` cover the version rule and the remote folder.

The workflow also runs on every pull request that changes `client/`, `tools/install-size/` or the workflow:
- The catalog version is `pr-<number>`; the build is unsigned and nothing is uploaded.
- The size report is in the run summary.
- It is not in `CI complete`, so a build over the budget shows as a failed check on the pull request without blocking the merge.
- A newer push cancels the running build of the same pull request.

Optional arguments stamp the release, and the build restores the project settings afterwards:
- `-appVersion <major.minor.patch>` sets the version name, the form the server compares with the minimum version (`proto/README.md`).
- `-buildNumber <n>` sets the Android version code, 1 to 2,100,000,000.
- `-applicationId <package>` sets the Android package name.
- `-androidSigning <file>` signs with the upload key: a JSON file with `keystore`, `keystorePass`, `keyAlias` and `keyAliasPass`.

Every run passes the workflow run number as the build number, so build numbers only grow. Pushing a tag `client-v<major>.<minor>.<patch>` runs the same workflow as a release:
- The version comes from the tag, and the catalog version is `<version>-<build number>`.
- The package name is the repository variable `ANDROID_APPLICATION_ID`.
- The upload key comes from the `ANDROID_KEYSTORE_*` and `ANDROID_KEYALIAS_*` secrets (`docs/environments.md` 2.3). It is written to `client/Signing` for the build only and removed afterwards.
- The signed AAB goes to the Play internal testing track with `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON`.

A tag run fails before the build when a variable or secret is missing. The iOS build and the TestFlight upload are not built yet.

## Safe area

`UI/Layout/SafeAreaFitter` sits on the root `RectTransform` of each screen and anchors it to `Screen.safeArea`, so camera holes, notches, the island and the home indicator never cover UI. It refits when the safe area or the resolution changes (rotation, split screen, a folding display) and does nothing on other frames. `SafeAreaAnchors` turns the safe area into anchors, clamped to the screen; an empty screen or safe area falls back to the full screen. The `SafeAreaTests` EditMode tests cover 20:9 Android with a camera hole, iPhone landscape and portrait, a 4:3 tablet and 16:9 without insets. The `AspectRatioTests` PlayMode tests render a `ScreenView` through a camera into a texture of each display size: every device of the matrix (`docs/performance.md` 1), 16:9, 21:9, a 4:3 tablet and a folding inner display, in portrait and landscape. On each one the short side must be 1080 units, the canvas must keep the display's aspect, and the safe content must keep its size and stay on screen. Orientation is still open (`docs/ui-motion.md` 9), so nothing here assumes one.

## Match layout

`UI/Match/MatchLayout` places the regions of the match screen inside the safe area (`docs/ui-motion.md`, Match layout): the seat panels, the hand along the bottom and the board in what is left. Sizes come from `TokenSize` and `TokenSpacing`; arranging again allocates nothing. `TableSides.Of` turns a seat into a side of the table, with the local player at the bottom and the next turn on the left; `Net/SeatOrder` gives the clockwise index of a `Seat` (North, East, South, West in 2v2, X, Y in 1v1) and places a viewer without a seat as South or X. `MatchLayoutView` sits on the safe-area root of the match screen, applies the layout to the board, hand and four seat `RectTransform`s when the root changes size, and hides the side seats in 1v1. The `MatchLayoutTests` EditMode tests check both modes on 16:9, 19.5:9, 20:9, 21:9, 4:3 and a folding display, in landscape and portrait, with and without insets: every region stays inside the safe area, nothing overlaps and the board is the largest region.

## Hand

`Game/Match/HandChoice` decides what the local player may do with the hand from the latest `MatchSnapshot` (`docs/ui-motion.md`, Hand). It never works out a legal move: while playing it offers the tiles and ends in `legal_moves`; while opening it offers the forced tile to its seat, or any tile to both players of the opening team. A tap places a tile with one legal end at once and selects a tile with several; `TryChooseEnd` then places it. `TryDrop` places a dragged tile at the nearest of its ends within `size.touchTargetMin`. End turn (R-31), draw (R-22, R-23) and pass (R-21) are offered from `turn_placements`, the mode and the boneyard size. After any intent the hand takes no input until the next view, so a double tap never sends twice. Reading a view and tapping allocate nothing; the `HandChoiceTests` EditMode tests cover each rule.

## Score panel

`Game/Match/ScoreBoard` reads the team meters from each `MatchSnapshot`: the local team first, each filled to its score over `target_score`, the mark at `hand_points_lock_threshold`, and the lock badge for the teams in `locked_teams` (`docs/ui-motion.md`, Score panel and hand summary). `Game/Match/HandSummary` opens once per hand end, on the view that carries `last_result.left`, and explains it from server values only: the ending, the tiles left per seat, the counted pips, the hand points (`awarded` minus `carried_in`), the carry-over taken or kept, and the lock. Both allocate nothing per view; `ScoreBoardTests` cover them and `MatchSnapshotTests` cover the new fields.

## Emoji

`UI/Emoji` holds the in-match emoji of `docs/ui-motion.md` (Emoji wheel, #151). It is not attached to a match HUD yet, because the match scene does not exist.
- **`EmojiWheel`:** holds the equipped emoji, up to the slot count, and opens and closes the wheel. `TrySend` returns the id to send as `EmojiIntent` and closes the wheel. After a send the wheel stays calm for `TokenEffect.EmojiDurationMs`, and `OwnEmojiShown` restarts that time when the player's own `EmojiShown` arrives.
- **`EmojiPop`:** the pop's scale and opacity at a given time, including reduce motion.
- **`EmojiArt`:** the static sprite and the loop frames with their frame rate. It picks the static sprite on `EffectLevel.Low` or while the loop is not loaded.
- **`SeatEmojiLayer`:** prewarms `TokenEffect.EmojiPooled` `EmojiPopup` instances in `Build`, so build it when the match scene loads. It shows one emoji per seat anchor (a new one restarts the seat's popup) and returns a popup to the pool when its pop ends.

`EmojiTests` (EditMode) cover the wheel, the easing, the pop and the art choice. `SeatEmojiLayerTests` (PlayMode) cover the prewarm, replacement, expiry, the static fallback and reduce motion, and assert that showing and expiring emoji does not allocate.

## Device risk signals

`Core/DeviceRisk/DeviceRiskSignals` turns Android `Build` fields and known root, emulator and jailbreak paths into signal names (`docs/anti-cheat.md` 5.4, #185). `Game/DeviceRiskProbe.Collect` reads them on the device. The client only reports them; the server decides. The `DeviceRiskSignalsTests` EditMode tests cover retail devices, root, test keys, three emulators and a jailbreak path. `HookSignals` names instrumentation, hook framework and debugger signals from the process's mapped libraries (#206); `DeviceRiskProbe.CollectHooks` reads `/proc/self/maps` on Android, and the `HookSignalsTests` EditMode tests cover Frida, Xposed, LSPosed, Substrate and a debugger.

## Error reports

`Core/Diagnostics/ErrorReports` is the consent gate of #182 (`docs/privacy.md` 6.2, OQ-5). Its `OnLog` matches `Application.LogCallback`: while the `diagnostics` purpose is off nothing is sent, and after `Consent()` every exception, error and assert goes to an `IErrorSender` as an `ErrorReport` (type, message, stack trace), without account or analytics id. `Withdraw()` stops it at once. The crash SDK behind `IErrorSender`, its DSN and its symbol upload are owner setup: the DSN comes from build config, never from code. The `ErrorReportsTests` EditMode tests cover the gate.

## Analytics

`Core/Analytics` is the client side of `docs/analytics.md` (#181). `AnalyticsQueue` takes events only while the `analytics` consent purpose is on (`docs/privacy.md` 6.4): before `Consent(analyticsId)` and after `Withdraw()` nothing is queued, and `Withdraw` drops what was queued. `Track` accepts only the client event names of `AnalyticsEvents`, with one label and one value. `Flush` hands one batch of at most `AnalyticsBatchSize` events to an `IAnalyticsSender`, one send at a time; a failed send keeps the batch, and a full queue (`AnalyticsQueueLimit`) drops its oldest event. Both sizes come from admin config. The sender and the flush timer arrive with the network layer. The `AnalyticsQueueTests` EditMode tests cover these rules.

## Pooling and allocations

Anything that spawns more than once is pooled; a match frame allocates 0 B (`docs/performance.md` 3).

- Plain objects and collections use the engine pools in `UnityEngine.Pool`: `ObjectPool<T>`, `ListPool<T>`, `HashSetPool<T>`, `DictionaryPool<TKey, TValue>`. Warm each pool at load time so the frame loop only gets and releases.
- Scene objects use `Core/Pooling/PrefabPool<T>`: it instantiates `prewarm` inactive copies under one parent at load time, activates on `Get`, deactivates on `Release` and destroys only what exceeds `maxSize`. Size it for the worst case of the scene so a frame never instantiates.
- Runtime assemblies do not use `System.Linq`; iterate arrays and `List<T>` with `for`. `RuntimeCodeRulesTests` fails on any `System.Linq` outside `Editor`, `Generated` and `Tests`.
- No string concatenation, interpolation or `ToString()` in `Update`, input or animation code; text changes only when its value changes.
- The `PoolingTests` PlayMode tests assert with `Is.Not.AllocatingGCMemory()` that warmed pools cycle without allocating. Every new per-frame path adds the same assertion.
