# table.starter

The default table (#169): the `common` built-in table of `server/internal/cosmetics/builtin.json`, address `cosmetic/table/starter/static`, which had no art until now. It follows `docs/art-direction.md` 5.4: a red and indigo carpet with an original geometric and floral pattern and a calm centre, inside a wood rim with a gold inlay line. Every layer is computed by one deterministic script; no image generator was used. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: handmade (no image in `concept/` or `work/` came from a generator; `builtin.json` already lists `handmade`)
- Authors: script written and run by an AI agent (Claude Code) on 2026-11-15. No person has done W-3
- Generator: none
- Terms: not applicable, no generator
- Canonical concept: none. The design follows the description of the default table in `docs/art-direction.md` sections 1 and 5.4; the prototype carpet of board panel B-6 was looked at for colour and was not copied, traced or used as an input (its own provenance is not recorded, so it cannot ship)
- Inputs uploaded to the generator: none
- Similarity check: not done (RC-3, O-7). The pattern is built from three primitives in the script: a lozenge lattice, a four-petal flower and a stepped lozenge
- Human work: none. RC-5 does not apply to art with no generated content, but the art team should still paint over the rim and the carpet (see W-3 step 2 below)
- Self-check (docs/localization.md 6.1 step 1): not done. Notes from the agent for whoever signs: no script, letters, numbers, flags, emblems, religious symbols, stars, casino cues or team hues; the pattern is lozenges and four-petal flowers in red, indigo, ivory and ochre
- Cultural review: not done. `builtin.json` says `approved` for this item; that line was written before any art existed, so this art still needs the review of `docs/localization.md` 6.1 before it ships

Proposed for the admin draft, not decided: rarity `common` and source `default`, as `builtin.json` already has them.

## Layers (W-3) and export (W-5)

Run on 2026-11-15 by Claude Code with Python 3.12.10, NumPy 2.4.6, SciPy 1.15.3 and Pillow 12.1.1, from the repository root:

```bash
python -I assets-src/cosmetic/table/starter/work/starter_carpet.py docs/design-tokens.json assets-src/cosmetic/table/starter/work assets-src/cosmetic/table/starter/export
```

A second run into an empty folder wrote byte-identical files.

| Layer | File | How it is made |
|---|---|---|
| back | `work/starter_back.png` | The carpet, 1920 x 1080, drawn on a grid of 4 px knots (480 x 270). Field: deep red with a low-contrast lozenge lattice (period 30 knots) and small four-petal flowers in two near-field reds; a plain central lozenge medallion (half axes 150 and 84 knots) with a darker outline and a thin second line, so the centre where tiles sit is calm. Border, from the edge in: dark indigo under the rim, a red guard with small dark dots, an ivory line, the main indigo band (24 knots) with alternating stepped ochre lozenges and ivory four-petal flowers, whole at every corner, an ivory line and a thin red guard. Wool: per-knot shade jitter of 3.5%, a slow horizontal colour drift (abrash) of 3%, darker knot edges, a 0.6 px blur and per-pixel fibre noise from a fixed seed (169). Light: one lamp above the table, brighter in the centre and up to 30% darker toward the rim (`docs/art-direction.md` section 4) |
| middle | `work/starter_middle.png` | The gold inlay line of `table.crimson_carpet`: a 4 px line 54 px inside the edge with a small lozenge at each corner, shaded from `color.gold.lo` to `color.gold.hi`, read from `docs/design-tokens.json` |
| front | `work/starter_front.png` | The wood rim of `table.crimson_carpet`: a 40 px band with a 28 px corner radius, shaded from `color.wood.lo` to `color.wood.hi`, read from `docs/design-tokens.json` |
| static | `export/static.png` | back, middle and front composited in that order; 1920 x 1080 RGBA, fully opaque |

The carpet's own palette (art-direction section 2, cosmetic illustration art) is in the script: field reds, indigo, ivory and ochre. Wood and gold come only from the tokens. The middle and front layers match the set-1 tables, so tiles and seats sit the same on every table. The table is static, so the layers stay in `work/` and only `export/static.png` is exported (`docs/cosmetics.md` section 4).

| Step of `docs/asset-workflow.md` 5.1 | Done | How, or what is missing |
|---|---|---|
| 1. Class canvas and pivot | Yes, canvas open | 1920 x 1080, the canvas of the set-1 tables; how the table covers 4:3 to 21:9 screens is open (O-6) |
| 2. Paint-over | No | The carpet is flat and graphic, the rim has no grain; an artist paints over the rim, the inlay and the carpet's wool |
| 3. Layers | Yes | back, middle and front in `work/` |
| 4. Nothing art-direction section 8 forbids | Checked by the agent | No text, glow, drop shadow, rarity colour or team hue |
| 5. Silhouette and smallest-size tests | Tile readability only | The 28 faces of `tile.ivory` at their smallest display size on this table with the runtime shadow (`assets-src/tile/ivory/provenance.md`, Readability check): centre field relative luminance median 0.034 (90th percentile 0.044), ivory face 0.765, contrast 9.7:1. The centre is a plain medallion inside a low-contrast lattice |

`npm run asset:check -- assets-src/cosmetic/table/starter`: 1 sprite, ASTC 6x6 estimate 921,600 bytes, no limit set for a static table (O-5).

Not done: W-6 and W-7. `npm run asset:add` refuses the item until the Cultural review line above ends in `approved`, so nothing is copied to `client/Assets/Cosmetics/` and no Unity import was run.

## Files

| File | sha256 | What it is |
|---|---|---|
| work/starter_carpet.py | `167a982ee6f2d094d2846589154ff55548d015e3b77254bbc0eae21d223eb8d6` | The W-3 and W-5 script |
| work/starter_back.png | `f0d40be0e14bddf3fb993bfa0c215658e1e7de7c31acf8d8d0c5e346474231e3` | 1920 x 1080 RGBA, back layer |
| work/starter_middle.png | `1d1dbaf8ca4b737007b8a911d7f1b77afc24b2939e243d44a683be0518d41b6d` | 1920 x 1080 RGBA, middle layer |
| work/starter_front.png | `54646a5f296f37b72713f58da1cfbc969456c26e712620bb46450f0f0c266e11` | 1920 x 1080 RGBA, front layer |
| export/static.png | `b38571cea052771fd5d5adeb74eec8f52f015d361aa468031bab9053bc059856` | 1920 x 1080 RGBA static, flattened layers |
