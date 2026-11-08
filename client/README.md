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
| `Assets/HypeDomino/Game` | `HypeDomino.Game` | Core, Net, UI, Cosmetics |
| `Assets/HypeDomino/Generated` | `HypeDomino.Proto` | `Google.Protobuf.dll` only; output of `npm run proto:gen`, never edited |
| `Assets/HypeDomino/Tests/EditMode` | `HypeDomino.Tests.EditMode` | Editor only; references Core and UI |
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

`Scenes/Boot.unity` holds an orthographic camera and `FrameTimingSampler`. The sampler measures at the frame rate `GraphicsBoot` set (`docs/performance.md` 1.2; the highest refresh rate when none is set) and logs every 1200 frames:

```
FrameTiming target=<Hz> frames=1200 workP50=<ms>ms workP95=<ms>ms missedVsync=<percent>% gcFrames=<n> gcMax=<bytes>B gcCollections=<n> batchesMax=<n>
```

Work time is `max(cpuFrameTime, gpuFrameTime)`; a missed vsync is a presented interval longer than 1.5 refresh periods (`docs/performance.md` 2). Read it with `adb logcat -s Unity`.

`gcFrames` counts frames whose `ProfilerRecorder` "GC Allocated In Frame" was above 0, `gcMax` is the largest of them, `gcCollections` is the `GC.CollectionCount(0)` delta and `batchesMax` the largest "Batches Count" (`docs/performance.md` 3). A counter the player does not expose prints `n/a`; the Profiler counters need a development build. The frame that writes the log line allocates that line, so its GC value is not counted. `FrameWindow` holds the window and the `FrameWindowTests` EditMode tests cover it.

Size: the `Release build` workflow below; the Build Report (`Library/LastBuild.buildreport`) of a local release build gives the content breakdown.

## Shop and wallet

`UI/Shop` turns server replies into what the shop shows, without computing anything (`docs/economy.md` 3.4, 6.1). `WalletReply`, `ShopReply` and `PurchaseReply` read `GET /wallet`, `GET /shop` and `POST /shop/buy` with `JsonUtility`. `Balances` keeps coins and crystals apart and takes the new balance of a purchase's `currency` from the reply. `ChestCard` copies a `chest` offer's pool, owned items, odds (`[{"rarity","percent"}]`), opens left to the pity, price and refusal reason as sent. `ChestReveal` holds only the granted item, its rarity key and the opens left to the pity, so a reveal can show nothing but the result: no reels, wheels, slot sounds or near-miss effects (`docs/localization.md` section 6). Names are keys: `currency.coin.name`, `currency.crystal.name`, `rarity.<id>.name`.

`UI/Localization/builtin-strings.json` is the built-in bundle (`docs/localization.md` 2.2) for keys the client needs before a published bundle arrives, starting with the currency names of `docs/economy.md` OQ-13. `BuiltinStrings` looks a key up along the 1.2 fallback chain. The `ShopTests` EditMode tests cover these types.

## Graphics levels

`UI/Graphics` holds the device class rule and `GraphicsQuality`: Auto or a chosen level, the highest level the device tier allows, its effect level and render scale (`docs/performance.md` 1.1). Values come from `TokenGraphics`. `Game/GraphicsBoot` runs before the first scene: it classes the device, reads the saved choice, sets the render scale on the URP asset outside the editor and logs one `Graphics tier=... level=...` line. `Game/GraphicsProbe` is the first-match probe: the match scene adds it, and while the choice is Auto and the probe is not settled it measures work time and hands the p95 to `GraphicsBoot.SettleProbe`, which saves the drops with the app version and logs one `GraphicsProbe ...` line. `Game/ThermalWatch` starts with it on Android 10 and later: a `PowerManager` thermal status listener feeds `GraphicsQuality.UpdateHeat` every frame, and each level change logs one `Graphics level=... heat=...` line. `FrameRatePolicy` holds the frame rate (`docs/performance.md` 1.2): the rates offered for the level in use and the display, Auto or the chosen rate, the menu cap and the battery saver. `GraphicsBoot` sets `Application.targetFrameRate` to its `Target` and logs one `Graphics targetFps=...` line per change; the match scene calls `FrameRate.SetInMatch`, and `Game/PowerSaveWatch` reads the Android system saver at start and on focus. The `GraphicsQualityTests` and `FrameRatePolicyTests` EditMode tests cover the class rule on the matrix devices, the choice rules and the frame rate rules.

## Avatars

`UI/Cosmetics` holds `AvatarView` and the rules for choosing static or animated art (`docs/cosmetics.md` 8.3). It takes art through `ICosmeticArtSource`, which has no Addressables implementation until #68.

