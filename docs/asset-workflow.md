# Asset workflow

Status: proposal; the owner approves it by merging. How an asset goes from an AI-assisted concept to a published cosmetic: prompt templates and consistency, rights record, naming and folders in `assets-src/`, the clean-up and animation steps, and the sample character walkthrough. It defines no colour, size, duration or budget. Look and rules come from `docs/art-direction.md`, item data and addresses from `docs/cosmetics.md`, rarity from `docs/rarity.md`, delivery from ADR 0002, budgets from `docs/performance.md` and `docs/ui-motion.md` section 7, upload and publish from `docs/admin.md` section 5.2, the content checklist from `docs/localization.md` section 6. `node tools/asset-naming/asset-naming.mjs` checks `assets-src/` against section 4; CI runs its tests through `npm run test:tools`.

## 1. Stages

| Id | Stage | Tool | Output | Gate |
|---|---|---|---|---|
| W-1 | Concept | ChatGPT in the browser, company account (O-1) | Raw images in `concept/`, prompt log in `provenance.md` | Section 2 templates used; section 3 record started |
| W-2 | Concept review | Reference board (`docs/art-direction.md` section 9) | One approved concept sheet per item | Silhouette, smallest-size and palette checks of art-direction sections 2, 3, 5; similarity check (section 3) |
| W-3 | Clean-up | Illustrator, Photoshop | Layered master in `work/` | Section 5.1 |
| W-4 | Animation | Unity 2D rig, After Effects or Blender, as #93 decides | Clips or loop in `work/` | Section 5.2 |
| W-5 | Export | Section 5.3 | `export/` mirrors the Addressables address (section 4.2) | `node tools/asset-naming/asset-naming.mjs` passes |
| W-6 | Pack | Unity Addressables, Sprite Atlas, ASTC | Item group and atlas (`docs/cosmetics.md` 8.2) | Section 5.4 |
| W-7 | Upload and review | Admin assets (`docs/admin.md` 5.2) | Asset records, item draft | Self-check by the author, cultural review by another account (`docs/localization.md` 6.1) |
| W-8 | Publish | Admin change set | Manifest entry | V-1 to V-19 (`docs/cosmetics.md` section 9) |

Raw AI output never reaches W-5. Every exported pixel has passed W-3.

## 2. Prompts and consistency

### 2.1 Rules

- English prompts. The style key (2.2) is pasted verbatim at the start of every prompt; it is never paraphrased or shortened.
- Never name an artist, studio, game, film, brand, mascot or real person, and never ask for a look "like" one (CLAUDE.md, `docs/localization.md` 6, Likeness and brands). Describe shapes, proportions and materials instead.
- Never upload an image we do not own as input. Allowed inputs: our own concept sheets, our own sketches, our own exported sprites.
- No words, letters, numbers in a language, flags, rarity marks, glow or drop shadows in the requested image (`docs/art-direction.md` section 8). They are added at runtime or not at all.
- Colours in prompts are described by the item's own palette in plain words. UI, team, score and role hues are never requested for a cosmetic (`docs/art-direction.md` section 2).
- One change per follow-up prompt (pose, or expression, or outfit), so a drift can be traced to its cause.
- Every prompt and every downloaded image is logged in the item's `provenance.md` (section 3) at the time it is made.

### 2.2 Style key

```text
Original 2D game art for an all-ages mobile game. Flat cel shading with exactly three tones per material: base, one shade, one highlight. One light from directly above. Solid dark outer outline of even weight around the whole silhouette, thinner inner lines only where shapes overlap. Big simple shapes with rounded corners, no sharp points, no fine texture, no gradients, no photographic detail. No text, letters, numbers, logos, flags or symbols. No glow, no drop shadow, no cast shadow, no background scenery. Isolated on a plain flat light grey background that does not appear anywhere in the subject.
```

The plain background makes extraction in W-3 a single selection. The outline is recoloured to `color.shadow` in W-3; prompts never try to hit a token value.

### 2.3 Templates

`{style key}` is 2.2. Placeholders in braces are filled per item; nothing else changes.

