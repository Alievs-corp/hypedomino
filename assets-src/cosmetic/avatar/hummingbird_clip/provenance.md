# avatar.hummingbird_clip

Animated avatar of set 1 (#171), batch A, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-15 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: `concept/20261115-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261115-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a girl of about sixteen with deep brown skin, a voluminous round afro held back by a wide ivory silk headband with a small enamel hummingbird clip, small crystal drop earrings, and a high-collared sage green silk blouse with pearl buttons"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt. Not asked for: freckles and gold studs above the crystal drops. The enamel hummingbird has a red throat and green and blue wings; the blue is a small accent and does not dominate.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of -0.12, -0.19 px, scale 1.0005 and rotation 0.0003 rad (mean squared grey difference outside the eyes 0.00107 before, 0.00102 after), and the eye patch was offset by [-4.7, -1.9, 1.1] in RGB to meet the canonical skin. The edit also lowered the chin slightly; outside the eye mask and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `legendary`, source `shop:offer.avatar_hummingbird_clip`. The offer id is a placeholder until the shop record exists (V-15). Animated avatars are proposed at `epic` and above, as the batch leans to the higher tiers of `docs/rarity.md`; this batch proposes 5 epic, 4 legendary, 1 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signatures were not validated with a C2PA tool.

| Field | `concept/20261115-01.png` | `concept/20261115-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-15T16:59:29Z | `c2pa.created`, 2026-11-15T17:01:15Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-15 16:58:21 | 2026-11-15 17:00:02 |
| Codex session | `01a1e025-472a-7023-8a4b-ee9666b5b76b` | `01a1e026-d07f-7f33-ab9a-db5c53ae6d3a` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked with `cmp`). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/hummingbird_clip/work/hummingbird_clip_cleanup.py hummingbird_clip assets-src/cosmetic/avatar/hummingbird_clip/concept/20261115-01.png assets-src/cosmetic/avatar/hummingbird_clip/work
```

It printed background [189.0, 189.0, 189.0], pockets 1315 px, crop (137, 0, 1168, 1031), portrait 768. The script is `amethyst_courtier_cleanup.py` of batch A of the static avatars with three changes, all for the same reason: on these concepts grey background stayed opaque between loose strands and the neck, and hair edges kept a light outline. (1) It writes the master and a square 768 x 768 portrait crop (`work/hummingbird_clip_portrait.png`, the same crop rule, no circle) instead of the 512 static, because the static of an animated avatar is a loop key frame. (2) Enclosed background pockets are taken within 24 of the grey (10 before) and up to 140 px inside the outline (80 before), and a pocket counts only if its median colour is within 12 of the grey, so pearls, crystal and silver that come close to the grey stay. (3) In a 6 px band along the background the alpha is capped by projecting each pixel between the grey and the mean colour of the nearby solid subject (Gaussian sigma 6); where that colour is itself within 40 of the grey, by the colour distance to the grey (0 at 14, 1 at 60). The colour of edge pixels is unmixed from the grey and blended toward the nearby subject colour by one minus the alpha, so faint edge pixels take the subject's colour instead of a light speck.

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
python -I assets-src/cosmetic/avatar/hummingbird_clip/work/hummingbird_clip_loop.py assets-src/cosmetic/avatar/hummingbird_clip/concept assets-src/cosmetic/avatar/hummingbird_clip/work assets-src/cosmetic/avatar/hummingbird_clip/export
```

It printed `hummingbird_clip_loop: blink edit registered (tx, ty, scale, angle) = (-0.1207, -0.1909, 1.0005, 0.0003), residual 0.00107 -> 0.00102, eye colour offset [-4.7, -1.9, 1.1], parts 2, filled 4629 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run into empty folders wrote files byte-identical to each other and to the committed ones.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes. Half-closed frames made by cross-fading showed two sets of lids and were dropped | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.4 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `earring_left` (layer `work/hummingbird_clip_earring_left.png`, pivot (251, 474)) turns about its pivot by up to 6.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.4, the turn growing from 0 at the pivot to full at 50 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot; `earring_right` (layer `work/hummingbird_clip_earring_right.png`, pivot (523, 431)) turns about its pivot by up to 6.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.8, the turn growing from 0 at the pivot to full at 50 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot | The part lies over hair or skin, so the hole it leaves in the base layer is filled by normalized convolution from its surroundings (Gaussian sigma 48 down to 3); only a sliver of it shows while the part is off its rest position |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 3.68 of 255, against 3.1 on average and 4.0 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Re-crop, 2026-11-16

Owner feedback from the phone on 2026-11-16: inside the frames, the frame's top ornament band covered the top of the head and hair. The owner asked for the portrait "very slightly smaller and a bit lower"; `docs/art-direction.md` 5.1 keeps anything essential out of the outer band the frame covers. Only the crop framing changed. The cut-out master was re-made by the same run and is byte-identical; masks, pockets and the edge re-matte are unchanged.

- `CROP_SHARE` 0.86 to 0.935 (0.86 / 0.92: the subject is 8% smaller in the circle)
- `HEADROOM` 0.07 to 0.17, the same for every avatar except four animated ones whose moving part would leave the circle
- The crop may now reach above the top of the concept: `upper = min(top - HEADROOM * side, height - side)` instead of clipping it at 0, the rule `avatar.tawny_quill` already used. The subject starts within a few pixels of the concept's top edge, so a lower head needs transparent canvas above it; the bottom and side clamps are as before
- Printed by the clean-up: `hummingbird_clip: background [189.0, 189.0, 189.0], pockets 1315 px, crop (92, -183, 1213, 938), portrait 768`
- Crop `(137, 0, 1168, 1031)` to `(92, -183, 1213, 938)` on the concept
- Scale k = 0.9197. The rig geometry in portrait pixels (`HEAD`, `NECK_Y`, `NECK_BAND`, `SHOULDER_Y`, `BREATH`, `NOD`, `EYES`, `EYE_FEATHER`, the part polygons, pivots, `length` and `grow`, the glints and the ripples) is mapped through the crop change (x' = k x + offset, sizes and pixel amplitudes times k), so every mask stays on the same anatomy and the motion is the same relative to the subject. Angles, timing, blink frames and `STATIC_FRAME` are unchanged
- `register` now clips a crop edge above the concept to 0 when it marks the registration region, as the later loop scripts of this batch do; with the old slice a negative top would have marked the wrong rows
- Printed by the loop: `hummingbird_clip_loop: blink edit registered (tx, ty, scale, angle) = (-0.1208, -0.2336, 1.0005, 0.0003), residual 0.00095 -> 0.00090, eye colour offset [-4.6, -1.9, 1.1], parts 2, filled 4081 px`
- Head top (first row of `export/static.png` with alpha above 0.5, as a share of the canvas height): 0.010 before, 0.172 after; highest over the 16 loop frames 0.005 before, 0.167 after

Run on 2026-11-16 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/hummingbird_clip/work/hummingbird_clip_cleanup.py hummingbird_clip assets-src/cosmetic/avatar/hummingbird_clip/concept/20261115-01.png assets-src/cosmetic/avatar/hummingbird_clip/work
python -I assets-src/cosmetic/avatar/hummingbird_clip/work/hummingbird_clip_loop.py assets-src/cosmetic/avatar/hummingbird_clip/concept assets-src/cosmetic/avatar/hummingbird_clip/work assets-src/cosmetic/avatar/hummingbird_clip/export
```

Determinism: two more runs of the same commands into two empty folders wrote the same 26 files byte for byte, and both match the files committed here.

`npm run asset:check -- assets-src/cosmetic/avatar/hummingbird_clip` after the re-crop: 17 sprites, ASTC 6x6 estimate 278,528 bytes, limit under 300,000 (#171, animated avatars under 300 KB each): pass.

The Files table below keeps the hashes of the first export. The files the re-crop changed now hash as follows:

| File | sha256 after the re-crop |
|---|---|
| export/anim/0001.png | `3e26c896c802e090ccd6129017ef6eb08072fecf86208477135d0f84abcee374` |
| export/anim/0002.png | `7d9952812efe40b24f1fe07727d500ae808fba4b1acd0a430496e0bdd6ba8f88` |
| export/anim/0003.png | `8051300154a2e86ce91359f88e1aa734f20ea90c1580a13a018f72ed2b2b4bb3` |
| export/anim/0004.png | `b21127a4a7c3766795af004f2ef2bcdcac0338944874a5b743ae9372a6147167` |
| export/anim/0005.png | `5370b6858b0eb96c23fa376be7a6c5fa7bf235fdf5e71d627b501dcd535e65ec` |
| export/anim/0006.png | `6c0f96c8ec4ec6e741bd78925c3fd88e42b8098dbe9bef4b5560a27fc73ae30f` |
| export/anim/0007.png | `c7767e2e605e932cee2acf2de85a6e6f8ad4b11f33e904f89c32ef05adb29715` |
| export/anim/0008.png | `9a47c54aeebee25a457270fbb5ec681b44f614988b022fde2f12773b12a5105c` |
| export/anim/0009.png | `5182ea583b7409bd4e948e009c1b5fa2acc7839f6669b9819d3dff0c21961867` |
| export/anim/0010.png | `98f1a25e6c2805bac4e312ec9af393f706ce84bc372bd74e893f326652a600f1` |
| export/anim/0011.png | `1a6ff408ad89e1c19cb123a2dbf2e67c5771e5bd7939bad798b4fff137429fb2` |
| export/anim/0012.png | `cd5500a53f7fc6193bde092360d600295d7327f2febbc1d8e97f9889c7ef4610` |
| export/anim/0013.png | `e3f7564710bff4873b9eea13a659e5b5acc59cd247565ab1d2f7dedad19dba01` |
| export/anim/0014.png | `f1ec79cf707e5c19bf92ea1330e529c9753af5e23d398996761c0a167c8d1c79` |
| export/anim/0015.png | `1e305e16a3076635673f166d8bc64b39af3e97d9d03e37e519b5481732543fb5` |
| export/anim/0016.png | `a9d834ae69b67e4868b1a773e11ca1394f1a2d4e20b3f021372dc78878e7fa3b` |
| export/static.png | `5182ea583b7409bd4e948e009c1b5fa2acc7839f6669b9819d3dff0c21961867` |
| work/hummingbird_clip_base.png | `b1bd1175f83f4b784fc065d7955ddb9f1a97de4cdf39319bdc1ff6b1b6a3314e` |
| work/hummingbird_clip_cleanup.py | `d699aa8b38c6c66da5f607232e08ae8c3bd0032a3517e9a07c89d4ef7d5761ab` |
| work/hummingbird_clip_earring_left.png | `693510b2cf00f18d70144f9221859a09968ba97393a2b048f11db932c3b677a2` |
| work/hummingbird_clip_earring_right.png | `80637bc6edcc1f0fd4eaefa692f7f8c79215b77ca7204c24d9642ce110464bf1` |
| work/hummingbird_clip_eyes_closed.png | `553c38520e0b2e8768e555f8eaf8fdc2a2dbd1358fc9cd8acef9bbe4361f006b` |
| work/hummingbird_clip_eyes_open.png | `9092b21b787abeeccbc5359e71f47be2f0585a6c92e82c9b5050817bcfdb879d` |
| work/hummingbird_clip_head.png | `2d92aa17fe8167ffef42749145d818ebe22786aa3aa93daacb917a61e582651f` |
| work/hummingbird_clip_loop.py | `94f442bb7f26e9cfb8e5063727da3132f588ace250d804fc5a34e40669299f98` |
| work/hummingbird_clip_portrait.png | `74c76852a5a96fc02ba958ff38406246eb011c2cab6caa588ffc6e4de5f81a74` |
| work/hummingbird_clip_torso.png | `656118196fe0b48152d7b1613d5c60a4f8efa0485358f9b4ae01ea09cf806111` |

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `f730ea001abfff7fac435d45efef15f28b2a7182d58d36def1b177b87ac79f0f` | 1312 x 1199 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e025-472a-7023-8a4b-ee9666b5b76b` |
| concept/20261115-02.png | `d0f5fb117b4976916b9522681b36887714f1c57f4614b2d9b0f86b91a7465d71` | 1312 x 1199 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e026-d07f-7f33-ab9a-db5c53ae6d3a` |
| work/hummingbird_clip_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-15, from `amethyst_courtier_cleanup.py` |
| work/hummingbird_clip_loop.py | `784de8fbf656f3abc6380923c5af4427578bac482775a98c78eadf1243105672` | W-4 and W-5 script | Written by Claude Code, 2026-11-15; the same script for every avatar of this batch with the item's geometry |
| work/hummingbird_clip_master.png | `6cdea259a9db0c35d0b5d3eed177317e37f51ae605d1056b03b58c5e59f12462` | 1312 x 1199 RGBA, the cut-out master | `hummingbird_clip_cleanup.py` |
| work/hummingbird_clip_base.png | `a0433ebfc6b6febe4bd3c1b416eb4cb3d03d67025a5ae702800b5cb5bbf5a8f6` | The portrait with the moving parts taken out (filled or cleared) | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_earring_left.png | `d7ab4d89e87f63cb1d7817f689998ece40d69cf6cb13976edb7ffa8da8947dda` | Moving part `earring_left`: the portrait through its feathered polygon | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_earring_right.png | `abf8fad7eed6a569b01f33ea6874af33b0386c364e9b77f51fde03683ff3b25f` | Moving part `earring_right`: the portrait through its feathered polygon | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_eyes_closed.png | `a98b82dc8b295de51a38702a0b64a27de88e8d0b6d64690039d1704f6bf33a1f` | Eye region, closed (aligned edit) through the same mask | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_eyes_open.png | `f9e8c2e01621fb45b87ad565a6544598c79a72ad1673a3fd73db6b049d4557d0` | Eye region, open (canonical concept) through the feathered eye mask | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_head.png | `93f53ccc6db1706686be5a8020bb4d2fd1b741fd69b575791b3740e6a54da63e` | Head and hair: the portrait weighted by the head field of the warp | `hummingbird_clip_loop.py` |
| work/hummingbird_clip_portrait.png | `639f2c9713ee53d0893e8b3ec80fd979ed34324c0486e3971339131a7bd5f4ad` | 768 x 768 RGBA portrait crop, no circle | `hummingbird_clip_cleanup.py` |
| work/hummingbird_clip_torso.png | `143490561a97b891ed75099984ae13e66d4b55ee7201b9809a8ea6aeefa4bb55` | Torso and costume: the portrait weighted by one minus the head field | `hummingbird_clip_loop.py` |
| export/static.png | `f42aafa21bc78d8a01ad375e748baa44b16271cbf319b9bcfc91b22ab8a80a19` | 192 x 192 RGBA static, loop frame 9 | `hummingbird_clip_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `hummingbird_clip_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 16:58:21 UTC; `c2pa.created` 2026-11-15T16:59:29Z |
| 2026-11-15 | concept/20261115-02.png | Edit of `concept/20261115-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 17:00:02 UTC; `c2pa.created` 2026-11-15T17:01:15Z |
| 2026-11-15 | work/, export/ | The scripts above |
| 2026-11-16 | export/static.png, work/hummingbird_clip_base.png, work/hummingbird_clip_cleanup.py, work/hummingbird_clip_earring_left.png, work/hummingbird_clip_earring_right.png, work/hummingbird_clip_eyes_closed.png, work/hummingbird_clip_eyes_open.png, work/hummingbird_clip_head.png, work/hummingbird_clip_loop.py, work/hummingbird_clip_portrait.png, work/hummingbird_clip_torso.png, export/anim/0001.png to 0016.png | Re-crop: smaller and lower per owner feedback. `CROP_SHARE` 0.86 to 0.935, `HEADROOM` 0.07 to 0.17, crop `(137, 0, 1168, 1031)` to `(92, -183, 1213, 938)`; see "Re-crop, 2026-11-16" |

Prompt of `concept/20261115-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a girl of about sixteen with deep brown skin, a voluminous round afro held back by a wide ivory silk headband with a small enamel hummingbird clip, small crystal drop earrings, and a high-collared sage green silk blouse with pearl buttons, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261115-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a girl of about sixteen with deep brown skin, a voluminous round afro held back by a wide ivory silk headband with a small enamel hummingbird clip, small crystal drop earrings, and a high-collared sage green silk blouse with pearl buttons, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261115-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261115-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-15.
