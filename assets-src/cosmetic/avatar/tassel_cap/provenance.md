# avatar.tassel_cap

Animated avatar of set 1 (#171), batch A, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-15 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: `concept/20261115-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261115-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a boy of about fourteen with warm tan skin and short wavy black hair, a small round burgundy felt cap with a long gold silk tassel hanging down beside his ear, an ivory shirt and a mustard yellow silk waistcoat with gold piping"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt. Not asked for: freckles and a braided gold cord around the cap. The small round brimless cap with a tassel resembles traditional caps that carry cultural or religious meaning in some markets; that is a point for the cultural review.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.09, 0.12 px, scale 0.9999 and rotation -0.0000 rad (mean squared grey difference outside the eyes 0.00109 before, 0.00108 after), and the eye patch was offset by [-7.9, -0.4, 3.1] in RGB to meet the canonical skin. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly; point for the cultural review: the round brimless cap with a tassel
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `epic`, source `shop:offer.avatar_tassel_cap`. The offer id is a placeholder until the shop record exists (V-15). Animated avatars are proposed at `epic` and above, as the batch leans to the higher tiers of `docs/rarity.md`; this batch proposes 5 epic, 4 legendary, 1 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signatures were not validated with a C2PA tool.

| Field | `concept/20261115-01.png` | `concept/20261115-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-15T16:43:27Z | `c2pa.created`, 2026-11-15T16:44:25Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-15 16:42:44 | 2026-11-15 16:43:49 |
| Codex session | `01a1e016-fac0-7992-9360-e593494fb21e` | `01a1e017-f7d3-76c3-b8cf-723a74428734` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked with `cmp`). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/tassel_cap/work/tassel_cap_cleanup.py tassel_cap assets-src/cosmetic/avatar/tassel_cap/concept/20261115-01.png assets-src/cosmetic/avatar/tassel_cap/work
```

It printed background [199.0, 198.0, 198.0], pockets 2690 px, crop (76, 0, 1154, 1078), portrait 768. The script is `amethyst_courtier_cleanup.py` of batch A of the static avatars with three changes, all for the same reason: on these concepts grey background stayed opaque between loose strands and the neck, and hair edges kept a light outline. (1) It writes the master and a square 768 x 768 portrait crop (`work/tassel_cap_portrait.png`, the same crop rule, no circle) instead of the 512 static, because the static of an animated avatar is a loop key frame. (2) Enclosed background pockets are taken within 24 of the grey (10 before) and up to 140 px inside the outline (80 before), and a pocket counts only if its median colour is within 12 of the grey, so pearls, crystal and silver that come close to the grey stay. (3) In a 6 px band along the background the alpha is capped by projecting each pixel between the grey and the mean colour of the nearby solid subject (Gaussian sigma 6); where that colour is itself within 40 of the grey, by the colour distance to the grey (0 at 14, 1 at 60). The colour of edge pixels is unmixed from the grey and blended toward the nearby subject colour by one minus the alpha, so faint edge pixels take the subject's colour instead of a light speck.

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
python -I assets-src/cosmetic/avatar/tassel_cap/work/tassel_cap_loop.py assets-src/cosmetic/avatar/tassel_cap/concept assets-src/cosmetic/avatar/tassel_cap/work assets-src/cosmetic/avatar/tassel_cap/export
```

It printed `tassel_cap_loop: blink edit registered (tx, ty, scale, angle) = (0.0861, 0.1229, 0.9999, -0.0000), residual 0.00109 -> 0.00108, eye colour offset [-7.9, -0.4, 3.1], parts 1, filled 13229 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run into empty folders wrote files byte-identical to each other and to the committed ones.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes. Half-closed frames made by cross-fading showed two sets of lids and were dropped | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `tassel` (layer `work/tassel_cap_tassel.png`, pivot (177, 280)) turns about its pivot by up to 4.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.4, the turn growing from 0 at the pivot to full at 200 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot | The part lies over hair or skin, so the hole it leaves in the base layer is filled by normalized convolution from its surroundings (Gaussian sigma 48 down to 3); only a sliver of it shows while the part is off its rest position |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 3.24 of 255, against 2.78 on average and 3.57 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `98943ffa99a7af7f8bd418e7a24718670d73cb85d7880b66cd087d7f08632540` | 1254 x 1254 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e016-fac0-7992-9360-e593494fb21e` |
| concept/20261115-02.png | `2b04fe7784bc1c7e14cfaef6aa619b26da5f03fee6cbf6ad866bdd89fd8a527e` | 1254 x 1254 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e017-f7d3-76c3-b8cf-723a74428734` |
| work/tassel_cap_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-15, from `amethyst_courtier_cleanup.py` |
| work/tassel_cap_loop.py | `30cf5ae739ee3941d9c905119ab3fbbb194bd7bd7b369461719df9723fd980ee` | W-4 and W-5 script | Written by Claude Code, 2026-11-15; the same script for every avatar of this batch with the item's geometry |
| work/tassel_cap_master.png | `644b81763889e3b40d17511460be6e5b1870bd3c475a209f3fbbf617f52a0e97` | 1254 x 1254 RGBA, the cut-out master | `tassel_cap_cleanup.py` |
| work/tassel_cap_base.png | `3c2ba4b3fb07aef3fcb9099fe38fc8af520373c88a7e213ea40b7ecefea71a4a` | The portrait with the moving parts taken out (filled or cleared) | `tassel_cap_loop.py` |
| work/tassel_cap_eyes_closed.png | `7374bfefd3ecc3c88337b640cb062ca117b70e5f8e5796b2b9f15364e157c688` | Eye region, closed (aligned edit) through the same mask | `tassel_cap_loop.py` |
| work/tassel_cap_eyes_open.png | `6f5eb0dfe4e4fae2632c673dc200b0533431b225d37a2cb423d77646afd02d6f` | Eye region, open (canonical concept) through the feathered eye mask | `tassel_cap_loop.py` |
| work/tassel_cap_head.png | `529c50cb5df18c7d3c3ce652c424997198e2ea654086501a2683597e17f49c7f` | Head and hair: the portrait weighted by the head field of the warp | `tassel_cap_loop.py` |
| work/tassel_cap_portrait.png | `0015ce0ca0012d2436891a6042420616e2cff897fdc5116d158823512b6184af` | 768 x 768 RGBA portrait crop, no circle | `tassel_cap_cleanup.py` |
| work/tassel_cap_tassel.png | `265ce0c467c89e640d0233ca2d0b3bf2c2de6e72e6ff65c7ae55890eefa0fe21` | Moving part `tassel`: the portrait through its feathered polygon | `tassel_cap_loop.py` |
| work/tassel_cap_torso.png | `b66c634cd6c1b242c702716ae8d5a4b75bcf16a8243ebb20bee27687fa95660c` | Torso and costume: the portrait weighted by one minus the head field | `tassel_cap_loop.py` |
| export/static.png | `8ff747c0b13c981245adaf8f992dbb2b8019283c8bdd152729279fdd531dc368` | 192 x 192 RGBA static, loop frame 9 | `tassel_cap_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `tassel_cap_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 16:42:44 UTC; `c2pa.created` 2026-11-15T16:43:27Z |
| 2026-11-15 | concept/20261115-02.png | Edit of `concept/20261115-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 16:43:49 UTC; `c2pa.created` 2026-11-15T16:44:25Z |
| 2026-11-15 | work/, export/ | The scripts above |

Prompt of `concept/20261115-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a boy of about fourteen with warm tan skin and short wavy black hair, a small round burgundy felt cap with a long gold silk tassel hanging down beside his ear, an ivory shirt and a mustard yellow silk waistcoat with gold piping, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261115-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a boy of about fourteen with warm tan skin and short wavy black hair, a small round burgundy felt cap with a long gold silk tassel hanging down beside his ear, an ivory shirt and a mustard yellow silk waistcoat with gold piping, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261115-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261115-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-15.
