# avatar.charcoal_peony

Animated avatar of set 1 (#171), batch C, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-16 (UTC) while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-16 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`, final runs on 2026-11-16. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-16 for the Plus plan and records the date here
- Canonical concept: concept/20261116-01.png
- Inputs uploaded to the generator: `concept/20261116-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261116-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a man in his eighties with medium-brown skin, a bald head and a full white moustache, a charcoal velvet jacket over an ivory shirt, holding a single long-stemmed pale pink peony up beside his shoulder, the whole flower, stem and hand fully visible"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt: he holds a single long-stemmed pale pink peony upright in his right hand (the viewer's left). Not asked for: a burgundy brocade waistcoat, a ruby cravat pin, a crystal fob on a chain, a burgundy pocket square and a ruby cufflink. The face sits right of the centre of the circle, because the crop already starts at the left edge of the concept to keep the flower inside.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.37, 0.18 px, scale 1.0003 and rotation -0.0001 rad (mean squared grey difference outside the eyes 0.00159 before, 0.00149 after), and the eye patch was offset by [-1.1, 3.3, 5.4] in RGB to meet the canonical skin. The changes outside the eye mask are small and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `epic`, source `shop:offer.avatar_charcoal_peony`. The offer id is a placeholder until the shop record exists (V-15). One ambient motion, the peony nodding on its stem. Animated avatars stay at `epic` and above as in batches A and B; batch C proposes 5 epic, 3 legendary, 2 mythic, which brings the 30 animated avatars of set 1 to 15 epic, 11 legendary and 4 mythic.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-16; the signatures were not validated with a C2PA tool.

| Field | `concept/20261116-01.png` | `concept/20261116-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-16T08:36:19Z | `c2pa.created`, 2026-11-16T08:51:53Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-16 08:35:24 | 2026-11-16 08:50:36 |
| Codex session | `01a1e37f-29f9-73c1-a820-9cbb02c7b9ab` | `01a1e38d-17d7-7082-acd0-76d072a89059` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked by a byte comparison). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-16 (UTC) by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/charcoal_peony/work/charcoal_peony_cleanup.py charcoal_peony assets-src/cosmetic/avatar/charcoal_peony/concept/20261116-01.png assets-src/cosmetic/avatar/charcoal_peony/work
```

It printed background [203.0, 203.0, 203.0], pockets 2444 px, crop (0, 0, 1031, 1031), portrait 768. The script is `cornflower_ribbon_cleanup.py` of batch B unchanged (the batch A clean-up of `avatar.ruby_drop` with its 768 crop, grey-pocket rule and edge re-matte). It writes the master and a square 768 x 768 portrait crop (`work/charcoal_peony_portrait.png`, no circle); the enclosed grey pockets, the edge re-matte and the fringe removal work as described in the batch A provenance of `avatar.ruby_drop`.

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
python -I assets-src/cosmetic/avatar/charcoal_peony/work/charcoal_peony_loop.py assets-src/cosmetic/avatar/charcoal_peony/concept assets-src/cosmetic/avatar/charcoal_peony/work assets-src/cosmetic/avatar/charcoal_peony/export
```

