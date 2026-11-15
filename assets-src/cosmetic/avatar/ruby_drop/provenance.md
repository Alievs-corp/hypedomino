# avatar.ruby_drop

Animated avatar of set 1 (#171), batch A, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-15 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: `concept/20261115-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261115-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a woman in her thirties with light brown skin and freckles, short cropped black curly hair, long gold chandelier earrings with ruby drops, and a high-necked wine-red velvet gown with a gold lace collar"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt. Not asked for but present: a large ruby and pearl brooch with a pearl drop on the lace collar, puffed sleeves and gold embroidered bands down the gown.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.35, 0.19 px, scale 1.0002 and rotation -0.0002 rad (mean squared grey difference outside the eyes 0.00245 before, 0.00233 after), and the eye patch was offset by [-6.7, -2.3, -0.9] in RGB to meet the canonical skin. The edit also softened the smile; that is outside the eye mask and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `legendary`, source `shop:offer.avatar_ruby_drop`. The offer id is a placeholder until the shop record exists (V-15). Animated avatars are proposed at `epic` and above, as the batch leans to the higher tiers of `docs/rarity.md`; this batch proposes 5 epic, 4 legendary, 1 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signatures were not validated with a C2PA tool.

| Field | `concept/20261115-01.png` | `concept/20261115-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-15T16:38:13Z | `c2pa.created`, 2026-11-15T16:39:37Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-15 16:37:17 | 2026-11-15 16:38:36 |
| Codex session | `01a1e011-fe7d-76c3-8400-6542666754e9` | `01a1e013-32d5-7843-8dce-5e10ee152afc` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked with `cmp`). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/ruby_drop/work/ruby_drop_cleanup.py ruby_drop assets-src/cosmetic/avatar/ruby_drop/concept/20261115-01.png assets-src/cosmetic/avatar/ruby_drop/work
```

It printed background [192.0, 192.0, 191.0], pockets 1547 px, crop (62, 0, 1093, 1031), portrait 768. The script is `amethyst_courtier_cleanup.py` of batch A of the static avatars with three changes, all for the same reason: on these concepts grey background stayed opaque between loose strands and the neck, and hair edges kept a light outline. (1) It writes the master and a square 768 x 768 portrait crop (`work/ruby_drop_portrait.png`, the same crop rule, no circle) instead of the 512 static, because the static of an animated avatar is a loop key frame. (2) Enclosed background pockets are taken within 24 of the grey (10 before) and up to 140 px inside the outline (80 before), and a pocket counts only if its median colour is within 12 of the grey, so pearls, crystal and silver that come close to the grey stay. (3) In a 6 px band along the background the alpha is capped by projecting each pixel between the grey and the mean colour of the nearby solid subject (Gaussian sigma 6); where that colour is itself within 40 of the grey, by the colour distance to the grey (0 at 14, 1 at 60). The colour of edge pixels is unmixed from the grey and blended toward the nearby subject colour by one minus the alpha, so faint edge pixels take the subject's colour instead of a light speck.

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
python -I assets-src/cosmetic/avatar/ruby_drop/work/ruby_drop_loop.py assets-src/cosmetic/avatar/ruby_drop/concept assets-src/cosmetic/avatar/ruby_drop/work assets-src/cosmetic/avatar/ruby_drop/export
```

It printed `ruby_drop_loop: blink edit registered (tx, ty, scale, angle) = (0.3499, 0.1851, 1.0002, -0.0002), residual 0.00245 -> 0.00233, eye colour offset [-6.7, -2.3, -0.9], parts 2, filled 12854 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run into empty folders wrote files byte-identical to each other and to the committed ones.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes. Half-closed frames made by cross-fading showed two sets of lids and were dropped | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `earring_left` (layer `work/ruby_drop_earring_left.png`, pivot (285, 408)) turns about its pivot by up to 5.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.5, the turn growing from 0 at the pivot to full at 90 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot; `earring_right` (layer `work/ruby_drop_earring_right.png`, pivot (560, 400)) turns about its pivot by up to 5.0 degrees with sin(lag) - sin(2 pi (f - 8) / 16 + lag), lag 0.8, the turn growing from 0 at the pivot to full at 90 px (768 scale), so the top stays attached and the end swings; the part also follows the head or torso warp at its pivot | The part hangs over the background, so the hole it leaves in the base layer is transparent |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 2.88 of 255, against 2.72 on average and 3.09 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `faad2ccb35f1f77d29d38e54709dd9806f255b45c11b69657d810659ab1b7d58` | 1199 x 1312 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e011-fe7d-76c3-8400-6542666754e9` |
| concept/20261115-02.png | `765ffa799427fd22a81b15250b2c5a5fbf00fcb54d80be0b5f1f64015cb5f141` | 1199 x 1312 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e013-32d5-7843-8dce-5e10ee152afc` |
| work/ruby_drop_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-15, from `amethyst_courtier_cleanup.py` |
| work/ruby_drop_loop.py | `1d933e11075b077f1f2e17a8717de9495e7ca2f6afb5de50dde8dedeb496e5d6` | W-4 and W-5 script | Written by Claude Code, 2026-11-15; the same script for every avatar of this batch with the item's geometry |
| work/ruby_drop_master.png | `6738ea18002dbd05baad3ea9f0130e9b2937684354783341f284837efc94e87c` | 1199 x 1312 RGBA, the cut-out master | `ruby_drop_cleanup.py` |
| work/ruby_drop_base.png | `61e9015a9e4b66de9d7d8b8d600bfaf68f458504f91ac4c41549f109b9179475` | The portrait with the moving parts taken out (filled or cleared) | `ruby_drop_loop.py` |
| work/ruby_drop_earring_left.png | `3344822b6e3c9231e618f53e9cb494a85ccf7fa2a129a373fe439c1c1d502ab9` | Moving part `earring_left`: the portrait through its feathered polygon | `ruby_drop_loop.py` |
| work/ruby_drop_earring_right.png | `8408bba12038d99ebbf69e2aa55402c55009080275f371cf7978040e03c412ad` | Moving part `earring_right`: the portrait through its feathered polygon | `ruby_drop_loop.py` |
| work/ruby_drop_eyes_closed.png | `939b4d2e269e130bda748bca954dbc2e253b04a2c47e90aa7193760925763ea0` | Eye region, closed (aligned edit) through the same mask | `ruby_drop_loop.py` |
| work/ruby_drop_eyes_open.png | `0f427c3d0b856b7a14f4ffe67d74f60413f4820f55a6aa336296ab26ab000bfb` | Eye region, open (canonical concept) through the feathered eye mask | `ruby_drop_loop.py` |
| work/ruby_drop_head.png | `cdd9e776b994e49add5c00fec5cb14f5cccb319a5c23ba04b722a7b23fd5caff` | Head and hair: the portrait weighted by the head field of the warp | `ruby_drop_loop.py` |
| work/ruby_drop_portrait.png | `6fe0df714347b4fe7c528aaa9c35f2d0c4d1ab4dffdcdb826d81520887928b45` | 768 x 768 RGBA portrait crop, no circle | `ruby_drop_cleanup.py` |
| work/ruby_drop_torso.png | `5ea64b7cfbeb1364734001d71ee38d6a4ea98b27f411111dc70dc4605786fe2e` | Torso and costume: the portrait weighted by one minus the head field | `ruby_drop_loop.py` |
| export/static.png | `b2d67425a77414b9990cfa29a4824a726e1754f3b41ef1e7ab510a34f90433e8` | 192 x 192 RGBA static, loop frame 9 | `ruby_drop_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `ruby_drop_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 16:37:17 UTC; `c2pa.created` 2026-11-15T16:38:13Z |
| 2026-11-15 | concept/20261115-02.png | Edit of `concept/20261115-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 16:38:36 UTC; `c2pa.created` 2026-11-15T16:39:37Z |
| 2026-11-15 | work/, export/ | The scripts above |

Prompt of `concept/20261115-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her thirties with light brown skin and freckles, short cropped black curly hair, long gold chandelier earrings with ruby drops, and a high-necked wine-red velvet gown with a gold lace collar, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261115-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her thirties with light brown skin and freckles, short cropped black curly hair, long gold chandelier earrings with ruby drops, and a high-necked wine-red velvet gown with a gold lace collar, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261115-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261115-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-15.
