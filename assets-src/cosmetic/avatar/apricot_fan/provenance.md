# avatar.apricot_fan

Animated avatar of set 1 (#171), batch C, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-16 (UTC) while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-16 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`, final runs on 2026-11-16. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-16 for the Plus plan and records the date here
- Canonical concept: concept/20261116-02.png
- Inputs uploaded to the generator: `concept/20261116-02.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261116-03.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a woman in her thirties with light olive skin, dark brown hair in a soft low chignon, small pearl studs, an apricot silk gown with an ivory lace neckline, holding a small open ivory lace folding fan just below her chin in front of her collarbone, the whole fan and her hand fully visible and well inside the frame"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt: the open ivory lace fan is held in her right hand (the viewer's left) below her chin, in front of the lace neckline, well inside the frame. Not asked for: a jewelled leaf hair ornament with a small drop, a jewelled brooch with chains on the bodice and a crystal drop on the fan's rivet.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.20, 0.06 px, scale 1.0003 and rotation -0.0002 rad (mean squared grey difference outside the eyes 0.00148 before, 0.00145 after), and the eye patch was offset by [-8.6, -2.1, 2.3] in RGB to meet the canonical skin. The changes outside the eye mask are small and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `legendary`, source `shop:offer.avatar_apricot_fan`. The offer id is a placeholder until the shop record exists (V-15). A large prop in front of the costume: the lace fan opens and closes while it flutters about its rivet. Animated avatars stay at `epic` and above as in batches A and B; batch C proposes 5 epic, 3 legendary, 2 mythic, which brings the 30 animated avatars of set 1 to 15 epic, 11 legendary and 4 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-16; the signatures were not validated with a C2PA tool.

| Field | `concept/20261116-02.png` | `concept/20261116-03.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-16T08:29:25Z | `c2pa.created`, 2026-11-16T08:46:00Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-16 08:28:34 | 2026-11-16 08:44:57 |
| Codex session | `01a1e378-ea44-7cf2-ae24-30a896c6a9ee` | `01a1e387-eb42-78f2-8900-5234738a143f` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked by a byte comparison). One earlier generation for this item was rejected and is not committed: `20261116-01.png` (Codex session `01a1e372-e08c-7b83-b2ac-fd0b3aac4339`, prompt sent 2026-11-16 08:21:58 UTC), with the first prompt quoted below. It met the prompt, but the fan was held up beside her shoulder and reached the right edge of the concept, so with the standard crop it lay partly outside the circle and its motion would have been cut by the circle edge. The prompt was changed in one place, the position and size of the fan. The next generation became the canonical; the file numbers of this item therefore start at 02.

## Clean-up (W-3)

Run on 2026-11-16 (UTC) by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/apricot_fan/work/apricot_fan_cleanup.py apricot_fan assets-src/cosmetic/avatar/apricot_fan/concept/20261116-02.png assets-src/cosmetic/avatar/apricot_fan/work
```

It printed background [213.0, 212.0, 212.0], pockets 741 px, crop (63, 0, 1141, 1078), portrait 768. The script is `cornflower_ribbon_cleanup.py` of batch B unchanged (the batch A clean-up of `avatar.ruby_drop` with its 768 crop, grey-pocket rule and edge re-matte). It writes the master and a square 768 x 768 portrait crop (`work/apricot_fan_portrait.png`, no circle); the enclosed grey pockets, the edge re-matte and the fringe removal work as described in the batch A provenance of `avatar.ruby_drop`.

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Background, trim, class canvas and pivot | Yes, canvas open | As in batches A and B. The loop script puts a circular alpha (soft edge of 2.25 px at 768, 0.56 px at 192) on every frame, because the portrait is shown in a circle (`docs/art-direction.md` 5.1); the class canvas is open (O-6) and D-1 sets 192 x 192 |
| 2. Paint-over, AI artefacts | No | No paint-over |
| 3. Moving parts on their own layers | Yes | Written by the loop script into `work/` (table below); the ripples are warps of the base layer and need no layer of their own |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow, rarity colour or team hue seen in the concept, the master or the frames |
| 5. Silhouette and smallest-size tests | In part | The agent looked at every frame at 192 px and enlarged on a dark backdrop; face, expression and the moving part read. The smallest display size and frame band of an avatar are open (#168, O-6), so the test is not final |

## Animation (W-4) and export (W-5)

Run on 2026-11-16 (UTC) by Claude Code with the same versions, after the clean-up:

```bash
python -I assets-src/cosmetic/avatar/apricot_fan/work/apricot_fan_loop.py assets-src/cosmetic/avatar/apricot_fan/concept assets-src/cosmetic/avatar/apricot_fan/work assets-src/cosmetic/avatar/apricot_fan/export
```

It printed `apricot_fan_loop: blink edit registered (tx, ty, scale, angle) = (0.2014, 0.0578, 1.0003, -0.0002), residual 0.00148 -> 0.00145, eye colour offset [-8.6, -2.1, 2.3], parts 1, ripples 0, filled 71822 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run of the clean-up and the loop into empty folders wrote files byte-identical to each other and to the committed ones.

