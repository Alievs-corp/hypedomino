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
| Quality levels | No per-level pipeline; every level uses the default URP asset | One renderer until #159 adds graphics levels |
| Profiling | Frame Timing Stats on | `FrameTimingManager` measurements, `docs/performance.md` 2 |
| Color space | Linear | Lit 3D characters, ADR 0009 |

## Textures and atlases

- Android texture compression is ASTC only; the build subtarget is ASTC (ADR 0002).
- `Settings/SpriteTextureASTC.preset` is the default preset for every new texture: Sprite, no mipmaps, ASTC 6x6 on Android and iOS. ASTC 6x6 is the base-colour block size of ADR 0009. Per-class block and maximum sizes are open item O-5 in `docs/asset-workflow.md` (#93, #135); a class that needs other values adds a preset with a folder filter above this one in the Preset Manager.
- Sprite Atlas V2 is the packer. One atlas per cosmetic item inside its Addressables group and effect sprites only in the shared effects atlas (`docs/cosmetics.md` 8.2, `docs/art-direction.md` 6).

## Measuring the empty build

`Scenes/Boot.unity` holds an orthographic camera and `FrameTimingSampler`. The sampler sets `Application.targetFrameRate` to the highest refresh rate the display reports and logs every 1200 frames:

```
FrameTiming target=<Hz> frames=1200 workP50=<ms>ms workP95=<ms>ms missedVsync=<percent>% gcFrames=<n> gcMax=<bytes>B gcCollections=<n> batchesMax=<n>
```

Work time is `max(cpuFrameTime, gpuFrameTime)`; a missed vsync is a presented interval longer than 1.5 refresh periods (`docs/performance.md` 2). Read it with `adb logcat -s Unity`.

`gcFrames` counts frames whose `ProfilerRecorder` "GC Allocated In Frame" was above 0, `gcMax` is the largest of them, `gcCollections` is the `GC.CollectionCount(0)` delta and `batchesMax` the largest "Batches Count" (`docs/performance.md` 3). A counter the player does not expose prints `n/a`; the Profiler counters need a development build. The frame that writes the log line allocates that line, so its GC value is not counted. `FrameWindow` holds the window and the `FrameWindowTests` EditMode tests cover it.

Size: build a release Android player (File > Build Profiles > Android, Development Build off) and read the APK size and the Build Report.

## Shop and wallet

`UI/Shop` turns server replies into what the shop shows, without computing anything (`docs/economy.md` 3.4, 6.1). `WalletReply`, `ShopReply` and `PurchaseReply` read `GET /wallet`, `GET /shop` and `POST /shop/buy` with `JsonUtility`. `Balances` keeps coins and crystals apart and takes the new balance of a purchase's `currency` from the reply. `ChestCard` copies a `chest` offer's pool, owned items, odds (`[{"rarity","percent"}]`), opens left to the pity, price and refusal reason as sent. `ChestReveal` holds only the granted item, its rarity key and the opens left to the pity, so a reveal can show nothing but the result: no reels, wheels, slot sounds or near-miss effects (`docs/localization.md` section 6). Names are keys: `currency.coin.name`, `currency.crystal.name`, `rarity.<id>.name`.

`UI/Localization/builtin-strings.json` is the built-in bundle (`docs/localization.md` 2.2) for keys the client needs before a published bundle arrives, starting with the currency names of `docs/economy.md` OQ-13. `BuiltinStrings` looks a key up along the 1.2 fallback chain. The `ShopTests` EditMode tests cover these types.

## Avatars

`UI/Cosmetics` holds `AvatarView` and the rules for choosing static or animated art (`docs/cosmetics.md` 8.3). It takes art through `ICosmeticArtSource`, which has no Addressables implementation until #68.

## Rarity marks

`UI/Rarity` draws the rarity diamond and halo from a `RarityTable` (`docs/rarity.md` 2.2.1). `rarity-defaults.json` there is generated by `npm run tokens:gen`; never edit it by hand.

## Design tokens

`UI/Tokens/DesignTokens.cs` is generated from `docs/design-tokens.json` by `npm run tokens:gen`; edit the JSON and regenerate, never the file. It holds `TokenColor` (`Color32`), `TokenOpacity`, `TokenSpacing`, `TokenRadius` and `TokenSize` (UI units on the 1080-unit short side), `TokenDurationMs`, `TokenEasing` (cubic-bezier points as `Vector4`), `TokenMotion` and `TokenEffect` (each effect's `DurationMs` and `Pooled` count, the largest over the levels). `UI/Motion/Easing` evaluates a `TokenEasing` curve. Rarity colours are not constants: they arrive with the admin-published rarity config (`docs/admin.md`). `npm run test:tools` fails when the file is stale, and `RuntimeCodeRulesTests` fails on `new Color(...)`, `new Color32(...)` or a named `Color` such as `Color.white` in runtime code outside it.

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
