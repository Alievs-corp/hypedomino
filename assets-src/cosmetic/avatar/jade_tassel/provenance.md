# avatar.jade_tassel

Animated avatar of set 1 (#171), batch A, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-15 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: `concept/20261115-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261115-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a girl of about eight with light golden skin, black hair in two round buns tied with jade green silk ribbons that end in small gold tassels hanging beside her cheeks, and a quilted cream silk jacket with jade toggle buttons"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt. Not asked for but present: jade drop earrings, crystal drops under the floral hair ornaments, fleece trim on the collar and gold scroll ornaments with crystal and jade drops on the shoulders.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.11, -0.32 px, scale 1.0000 and rotation 0.0001 rad (mean squared grey difference outside the eyes 0.00165 before, 0.00155 after), and the eye patch was offset by [2.6, 4.7, 4.9] in RGB to meet the canonical skin. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly; the scroll ornaments on the collar and shoulders are cloud-like scrolls, judged decorative by the agent, left to the cultural review
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `epic`, source `shop:offer.avatar_jade_tassel`. The offer id is a placeholder until the shop record exists (V-15). Animated avatars are proposed at `epic` and above, as the batch leans to the higher tiers of `docs/rarity.md`; this batch proposes 5 epic, 4 legendary, 1 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signatures were not validated with a C2PA tool.

| Field | `concept/20261115-01.png` | `concept/20261115-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-15T16:29:52Z | `c2pa.created`, 2026-11-15T16:32:13Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-15 16:29:07 | 2026-11-15 16:31:36 |
| Codex session | `01a1e00a-834f-7583-b533-bacacdf3e4ce` | `01a1e00c-c7b7-79b3-b05a-290ec17c6f69` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked with `cmp`). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/jade_tassel/work/jade_tassel_cleanup.py jade_tassel assets-src/cosmetic/avatar/jade_tassel/concept/20261115-01.png assets-src/cosmetic/avatar/jade_tassel/work
```

It printed background [206.0, 206.0, 205.0], pockets 11535 px, crop (68, 0, 1146, 1078), portrait 768. The script is `amethyst_courtier_cleanup.py` of batch A of the static avatars with three changes, all for the same reason: on these concepts grey background stayed opaque between loose strands and the neck, and hair edges kept a light outline. (1) It writes the master and a square 768 x 768 portrait crop (`work/jade_tassel_portrait.png`, the same crop rule, no circle) instead of the 512 static, because the static of an animated avatar is a loop key frame. (2) Enclosed background pockets are taken within 24 of the grey (10 before) and up to 140 px inside the outline (80 before), and a pocket counts only if its median colour is within 12 of the grey, so pearls, crystal and silver that come close to the grey stay. (3) In a 6 px band along the background the alpha is capped by projecting each pixel between the grey and the mean colour of the nearby solid subject (Gaussian sigma 6); where that colour is itself within 40 of the grey, by the colour distance to the grey (0 at 14, 1 at 60). The colour of edge pixels is unmixed from the grey and blended toward the nearby subject colour by one minus the alpha, so faint edge pixels take the subject's colour instead of a light speck.

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Background, trim, class canvas and pivot | Yes, canvas open | As in the static batch A with the changes above. The loop script puts a circular alpha (soft edge of 2.25 px at 768, 0.56 px at 192) on every frame, because the portrait is shown in a circle (`docs/art-direction.md` 5.1); the class canvas is open (O-6) and D-1 sets 192 x 192 |
| 2. Paint-over, AI artefacts | No | No paint-over |
| 3. Moving parts on their own layers | Yes | Written by the loop script into `work/` (table below) |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow, rarity colour or team hue seen in the concept, the master or the frames |
| 5. Silhouette and smallest-size tests | In part | The agent looked at every frame at 192 px and at 2x on a dark backdrop and at frame 1 on a light one; face, expression and the moving part read. The smallest display size and frame band of an avatar are open (#168, O-6), so the test is not final |

## Animation (W-4) and export (W-5)

Run on 2026-11-15 by Claude Code with the same versions, after the clean-up:

```bash
python -I assets-src/cosmetic/avatar/jade_tassel/work/jade_tassel_loop.py assets-src/cosmetic/avatar/jade_tassel/concept assets-src/cosmetic/avatar/jade_tassel/work assets-src/cosmetic/avatar/jade_tassel/export
```

It printed `jade_tassel_loop: blink edit registered (tx, ty, scale, angle) = (0.1130, -0.3151, 1.0000, 0.0001), residual 0.00165 -> 0.00155, eye colour offset [2.6, 4.7, 4.9], parts 2, filled 22229 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run into empty folders wrote files byte-identical to each other and to the committed ones.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes. Half-closed frames made by cross-fading showed two sets of lids and were dropped | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `tassel_left` (layer `work/jade_tassel_tassel_left.png`, pivot (145, 462)) turns about its pivot by up to 7.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.6, the turn growing from 0 at the pivot to full at 120 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot; `tassel_right` (layer `work/jade_tassel_tassel_right.png`, pivot (669, 466)) turns about its pivot by up to 7.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.9, the turn growing from 0 at the pivot to full at 120 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot | The part hangs over the background, so the hole it leaves in the base layer is transparent |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 5.31 of 255, against 4.52 on average and 5.69 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Re-crop, 2026-11-16

Owner feedback from the phone on 2026-11-16: inside the frames, the frame's top ornament band covered the top of the head and hair. The owner asked for the portrait "very slightly smaller and a bit lower"; `docs/art-direction.md` 5.1 keeps anything essential out of the outer band the frame covers. Only the crop framing changed. The cut-out master was re-made by the same run and is byte-identical; masks, pockets and the edge re-matte are unchanged.

- `CROP_SHARE` 0.86 to 0.935 (0.86 / 0.92: the subject is 8% smaller in the circle)
- `HEADROOM` 0.07 to 0.12, not the 0.17 of the other avatars: at 0.17 the right tassel layer was 21.5% outside the circle; at 0.12 it is 0.0%
- The crop may now reach above the top of the concept: `upper = min(top - HEADROOM * side, height - side)` instead of clipping it at 0, the rule `avatar.tawny_quill` already used. The subject starts within a few pixels of the concept's top edge, so a lower head needs transparent canvas above it; the bottom and side clamps are as before
- Printed by the clean-up: `jade_tassel: background [206.0, 206.0, 205.0], pockets 11535 px, crop (21, -134, 1193, 1038), portrait 768`
- Crop `(68, 0, 1146, 1078)` to `(21, -134, 1193, 1038)` on the concept
- Scale k = 0.9198. The rig geometry in portrait pixels (`HEAD`, `NECK_Y`, `NECK_BAND`, `SHOULDER_Y`, `BREATH`, `NOD`, `EYES`, `EYE_FEATHER`, the part polygons, pivots, `length` and `grow`, the glints and the ripples) is mapped through the crop change (x' = k x + offset, sizes and pixel amplitudes times k), so every mask stays on the same anatomy and the motion is the same relative to the subject. Angles, timing, blink frames and `STATIC_FRAME` are unchanged
- `register` now clips a crop edge above the concept to 0 when it marks the registration region, as the later loop scripts of this batch do; with the old slice a negative top would have marked the wrong rows
- Printed by the loop: `jade_tassel_loop: blink edit registered (tx, ty, scale, angle) = (0.1099, -0.3107, 1.0000, 0.0001), residual 0.00149 -> 0.00140, eye colour offset [2.5, 4.7, 4.8], parts 2, filled 19116 px`
- Head top (first row of `export/static.png` with alpha above 0.5, as a share of the canvas height): 0.036 before, 0.120 after; highest over the 16 loop frames 0.031 before, 0.115 after

Run on 2026-11-16 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/jade_tassel/work/jade_tassel_cleanup.py jade_tassel assets-src/cosmetic/avatar/jade_tassel/concept/20261115-01.png assets-src/cosmetic/avatar/jade_tassel/work
python -I assets-src/cosmetic/avatar/jade_tassel/work/jade_tassel_loop.py assets-src/cosmetic/avatar/jade_tassel/concept assets-src/cosmetic/avatar/jade_tassel/work assets-src/cosmetic/avatar/jade_tassel/export
```

Determinism: two more runs of the same commands into two empty folders wrote the same 26 files byte for byte, and both match the files committed here.

`npm run asset:check -- assets-src/cosmetic/avatar/jade_tassel` after the re-crop: 17 sprites, ASTC 6x6 estimate 278,528 bytes, limit under 300,000 (#171, animated avatars under 300 KB each): pass.

The Files table below keeps the hashes of the first export. The files the re-crop changed now hash as follows:

| File | sha256 after the re-crop |
|---|---|
| export/anim/0001.png | `1f30a655b3211523f24993b710faffd4e76fb3fe772c8ac9eb69b7c4d2cfa0e1` |
| export/anim/0002.png | `9cef680689fbedf85c9a10f695cfbf70aa6135b0349120f214f481c5c58a0f82` |
| export/anim/0003.png | `2194a2a92a81c37d782264db23feeacd3fce4e9508db6b39e0932c4bbbb08b05` |
| export/anim/0004.png | `8759b569e66716ebbab9186cd124b7b55c33c6675cc5553ba9f973e54b62e44d` |
| export/anim/0005.png | `cc9cf3871912cc813349ee6c01d67037f9d55a4da494444e30107fc033920f54` |
| export/anim/0006.png | `62d3d12b3296c6734a922fe8a8e174d0f287287c26e9f3d58809d05bb8bff6a3` |
| export/anim/0007.png | `a8c7867a7480de090934c183191405f1fe3eee23dc143ada92689340c961261e` |
| export/anim/0008.png | `e13915c0bce481e68c8b839c441c9615794e039e09391b758fe83d681ff6ea3b` |
| export/anim/0009.png | `e488aa952387aeb62c04dd1a67291a83e829275548039282a3a6998770df5c20` |
| export/anim/0010.png | `0617ab8880a3322c41b11c0ff633a1fe47a3c151c373f7173aacf1a0bfcd5903` |
| export/anim/0011.png | `1b07ab8774b80afb712d685ad0ceecb93370faefd2723fe4af62abaa5194bb8d` |
| export/anim/0012.png | `8d95e6b73a211624f0a221ce2a4dc19e7cebd027fc17d98c6280b2263d509739` |
| export/anim/0013.png | `fd297a1853d708accba62f0980b6806187888a37c58fb7e036f8cb5629688773` |
| export/anim/0014.png | `4767fd0e625b10e9ba3499afa463b88fd9c9ba2bd6ddcc45deeefdc3321e96c6` |
| export/anim/0015.png | `5213c8baba0f5c79ec8ba2ed6929e9f9efb6d67439e8f2b61d77be8d159d4b86` |
| export/anim/0016.png | `a469d6bb20b9c88193bb8a8273b5225d6ca88e4c645330621a39a13c9080a149` |
| export/static.png | `e488aa952387aeb62c04dd1a67291a83e829275548039282a3a6998770df5c20` |
| work/jade_tassel_base.png | `65c88eacd5906706a93df296e076a166b43bec98b060f6a6168f6169d6108570` |
| work/jade_tassel_cleanup.py | `18a09c5cb88e9deea6bacde53bf5efc038ab7d778bc75123651cf1298b91226b` |
| work/jade_tassel_eyes_closed.png | `40c828ff261779c6a7906bdb192f5dfb555eceb94eda5c66a76792ec457d87a4` |
| work/jade_tassel_eyes_open.png | `4384a7521149e7de217694d2b3180506d2f3ee9489c6c7acf1ebccb0f79a602b` |
| work/jade_tassel_head.png | `6daa149f8e59cfd1a25523d51c765fdec46da70e468d2a0d97a28c37a17e3580` |
| work/jade_tassel_loop.py | `d907d383fdd8116c1c4b2bc46fb4d6e19ac29d78470b6c022669f7f94228571d` |
| work/jade_tassel_portrait.png | `76f73fcc96f10c5e12a1353b70872beb25b5c2f2948eea7aac20c3a9ccba0916` |
| work/jade_tassel_tassel_left.png | `dc9565475c251fd100a154be5219e8a1fce7ca49b5b3c985a2f877a8aef55e81` |
| work/jade_tassel_tassel_right.png | `e473728a637a15c183a4873d58472af80f5a4e726d618c5fcaee53137149e419` |
| work/jade_tassel_torso.png | `a07d2b733462ed0386255ce140e8ddd48b84dbc9ff53fff8c05d16ef8a2a44bd` |

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `b9e8a81c1128de5dd13807fb2e60bce4027947088d40ed4cb5b705b1d03933f0` | 1254 x 1254 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e00a-834f-7583-b533-bacacdf3e4ce` |
| concept/20261115-02.png | `3c5d5f94e72ffc59bcb71e7da94b58c8c17a96e06fe0ce1c6dea77a89b09fcfb` | 1254 x 1254 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e00c-c7b7-79b3-b05a-290ec17c6f69` |
| work/jade_tassel_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-15, from `amethyst_courtier_cleanup.py` |
| work/jade_tassel_loop.py | `70224de1e285b50d80319dbdfb7ee6b3d03c5214480008a6d17c2fab02890e2d` | W-4 and W-5 script | Written by Claude Code, 2026-11-15; the same script for every avatar of this batch with the item's geometry |
| work/jade_tassel_master.png | `3b6c534a07071ba8177b4eeb4822543e957ffcc2492e77db0c75cf2843e50f5d` | 1254 x 1254 RGBA, the cut-out master | `jade_tassel_cleanup.py` |
| work/jade_tassel_base.png | `63c87c48de7a6ab3f25cedbee8a5cd0ddeb82fef0ef3fa623223278cba8e1a08` | The portrait with the moving parts taken out (filled or cleared) | `jade_tassel_loop.py` |
| work/jade_tassel_earring_left.png | `3d181db3052dfb361eef86b0e2a90b813ca12818a1ba3fec91cd8ef7baadaf2c` | Layer | `jade_tassel_loop.py` |
| work/jade_tassel_earring_right.png | `1b2311450e6c20fd5fb9d567135ebdffef6e12dbe29b304c2312a392db91350c` | Layer | `jade_tassel_loop.py` |
| work/jade_tassel_eyes_closed.png | `f67849fffe592897453f0b099fdc980a18f8437eb75ae4d613187ce529520099` | Eye region, closed (aligned edit) through the same mask | `jade_tassel_loop.py` |
| work/jade_tassel_eyes_open.png | `181eb9b5449385d59a29697b2248b85cd5f956cc376354e33ef936c86723334b` | Eye region, open (canonical concept) through the feathered eye mask | `jade_tassel_loop.py` |
| work/jade_tassel_head.png | `aa45671d9add70f5b8ab22d5d232be0b39c7f33ed8f906582324d046e6e77258` | Head and hair: the portrait weighted by the head field of the warp | `jade_tassel_loop.py` |
| work/jade_tassel_portrait.png | `971233af71e0611c8f075ff0ac4e3657b3fd69e08b325b81be0e71f8835296db` | 768 x 768 RGBA portrait crop, no circle | `jade_tassel_cleanup.py` |
| work/jade_tassel_tassel_left.png | `3d50b816368912f22d989d1a7792690b588c107937313b2472ff3e0ee614c425` | Moving part `tassel_left`: the portrait through its feathered polygon | `jade_tassel_loop.py` |
| work/jade_tassel_tassel_right.png | `643fe05cf1bcd04b3835e376ac04f0cd96907867dd5aa50403de922bec79bac4` | Moving part `tassel_right`: the portrait through its feathered polygon | `jade_tassel_loop.py` |
| work/jade_tassel_torso.png | `2565c9ca2870ece57f7707f2ffda4d6021b2e80298e390ef9e9136f6d7ed2784` | Torso and costume: the portrait weighted by one minus the head field | `jade_tassel_loop.py` |
| export/static.png | `b5ffbd77c5982ee29b7ad06b483dc080c8b4245823c70d61316e0fffd5f816d8` | 192 x 192 RGBA static, loop frame 9 | `jade_tassel_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `jade_tassel_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 16:29:07 UTC; `c2pa.created` 2026-11-15T16:29:52Z |
| 2026-11-15 | concept/20261115-02.png | Edit of `concept/20261115-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 16:31:36 UTC; `c2pa.created` 2026-11-15T16:32:13Z |
| 2026-11-15 | work/, export/ | The scripts above |
| 2026-11-16 | export/static.png, work/jade_tassel_base.png, work/jade_tassel_cleanup.py, work/jade_tassel_eyes_closed.png, work/jade_tassel_eyes_open.png, work/jade_tassel_head.png, work/jade_tassel_loop.py, work/jade_tassel_portrait.png, work/jade_tassel_tassel_left.png, work/jade_tassel_tassel_right.png, work/jade_tassel_torso.png, export/anim/0001.png to 0016.png | Re-crop: smaller and lower per owner feedback. `CROP_SHARE` 0.86 to 0.935, `HEADROOM` 0.07 to 0.12, crop `(68, 0, 1146, 1078)` to `(21, -134, 1193, 1038)`; see "Re-crop, 2026-11-16" |

Prompt of `concept/20261115-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a girl of about eight with light golden skin, black hair in two round buns tied with jade green silk ribbons that end in small gold tassels hanging beside her cheeks, and a quilted cream silk jacket with jade toggle buttons, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261115-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a girl of about eight with light golden skin, black hair in two round buns tied with jade green silk ribbons that end in small gold tassels hanging beside her cheeks, and a quilted cream silk jacket with jade toggle buttons, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261115-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261115-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-15.