| Id | Use | Prompt after the style key |
|---|---|---|
| P-1 | Character concept sheet (new character) | `Character design sheet of {one-line identity: who, mood, outfit, two signature shapes}. About three heads tall, large head and hands, short limbs, compact torso, eyes set low in the head. Show the same character twice side by side at the same scale: front view and three-quarter view. Neutral standing pose, arms slightly away from the body. Palette: {3 to 5 plain colour words}.` |
| P-2 | Character key pose | Attach the approved P-1 sheet. `Same character as the attached sheet, identical proportions, outfit, palette and outline. Three-quarter view. Pose: {idle / win / lose / reaction line from docs/art-direction.md 5.2}. One clear line of action, full body visible, nothing cropped.` |
| P-3 | Avatar | Attach the approved P-1 sheet if the avatar is of a character. `Head and shoulders portrait of {identity}, face centred, looking slightly toward the viewer, friendly expression. Keep the outer band of the image empty of anything essential.` |
| P-4 | Frame | `Decorative square border for a round portrait, {theme in plain words}, symmetric, open centre left empty, thin enough not to cover a face. Palette: {colour words}.` |
| P-5 | Table layer | `Top-down view of a game table surface, {theme}, {back / middle / front} layer only, calm low-contrast centre where game pieces will sit, decoration only toward the edges. Palette: {colour words}.` |
| P-6 | Emoji | `Single expressive reaction icon: {emotion}, readable as a silhouette, one subject, no hands making any sign.` |

Effect sprites (`docs/art-direction.md` section 6) are white or greyscale shapes; they are drawn by hand in W-3 and need no template.

### 2.4 Keeping a character consistent

1. The approved P-1 sheet is the canonical reference, saved as `concept/<date>-<nn>.png` and marked as canonical in `provenance.md`.
2. Every later prompt for the item (P-2, P-3) attaches the canonical sheet and repeats the identity line of P-1 word for word.
3. All poses of one item are generated in one conversation started from the canonical sheet. A new conversation starts again from step 2.
4. A generation that changes proportions, outfit, palette or outline weight is rejected, not fixed in the prompt chain. Small drift (a hand, a fold) is fixed in W-3.
5. The final consistency comes from W-3 and W-4: one cleaned master, and every pose and the static fallback is built from its parts or redrawn over it at the same canvas and pivot (`docs/art-direction.md` 5.1, section 7). AI output is reference for the poses, not the source of the shipped frames.

## 3. Rights and provenance record

Every item folder has a `provenance.md`. It is written during W-1 to W-3 and is checked at W-7 before the cultural review. `artSource` in the manifest (`docs/cosmetics.md` section 3) is `ai_assisted` if any image in `concept/` or `work/` came from a generator, `handmade` otherwise.

```markdown
# <item id>

- artSource: ai_assisted | handmade
- Authors: <names of the people who made W-1 to W-4>
- Generator: <product and model name as shown in the tool>, account <company account id, never a password or token>
- Terms: <URL of the generator's terms in force>, read on <YYYY-MM-DD>
- Canonical concept: concept/<YYYYMMDD>-<nn>.png
- Inputs uploaded to the generator: <files and their origin; only our own art>
- Similarity check: <method>, <YYYY-MM-DD>, <result; any close match and what was changed>
- Human work: <what was redrawn, recoloured, rigged, animated, in which work/ files>
- Self-check (docs/localization.md 6.1 step 1): <YYYY-MM-DD>, <author>, pass
- Cultural review: <YYYY-MM-DD>, <reviewer>, approved | rejected

| Date | File | Prompt or edit |
|---|---|---|
| YYYY-MM-DD | concept/YYYYMMDD-01.png | P-1 with {filled placeholders} |
```

What must be verified, per item:

| Id | Check | Evidence in `provenance.md` |
|---|---|---|
| RC-1 | The generation used the company account named in O-1, and the generator's terms in force on that day were read | Generator, Terms lines |
| RC-2 | No input image, prompt or file name names or depicts a third party's artist, studio, game, brand, mascot, logo or real person | Prompt log, Inputs line |
| RC-3 | A reverse image search of the canonical concept and the final static finds no close match to existing work | Similarity check line |
| RC-4 | No AI artefacts and no resemblance to a real person in any generated person (`docs/localization.md` 6, Representation) | Self-check line |
| RC-5 | Substantial human work is documented and its layered source files are kept in `work/` | Human work line, `work/` files |
| RC-6 | The item passed `docs/localization.md` section 6 | Self-check and Cultural review lines |

This document draws no legal conclusion on ownership, copyright protection or licensing of AI-generated images; those are owner items O-1 to O-4. Until they are closed, the record above is kept for every item so the answer can be applied to existing assets later.

