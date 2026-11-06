# 0012. Animated 2D cosmetics: sprite sheet loops

Status: Proposed. The format follows from constraints checked below; its device measurement is owed before this ADR is accepted.
Date: 2026-11-06
Issue: #93

Characters are real-time 3D and stay under ADR 0009. This ADR covers the other animated cosmetics: avatars, frames, emoji and table layers.

## Context

Animated avatars appear in lists, the shop and the profile as well as at the seats. They have to work inside a Canvas, including under scrolling and masking. Each avatar has a size limit of 300 KB (#171, a loop of 1 to 2 s), and the install limit in ADR 0002 keeps the art out of the install package. An item's art goes through the asset pipeline of #135 and `docs/asset-workflow.md` 5.2, where After Effects and Blender already deliver PNG sequences. `LoopArt` already plays emoji from an array of sprites.

## Decision

- An animated 2D item is a loop of whole frames in the item's one atlas (`docs/cosmetics.md` 8.2), compressed with ASTC, without mipmaps, with its static sprite in the same group.
- Frames keep the class canvas and pivot untrimmed, so swapping a frame or the static shows no jump (`docs/art-direction.md` section 7).
- `AnimatedImage` plays the loop on a UI `Image`. It sets the sprite only when the frame changes and allocates nothing while playing.
- It shows the static on the low effect level, with reduce motion, for an item with no loop, and when the screen's budget is spent (`docs/cosmetics.md` 8.3). A row that is disabled frees its slot and rests on the static.
- Each screen holds one `AnimationBudget`. Its initial limit is `maxAnimatedCosmetics` in `docs/design-tokens.json`: 0, 10 and 10 for low, medium and high. The 10 comes from the acceptance criterion of #93, not from a measurement, and the measurement below confirms or changes it by PR.
- A screen puts its animated images under their own nested Canvas, so a frame change does not rebuild the rest of the screen.

### Size rule

ASTC stores 16 bytes per block, so the texture size follows from the atlas and the block size. These are before bundle compression and include no bundle overhead:

| Atlas | Block | Bytes | Fits the 300 KB of #171 |
|---|---|---|---|
| 1024 x 1024 | 6x6 | 467,856 | No |
| 768 x 768 | 6x6 | 262,144 | Yes |
| 1024 x 1024 | 8x8 | 262,144 | Yes |

As an illustration, 768 x 768 holds 16 frames of 192 x 192. The canvas per class, the block size per class and the frame rate stay open (`docs/asset-workflow.md` O-5, O-6). 8x8 may not hold the outline of `docs/art-direction.md` section 4; that needs a look at real art on the reference panel.

## Alternatives considered and why rejected

None of these was measured on a device. Each is rejected on a constraint.

- Unity 2D Animation rig: `SpriteSkin` carries `[RequireComponent(typeof(SpriteRenderer))]` in version 10.1 of the package, so it does not drive a UI `Image`. Avatars in a Canvas could not use it without a separate world-space path for lists and masks. It would also add a package the project does not have.
- Spine: it has a UI component, but needs a third-party runtime and a commercial licence, which is a purchase for the owner. Price and licence terms were not checked. It would be reconsidered if sprite sheets fail the size rule on real art.
- Lottie: no first-party Unity player, so it adds a third-party runtime that rasterises vectors at run time. Painted art from the pipeline in `docs/asset-workflow.md` is not vector art.
- Video: already rejected for in-match effects in ADR 0002.

## Measurement owed

Before this ADR is accepted, on a Mid reference device of `docs/performance.md` section 1 under its section 2 conditions:

- 10 animated avatars on one screen hold 60 fps (acceptance criterion of #93), reporting work time, missed vsyncs, batches, SetPass calls, GC and memory.
- The bundle of one real avatar, against the 300 KB of #171.
- With and without the nested Canvas, to confirm that rule.

A device outside the matrix, such as the High-class device of ADR 0009, shows headroom but does not close the criterion.

## Consequences

- Easier: one player for avatars, frames, emoji and table layers; no new package; the admin can measure frames and atlas size from the file (#135); art needs no Unity rig.
- Harder: every frame is texture memory, so bytes grow with length and size, and a smooth long loop does not fit the limit. Reuse of parts between frames is not possible as it would be with a rig.
- `docs/asset-workflow.md` 5.2 drops the Unity 2D rig path for these types.