It printed `charcoal_peony_loop: blink edit registered (tx, ty, scale, angle) = (0.3680, 0.1834, 1.0003, -0.0001), residual 0.00159 -> 0.00149, eye colour offset [-1.1, 3.3, 5.4], parts 1, ripples 0, filled 36667 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run of the clean-up and the loop into empty folders wrote files byte-identical to each other and to the committed ones.

The script is the batch B loop script (`cornflower_ribbon_loop.py`, which carries the batch A rig of `ruby_drop_loop.py` with `flap`, `RIPPLES`, the clearing hole fill, clipped glints and the clipped registration region) with this item's geometry and one addition, the same for every avatar of batch C: a part can spin (`spin`, whole turns per loop added to its turn angle, so a full turn ends where it started). Parts without `spin` move exactly as in batch B.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `peony` (layer `work/charcoal_peony_peony.png`, pivot (220, 512)) turns about its pivot by 5.0 degrees x (sin(lag) - sin(t + lag) + 0.3 sin(2t)), lag 0.0, the turn growing from 0 at the pivot to full at 220 px (768 scale), so the end near the pivot stays attached and the far end swings; the part also follows the head or torso warp at its pivot; the hole it leaves in the base layer is filled from the surrounding pixels, and where that estimate is less than half covered it is cleared instead, so no smear shows over the background | Bloom, leaves and the stem above the hand are one layer over the background; the turn grows from the hand along the stem, so the bloom nods, inside the circle |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 4.30 of 255, against 3.49 on average and 4.34 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261116-01.png | `e93cd8a2135b6c21819e076f2760392948ca61ae5a5ff821f6ad54f71656b7c4` | 1199 x 1312 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e37f-29f9-73c1-a820-9cbb02c7b9ab` |
| concept/20261116-02.png | `abada59b850a199a2ebb54f5095515af09412e3317d2849ce339a7e9162eee3d` | 1199 x 1312 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e38d-17d7-7082-acd0-76d072a89059` |
| work/charcoal_peony_cleanup.py | `8c39fa1eb5065d0824cb0eaf53942d85348631f0b7c1a41fc272137a55333af3` | W-3 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_cleanup.py` |
| work/charcoal_peony_loop.py | `0fe4c37da5662944f3777e98f199c807ffc8ae885972570963b4649b2ab25b58` | W-4 and W-5 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_loop.py`; the same script for every avatar of this batch with the item's geometry |
| work/charcoal_peony_master.png | `5d0dc9999b109e9e4c5a6411f7410106fe723abdffae55e364f719a346f47d76` | 1199 x 1312 RGBA, the cut-out master | `charcoal_peony_cleanup.py` |
| work/charcoal_peony_base.png | `632fa3684107e1a85ca250dafaa451702cad298bbd0f1a96e51412efb5199570` | The portrait with the moving parts taken out (filled or cleared) | `charcoal_peony_loop.py` |
| work/charcoal_peony_eyes_closed.png | `3aa2d183f6b29a38981524953ee6aa5d7ef8ed463ea415a82c054ef757c2fd14` | Eye region, closed (aligned edit) through the same mask | `charcoal_peony_loop.py` |
| work/charcoal_peony_eyes_open.png | `17258dd8ba5eaf705deb209a731de3f75a5b1b6d1ff8d1d478c15ed0ed0f9dbb` | Eye region, open (canonical concept) through the feathered eye mask | `charcoal_peony_loop.py` |
| work/charcoal_peony_head.png | `520504807ffdd7bd67142b0fa8e19a2b62aa2267a7dfc9519f6788ca031e1d41` | Head and hair: the portrait weighted by the head field of the warp | `charcoal_peony_loop.py` |
| work/charcoal_peony_peony.png | `b6ee342a2ac7313b33fbca94f1a2a922fb8f08167a21bc5c527aa7be995f0b68` | Moving part `peony`: the portrait through its feathered polygon | `charcoal_peony_loop.py` |
| work/charcoal_peony_portrait.png | `438a8e3edf220430945a35d8120fd8afa8101b9f8744b1636271ba51640c6e57` | 768 x 768 RGBA portrait crop, no circle | `charcoal_peony_cleanup.py` |
| work/charcoal_peony_torso.png | `0afecbf6622ddf0d0f60a07402aa1a5393b34ba9a257ce81dbf9b7f5485a8282` | Torso and costume: the portrait weighted by one minus the head field | `charcoal_peony_loop.py` |
| export/static.png | `86538671eb313bed6d9dcfbab88d8fd3babb66d514349b04a811ca405da210f0` | 192 x 192 RGBA static, loop frame 9 | `charcoal_peony_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `charcoal_peony_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-16 | concept/20261116-01.png | P-3 with the style key, quoted below. Sent 08:35:24 UTC; `c2pa.created` 2026-11-16T08:36:19Z |
| 2026-11-16 | concept/20261116-02.png | Edit of `concept/20261116-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 08:50:36 UTC; `c2pa.created` 2026-11-16T08:51:53Z |
| 2026-11-16 | work/, export/ | The scripts above |

Prompt of `concept/20261116-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a man in his eighties with medium-brown skin, a bald head and a full white moustache, a charcoal velvet jacket over an ivory shirt, holding a single long-stemmed pale pink peony up beside his shoulder, the whole flower, stem and hand fully visible, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261116-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261116-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a man in his eighties with medium-brown skin, a bald head and a full white moustache, a charcoal velvet jacket over an ivory shirt, holding a single long-stemmed pale pink peony up beside his shoulder, the whole flower, stem and hand fully visible, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261116-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261116-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-16.