## Rarity marks

`UI/Rarity` draws the rarity diamond and halo from a `RarityTable` (`docs/rarity.md` 2.2.1). `rarity-defaults.json` there is generated by `npm run tokens:gen`; never edit it by hand.

## Design tokens

`UI/Tokens/DesignTokens.cs` is generated from `docs/design-tokens.json` by `npm run tokens:gen`; edit the JSON and regenerate, never the file. It holds `TokenColor` (`Color32`), `TokenOpacity`, `TokenSpacing`, `TokenRadius` and `TokenSize` (UI units on the 1080-unit short side), `TokenDurationMs`, `TokenEasing` (cubic-bezier points as `Vector4`), `TokenMotion` and `TokenEffect` (each effect's `DurationMs` and `Pooled` count, the largest over the levels). `UI/Motion/Easing` evaluates a `TokenEasing` curve. Rarity colours are not constants: they arrive with the admin-published rarity config (`docs/admin.md`). `npm run test:tools` fails when the file is stale, and `RuntimeCodeRulesTests` fails on `new Color(...)`, `new Color32(...)` or a named `Color` such as `Color.white` in runtime code outside it.

## Screens

`UI/Screens` holds the screen stack of `docs/ui-motion.md` 5 "Screens". A `ScreenView` is one screen: its own overlay `Canvas` scaled so the short side is 1080 units, a `CanvasGroup` content root that fades and scales, and for a modal a full-screen scrim `Image` set by `Configure`. `ScreenStack` pushes and pops them, hands input to the top screen at once (the `GraphicRaycaster` of every other screen is off), hides a covered screen by disabling its `Canvas` only, and pops on the Android back key. Transitions are `ScreenFade` values driven by unscaled time, so an interrupted one reverses from where it is. `ScreenStack.ReduceMotion` drops the scale. The `ScreenStackTests` PlayMode tests cover push, pop, reversal both ways, modals, the root screen, reduce motion and the canvas scale.

## Content

`Content` holds the client side of `docs/content-delivery.md` on Addressables 2.10.3, the version Unity 6000.3 ships. `CatalogUpdater` decides when a remote catalog is downloaded, committed, retried (2 s, 8 s, 30 s) or held back until a match ends; it is plain C# behind `ICatalogSource` and covered by the `CatalogUpdaterTests` EditMode tests. `AddressablesCatalogSource` downloads the catalog to `persistentDataPath/content`, keeps it there as `catalog_<version>.bin`, loads it with `Addressables.LoadContentCatalogAsync`, keeps only the last remote locator and stores its version and root in `PlayerPrefs` (`content.catalogVersion`, `content.baseUrl`). Every successful load runs `Addressables.CleanBundleCache`, which reads the loaded catalogs from their files and removes cached bundles none of them references (CD-24). Remote bundles are built with the load path prefix `hd-content:/`, and `ContentUrls.Resolve` rewrites it to `<content_base_url>/addressables/` through `Addressables.InternalIdTransformFunc`. `AddressablesArtSource<T>` is the `ICosmeticArtSource<T>` for `CosmeticArtCache`: one handle per address, released on the cache's last release, and a failed load is released at once. `ContentBoot` starts all of this before the first scene in a player build; in the editor Addressables reads from the AssetDatabase and no catalog is downloaded. The network layer calls `ContentBoot.Catalog.Announce(catalog_version, content_base_url)` after an accepted `ServerHello` and `SetInMatch` around a match.

`ContentCache` is the disk cache of `docs/content-delivery.md` 7: `Start` caps Unity's AssetBundle cache at 1,000,000,000 bytes and 30 days unused and runs a clear left pending from the last session, `OccupiedBytes` is the size to show, and `Clear` answers `Cleared`, `AtNextStart` (a bundle is in use) or `NotInMatch`. It is plain C# behind `ICacheStore`, covered by `ContentCacheTests`; `UnityCacheStore` is the `Caching` and `PlayerPrefs` (`content.clearAtStart`) side. The settings screen (#157) reads `ContentBoot.Cache`.

`Editor/ContentSetup` (menu HypeDomino > Apply Content Setup, also run by Apply Project Setup) writes the Addressables settings under `Assets/AddressableAssetsData`: the remote load path `hd-content:/[BuildTarget]`, a binary remote catalog that is never checked at start (the client loads catalogs itself), content built with every player build, the install-time group `cosmetic-builtin` as the default group, and the group template `Cosmetic Item` for remote item groups: one bundle per entry, a content hash in the file name, 2 retries and a 20 s timeout without progress.

## Editor methods in CI

Without a local Unity license, an editor method that creates or changes assets runs in CI: `gh workflow run unity-editor.yml --ref <branch> -f method=HypeDomino.Editor.<Class>.<Method>`. The workflow runs the method in batch mode and uploads everything it changed under `client/` as `editor-changes.patch` (artifact `editor-changes`); `gh run download <run id> -n editor-changes` and `git apply editor-changes.patch` bring it into the branch. The method runs its work through `EditorMethod.Run`, as `ContentSetup.ApplyAndExit` does: it exits the editor and writes `Logs/editor-method-succeeded` on success. The builder step always reports a failed build, because no player is built, so the run fails only when that file is missing.

## Release build

`Editor/ReleaseBuild.BuildAndroid` builds the release AAB for one catalog version (`docs/content-delivery.md` 6): it sets the Addressables player version to `-catalogVersion <version>`, builds the scenes in Build Settings into `-customBuildPath` (default `Builds/hypedomino.aab`) together with the content, and fails when `ServerData/Android/catalog_<version>.bin` is missing. It restores the player version and the AAB switch afterwards. In CI: `gh workflow run release-build.yml -f catalog_version=<version>`. The workflow measures the download size with `bundletool` on the low, mid and high device specs in `tools/install-size/specs`, fails at 150 MB or more (`docs/performance.md` 3) and keeps the AAB, the size report and the remote folder as artifacts. `ReleaseBuildTests` cover the version rule and the remote folder.

## Safe area

`UI/Layout/SafeAreaFitter` sits on the root `RectTransform` of each screen and anchors it to `Screen.safeArea`, so camera holes, notches, the island and the home indicator never cover UI. It refits when the safe area or the resolution changes (rotation, split screen, a folding display) and does nothing on other frames. `SafeAreaAnchors` turns the safe area into anchors, clamped to the screen; an empty screen or safe area falls back to the full screen. The `SafeAreaTests` EditMode tests cover 20:9 Android with a camera hole, iPhone landscape and portrait, a 4:3 tablet and 16:9 without insets. Orientation is still open (`docs/ui-motion.md` 9), so nothing here assumes one.

## Emoji

`UI/Emoji` holds the in-match emoji of `docs/ui-motion.md` (Emoji wheel, #151). It is not attached to a match HUD yet, because the match scene does not exist.
- **`EmojiWheel`:** holds the equipped emoji, up to the slot count, and opens and closes the wheel. `TrySend` returns the id to send as `EmojiIntent` and closes the wheel. After a send the wheel stays calm for `TokenEffect.EmojiDurationMs`, and `OwnEmojiShown` restarts that time when the player's own `EmojiShown` arrives.
- **`EmojiPop`:** the pop's scale and opacity at a given time, including reduce motion.
- **`EmojiArt`:** the static sprite and the loop frames with their frame rate. It picks the static sprite on `EffectLevel.Low` or while the loop is not loaded.
- **`SeatEmojiLayer`:** prewarms `TokenEffect.EmojiPooled` `EmojiPopup` instances in `Build`, so build it when the match scene loads. It shows one emoji per seat anchor (a new one restarts the seat's popup) and returns a popup to the pool when its pop ends.

`EmojiTests` (EditMode) cover the wheel, the easing, the pop and the art choice. `SeatEmojiLayerTests` (PlayMode) cover the prewarm, replacement, expiry, the static fallback and reduce motion, and assert that showing and expiring emoji does not allocate.

## Pooling and allocations

Anything that spawns more than once is pooled; a match frame allocates 0 B (`docs/performance.md` 3).

- Plain objects and collections use the engine pools in `UnityEngine.Pool`: `ObjectPool<T>`, `ListPool<T>`, `HashSetPool<T>`, `DictionaryPool<TKey, TValue>`. Warm each pool at load time so the frame loop only gets and releases.
- Scene objects use `Core/Pooling/PrefabPool<T>`: it instantiates `prewarm` inactive copies under one parent at load time, activates on `Get`, deactivates on `Release` and destroys only what exceeds `maxSize`. Size it for the worst case of the scene so a frame never instantiates.
- Runtime assemblies do not use `System.Linq`; iterate arrays and `List<T>` with `for`. `RuntimeCodeRulesTests` fails on any `System.Linq` outside `Editor`, `Generated` and `Tests`.
- No string concatenation, interpolation or `ToString()` in `Update`, input or animation code; text changes only when its value changes.
- The `PoolingTests` PlayMode tests assert with `Is.Not.AllocatingGCMemory()` that warmed pools cycle without allocating. Every new per-frame path adds the same assertion.
