# client

Unity 6 LTS game client (URP Universal Renderer with 2D sprites and real-time 3D characters, ADR 0009; C#, Addressables). It renders state and sends player intents only; the server decides every rule, score and reward.

## Editor

Unity `6000.3.27f1` (Unity 6.3 LTS), pinned in `ProjectSettings/ProjectVersion.txt`. Install it with Unity Hub together with Android Build Support (OpenJDK, Android SDK & NDK Tools). An upgrade is its own PR with a measurement (ADR 0001).

## Layout

| Path | Assembly | References |
|---|---|---|
| `Assets/HypeDomino/Core` | `HypeDomino.Core` | none |
| `Assets/HypeDomino/Net` | `HypeDomino.Net` | Core, Proto |
| `Assets/HypeDomino/UI` | `HypeDomino.UI` | Core |
| `Assets/HypeDomino/Cosmetics` | `HypeDomino.Cosmetics` | Core |
| `Assets/HypeDomino/Game` | `HypeDomino.Game` | Core, Net, UI, Cosmetics |
| `Assets/HypeDomino/Generated` | `HypeDomino.Proto` | `Google.Protobuf.dll` only; output of `npm run proto:gen`, never edited |
| `Assets/HypeDomino/Tests/EditMode` | `HypeDomino.Tests.EditMode` | Editor only |
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
FrameTiming target=<Hz> frames=1200 workP50=<ms>ms workP95=<ms>ms missedVsync=<percent>%
```

Work time is `max(cpuFrameTime, gpuFrameTime)`; a missed vsync is a presented interval longer than 1.5 refresh periods (`docs/performance.md` 2). Read it with `adb logcat -s Unity`.

Size: build a release Android player (File > Build Profiles > Android, Development Build off) and read the APK size and the Build Report.
