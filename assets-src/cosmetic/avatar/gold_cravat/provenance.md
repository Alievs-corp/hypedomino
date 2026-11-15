# avatar.gold_cravat

Static avatar of set 1 (#172), batch A, a painted portrait made with template P-3 of `docs/asset-workflow.md` section 2. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted
- Authors: W-1 concept generated on 2026-11-15 by an AI agent (Claude Code) running Codex CLI on the project owner's computer. W-3 clean-up and W-5 export by the same agent with `work/gold_cravat_cleanup.py`. No person has done W-3
- Generator: Codex CLI 0.169.0 built-in image generation (`$imagegen`), model `gpt-image-2` per OpenAI docs (learn.chatgpt.com/docs/image-generation); the C2PA manifest in the PNG names `ChatGPT` / `gpt-image` (below). Account holder: the project owner's ChatGPT account, plan Plus (read from the Codex login on 2026-11-15)
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open for the terms, O-2). The account holder reads the terms in force on 2026-11-15 for the Plus plan and records the date here
- Canonical concept: concept/20261115-01.png
- Inputs uploaded to the generator: none; one prompt, one image, no edits, in a new Codex session (`01a1dfee-935e-7970-856f-58745dcccba8`)
- Prompt: P-3 with the style key of section 2.2 pasted verbatim, quoted below
- Concept against the prompt: Matches the prompt. Not asked for but present: gold scroll embroidery on the waistcoat; the white sleeves are full and puffed.
- Similarity check: not done (RC-3 open, O-7)
- Real-person check (RC-4): not done by a person. The face is semi-realistic, so `docs/art-direction.md` 5.1 makes this check matter; the reviewer confirms it resembles no real person
- Human work: none. Clean-up and export are scripted image operations run by an AI agent, so RC-5 stays open; the master and the script are in `work/`
- Self-check (docs/localization.md 6.1 step 1): not done. The asset author signs it, and W-3 has no human author yet. Notes from the agent for whoever signs: no text, letters, numbers, flags, emblems, religious symbols, weapons, casino cues or team hues seen in the art; the expression is friendly; the costume cues are salon dress, not a uniform or a symbol
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `rare`, source `shop:offer.avatar_gold_cravat`. The offer id is a placeholder until the shop record exists (V-15).

## Embedded C2PA manifest

`concept/20261115-01.png` carries a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-15; the signature was not validated with a C2PA tool.

| Field | Value |
|---|---|
| Action | `c2pa.created`, 2026-11-15T15:59:24Z |
| Software agent | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC |

The prompt was sent at 15:58:36 UTC on 2026-11-15 (agent clock), shortly before `c2pa.created`. Codex saved the image under `~/.codex/generated_images/01a1dfee-935e-7970-856f-58745dcccba8/` and copied it into `concept/`; the two files are byte-identical (checked with `cmp`).

## Clean-up (W-3) and export (W-5)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/avatar/gold_cravat/work/gold_cravat_cleanup.py gold_cravat assets-src/cosmetic/avatar/gold_cravat/concept/20261115-01.png assets-src/cosmetic/avatar/gold_cravat/work assets-src/cosmetic/avatar/gold_cravat/export
```

It printed background `[202.0, 202.0, 202.0]`, pockets 101 px, crop `(84, 0, 1162, 1078)`, canvas 512. Two more runs into two empty folders wrote files byte-identical to each other and to the committed ones.

The script is `court_poet_cleanup.py` of set 1 with one added step: grey background that the 3 px opening cuts off from the outside (between a loose strand and the neck, for example) is a pocket too, so the hole filling does not paint it back as subject. Without it, grey patches beside the neck stayed opaque on `avatar.amethyst_courtier`.

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Background, trim, class canvas and pivot | Yes, canvas open | The flat grey background is the median of a 6 px border strip. Pixels within 30 of it (RGB distance) that connect to the top, left or right edge are background, opened by 3 px; enclosed pockets (within 10 of the background colour, at least 20 px, at most 80 px inside the subject's outline) and background cut off by the opening are background too; the largest subject part is kept, its other holes filled, eroded by 2 px and feathered (Gaussian sigma 1), and edge pixels are unmixed from the grey. Crop: a square of 0.86 of the shorter side, centred on the head (the top 35% of the subject), 7% headroom, clamped to the image. Canvas 512 x 512 with a circular alpha (1.5 px soft edge), because the portrait is shown in a circle (`docs/art-direction.md` 5.1); the class canvas is open (O-6) |
| 2. Paint-over, AI artefacts | No | No paint-over. Hair and fur edges keep a thin light rim where the painted rim light and the grey background meet, and fine loose strands keep some grey; an artist cleans the edge |
| 3. Moving parts on their own layers | Not needed | Static item |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow, rarity colour or team hue seen |
| 5. Silhouette and smallest-size tests | In part | The agent looked at the static on a dark backdrop and at 256 px and 64 px; face and expression read. The smallest display size and frame band of an avatar are open (#168, O-6), so the test is not final |

`npm run asset:check -- assets-src/cosmetic/avatar/gold_cravat`: 1 sprite, ASTC 6x6 estimate 118,336 bytes, no limit set for a static avatar (O-5).

Not done: W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261115-01.png | `a8678eb630684ab9b2246085f13084b2cbe20b6f4e32b52a87e816b1fc68ef5d` | 1254 x 1254 RGB, the generator's output, unchanged | Codex CLI `$imagegen`, copied byte for byte by Codex into `concept/` |
| work/gold_cravat_cleanup.py | `3492a2b5d061b1c2729a06914305614c4a252a7c2579696bd081c98668dc9e18` | W-3 and W-5 script | Written by Claude Code, 2026-11-15, from `court_poet_cleanup.py` of set 1 |
| work/gold_cravat_master.png | `3bf2a872a255beb2c563dce3d70aaa86b34c8aa06d4876f7531005ef682d9c21` | 1254 x 1254 RGBA, the cut-out master | `gold_cravat_cleanup.py` |
| export/static.png | `24bbade011bf19354d3a8d6a4b0fce75c7c95d1aa762e35d4f6f85bb4ff862b7` | 512 x 512 RGBA static, circular alpha | `gold_cravat_cleanup.py` |

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-15 | concept/20261115-01.png | P-3 with the style key, quoted below. Sent 15:58:36 UTC by Claude Code, Codex session `01a1dfee-935e-7970-856f-58745dcccba8`; `c2pa.created` 2026-11-15T15:59:24Z |
| 2026-11-15 | work/gold_cravat_master.png, export/static.png | `gold_cravat_cleanup.py` on the concept |

Prompt of `concept/20261115-01.png`, verbatim (the Codex call wraps it in `$imagegen <prompt> Generate exactly one image. Save the final PNG in the current working directory as 20261115-01.png. Do not use any API key and do not edit the image.`):

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Painted splash-art head and shoulders portrait of a man in his forties with dark brown skin, a shaved head, a neat salt-and-pepper goatee, a charcoal velvet waistcoat over a crisp white shirt with a gold silk cravat and a gold watch chain, face centred, looking slightly toward the viewer, friendly expression, rich costume detail. Keep the outer band of the image empty of anything essential.
```

RC-2 against these prompts and the Inputs line: they name no artist, studio, game, brand, mascot, logo or real person, and no image was uploaded. Checked by Claude Code on 2026-11-15.
