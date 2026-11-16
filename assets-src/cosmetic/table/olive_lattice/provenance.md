# table.olive_lattice

Static table of set 1 (#174): an olive carpet: an olive field with a low-contrast lozenge lattice of small four-petal flowers and a plain central lozenge medallion, inside an umber band of alternating ochre-and-madder stepped diamonds and cream leaves between cream lines, with madder guards dotted in umber, inside the wood rim and gold inlay line of every table (`docs/art-direction.md` 5.4). Every layer is computed by one deterministic script; no image generator was used. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: handmade
- Authors: script written and run by an AI agent (Claude Code) on 2026-11-15. No person has done W-3
- Generator: none
- Terms: not applicable, no generator
- Canonical concept: none. The design follows `docs/art-direction.md` sections 1 and 5.4; no image was used as an input
- Inputs uploaded to the generator: none
- Similarity check: not done (RC-3, O-7). The pattern is built from the primitives in the script (lozenge lattice, four- and eight-petal flowers, stepped diamonds, octagonal rosettes, leaves, chevrons)
- Human work: none. RC-5 does not apply to art with no generated content, but the art team should still paint over the rim and the carpet (W-3 step 2 below)
- Self-check (docs/localization.md 6.1 step 1): not done. Notes from the agent for whoever signs: no script, letters, numbers, flags, emblems, religious symbols, stars, casino cues or team hues; the pattern is lozenges, four-petal flowers, stepped diamonds and leaves in olive, umber, madder, cream and ochre
- Cultural review: not done. W-7 asks for it by another account than the author's (docs/localization.md 6.1)

Proposed for the admin draft, not decided: rarity `common` (a scripted pattern with no generated art, below the painted tables; #174 asks for richer art at higher rarity), source `shop:offer.table_olive_lattice`. The offer id is a placeholder until the shop record exists (V-15). `table.starter` stays the `common` default table.

## Layers (W-3) and export (W-5)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/table/olive_lattice/work/olive_lattice_table.py assets-src/cosmetic/table/olive_lattice/work/olive_lattice_table.json docs/design-tokens.json assets-src/cosmetic/table/olive_lattice assets-src/cosmetic/table/olive_lattice
```

It printed `olive_lattice: knot 4 px, canvas 1920x1080, seed 17402, static`. Two more runs into two empty folders wrote files byte-identical to each other and to the committed ones.

The script reads its settings from `work/olive_lattice_table.json`; every table of this set uses the same script with its own settings file. Wood and gold come only from `docs/design-tokens.json`, so the rim and the inlay line are those of `table.starter` and the set-1 tables at every canvas size (`docs/art-direction.md` 5.4: the same layout on every table).

| Layer | File | How it is made |
|---|---|---|
| back | `work/olive_lattice_back.png` | The carpet, 1920 x 1080, drawn on a grid of 4 px knots (480 x 270) from the pattern in `work/olive_lattice_table.json`: an olive carpet: an olive field with a low-contrast lozenge lattice of small four-petal flowers and a plain central lozenge medallion, inside an umber band of alternating ochre-and-madder stepped diamonds and cream leaves between cream lines, with madder guards dotted in umber. Wool as in `table.starter`: per-knot shade jitter of 3.5%, a slow colour drift (abrash) of 3%, darker knot edges, a 0.6 px blur and per-pixel fibre noise from the seed 17402. Light: one lamp above the table, brighter in the centre and up to 30% darker toward the rim (`docs/art-direction.md` section 4). The carpet's own palette (olive, deep olive, umber, madder, cream and ochre) is in the settings file (art-direction section 2, cosmetic illustration art) |
| middle | `work/olive_lattice_middle.png` | The gold inlay line of `table.crimson_carpet`: a 4 px line 54 px inside the edge with a small lozenge at each corner (at the 1920 x 1080 scale), shaded with light from above from `color.gold.lo` through `color.gold.mid` and `color.gold.default` to `color.gold.hi`, drawn at 2x and box-filtered down |
| front | `work/olive_lattice_front.png` | The wood rim of `table.crimson_carpet`: a 40 px band with a 28 px corner radius (at the 1920 x 1080 scale), shaded from `color.wood.lo` to `color.wood.hi`, drawn at 2x and box-filtered down |
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
python -I assets-src/cosmetic/table/olive_lattice/work/olive_lattice_check.py assets-src/cosmetic/table/olive_lattice assets-src/tile/ivory/export/face assets-src/effect/tile_shadow/export/static.png docs/design-tokens.json <out>
```

The check measures the static. Centre: the 2v2 landscape board region of `docs/tiles.md` section 3 (1516 x 521 of the 1080-unit canvas), centred on the table. Ivory: the median relative luminance (WCAG) of the opaque pixels of the 28 `tile.ivory` faces, box-filtered to 32 x 64 px, the smallest display size of `docs/tiles.md` section 3; this gives 0.743, lower and so stricter than the 0.765 of the full-size face. Pass rule of this set: ivory against the median of the centre at least 7:1, and against its 90th percentile (the light motifs) at least 4.5:1. The same measure gives `table.starter` 9.2:1 and 8.3:1.

Output on 2026-11-15: `ivory face 0.743; worst scene static: centre median 0.032, 90th percentile 0.038; contrast 9.7:1 against the median (needs 7.0:1), 9.0:1 against the 90th percentile (needs 4.5:1): pass`

The script also writes `check.png`: the static scaled to the 828 px short side of the iPhone XR with a chain of nine `tile.ivory` faces at 32 x 64 px and their runtime shadow (`effect.tile_shadow`, `color.shadow` at `elevation.e1`) in the centre, and `check_x3.png`, its centre enlarged three times. The agent looked at both: pips and faces read on the carpet, and no motif in the centre competes with them.

Team hues (`docs/art-direction.md` section 2): share of static pixels with saturation and value of at least 0.5 within 15 degrees of hue of `color.role.teamA` and `color.role.teamB`: teamA 0.00%, teamB 0.38%. The same measure gives `table.crimson_carpet` teamB 17.42%, `table.indigo_carpet` 8.47% and `table.starter` 0.20% (teamA 0.00% on all three); the teamB share counts ochre, saffron and copper, which sit next to the team-B orange on the hue circle. Not a pass rule; the board review of section 9 decides.

`npm run asset:check -- assets-src/cosmetic/table/olive_lattice`: 1 sprites, ASTC 6x6 estimate 921,600 bytes, no size limit set for this type and kind (docs/asset-workflow.md O-5)

Not done: W-6. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| work/olive_lattice_table.py | `82dda447dd99dcf47e07a049f114a8f169e87e96555766d3eda1d4bf88a4f48a` | W-3 and W-5 script | Written by Claude Code, 2026-11-15, from `starter_carpet.py` and `crimson_carpet_cleanup.py` of set 1; the same script for every table of this set |
| work/olive_lattice_table.json | `8ea536013ad2aa6e39c38eb737b0d2caeaeefc31593bd9649f69337903cd5d7d` | The item's settings: canvas, seed, pattern and palette | Written by Claude Code, 2026-11-15 |
| work/olive_lattice_check.py | `34c3c89c3df7b87a1711fbe2451b7109eb1da623991e79755bfe7953d61b5421` | The readability check | Written by Claude Code, 2026-11-15; the same script for every table of this set |
| work/olive_lattice_back.png | `1620c1cead17f093d1badb79a681d54e0d6d7d68b2ff0f5e0cb5de57dfec26f8` | 1920 x 1080 RGBA, back layer | `olive_lattice_table.py` |
| work/olive_lattice_middle.png | `1d1dbaf8ca4b737007b8a911d7f1b77afc24b2939e243d44a683be0518d41b6d` | 1920 x 1080 RGBA, middle layer, gold inlay | `olive_lattice_table.py` |
| work/olive_lattice_front.png | `54646a5f296f37b72713f58da1cfbc969456c26e712620bb46450f0f0c266e11` | 1920 x 1080 RGBA, front layer, wood rim | `olive_lattice_table.py` |
| export/static.png | `ce6a6ddb419ffa821a6e23d8e106eb9f819359bf877752fd1da216029df5df6b` | 1920 x 1080 RGBA static, flattened layers | `olive_lattice_table.py` |