## 4. Naming and folders

### 4.1 Tree

```text
assets-src/
  README.md
  cosmetic/<type>/<slug>/     type and slug of the item id <type>.<slug>
    provenance.md
    concept/<YYYYMMDD>-<nn>.<png|jpg|jpeg|webp>
    work/<name>.<ext>
    export/static.png
    export/anim/...           avatar, frame, emoji
    export/clip/<clip>/...    character: idle, win, lose, reaction
    export/layer/<layer>/...  table
  medal/<slug>/               same item layout: provenance.md, concept/, work/, export/
  effect/<slug>/
  ui/<slug>/
  tile/<slug>/
  board/<slug>/               reference board panels, slug b_1 to b_10 (docs/art-direction.md 9)
  sound/<slug>/
```

| Rule | Detail |
|---|---|
| Characters | Every folder and file name is lowercase `[a-z0-9_]`, plus `-` only in concept file names and one `.` before the extension. No spaces, no capitals |
| Slug | The slug of the item id (`[a-z0-9_]+`, `docs/cosmetics.md` section 3). It is stable like the id: a folder is never renamed or reused for another item |
| Versions | No `_v2`, `_final`, `_new` in names. Git is the history; a changed asset replaces the file and the item gets a new `revision` at publish |
| Concept files | `<YYYYMMDD>-<nn>`, date of generation, `nn` from `01` per day |
| Work files | `<slug>` or `<slug>_<part>` (for example `comet_fox_head`), any tool format; heavy formats go through Git LFS (`.gitattributes`) |
| Item folder | Holds only `provenance.md`, `concept/`, `work/` and `export/`. `provenance.md` exists from the first concept on |
| Classes other than `cosmetic` | Folder roots only. Their addresses come from their own specs: medals `medal/<slug>/static` (`docs/admin.md` 5.9), UI kit #168, tiles #169, effects #176, sound #177 |

### 4.2 Export to Addressables

For cosmetics the path under `export/` is the address tail, so the mapping needs no table lookup:

| Export path in `assets-src/cosmetic/<type>/<slug>/` | Address (`docs/cosmetics.md` 8.1) | Manifest field |
|---|---|---|
| `export/static.png` | `cosmetic/<type>/<slug>/static` | `staticAddress` |
| `export/anim/` | `cosmetic/<type>/<slug>/anim` | `animationAddress` |
| `export/clip/<clip>/` | `cosmetic/character/<slug>/clip/<clip>` | `clips.<clip>` |
| `export/layer/<layer>/` | `cosmetic/table/<slug>/layer/<layer>` | `layers[].address` |

Item id `<type>.<slug>`, group `cosmetic-<type>-<slug>` (or `cosmetic-builtin` for a `default` item), one atlas per item inside its group, labels `cosmetic`, `cosmetic-<type>` and `cosmetic-static` on the static sprite (`docs/cosmetics.md` 8.2). Once `export/` exists it holds `export/static.png`, as V-2 requires `staticAddress` for every item. The files inside `anim/`, `clip/<clip>/` and `layer/<layer>/` are in the format #93 picks. Automating W-5 and W-6 with one command is #179.

## 5. From raw image to game

### 5.1 Clean-up (W-3)

