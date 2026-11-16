# table.rosewood_medallion

Static table of set 1 (#174): a rosewood carpet: a rosewood field with a fine lozenge lattice and small flowers at its nodes, a plain octagonal medallion outlined in ink with an ochre second line, ink corner spandrels outlined in ochre with one ivory eight-petal flower each, and an ink border of ivory eight-petal flowers alternating with sage octagonal rosettes between ivory lines, with an ochre zigzag guard and an ochre-dotted guard, inside the wood rim and gold inlay line of every table (`docs/art-direction.md` 5.4). Every layer is computed by one deterministic script; no image generator was used. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: handmade
- Authors: script written and run by an AI agent (Claude Code) on 2026-11-15. No person has done W-3
- Generator: none
- Terms: not applicable, no generator
- Canonical concept: none. The design follows `docs/art-direction.md` sections 1 and 5.4; no image was used as an input
- Inputs uploaded to the generator: none
- Similarity check: not done (RC-3, O-7). The pattern is built from the primitives in the script (lozenge lattice, four- and eight-petal flowers, stepped diamonds, octagonal rosettes, leaves, chevrons)
- Human work: none. RC-5 does not apply to art with no generated content, but the art team should still paint over the rim and the carpet (W-3 step 2 below)
- Self-check (docs/localization.md 6.1 step 1): not done. Notes from the agent for whoever signs: no script, letters, numbers, flags, emblems, religious symbols, stars, casino cues or team hues; the pattern is lozenges, octagons, eight-petal flowers and octagonal rosettes in rosewood, ink, ivory, sage and ochre
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `common` (a scripted pattern with no generated art, below the painted tables; #174 asks for richer art at higher rarity), source `shop:offer.table_rosewood_medallion`. The offer id is a placeholder until the shop record exists (V-15). `table.starter` stays the `common` default table.

## Layers (W-3) and export (W-5)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/table/rosewood_medallion/work/rosewood_medallion_table.py assets-src/cosmetic/table/rosewood_medallion/work/rosewood_medallion_table.json docs/design-tokens.json assets-src/cosmetic/table/rosewood_medallion assets-src/cosmetic/table/rosewood_medallion
```

It printed `rosewood_medallion: knot 4 px, edges depth 192 step 210 start 435 score 0.5393, canvas 1920x1080, seed 17403, static` (run again on 2026-11-16). Two more runs into two empty folders wrote files byte-identical to each other and to the committed ones.

The script reads its settings from `work/rosewood_medallion_table.json`; every table of this set uses the same script with its own settings file. Wood and gold come only from `docs/design-tokens.json`, so the rim and the inlay line are those of `table.starter` and the set-1 tables at every canvas size (`docs/art-direction.md` 5.4: the same layout on every table).

| Layer | File | How it is made |
|---|---|---|
| back | `work/rosewood_medallion_back.png` | The carpet, 1920 x 1080, drawn on a grid of 4 px knots (480 x 270) from the pattern in `work/rosewood_medallion_table.json`: a rosewood carpet: a rosewood field with a fine lozenge lattice and small flowers at its nodes, a plain octagonal medallion outlined in ink with an ochre second line, ink corner spandrels outlined in ochre with one ivory eight-petal flower each, and an ink border of ivory eight-petal flowers alternating with sage octagonal rosettes between ivory lines, with an ochre zigzag guard and an ochre-dotted guard. Wool as in `table.starter`: per-knot shade jitter of 3.5%, a slow colour drift (abrash) of 3%, darker knot edges, a 0.6 px blur and per-pixel fibre noise from the seed 17403. Light: one lamp above the table, brighter in the centre and up to 30% darker toward the rim (`docs/art-direction.md` section 4). The carpet's own palette (rosewood, deep rosewood, ink, ivory, sage and ochre) is in the settings file (art-direction section 2, cosmetic illustration art) |
| middle | `work/rosewood_medallion_middle.png` | The gold inlay line of `table.crimson_carpet`: a 4 px line 54 px inside the edge with a small lozenge at each corner (at the 1920 x 1080 scale), shaded with light from above from `color.gold.lo` through `color.gold.mid` and `color.gold.default` to `color.gold.hi`, drawn at 2x and box-filtered down |
| front | `work/rosewood_medallion_front.png` | The wood rim of `table.crimson_carpet`: a 40 px band with a 28 px corner radius (at the 1920 x 1080 scale), shaded from `color.wood.lo` to `color.wood.hi`, drawn at 2x and box-filtered down |
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
python -I assets-src/cosmetic/table/rosewood_medallion/work/rosewood_medallion_check.py assets-src/cosmetic/table/rosewood_medallion assets-src/tile/ivory/export/face assets-src/effect/tile_shadow/export/static.png docs/design-tokens.json <out>
```

The check measures the static. Centre: the 2v2 landscape board region of `docs/tiles.md` section 3 (1516 x 521 of the 1080-unit canvas), centred on the table. Ivory: the median relative luminance (WCAG) of the opaque pixels of the 28 `tile.ivory` faces, box-filtered to 32 x 64 px, the smallest display size of `docs/tiles.md` section 3; this gives 0.743, lower and so stricter than the 0.765 of the full-size face. Pass rule of this set: ivory against the median of the centre at least 7:1, and against its 90th percentile (the light motifs) at least 4.5:1. The same measure gives `table.starter` 9.2:1 and 8.3:1.

Output on 2026-11-16: `ivory face 0.743; worst scene static: centre median 0.024, 90th percentile 0.029; contrast 10.7:1 against the median (needs 7.0:1), 10.0:1 against the 90th percentile (needs 4.5:1): pass`

The script also writes `check.png`: the static scaled to the 828 px short side of the iPhone XR with a chain of nine `tile.ivory` faces at 32 x 64 px and their runtime shadow (`effect.tile_shadow`, `color.shadow` at `elevation.e1`) in the centre, and `check_x3.png`, its centre enlarged three times. The agent looked at both: pips and faces read on the carpet, and no motif in the centre competes with them.

Team hues (`docs/art-direction.md` section 2): share of static pixels with saturation and value of at least 0.5 within 15 degrees of hue of `color.role.teamA` and `color.role.teamB`: teamA 0.00%, teamB 1.11%. The same measure gives `table.crimson_carpet` teamB 17.42%, `table.indigo_carpet` 8.47% and `table.starter` 0.20% (teamA 0.00% on all three); the teamB share counts ochre, saffron and copper, which sit next to the team-B orange on the hue circle. Not a pass rule; the board review of section 9 decides.

`npm run asset:check -- assets-src/cosmetic/table/rosewood_medallion`: 1 sprites, ASTC 6x6 estimate 921,600 bytes, no size limit set for this type and kind (docs/asset-workflow.md O-5)

Not done: W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Any screen shape (2026-11-16)

Owner decision of 2026-11-16, after the Art Bench check of #724 on a 20:9 phone: the rim, the inlay and the carpet border must show whole on every screen shape (16:9 to 21:9, 4:3, nearly square). Done on 2026-11-16 by Claude Code in `work/rosewood_medallion_table.py` (`periodic_edges`, `write_slice`), with no new canvas and no new layer; the rule is `docs/cosmetics.md` 8.4 and `docs/art-direction.md` 5.4.

- Ring depth 192 px at the 1920 x 1080 scale: the rim, the inlay, the carpet border and its guards, the same on all four sides. The octagon medallion and the dark corners of the field stay in the field; the corners may be cropped on a shape other than 16:9.
- The top and bottom borders now repeat. Each band is split into its light (a Gaussian blur of 120 px along the edge, made symmetric left to right) and its detail. The script tries steps of 150 to 480 px in steps of 6 (so a 1280 x 720 canvas keeps whole pixels) and corner pieces of at least the depth plus 30 px and at most 520 px less a quarter step, scores each by how well the art matches at both ends of one step (windows of 12 and 48 px), how mirror-symmetric it is at the start, and how far the repeated band drifts from the art, and keeps the lowest score. Chosen: step 210 px, corner pieces 0 to 435 px and 1485 to 1920 px, score 0.5393. Between the corner pieces both bands are the run from 435 to 645 px repeated 5 times, joined by 32 px cross-fades, and they fade into the untouched field over the 16 px inside the ring. The corner pieces and the side bands keep the art as it was; a motif unique to the middle of the top or bottom edge is replaced by the run.
- `export/slice.json`, `{"depth": 192, "step": 210}`, records the depth and the step in pixels of the static; `CosmeticContent` sets them as the sprite border of the static and of every layer frame, and `TableImage` draws the table at the screen's shape. Drawn by the same rule as `TableCloth` (scratch render, not committed): 16:9 steps +0 x1.000; 20:9 steps +2 x1.026; 21:9 steps +3 x0.988; 4:3 steps -2 x0.960 (steps gained or lost in the middle of the long edges, then the horizontal scale of the whole table).
- Readability check of #708 run again on the new static (section above): pass. `npm run asset:check` gives the same sprite count and estimate as before.
- Not done: the look of the repeated edges is not reviewed by a person; the device check on the Art Bench comes with the bench build of this change.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| work/rosewood_medallion_table.py | `f194b3c59457057f156fd4c8a77a477594a7c9d29783f205825c38e023cf07ba` | W-3 and W-5 script | Written by Claude Code, 2026-11-15, from `starter_carpet.py` and `crimson_carpet_cleanup.py` of set 1; the same script for every table of this set. `periodic_edges` and `write_slice` added on 2026-11-16 (section Any screen shape) |
| work/rosewood_medallion_table.json | `ff292c43e3bf881732e7f454f595fb246034a56808c89dad663adad1ca695de8` | The item's settings: canvas, seed, pattern and palette | Written by Claude Code, 2026-11-15 |
| work/rosewood_medallion_check.py | `34c3c89c3df7b87a1711fbe2451b7109eb1da623991e79755bfe7953d61b5421` | The readability check | Written by Claude Code, 2026-11-15; the same script for every table of this set |
| work/rosewood_medallion_back.png | `a78faa1aa40452508eb7004ba3bbfc20cbdbed8055c7b553ccd17d4ec28ec89f` | 1920 x 1080 RGBA, back layer | `rosewood_medallion_table.py` |
| work/rosewood_medallion_middle.png | `1d1dbaf8ca4b737007b8a911d7f1b77afc24b2939e243d44a683be0518d41b6d` | 1920 x 1080 RGBA, middle layer, gold inlay | `rosewood_medallion_table.py` |
| work/rosewood_medallion_front.png | `54646a5f296f37b72713f58da1cfbc969456c26e712620bb46450f0f0c266e11` | 1920 x 1080 RGBA, front layer, wood rim | `rosewood_medallion_table.py` |
| export/static.png | `b56131d1aaafed715e3ba0cc6df5ef4e031aa7b65339c49b7285ffb81689fb5b` | 1920 x 1080 RGBA static, flattened layers | `rosewood_medallion_table.py` |
| export/slice.json | `6308037bee79bef3ec697afff48c5fabd3f4f71e0ec61c6c33059f1d91943f73` | `{"depth": 192, "step": 210}`: the ring depth and the edge step in pixels of the static (`docs/cosmetics.md` 8.4) | `rosewood_medallion_table.py`, 2026-11-16 |
