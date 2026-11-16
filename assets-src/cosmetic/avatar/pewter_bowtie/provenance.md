# avatar.pewter_bowtie

Animated avatar of set 1 (#171), batch B, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 and 2026-11-16 (UTC) while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-15 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`, final runs on 2026-11-16. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: `concept/20261115-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261115-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a man in his forties with medium brown skin and South Asian features, wavy black hair and a short neat black beard, a black velvet dinner jacket over a white wing-collar shirt, and a wide pewter-grey silk bow tie with long loose trailing ends"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt. Not asked for: a pleated shirt front with white studs.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.27, 0.11 px, scale 0.9999 and rotation -0.0001 rad (mean squared grey difference outside the eyes 0.00058 before, 0.00055 after), and the eye patch was offset by [-5.7, -1.4, -0.5] in RGB to meet the canonical skin. The changes outside the eye mask are small and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `epic`, source `shop:offer.avatar_pewter_bowtie`. The offer id is a placeholder until the shop record exists (V-15). One ambient motion, the trailing ends of the bow. Animated avatars stay at `epic` and above as in batch A; batch B proposes 5 epic, 4 legendary, 1 mythic, the same split as batch A.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signatures were not validated with a C2PA tool.

| Field | `concept/20261115-01.png` | `concept/20261115-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-15T17:34:42Z | `c2pa.created`, 2026-11-15T17:50:42Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-15 17:34:07 | 2026-11-15 17:49:40 |
| Codex session | `01a1e046-04a9-7240-81ad-37083f6b3354` | `01a1e054-417a-78d3-9295-d7e6ae840c49` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked with `cmp`). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-16 (UTC) by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/pewter_bowtie/work/pewter_bowtie_cleanup.py pewter_bowtie assets-src/cosmetic/avatar/pewter_bowtie/concept/20261115-01.png assets-src/cosmetic/avatar/pewter_bowtie/work
```

It printed background [181.0, 180.0, 180.0], pockets 1336 px, crop (64, 0, 1142, 1078), portrait 768. The script is `ruby_drop_cleanup.py` of batch A unchanged. It writes the master and a square 768 x 768 portrait crop (`work/pewter_bowtie_portrait.png`, no circle); the enclosed grey pockets, the edge re-matte and the fringe removal work as described in the batch A provenance of `avatar.ruby_drop`.

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Background, trim, class canvas and pivot | Yes, canvas open | As in batch A. The loop script puts a circular alpha (soft edge of 2.25 px at 768, 0.56 px at 192) on every frame, because the portrait is shown in a circle (`docs/art-direction.md` 5.1); the class canvas is open (O-6) and D-1 sets 192 x 192 |
| 2. Paint-over, AI artefacts | No | No paint-over |
| 3. Moving parts on their own layers | Yes | Written by the loop script into `work/` (table below); the ripples are warps of the base layer and need no layer of their own |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow, rarity colour or team hue seen in the concept, the master or the frames |
| 5. Silhouette and smallest-size tests | In part | The agent looked at every frame at 192 px and enlarged on a dark backdrop; face, expression and the moving part read. The smallest display size and frame band of an avatar are open (#168, O-6), so the test is not final |

## Animation (W-4) and export (W-5)

Run on 2026-11-16 (UTC) by Claude Code with the same versions, after the clean-up:

```bash
python -I assets-src/cosmetic/avatar/pewter_bowtie/work/pewter_bowtie_loop.py assets-src/cosmetic/avatar/pewter_bowtie/concept assets-src/cosmetic/avatar/pewter_bowtie/work assets-src/cosmetic/avatar/pewter_bowtie/export
```

It printed `pewter_bowtie_loop: blink edit registered (tx, ty, scale, angle) = (0.2680, 0.1099, 0.9999, -0.0001), residual 0.00058 -> 0.00055, eye colour offset [-5.7, -1.4, -0.5], parts 0, ripples 1, filled 0 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run into empty folders wrote files byte-identical to each other and to the committed ones.

The script is the batch A loop script (`ruby_drop_loop.py`) with this item's geometry and these additions, the same for every avatar of this batch: the concept file names are constants; a part can open and close (`flap`, a horizontal scale about its pivot); `RIPPLES` add a smooth sideways wave to the warp inside a feathered polygon, so cloth, ribbon and chain move without a cut-out layer and without seams; a hole behind a part is cleared where the fill estimate is less than half covered, which keeps a filled hole from leaving a smear over the background; a glint is clipped to the bright pixels of the portrait and its peak opacity is a constant (`GLINT_PEAK`, 0.55 as in batch A unless the item says otherwise); and the registration region is clipped to the image when the crop reaches past its edge.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient ripple | The two long trailing ends of the bow, below the knot: inside a polygon feathered by Gaussian sigma 14 (768 scale), the pixels shift sideways by 5 x (sin(2 pi y / 200 - 1 t) - sin(2 pi y / 200)) px at 768, faded in over 80 px below y 545, so the wave runs down once per loop and the attached top stays still | Warp only, so no seams and no hole to fill; the polygon ends before the circle edge |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 2.00 of 255, against 1.81 on average and 2.20 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `0828af8a575c9094fa8d503eb017afca4843250bfa60e9c7ad8898bbf0332040` | 1254 x 1254 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e046-04a9-7240-81ad-37083f6b3354` |
| concept/20261115-02.png | `c5aa8e0bb5deec16a3ea1c75638dd78b6488798c547b0ac314ec591a1ba9a804` | 1254 x 1254 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e054-417a-78d3-9295-d7e6ae840c49` |
| work/pewter_bowtie_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-15 and 2026-11-16, from `ruby_drop_cleanup.py` |
| work/pewter_bowtie_loop.py | `a5f8cadae9c02624d864f3c369e76c3bbd06369d5402847cb2a91eac5f5c3b2a` | W-4 and W-5 script | Written by Claude Code, 2026-11-15 and 2026-11-16, from `ruby_drop_loop.py`; the same script for every avatar of this batch with the item's geometry |
| work/pewter_bowtie_master.png | `8bab273855ae21e552f208e4214d41560db6349cbd66910c5aa214e6bd9e997c` | 1254 x 1254 RGBA, the cut-out master | `pewter_bowtie_cleanup.py` |
| work/pewter_bowtie_base.png | `85424f226a6b8517742cfbfe403d50089dc305afbb60966de93b2067f510017f` | The portrait with the moving parts taken out (filled or cleared) | `pewter_bowtie_loop.py` |
| work/pewter_bowtie_eyes_closed.png | `14bfcbe9fc692806fad3cf8d1927a78d17558ab79c288eba08001684e9dd1e3d` | Eye region, closed (aligned edit) through the same mask | `pewter_bowtie_loop.py` |
| work/pewter_bowtie_eyes_open.png | `824bdf4bdb29d8534509f489926f8dc9555315d8e2dd4b0ffe74dcc39058773d` | Eye region, open (canonical concept) through the feathered eye mask | `pewter_bowtie_loop.py` |
| work/pewter_bowtie_head.png | `91ea1d2c01173e82f8724eb8080aa76f883b4aa01b8ae706557988a7c1a10d34` | Head and hair: the portrait weighted by the head field of the warp | `pewter_bowtie_loop.py` |
| work/pewter_bowtie_portrait.png | `76908efd142bcd6429405ce3fe2f826dee803a72821dccc8cad02b2b8ae37154` | 768 x 768 RGBA portrait crop, no circle | `pewter_bowtie_cleanup.py` |
| work/pewter_bowtie_torso.png | `d7e9ff9ee91eb0e0be3d63d85d8961e2c74f693e5b0aea27208d469acf8d0b8d` | Torso and costume: the portrait weighted by one minus the head field | `pewter_bowtie_loop.py` |
| export/static.png | `b0c2a749cf19bdfd98b1a02cbb27dffa0a146c773b1d15b6194f6e9de9ce3163` | 192 x 192 RGBA static, loop frame 9 | `pewter_bowtie_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `pewter_bowtie_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 17:34:07 UTC; `c2pa.created` 2026-11-15T17:34:42Z |
| 2026-11-15 | concept/20261115-02.png | Edit of `concept/20261115-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 17:49:40 UTC; `c2pa.created` 2026-11-15T17:50:42Z |
| 2026-11-16 | work/, export/ | The scripts above |

Prompt of `concept/20261115-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a man in his forties with medium brown skin and South Asian features, wavy black hair and a short neat black beard, a black velvet dinner jacket over a white wing-collar shirt, and a wide pewter-grey silk bow tie with long loose trailing ends, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261115-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a man in his forties with medium brown skin and South Asian features, wavy black hair and a short neat black beard, a black velvet dinner jacket over a white wing-collar shirt, and a wide pewter-grey silk bow tie with long loose trailing ends, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261115-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261115-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-16.