1. Photoshop: remove the background, trim, place on the class canvas at the class pivot (`docs/art-direction.md` 5.1; sizes per class: O-6).
2. Illustrator: redraw or trace the approved concept as flat shapes, three tones per material (`docs/art-direction.md` section 4). Fix AI artefacts: hands, eyes, asymmetry, merged or extra parts.
3. Outer outline in `color.shadow`, at the one weight of the asset class (#168 to #175); inner lines thinner, in the item palette.
4. Remove anything art-direction section 8 forbids: text, glow, drop shadow, rarity colour, team or score hue as decoration.
5. Silhouette test (filled black) and smallest-display-size test (`docs/art-direction.md` sections 3 and 5) on the reference board.
6. For a rigged character: separate parts on named layers (`head`, `torso`, `arm_upper_l`, ...), with hidden overlap painted under each joint. Save as `work/<slug>.psb` so the Unity PSD Importer can build the rig from layers.

### 5.2 Animation (W-4)

The runtime format and the per-screen animation budget are #93; this section only fixes what each path must deliver.

| Path | Steps | Delivers |
|---|---|---|
| Unity 2D skeletal rig | Import `work/<slug>.psb`, bone the parts, weight, animate each clip | One rig, clips sharing it |
| After Effects | Animate the layered master, render each clip as a PNG sequence with alpha, no effects that add glow, blur or shadow | Frame sequences for a sprite sheet |
| Blender | Orthographic camera at the class canvas, flat shading with three tones, outline matching 5.1 step 3, no post-processing, render PNG sequence with alpha | Frame sequences for a sprite sheet |

Per type: a character delivers `idle`, `win`, `lose`, `reaction` (`docs/art-direction.md` 5.2); avatar, frame and emoji deliver one loop; an animated table delivers its layers, back to front, at most `TableMaxLayers` (`docs/cosmetics.md` section 4). Each clip is checked with reduce motion and on the low effect level, where only the static shows.

Static fallback, every item: a chosen key pose or key frame, not frame 0, exported at the same canvas, pivot and framing as the animation; for a table the flattened rest frame of all layers (`docs/art-direction.md` section 7). It is reviewed side by side with the animation on board panel B-8.

### 5.3 Export (W-5)

- PNG, straight alpha, sRGB, trimmed to the class canvas, into the `export/` path of section 4.2.
- Effect sprites exported white or greyscale for runtime tint (`docs/art-direction.md` section 2).
- `node tools/asset-naming/asset-naming.mjs` passes before the commit.

### 5.4 Pack (W-6)

- One Sprite Atlas per item in its group; effect sprites only in the shared effects atlas (`docs/art-direction.md` section 6).
- ASTC on both platforms (ADR 0002); block size per asset class is set in the asset class spec (O-5).
- Size: a character stays under 1.5 MB (#170 acceptance criterion); limits for the other classes and the atlas limits are #93 and #135. The admin measures format, size, atlas and frame count on upload (`docs/admin.md` 5.2).
- A new or changed effect, or anything in the match scene, reports the measurements of `docs/performance.md` section 8.

## 6. Sample character walkthrough

The acceptance check of #167. The art team fills the evidence column for one character; its board panels are B-4 to B-8 (`docs/art-direction.md` section 9). The item is complete when every row has evidence.

| Step | Done when | Evidence |
|---|---|---|
| 1. Item id chosen | `character.<slug>` free in the manifest, folder `assets-src/cosmetic/character/<slug>/` created | |
| 2. Concept (W-1) | P-1 used with the style key; prompt log in `provenance.md` | |
| 3. Concept review (W-2) | Canonical sheet marked; silhouette and size tests passed; RC-2, RC-3 recorded | |
| 4. Poses | P-2 for `idle`, `win`, `lose`, `reaction` from the canonical sheet; B-6 drawn | |
| 5. Clean-up (W-3) | Layered master in `work/`; outline in `color.shadow`; B-4 and B-5 drawn | |
| 6. Animation (W-4) | Four clips in the #93 format | |
| 7. Static fallback | Key pose exported at the shared pivot; B-8 drawn | |
| 8. Export (W-5) | `export/static.png` and `export/clip/<clip>/` present; naming check passes | |
| 9. Pack (W-6) | Item group and atlas built with ASTC; size under the #170 limit | |
| 10. Upload (W-7) | Admin asset checks pass; self-check and cultural review by another account recorded | |
| 11. Publish (W-8) | Manifest entry valid with `artSource` set; item renders animated on medium and static on low on a Mid device | |

## 7. Open items

| Id | Item | Owner |
|---|---|---|
| O-1 | Which generator account is used (company plan, not personal), who holds it, and whether contractors may use their own accounts | Owner |
| O-2 | The current terms of the chosen generator on ownership and commercial use of output, read and confirmed for this use | Owner, with legal advice |
| O-3 | Whether copyright protection of AI-assisted art matters for the business, in which jurisdictions, and how much human work (RC-5) that needs | Owner, with legal advice |
| O-4 | Whether any store or market requires disclosure of AI-generated content beyond `artSource`; how long provenance records are kept | Owner |
| O-5 | ASTC block size, maximum texture and atlas size per asset class | #93, #135, class specs #168 to #175 |
| O-6 | Canvas size, pivot, outline weight and smallest display size per class | #168 to #175 (`docs/art-direction.md` section 10) |
| O-7 | Reverse image search tool for RC-3 | Art lead |
| O-8 | The sample character of section 6 and the board images B-1 to B-10 | Art team, #167 |
