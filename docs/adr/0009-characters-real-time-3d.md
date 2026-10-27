# 0009. Characters: real-time 3D in the Universal Renderer

Status: Accepted
Date: 2026-10-27
Issue: #166

Supersedes the renderer choice of ADR 0001 (URP 2D renderer). Everything else in ADR 0001 stands.

## Context
The owner chose a premium, semi-realistic look for characters (2026-10-26) and wants it to scale with the graphics setting: the highest quality at the top setting, lower resolution when the player lowers it (2026-10-27). A first production test built two characters through image-to-3D, rigging and rendered sprite sheets. It showed the limits of sprites for this look:
- Every quality level needs its own set of frames. At the size where a semi-realistic face reads, four clips of one character at the high level are tens of megabytes of frames, far over the 1.5 MB per character of #170 and a heavy load on the 150 MB install and on download.
- Frame size, frame rate and camera angle are frozen at render time. A higher setting cannot add detail that was not rendered, and the seat, lobby and profile views each need their own render.
- Skins that change parts of a character multiply the frame sets again.

A skinned 3D mesh scales by level of detail and texture size, shares one skeleton and one set of clips across characters and skins, and is drawn from any angle the screen needs. Budgets from CLAUDE.md still hold: 60 fps on Mid, under 150 draw calls in the match scene, 0 B GC per frame, cold start under 3 s, install under 150 MB.

## Decision
- The `character` cosmetic type (`docs/cosmetics.md`) is a real-time skinned 3D mesh. Avatars, frames, tables, emoji, tiles, UI and effects stay 2D sprites.
- The client uses URP with the Universal Renderer (Forward path) instead of the 2D renderer. Sprites keep the sprite shaders, atlases and batching; the 2D renderer's 2D lights are not used. Characters are drawn by the same camera as the scene; no camera stacking.
- Every character shares one character shader made in Shader Graph: base colour, an optional normal map on the high level, one main directional light, ambient from spherical harmonics and a rim term. No real-time shadows, no per-character material instance beyond the textures, no post-processing (`docs/ui-motion.md` section 4). The ground contact shadow stays the pooled shadow sprite of `docs/art-direction.md` section 4.
- One standard humanoid skeleton for all characters. Base clips are authored once and retargeted through Unity Humanoid; a skin may bring its own clips on the same skeleton.
- Detail follows the effect level, which follows the graphics setting (`docs/ui-motion.md` section 7):

| Effect level | Character shown | Mesh | Textures |
|---|---|---|---|
| low | Static sprite (`staticAddress`), as for every animated cosmetic | None | Static sprite only |
| medium | `modelAddress` | Seat LOD, up to 10,000 triangles | Base colour 1024 px |
| high | `modelHqAddress` where the item has one, otherwise `modelAddress` | Seat LOD in the match scene; hero LOD (up to 30,000 triangles) in lobby, profile and shop | Base colour 2048 px and normal map 2048 px |

- Initial asset limits, confirmed or changed by the measurement below in a PR to this ADR: at most 65 bones and 4 bone influences per vertex (2 on medium), one mesh and one material per LOD, ASTC 6x6 for base colour and 5x5 for normal maps (ADR 0002 sets block size per asset class). The `model` bundle of one character with its clips stays under the 1.5 MB of #170; `modelHq` is its own bundle in the item's group (entries are packed separately, `docs/cosmetics.md` 8.2), fetched only on the high level, and stays under 6 MB.
- Every character keeps a static sprite (ADR 0002, `docs/cosmetics.md` V-2). It is rendered from the model in a chosen key pose, so static and 3D show the same character.
- Production goes through `docs/asset-workflow.md` section 5.5: concept turnaround, image-to-3D, clean-up and retopology in Blender, texture bake, rig, clips, Unity import.

## Alternatives considered and why rejected
- Pre-rendered sprite sheets from the 3D model (the first plan): quality is fixed at render time and size grows with every level, clip and skin; see Context.
- Skeletal 2D rigs (Unity 2D Animation or Spine) from painted parts: light and cheap, but a semi-realistic face and three-quarter turns cannot be shown from flat parts without redrawing them per angle, and the look the owner chose is lost.
- Keep the 2D renderer and draw characters with an overlay camera: camera stacking adds a full pass and a second culling and sorting step on mobile; the Universal Renderer draws sprites and meshes in one pass.
- A physically based lit shader per character (URP Lit): more texture samples and light work than the look needs; one shared shader keeps SetPass calls and memory flat.
- Real-time shadows: a shadow map pass for up to four seats costs more than the pooled shadow sprite and adds nothing at seat size.

## Consequences
- Easier: one model serves seat, lobby, profile and shop at any size; quality scales with the graphics setting without new art; base clips are shared; a skin is a texture or part change on the same skeleton.
- Harder: the art pipeline needs 3D skills (retopology, UV, bake, rigging). A skinned mesh costs CPU or GPU skinning per frame; four seats plus effects must still fit the work-time and draw-call gates of `docs/performance.md`. The client must handle a missing or still-loading model with the static sprite.
- Must be verified in #93 before the first character ships: the replay of `docs/performance.md` section 3 with four animated characters, on the Mid reference devices at the medium level and on the High reference devices at the high level, reporting work time, batches, SetPass calls, GC and memory, plus the download size of one `model` and one `modelHq` bundle. A miss changes the limits above by PR.
