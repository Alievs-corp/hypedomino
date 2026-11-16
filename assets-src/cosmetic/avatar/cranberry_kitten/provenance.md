# avatar.cranberry_kitten

Animated avatar of set 1 (#171), batch C, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2 and animated offline (decision D-3 of the Epic 15 plan): the cleaned master split into parts on layers in `work/`, moved by a deterministic script into the sprite-sheet loop of ADR 0012. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-16 (UTC) while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept and the closed-eyes edit generated in the project owner's ChatGPT account on 2026-11-16 by an AI agent (Claude Code) through Codex CLI. W-3 clean-up, W-4 animation and W-5 export by the same agent with the scripts in `work/`, final runs on 2026-11-16. No person has done W-3 or W-4
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model gpt-image-2 per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifests in the PNGs name ChatGPT / gpt-image (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-16 for the Plus plan and records the date here
- Canonical concept: concept/20261116-01.png
- Inputs uploaded to the generator: `concept/20261116-01.png`, our own canonical concept of this item, attached with `codex exec -i` to the one edit that made `concept/20261116-02.png` (closed eyes for the blink). Nothing else was uploaded
- Prompt: P-3 with the style key of section 2.2 pasted verbatim; `{identity}` = "a woman in her eighties with deep brown skin, short white curls, small gold earrings, a cream and cranberry striped silk gown with a high lace collar, and a small ginger kitten sitting on her shoulder close to her neck with its tail hanging down, the whole kitten fully visible"; quoted below with the edit prompt
- Concept against the prompt: Matches the prompt: the ginger kitten sits on her left shoulder (the viewer's right), close to her cheek, with its tail hanging down over the sleeve. Not asked for: a ruby brooch on a lace jabot, pearl buttons and a cranberry shawl with gold embroidery across the lower edge.
- Blink edit against the canonical: Pose, framing, costume and colours kept; the alignment fitted a shift of 0.11, -0.31 px, scale 1.0001 and rotation -0.0001 rad (mean squared grey difference outside the eyes 0.00159 before, 0.00151 after), and the eye patch was offset by [-8.2, -2.4, -1.0] in RGB to meet the canonical skin. The changes outside the eye mask are small and not used. Used: only the two eye ellipses.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up, animation and export are scripted image operations run by an AI agent, so RC-5 stays open; the master, the layers and the scripts are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 and W-4 have no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; no glow or drop shadow painted in; the expression is friendly
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `legendary`, source `shop:offer.avatar_cranberry_kitten`. The offer id is a placeholder until the shop record exists (V-15). A living companion moving on its own on two layers: the kitten turns its head and swishes its tail. Animated avatars stay at `epic` and above as in batches A and B; batch C proposes 5 epic, 3 legendary, 2 mythic, which brings the 30 animated avatars of set 1 to 15 epic, 11 legendary and 4 mythic.

2026-11-16: owner decision — an animated item's rarity follows its animated detail (`docs/rarity.md` 1.1): 2 animated details (kitten head and tail (a creature counts 2)), so `legendary`, was `mythic`. Any tier split stated above is superseded.

## Embedded C2PA manifests

Both concept files carry a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-16; the signatures were not validated with a C2PA tool.

| Field | `concept/20261116-01.png` | `concept/20261116-02.png` |
|---|---|---|
| Action | `c2pa.created`, 2026-11-16T08:26:33Z | `c2pa.created`, 2026-11-16T08:43:00Z |
| Software agent | `ChatGPT`, version `gpt-image` | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC | OpenAI OpCo, LLC |
| Prompt sent (UTC, agent clock) | 2026-11-16 08:25:38 | 2026-11-16 08:42:15 |
| Codex session | `01a1e376-3c51-7502-ac0e-3f521c9b6ac6` | `01a1e385-704e-7fc3-a5f2-4b3457a87dbb` |

Codex saved each image under `~/.codex/generated_images/<session>/` and copied it into `concept/`; each pair of files is byte-identical (checked by a byte comparison). No generation for this item was rejected: these are the only two images generated for it.

## Clean-up (W-3)

Run on 2026-11-16 (UTC) by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/cranberry_kitten/work/cranberry_kitten_cleanup.py cranberry_kitten assets-src/cosmetic/avatar/cranberry_kitten/concept/20261116-01.png assets-src/cosmetic/avatar/cranberry_kitten/work
```

It printed background [184.0, 184.0, 183.0], pockets 2180 px, crop (32, 0, 1235, 1203), portrait 768. The script is `cornflower_ribbon_cleanup.py` of batch B with one change: `CROP_SHARE` is 0.96 instead of 0.86, so the square crop takes more of the concept. With 0.86 most of the hanging tail lay outside the circle; with 0.96 the tail lies inside it down to its last few pixels, and the face stays centred. It writes the master and a square 768 x 768 portrait crop (`work/cranberry_kitten_portrait.png`, no circle); the enclosed grey pockets, the edge re-matte and the fringe removal work as described in the batch A provenance of `avatar.ruby_drop`.

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
python -I assets-src/cosmetic/avatar/cranberry_kitten/work/cranberry_kitten_loop.py assets-src/cosmetic/avatar/cranberry_kitten/concept assets-src/cosmetic/avatar/cranberry_kitten/work assets-src/cosmetic/avatar/cranberry_kitten/export
```

It printed `cranberry_kitten_loop: blink edit registered (tx, ty, scale, angle) = (0.1113, -0.3139, 1.0001, -0.0001), residual 0.00159 -> 0.00151, eye colour offset [-8.2, -2.4, -1.0], parts 2, ripples 0, filled 24024 px`. It writes the layers below into `work/`, `export/anim/0001.png` to `0016.png` and `export/static.png`: RGBA PNG, 8 bit, straight alpha. A second and a third run of the clean-up and the loop into empty folders wrote files byte-identical to each other and to the committed ones.

The script is the batch B loop script (`cornflower_ribbon_loop.py`, which carries the batch A rig of `ruby_drop_loop.py` with `flap`, `RIPPLES`, the clearing hole fill, clipped glints and the clipped registration region) with this item's geometry and one addition, the same for every avatar of batch C: a part can spin (`spin`, whole turns per loop added to its turn angle, so a full turn ends where it started). Parts without `spin` move exactly as in batch B.

| Choice | Value | Source |
|---|---|---|
| Canvas | 192 x 192 | D-1 of the Epic 15 plan and the ADR 0012 illustration (16 frames of 192 x 192 in a 768 atlas). Every frame is rendered at 768 x 768 in premultiplied alpha from the portrait crop and box-downsampled by 4 |
| Frames | 16 | D-1, ADR 0012 first measurement |
| Frame rate | 12 fps, a loop of 1.33 s | D-1; within the 1 to 2 s of #171. The rate is not stored with the loop (O-5, O-6) |
| Blink | Frames 5 and 6 (two frames, 167 ms) show the closed eyes, every other frame the open eyes | The closed eyes come from the Codex edit, aligned to the canonical by a similarity transform fitted by least squares outside the eyes and blended in through two feathered ellipses (Gaussian sigma 5 at 768) with a colour offset measured in a ring around them |
| Breathing | The shoulders and chest rise by up to 8 px at 768 (2 px at 192) with (1 - cos)/2 over the loop, through a smooth vertical mesh warp that is 1 at the shoulder line, falls to 0 at the bottom edge and carries the head at 0.4; the chest widens by 1.2% at the same time | Warp only, so no seams |
| Head | A tilt of up to 1.6 degrees about the neck with sin over the loop, and a nod of 2 px at 768 twice per loop, weighted by a soft ellipse around head and hair that fades out at the neck | Warp only |
| Ambient part | `kitten_head` (layer `work/cranberry_kitten_kitten_head.png`, pivot (556, 286)) turns about its pivot by 7.0 degrees x (sin(lag) - sin(t + lag) + 0.5 sin(2t)), lag 0.0, the turn growing from 0 at the pivot to full at 50 px (768 scale), so the end near the pivot stays attached and the far end swings; the part also follows the head or torso warp at its pivot; the hole it leaves in the base layer is filled from the surrounding pixels, and where that estimate is less than half covered it is cleared instead, so no smear shows over the background | The head hangs over her hair and the background; the turn grows from the kitten's neck, so the head tilts while the body stays on the shoulder |
| Ambient part | `kitten_tail` (layer `work/cranberry_kitten_kitten_tail.png`, pivot (650, 414)) turns about its pivot by 7.0 degrees x (sin(lag) - sin(t + lag) + 0.4 sin(2t)), lag 1.0, the turn growing from 0 at the pivot to full at 90 px (768 scale), so the end near the pivot stays attached and the far end swings; the part also follows the head or torso warp at its pivot; the hole it leaves in the base layer is filled from the surrounding pixels, and where that estimate is less than half covered it is cleared instead, so no smear shows over the background | The tail hangs over the sleeve; the turn grows from the tail root, so it swishes from the body. Its tip touches the circle edge at the widest swing |
| Seamless loop | Mean channel change from frame 16 to frame 1 is 3.14 of 255, against 3.08 on average and 3.67 at most between the other neighbouring frames | Every motion is periodic in the 16 frames |
| Static | Frame 9 | The rest frame: eyes open, no breath, no tilt or nod, every moving part at its painted position. `export/static.png` is byte-identical to `export/anim/0009.png` |
| Size | 17 sprites, 278,528 bytes | `npm run asset:check`, ASTC 6x6 estimate (17 x 16,384), under the 300,000 bytes of #171 |

Not done: the reduce-motion and low-level check of 5.2 (Unity), the board panels B-8 and B-9, the bundle measurement of ADR 0012, and W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Re-crop, 2026-11-16

Owner feedback from the phone on 2026-11-16: inside the frames, the frame's top ornament band covered the top of the head and hair. The owner asked for the portrait "very slightly smaller and a bit lower"; `docs/art-direction.md` 5.1 keeps anything essential out of the outer band the frame covers. Only the crop framing changed. The cut-out master was re-made by the same run and is byte-identical; masks, pockets and the edge re-matte are unchanged.

- `CROP_SHARE` 0.96 to 1.0 (0.96 / 0.92 = 1.043 is wider than the concept, so it is clamped to 1.0, the shorter side of the concept; the subject is 4.0% smaller in the circle instead of 8%)
- `HEADROOM` 0.07 to 0.12, not the 0.17 of the other avatars: at 0.17 the kitten's tail layer was 46.1% outside the circle (0.1% before); at 0.12 it is 26.2% outside, because the cat sits at the bottom right of a square concept. The tail end is cut by the circle; open for the owner
- The crop may now reach above the top of the concept: `upper = min(top - HEADROOM * side, height - side)` instead of clipping it at 0, the rule `avatar.tawny_quill` already used. The subject starts within a few pixels of the concept's top edge, so a lower head needs transparent canvas above it; the bottom and side clamps are as before
- Printed by the clean-up: `cranberry_kitten: background [184.0, 184.0, 183.0], pockets 2180 px, crop (0, -148, 1254, 1106), portrait 768`
- Crop `(32, 0, 1235, 1203)` to `(0, -148, 1254, 1106)` on the concept
- Scale k = 0.9593. The rig geometry in portrait pixels (`HEAD`, `NECK_Y`, `NECK_BAND`, `SHOULDER_Y`, `BREATH`, `NOD`, `EYES`, `EYE_FEATHER`, the part polygons, pivots, `length` and `grow`, the glints and the ripples) is mapped through the crop change (x' = k x + offset, sizes and pixel amplitudes times k), so every mask stays on the same anatomy and the motion is the same relative to the subject. Angles, timing, blink frames and `STATIC_FRAME` are unchanged
- Printed by the loop: `cranberry_kitten_loop: blink edit registered (tx, ty, scale, angle) = (0.1081, -0.3166, 1.0001, -0.0001), residual 0.00152 -> 0.00144, eye colour offset [-8.3, -2.5, -1.0], parts 2, ripples 0, filled 22175 px`
- Head top (first row of `export/static.png` with alpha above 0.5, as a share of the canvas height): 0.005 before, 0.120 after; highest over the 16 loop frames 0.000 before, 0.115 after

Run on 2026-11-16 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/cranberry_kitten/work/cranberry_kitten_cleanup.py cranberry_kitten assets-src/cosmetic/avatar/cranberry_kitten/concept/20261116-01.png assets-src/cosmetic/avatar/cranberry_kitten/work
python -I assets-src/cosmetic/avatar/cranberry_kitten/work/cranberry_kitten_loop.py assets-src/cosmetic/avatar/cranberry_kitten/concept assets-src/cosmetic/avatar/cranberry_kitten/work assets-src/cosmetic/avatar/cranberry_kitten/export
```

Determinism: two more runs of the same commands into two empty folders wrote the same 26 files byte for byte, and both match the files committed here.

`npm run asset:check -- assets-src/cosmetic/avatar/cranberry_kitten` after the re-crop: 17 sprites, ASTC 6x6 estimate 278,528 bytes, limit under 300,000 (#171, animated avatars under 300 KB each): pass.

The Files table below keeps the hashes of the first export. The files the re-crop changed now hash as follows:

| File | sha256 after the re-crop |
|---|---|
| export/anim/0001.png | `ad3a19a110f826e5a11a96a15f2f1331418569ae7dfd97ef79f1ea91f3ef8960` |
| export/anim/0002.png | `0bb177292860e18c91312d8424f1411eb6ee6ba2e6ea040941e3284685e55ce8` |
| export/anim/0003.png | `7eff19d47ce73f6ca09c6da10c94b6d21b66375af5a28d3afe08fb19c12a590d` |
| export/anim/0004.png | `0a0b64fbde5384e4911a42e911affaf1b0537e983e12e4ad20e66bdb3186b48e` |
| export/anim/0005.png | `d391506abd9b9d636cc64e85d547dca0bd810599be76200b3f6a1417017a04d0` |
| export/anim/0006.png | `fbcd078ba6d16cfeba17d97635d8aa4286ae1b62bc5b019a9402b6ec0cae61ed` |
| export/anim/0007.png | `6aa7415c7aa07bf2ea03626e3fa15c0acb5d2ce8bd6d5d5efa103d6b3288b815` |
| export/anim/0008.png | `c7f71ffa41b09583c49f66ec23fc7d5e8e9bdbf504ab3b0ddfa6d3ac5850632d` |
| export/anim/0009.png | `77acad5d3493f804b2a1e8c761b3a423f72362ec7e3107f22b3efa080d9fa6ba` |
| export/anim/0010.png | `e29fcc6c63dbd96c12abb4c6a66aba4f3b29ba7ca517ee2d9c5530df9b3e90ec` |
| export/anim/0011.png | `7c8240236e02b824204e910e87d82f90aa5743e51d7c1d2d9b48c3ed6bec0c57` |
| export/anim/0012.png | `1697236b5a6cf0f56513cebb6010ebac368dcbe9c46805c56c06f98590907427` |
| export/anim/0013.png | `977dced67cff2c1e01459c9e01ac741f71302e4b947fb3a4d5bef8225d9505ca` |
| export/anim/0014.png | `f44460578acb2c94a2943b9bb4c6a5c89451da052bba0085ed943bddf696a52d` |
| export/anim/0015.png | `d8db4a43ad1b962c30744f4201863cb8598af06444eb2974c04b035daa6096c8` |
| export/anim/0016.png | `d3f07b3d21b96c2deb90823d4e9f3d56f29b85df4e16ce2d46830858003713a0` |
| export/static.png | `77acad5d3493f804b2a1e8c761b3a423f72362ec7e3107f22b3efa080d9fa6ba` |
| work/cranberry_kitten_base.png | `ef13f5f6c82ccafa0837e00a503ecbeee2c4240723669fa52b427141b7363085` |
| work/cranberry_kitten_cleanup.py | `c7158b7b24a7629ccbe1cd3750f6460d24bd57dd4c2802c105e2b978593786f4` |
| work/cranberry_kitten_eyes_closed.png | `b6ba8442b3283ed92a1a332e8df44323e63bea2e30661928b6bdcb5840cf993a` |
| work/cranberry_kitten_eyes_open.png | `5b5e4e9be0d106993d56254e5d75ba907e418e8ebfe9ea688ffc1a1ce1cebad6` |
| work/cranberry_kitten_head.png | `cf105edbc01f3754e2c3fd55fa19723fc9243f67f049c19acd5ac028751fd2cb` |
| work/cranberry_kitten_kitten_head.png | `108ab8e33a7454028acd1dab545fed5dc5ddd58ffbd619cc736d43f63bd678c4` |
| work/cranberry_kitten_kitten_tail.png | `3c86d84be76ca3415e8dccc9698c7a520faee82170588f05c3342fefb68dbd21` |
| work/cranberry_kitten_loop.py | `08aa5d41c50b1985f83c18299f320027b31197462dcfb9bcc797b0e4861c0519` |
| work/cranberry_kitten_portrait.png | `54b79a0e1aa2f7618bdc05965f415ac3e5d3d6c1048e7d53dbfb498e3b06d4eb` |
| work/cranberry_kitten_torso.png | `d55f15f1893108ab5edadd5ba7cd250c496208c7797049f3f2bdf13afce2e6cb` |

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261116-01.png | `a8d72b23c50a71574b3ba7cd12f62d5b5c10eb064182526466f03331c7e6c547` | 1254 x 1254 RGB, the canonical concept, unchanged | Codex CLI `$imagegen`, session `01a1e376-3c51-7502-ac0e-3f521c9b6ac6` |
| concept/20261116-02.png | `7d7c47ff225abf1ddaf158162a5cbb258964dbf351150bda15abd520debd69ec` | 1254 x 1254 RGB, closed-eyes edit of the canonical, unchanged | Codex CLI `$imagegen` with the canonical attached, session `01a1e385-704e-7fc3-a5f2-4b3457a87dbb` |
| work/cranberry_kitten_cleanup.py | `fe4b40559df930b29d309bc229bf30331aba753ba3f792c67667639b2e84089a` | W-3 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_cleanup.py` |
| work/cranberry_kitten_loop.py | `39b044726beafd697b2b1fba20140ad16a6ba13799960c0a1ac4a0c468727247` | W-4 and W-5 script | Written by Claude Code, 2026-11-16, from `cornflower_ribbon_loop.py`; the same script for every avatar of this batch with the item's geometry |
| work/cranberry_kitten_master.png | `cbdd4d585997f5f1c38eac13e7fb49090676994ff1a5437cbfc2f1b09c864924` | 1254 x 1254 RGBA, the cut-out master | `cranberry_kitten_cleanup.py` |
| work/cranberry_kitten_base.png | `de400524c502dbbe53698ba1b67e95e8bbe422b77c8ba67ccc955c79f6f5ec8c` | The portrait with the moving parts taken out (filled or cleared) | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_eyes_closed.png | `8e92bd4417a7cf3630b83edda6eb2723aff0f25aa30ebb7bc07b9c135b658a68` | Eye region, closed (aligned edit) through the same mask | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_eyes_open.png | `0a84d2b66fcd56c26b0898e4a7839fd84b39199908d043d27e71524cfa3074e8` | Eye region, open (canonical concept) through the feathered eye mask | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_head.png | `bad68c4b2a175e51fc0e6344cd0f0aa35c2e4fe4f51b4207f5651e2351b99c5f` | Head and hair: the portrait weighted by the head field of the warp | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_kitten_head.png | `f097e1781b359934bf29bfe0d17f9b569c6fc166362bd85db5e97a59b8578608` | Moving part `kitten_head`: the portrait through its feathered polygon | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_kitten_tail.png | `c2ecc4d0e3cd1830ccc183836912881392a8dbf3a37c2e7eb828e10ba23c246a` | Moving part `kitten_tail`: the portrait through its feathered polygon | `cranberry_kitten_loop.py` |
| work/cranberry_kitten_portrait.png | `011641d31561d5d54f1d05f8d2d52631b16b829e0746fe6bb0307d4a50e5bec8` | 768 x 768 RGBA portrait crop, no circle | `cranberry_kitten_cleanup.py` |
| work/cranberry_kitten_torso.png | `9fa1c0d651dc045e2123e0d047500fb13bbe8630f2d2adc9879fc2cb7dad03f1` | Torso and costume: the portrait weighted by one minus the head field | `cranberry_kitten_loop.py` |
| export/static.png | `472dbcd42e26d27ee3ce807e02efc5ac060249240db16ffbe0d875bbf488cbce` | 192 x 192 RGBA static, loop frame 9 | `cranberry_kitten_loop.py` |
| export/anim/0001.png to 0016.png | per file in git | 192 x 192 RGBA loop frames | `cranberry_kitten_loop.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-16 | concept/20261116-01.png | P-3 with the style key, quoted below. Sent 08:25:38 UTC; `c2pa.created` 2026-11-16T08:26:33Z |
| 2026-11-16 | concept/20261116-02.png | Edit of `concept/20261116-01.png` (attached): eyes closed, everything else identical, quoted below. Sent 08:42:15 UTC; `c2pa.created` 2026-11-16T08:43:00Z |
| 2026-11-16 | work/, export/ | The scripts above |
| 2026-11-16 | export/static.png, work/cranberry_kitten_base.png, work/cranberry_kitten_cleanup.py, work/cranberry_kitten_eyes_closed.png, work/cranberry_kitten_eyes_open.png, work/cranberry_kitten_head.png, work/cranberry_kitten_kitten_head.png, work/cranberry_kitten_kitten_tail.png, work/cranberry_kitten_loop.py, work/cranberry_kitten_portrait.png, work/cranberry_kitten_torso.png, export/anim/0001.png to 0016.png | Re-crop: smaller and lower per owner feedback. `CROP_SHARE` 0.96 to 1.0, `HEADROOM` 0.07 to 0.12, crop `(32, 0, 1235, 1203)` to `(0, -148, 1254, 1106)`; see "Re-crop, 2026-11-16" |

Prompt of `concept/20261116-01.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her eighties with deep brown skin, short white curls, small gold earrings, a cream and cranberry striped silk gown with a high lace collar, and a small ginger kitten sitting on her shoulder close to her neck with its tail hanging down, the whole kitten fully visible, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

Codex was given it inside this instruction: `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261116-01.png. Do not use any API key and do not edit the image.`

Prompt of `concept/20261116-02.png`, verbatim:

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a woman in her eighties with deep brown skin, short white curls, small gold earrings, a cream and cranberry striped silk gown with a high lace collar, and a small ginger kitten sitting on her shoulder close to her neck with its tail hanging down, the whole kitten fully visible, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential. Edit the attached image, which is our own canonical concept of this portrait: change only one thing, the eyes are gently closed as in a relaxed blink. Everything else stays identical: the same face, pose, head position, framing, size, costume, colours, light and plain flat light grey background.
```

Codex was given it with `-i concept/20261116-01.png` inside this instruction: `$imagegen <prompt> Generate exactly one image by editing the attached image. Save the final PNG in the current working directory as 20261116-02.png. Do not use any API key.`

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and the only uploaded image is our own canonical concept of this item. Checked by Claude Code on 2026-11-16.
