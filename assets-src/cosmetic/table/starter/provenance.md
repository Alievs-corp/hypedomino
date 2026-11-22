# table.starter

The default table (#169): the `common` built-in table of `server/internal/cosmetics/builtin.json`, address `cosmetic/table/starter/static`, which had no art until now. It follows `docs/art-direction.md` 5.4: a red and indigo carpet with an original geometric and floral pattern and a calm centre, inside a wood rim with a gold inlay line. Every layer is computed by one deterministic script; no image generator was used. Every line below is read from a file or a record, or marked as not done. Recorded on 2026-11-15 while the work was done.

- artSource: ai_assisted (since 2026-11-22: `concept/20261122-01.png` came from a generator; the shipped layers and the static are still the script output)
- Authors: script written and run by an AI agent (Claude Code) on 2026-11-15. No person has done W-3. W-1 concept generated on 2026-11-22 by an AI agent (Claude Code) through Codex CLI, in the project owner's ChatGPT account
- Generator: Codex CLI 0.171.0 built-in image generation (`$imagegen`, `docs/asset-workflow.md` 2.5), signed in with the project owner's ChatGPT account (`codex login status`: Logged in using ChatGPT); the plan was not read on 2026-11-22
- Terms: OpenAI Terms of Use, `https://openai.com/policies/terms-of-use/`, not read (RC-1 open, O-2). The account holder reads the terms in force on 2026-11-22 for the plan and records the date here
- Canonical concept: concept/20261122-01.png, a painted carpet concept for the back layer. Not yet used: W-3 from this concept (trim, fit, calm centre, layers, readability check, as `table.crimson_carpet` does) is not done, so `work/` and `export/` are unchanged
- Inputs uploaded to the generator: none; one prompt, one image, no edits
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

## Any screen shape (2026-11-16)

Owner decision of 2026-11-16, after the Art Bench check of #724 on a 20:9 phone: the rim, the inlay and the carpet border must show whole on every screen shape (16:9 to 21:9, 4:3, nearly square). Done on 2026-11-16 by Claude Code in `work/starter_carpet.py` (`periodic_edges`, `write_slice`), with no new canvas and no new layer; the rule is `docs/cosmetics.md` 8.4 and `docs/art-direction.md` 5.4.

- Ring depth 168 px at the 1920 x 1080 scale: the rim, the inlay, the carpet border and its guards, the same on all four sides. The lozenge medallion stays in the field; on 4:3 and narrower its left and right points are cropped.
- The top and bottom borders now repeat. Each band is split into its light (a Gaussian blur of 120 px along the edge, made symmetric left to right) and its detail. The script tries steps of 150 to 480 px in steps of 6 (so a 1280 x 720 canvas keeps whole pixels) and corner pieces of at least the depth plus 30 px and at most 520 px less a quarter step, scores each by how well the art matches at both ends of one step (windows of 12 and 48 px), how mirror-symmetric it is at the start, and how far the repeated band drifts from the art, and keeps the lowest score. Chosen: step 372 px, corner pieces 0 to 402 px and 1518 to 1920 px, score 0.3282. Between the corner pieces both bands are the run from 402 to 774 px repeated 3 times, joined by 32 px cross-fades, and they fade into the untouched field over the 16 px inside the ring. The corner pieces and the side bands keep the art as it was; a motif unique to the middle of the top or bottom edge is replaced by the run.
- `export/slice.json`, `{"depth": 168, "step": 372}`, records the depth and the step in pixels of the static; `CosmeticContent` sets them as the sprite border of the static and of every layer frame, and `TableImage` draws the table at the screen's shape. Drawn by the same rule as `TableCloth` (scratch render, not committed): 16:9 steps +0 x1.000; 20:9 steps +1 x1.047; 21:9 steps +2 x0.946; 4:3 steps -1 x0.930 (steps gained or lost in the middle of the long edges, then the horizontal scale of the whole table).
- Readability check of #708 run again on the new static (section above): pass. `npm run asset:check` gives the same sprite count and estimate as before.
- Not done: the look of the repeated edges is not reviewed by a person; the device check on the Art Bench comes with the bench build of this change.

## Files

| File | sha256 | What it is |
|---|---|---|
| work/starter_carpet.py | `a0f2f52a3ba0aa334122da51cbd04077b02ffed0801da253075c8fe9b573aba3` | The W-3 and W-5 script; `periodic_edges` and `write_slice` added on 2026-11-16 (section Any screen shape) |
| work/starter_back.png | `f41ace75aa5da3fc9cd26de93f36b9f792e1e311fc47ad943f20d5fbb1e2326d` | 1920 x 1080 RGBA, back layer |
| work/starter_middle.png | `1d1dbaf8ca4b737007b8a911d7f1b77afc24b2939e243d44a683be0518d41b6d` | 1920 x 1080 RGBA, middle layer |
| work/starter_front.png | `54646a5f296f37b72713f58da1cfbc969456c26e712620bb46450f0f0c266e11` | 1920 x 1080 RGBA, front layer |
| export/static.png | `8b6e0b70ff0b7f6a2f5f150446f724518edab99dbfc842f0eef9477a98310ad9` | 1920 x 1080 RGBA static, flattened layers |
| export/slice.json | `b072b29b49d7e3e4bab9d89d5b65ec3b825e2afd2dc459aa46fff328361e6011` | `{"depth": 168, "step": 372}`: the ring depth and the edge step in pixels of the static (`docs/cosmetics.md` 8.4) |

## Concept (2026-11-22)

Generated by Claude Code with `env -u OPENAI_API_KEY codex exec --skip-git-repo-check --sandbox workspace-write` in `concept/` (`docs/asset-workflow.md` 2.5). Codex session `01a2039c-5418-72f1-a1b0-36fd966ea663`, prompt sent 2026-11-22T14:15:04Z. `concept/20261122-01.png` is a byte copy of the file Codex keeps under `~/.codex/generated_images/01a2039c-5418-72f1-a1b0-36fd966ea663/`. The C2PA manifest was not read.

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-22 | concept/20261122-01.png | P-5 back layer with the style key, quoted below. SHA-256 `56caa5dcff63a7ea5c511a864fd66cdca544b799d22d038f2f46fa57fb4446ec` |

Prompt of `concept/20261122-01.png`, verbatim (the instruction to save the file follows it in the Codex call):

```text
Original 2D game art for an all-ages mobile game with a premium, aristocratic look. Painted, semi-realistic rendering with soft volume and rich materials: polished gold and brass, ivory, deep lacquer, gemstone, crystal, silk and wool. Fine ornament inside a clear, readable outer shape. One warm light from above. No text, letters, numbers, logos, flags, emblems or religious symbols. No weapons. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject. Top-down view of a woven carpet table cloth, a red and indigo salon carpet with a low-contrast lozenge lattice of small four-petal flowers, a plain central lozenge medallion, and an indigo border of small rosettes and stepped motifs, back (the whole carpet, field and border, filling the image edge to edge) layer only, original geometric and floral pattern, calm low-contrast centre where game pieces will sit. Palette: deep red, dark indigo, ivory and muted ochre.
```

Concept against the prompt, looked at by the agent: an edge-to-edge carpet, field and border, with a calm centre; no text, letters, numbers, glow, drop shadow, casino cue or team hue seen. Not checked: generator artefacts at full size, the readability rule of #708 (needs W-3 first).
