# table.slate_flatweave

Static table of set 1 (#174): a slate carpet in a flat-woven style: a slate field with a low-contrast lozenge lattice and small slate diamonds, inside a charcoal band of ochre chevrons between ivory lines, a rust guard with a charcoal zigzag and a slate-deep guard with rust dots, inside the wood rim and gold inlay line of every table (`docs/art-direction.md` 5.4). Every layer is computed by one deterministic script; no image generator was used. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted (since 2026-11-22: `concept/20261122-01.png` came from a generator; the shipped layers and the static are still the script output)
- Authors: script written and run by an AI agent (Claude Code) on 2026-11-15. No person has done W-3. W-1 concept generated on 2026-11-22 by an AI agent (Claude Code) through Codex CLI, in the project owner's ChatGPT account
- Generator: Codex CLI 0.171.0 built-in image generation (`$imagegen`, `docs/asset-workflow.md` 2.5), signed in with the project owner's ChatGPT account (`codex login status`: Logged in using ChatGPT); the plan was not read on 2026-11-22
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open, O-2). The account holder reads the terms in force on 2026-11-22 for the plan and records the date here
- Canonical concept: concept/20261122-01.png, a painted carpet concept for the back layer. Not yet used: W-3 from this concept (trim, fit, calm centre, layers, readability check, as `table.crimson_carpet` does) is not done, so `work/` and `export/` are unchanged
- Inputs uploaded to the generator: none; one prompt, one image, no edits
- Similarity check: not done (RC-3, O-7). The pattern is built from the primitives in the script (lozenge lattice, four- and eight-petal flowers, stepped diamonds, octagonal rosettes, leaves, chevrons)
- Human work: none. RC-5 does not apply to art with no generated content, but the art team should still paint over the rim and the carpet (W-3 step 2 below)
- Self-check (docs/localization.md 6.1 step 1): not done. Notes from the agent for whoever signs: no script, letters, numbers, flags, emblems, religious symbols, stars, casino cues or team hues; the pattern is lozenges, small diamonds and chevrons in slate, charcoal, rust, ivory and ochre
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `common` (a scripted pattern with no generated art, below the painted tables; #174 asks for richer art at higher rarity), source `shop:offer.table_slate_flatweave`. The offer id is a placeholder until the shop record exists (V-15). `table.starter` stays the `common` default table.

## Layers (W-3) and export (W-5)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/table/slate_flatweave/work/slate_flatweave_table.py assets-src/cosmetic/table/slate_flatweave/work/slate_flatweave_table.json docs/design-tokens.json assets-src/cosmetic/table/slate_flatweave assets-src/cosmetic/table/slate_flatweave
```

It printed `slate_flatweave: knot 4 px, edges depth 174 step 384 start 384 score 0.5222, canvas 1920x1080, seed 17401, static` (run again on 2026-11-16). Two more runs into two empty folders wrote files byte-identical to each other and to the committed ones.

The script reads its settings from `work/slate_flatweave_table.json`; every table of this set uses the same script with its own settings file. Wood and gold come only from `docs/design-tokens.json`, so the rim and the inlay line are those of `table.starter` and the set-1 tables at every canvas size (`docs/art-direction.md` 5.4: the same layout on every table).

| Layer | File | How it is made |
|---|---|---|
| back | `work/slate_flatweave_back.png` | The carpet, 1920 x 1080, drawn on a grid of 4 px knots (480 x 270) from the pattern in `work/slate_flatweave_table.json`: a slate carpet in a flat-woven style: a slate field with a low-contrast lozenge lattice and small slate diamonds, inside a charcoal band of ochre chevrons between ivory lines, a rust guard with a charcoal zigzag and a slate-deep guard with rust dots. Wool as in `table.starter`: per-knot shade jitter of 3.5%, a slow colour drift (abrash) of 3%, darker knot edges, a 0.6 px blur and per-pixel fibre noise from the seed 17401. Light: one lamp above the table, brighter in the centre and up to 30% darker toward the rim (`docs/art-direction.md` section 4). The carpet's own palette (slate, deep slate, charcoal, rust, ivory and ochre) is in the settings file (art-direction section 2, cosmetic illustration art) |
| middle | `work/slate_flatweave_middle.png` | The gold inlay line of `table.crimson_carpet`: a 4 px line 54 px inside the edge with a small lozenge at each corner (at the 1920 x 1080 scale), shaded with light from above from `color.gold.lo` through `color.gold.mid` and `color.gold.default` to `color.gold.hi`, drawn at 2x and box-filtered down |
| front | `work/slate_flatweave_front.png` | The wood rim of `table.crimson_carpet`: a 40 px band with a 28 px corner radius (at the 1920 x 1080 scale), shaded from `color.wood.lo` to `color.wood.hi`, drawn at 2x and box-filtered down |
| static | `export/static.png` | back, middle and front composited in that order; 1920 x 1080 RGBA, fully opaque |

The table is static: `docs/cosmetics.md` section 4 forbids `layers` on a static table, so the layers stay in `work/` and only `export/static.png` is exported.

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Class canvas and pivot | Yes, canvas open | 1920 x 1080, the canvas of the set-1 tables; how the table covers 4:3 to 21:9 screens is open (O-6) |
| 2. Paint-over | No | The carpet is flat and graphic, the rim has no grain; an artist paints over the rim, the inlay and the carpet's wool |
| 3. Moving parts on their own layers | Yes | back, middle and front in `work/`; no layer animates in this static item |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow or rarity colour seen; team hues in the readability section |
| 5. Silhouette and smallest-size tests | Tile readability only | Readability check below |

## Readability check (acceptance of #174)

```bash
python -I assets-src/cosmetic/table/slate_flatweave/work/slate_flatweave_check.py assets-src/cosmetic/table/slate_flatweave assets-src/tile/ivory/export/face assets-src/effect/tile_shadow/export/static.png docs/design-tokens.json <out>
```

The check measures the static. Centre: the 2v2 landscape board region of `docs/tiles.md` section 3 (1516 x 521 of the 1080-unit canvas), centred on the table. Ivory: the median relative luminance (WCAG) of the opaque pixels of the 28 `tile.ivory` faces, box-filtered to 32 x 64 px, the smallest display size of `docs/tiles.md` section 3; this gives 0.743, lower and so stricter than the 0.765 of the full-size face. Pass rule of this set: ivory against the median of the centre at least 7:1, and against its 90th percentile (the light motifs) at least 4.5:1. The same measure gives `table.starter` 9.2:1 and 8.3:1.

Output on 2026-11-16: `ivory face 0.743; worst scene static: centre median 0.030, 90th percentile 0.038; contrast 9.8:1 against the median (needs 7.0:1), 9.1:1 against the 90th percentile (needs 4.5:1): pass`

The script also writes `check.png`: the static scaled to the 828 px short side of the iPhone XR with a chain of nine `tile.ivory` faces at 32 x 64 px and their runtime shadow (`effect.tile_shadow`, `color.shadow` at `elevation.e1`) in the centre, and `check_x3.png`, its centre enlarged three times. The agent looked at both: pips and faces read on the carpet, and no motif in the centre competes with them.

Team hues (`docs/art-direction.md` section 2): share of static pixels with saturation and value of at least 0.5 within 15 degrees of hue of `color.role.teamA` and `color.role.teamB`: teamA 0.00%, teamB 0.59%. The same measure gives `table.crimson_carpet` teamB 17.42%, `table.indigo_carpet` 8.47% and `table.starter` 0.20% (teamA 0.00% on all three); the teamB share counts ochre, saffron and copper, which sit next to the team-B orange on the hue circle. Not a pass rule; the board review of section 9 decides.

`npm run asset:check -- assets-src/cosmetic/table/slate_flatweave`: 1 sprites, ASTC 6x6 estimate 921,600 bytes, no size limit set for this type and kind (docs/asset-workflow.md O-5)

Not done: W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Any screen shape (2026-11-16)

Owner decision of 2026-11-16, after the Art Bench check of #724 on a 20:9 phone: the rim, the inlay and the carpet border must show whole on every screen shape (16:9 to 21:9, 4:3, nearly square). Done on 2026-11-16 by Claude Code in `work/slate_flatweave_table.py` (`periodic_edges`, `write_slice`), with no new canvas and no new layer; the rule is `docs/cosmetics.md` 8.4 and `docs/art-direction.md` 5.4.

- Ring depth 174 px at the 1920 x 1080 scale: the rim, the inlay, the carpet border and its guards, the same on all four sides.
- The top and bottom borders now repeat. Each band is split into its light (a Gaussian blur of 120 px along the edge, made symmetric left to right) and its detail. The script tries steps of 150 to 480 px in steps of 6 (so a 1280 x 720 canvas keeps whole pixels) and corner pieces of at least the depth plus 30 px and at most 520 px less a quarter step, scores each by how well the art matches at both ends of one step (windows of 12 and 48 px), how mirror-symmetric it is at the start, and how far the repeated band drifts from the art, and keeps the lowest score. Chosen: step 384 px, corner pieces 0 to 384 px and 1536 to 1920 px, score 0.5222. Between the corner pieces both bands are the run from 384 to 768 px repeated 3 times, joined by 32 px cross-fades, and they fade into the untouched field over the 16 px inside the ring. The corner pieces and the side bands keep the art as it was; a motif unique to the middle of the top or bottom edge is replaced by the run.
- `export/slice.json`, `{"depth": 174, "step": 384}`, records the depth and the step in pixels of the static; `CosmeticContent` sets them as the sprite border of the static and of every layer frame, and `TableImage` draws the table at the screen's shape. Drawn by the same rule as `TableCloth` (scratch render, not committed): 16:9 steps +0 x1.000; 20:9 steps +1 x1.042; 21:9 steps +2 x0.938; 4:3 steps -1 x0.938 (steps gained or lost in the middle of the long edges, then the horizontal scale of the whole table).
- Readability check of #708 run again on the new static (section above): pass. `npm run asset:check` gives the same sprite count and estimate as before.
- Not done: the look of the repeated edges is not reviewed by a person; the device check on the Art Bench comes with the bench build of this change.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| work/slate_flatweave_table.py | `f194b3c59457057f156fd4c8a77a477594a7c9d29783f205825c38e023cf07ba` | W-3 and W-5 script | Written by Claude Code, 2026-11-15, from `starter_carpet.py` and `crimson_carpet_cleanup.py` of set 1; the same script for every table of this set. `periodic_edges` and `write_slice` added on 2026-11-16 (section Any screen shape) |
| work/slate_flatweave_table.json | `0c601fab47110402e93300c2d095e116abe016ae005f299c5e3338065a2615d2` | The item's settings: canvas, seed, pattern and palette | Written by Claude Code, 2026-11-15 |
| work/slate_flatweave_check.py | `34c3c89c3df7b87a1711fbe2451b7109eb1da623991e79755bfe7953d61b5421` | The readability check | Written by Claude Code, 2026-11-15; the same script for every table of this set |
| work/slate_flatweave_back.png | `e8b28e7f315a6d66f8e3e4db7ac51666b566dbf98890249f833a8f38545ee749` | 1920 x 1080 RGBA, back layer | `slate_flatweave_table.py` |
| work/slate_flatweave_middle.png | `1d1dbaf8ca4b737007b8a911d7f1b77afc24b2939e243d44a683be0518d41b6d` | 1920 x 1080 RGBA, middle layer, gold inlay | `slate_flatweave_table.py` |
| work/slate_flatweave_front.png | `54646a5f296f37b72713f58da1cfbc969456c26e712620bb46450f0f0c266e11` | 1920 x 1080 RGBA, front layer, wood rim | `slate_flatweave_table.py` |
| export/static.png | `c26603f9df452eec5ba175c71d81b4602346d458c8c8fdca03bbcfdd7ba45e6f` | 1920 x 1080 RGBA static, flattened layers | `slate_flatweave_table.py` |
| export/slice.json | `2348c30225cd3681940e4d507b2189452934c0bebb397ba54a952561ee760a7c` | `{"depth": 174, "step": 384}`: the ring depth and the edge step in pixels of the static (`docs/cosmetics.md` 8.4) | `slate_flatweave_table.py`, 2026-11-16 |

## Concept (2026-11-22)

Generated by Claude Code with `env -u OPENAI_API_KEY codex exec --skip-git-repo-check --sandbox workspace-write` in `concept/` (`docs/asset-workflow.md` 2.5). Codex session `01a2039a-e748-7172-9528-16831f3192dc`, prompt sent 2026-11-22T14:13:30Z. `concept/20261122-01.png` is a byte copy of the file Codex keeps under `~/.codex/generated_images/01a2039a-e748-7172-9528-16831f3192dc/`. The C2PA manifest was not read.

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-22 | concept/20261122-01.png | P-5 back layer with the style key, quoted below. SHA-256 `d90fbea05794697b14d8976dbb3fdd1658d0396e9e3f7366b984d2c3a7c07cb8` |

Prompt of `concept/20261122-01.png`, verbatim (the instruction to save the file follows it in the Codex call):

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Top-down view of a woven carpet table cloth, a slate flat-woven salon carpet with a low-contrast lozenge lattice of small diamonds, and a charcoal border of chevrons with thin zigzag and dotted guard stripes, back (the whole carpet, field and border, filling the image edge to edge) layer only, original geometric and floral pattern, calm low-contrast centre where game pieces will sit. Palette: slate grey-blue, charcoal, ochre, ivory and rust.
```

Concept against the prompt, looked at by the agent: an edge-to-edge carpet, field and border, with a calm centre; no text, letters, numbers, glow, drop shadow, casino cue or team hue seen. Not checked: generator artefacts at full size, the readability rule of #708 (needs W-3 first).