The script is the batch B loop script (`cornflower_ribbon_loop.py`, which carries the batch A rig of `ruby_drop_loop.py` with `flap`, `RIPPLES`, the clearing hole fill, clipped glints and the clipped registration region) with this item's geometry and one addition, the same for every avatar of batch C: a part can spin (`spin`, whole turns per loop added to its turn angle, so a full turn ends where it started). Parts without `spin` move exactly as in batch B.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `fan` (layer `work/apricot_fan_fan.png`, pivot (355, 668)) turns about its pivot by 4.0 degrees x (sin(lag) - sin(t + lag) + 0.6 sin(2t)), lag 0.0, the turn growing from 0 at the pivot to full at 80 px (768 scale), so the end near the pivot stays attached and the far end swings; at the same time its width about the vertical line through the pivot is scaled by 1 - 0.1 x (1 - cos 2t) / 2, so it folds in to 90% and opens again twice per loop; the part also follows the head or torso warp at its pivot; the hole it leaves in the base layer is filled from the surrounding pixels, and where that estimate is less than half covered it is cleared instead, so no smear shows over the background | The fan sits over the lace neckline and the sleeve, well inside the circle; the hand that holds it stays in the base layer, so the fan turns and folds about the rivet under the fingers |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 4.31 of 255, against 3.33 on average and 4.30 at most between the other neighbouring frames | Every motion is periodic in the 16 frames; the step from frame 16 to frame 1 is the largest by a hair because it lies half a loop from the rest frame, where the swing is fastest, not because of a jump |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261116-02.png | `10e138439dfc30b9f3bb942b69a06c95b424e43946c227748af875000757d8a2` | 1254 x 1254 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e378-ea44-7cf2-ae24-30a896c6a9ee` |
| concept/20261116-03.png | `be61b2b6e9d6306e92af0576528fe7e131af4689def851d61299dec0d81288d8` | 1254 x 1254 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e387-eb42-78f2-8900-5234738a143f` |
| work/apricot_fan_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_cleanup.py` |
| work/apricot_fan_loop.py | `3e009bb2bc3bd2be0a70ac457604fa4d21fee2728e9dc6c4eb7094ad1d0ed57e` | W-4 and W-5 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_loop.py`; the same script for every avatar of this batch with the item's geometry |
| work/apricot_fan_master.png | `c9f3da9c7d19406f270a1c1b1b091f21b67336ed7fa446230846088b1636b204` | 1254 x 1254 RGBA, the cut-out master | `apricot_fan_cleanup.py` |
| work/apricot_fan_base.png | `356b92aa19c0afad46b87ccfe21df1c95abf65a7c267127c01933eb31fbec593` | The portrait with the moving parts taken out (filled or cleared) | `apricot_fan_loop.py` |
| work/apricot_fan_eyes_closed.png | `952b6d9a59eeba83bbc229e350fc98b6d5544f3846ba36a413d3c25e9fbd1bec` | Eye region, closed (aligned edit) through the same mask | `apricot_fan_loop.py` |
| work/apricot_fan_eyes_open.png | `dca95191b0a6a1b72286885d02c89702091d05ea68cf332b193ca0b0ef7fa160` | Eye region, open (canonical concept) through the feathered eye mask | `apricot_fan_loop.py` |
| work/apricot_fan_fan.png | `6fd1607d714cce323e18a81768ad18185037556f838ab668f727945b8e9df54a` | Moving part `fan`: the portrait through its feathered polygon | `apricot_fan_loop.py` |
| work/apricot_fan_head.png | `67f21558eb628acafff869d68f9800f55e9c6c89bd830eec4f4edf8c15137793` | Head and hair: the portrait weighted by the head field of the warp | `apricot_fan_loop.py` |
| work/apricot_fan_portrait.png | `1256d3ff22b52e977a5c5b9ea3f4cbe4e58d31d9c2a6898603ae75248fce39a0` | 768 x 768 RGBA portrait crop, no circle | `apricot_fan_cleanup.py` |
| work/apricot_fan_torso.png | `8c4887731aa79ab17766ff634333512f38f918f9a9ef522472dce827c0fc6b3c` | Torso and costume: the portrait weighted by one minus the head field | `apricot_fan_loop.py` |
| export/static.png | `24993395c8e42ef9c80a38c938b09a516039a5914cb1ef821f5403cf55dd484e` | 192 x 192 RGBA static, loop frame 9 | `apricot_fan_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `apricot_fan_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-16 | concept/20261116-01.png (rejected, not committed) | P-3 with the style key and the first identity, quoted below. Sent 08:21:58 UTC; Codex session `01a1e372-e08c-7b83-b2ac-fd0b3aac4339` |
| 2026-11-16 | concept/20261116-02.png | P-3 with the style key, quoted below. Sent 08:28:34 UTC; `c2pa.created` 2026-11-16T08:29:25Z |
| 2026-11-16 | concept/20261116-03.png | Edit of `concept/20261116-02.png` (attached): eyes closed, everything else identical, quoted below. Sent 08:44:57 UTC; `c2pa.created` 2026-11-16T08:46:00Z |
| 2026-11-16 | work/, export/ | The scripts above |

Prompt of `concept/20261116-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her thirties with light olive skin, dark brown hair in a soft low chignon, small pearl studs, an apricot silk gown with an ivory lace neckline, holding a small open ivory lace folding fan just below her chin in front of her collarbone, the whole fan and her hand fully visible and well inside the frame, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261116-02.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261116-03.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her thirties with light olive skin, dark brown hair in a soft low chignon, small pearl studs, an apricot silk gown with an ivory lace neckline, holding a small open ivory lace folding fan just below her chin in front of her collarbone, the whole fan and her hand fully visible and well inside the frame, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261116-02.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261116-03.png. Do not use any API key.`

Prompt of the rejected `concept/20261116-01.png`, verbatim (the identity differs only in where the fan is held):

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her thirties with light olive skin, dark brown hair in a soft low chignon, small pearl studs, an apricot silk gown with an ivory lace neckline, holding an open ivory lace folding fan up beside her shoulder, the whole fan and her hand fully visible, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-16.
