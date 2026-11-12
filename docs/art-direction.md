# Art direction

Status: proposal; the owner approves it by merging. How every shipped image looks: palette use, shape, light, outline, avatars, effects, static fallbacks and the reference board. It defines no colour, duration or budget value. Colours are named by their token path in `docs/design-tokens.json`; motion and effect budgets come from `docs/ui-motion.md` and `docs/performance.md`; delivery follows ADR 0002; item data follows `docs/cosmetics.md`; content limits follow `docs/localization.md` sections 6 and 7 (PEGI 3 target). `node tools/art-direction/art-direction.mjs` checks that this file contains no colour literal and that every token path it names exists; CI runs its tests through `npm run test:tools`.

Production workflow, prompt templates, naming and folders are `docs/asset-workflow.md` (#167). UI kit is #168, tile set and table base art #169, effects library #176, animated asset format and per-screen animation budget #93, rarity and status colours `docs/rarity.md`.

## 1. Pillars

| Pillar | Means |
|---|---|
| Readable first | Board state (pips, legal ends, turn owner, score) reads at phone size in one glance. Decoration never competes with it |
| Energy from motion and light | Calm, flat shapes at rest; the hype comes from pops, bursts and glow on events, inside their budgets |
| Cheap by construction | 2D art only: flat shading, baked outlines, white effect sprites tinted at runtime, shared atlases. No 3D models (owner decision 2026-11-12). A look that needs a post-process or a material instance is rejected |
| All ages, original | Every design is ours, passes the cultural checklist and reads the same in all four launch markets |

## 2. Palette usage

Dark layered ground, vivid accents. The accent hues carry meaning, so art never spends them on decoration.

| Layer | Tokens | Rule for art |
|---|---|---|
| Ground | `color.background.base`, `color.background.raised` | Everything is composed to sit on these. Art never fills the ground with a bright flat colour |
| Panels behind art | `color.surface.default`, `color.surface.raised` | Icons, avatars and items in lists and shop cards are checked on both |
| Shadow and outline | `color.shadow` | Baked elevation shadows (ui-motion section 4) and the outer outline of every sprite (section 4 here) |
| Dimming | `color.scrim` at `opacity.scrim` | Modal backdrop and the `block` board dim; never painted into an asset |
| Main action | `color.role.primary` | Only the one primary action per screen and its `glow.soft`. Not a decoration hue |
| Selection, links | `color.role.secondary` | Selection and legal-end highlights |
| Outcome | `color.role.success`, `color.role.warning`, `color.role.danger` | Only with their meaning and always paired with an icon or shape |
| Points | `color.role.score` | Score popups, score particles, points flying to the meter. Nothing else is score yellow in a match |
| Teams | `color.role.teamA`, `color.role.teamB` | Team seats, meters, the `domino` banner and `matchWin` confetti. A cosmetic is never dominated by a team hue, so it cannot be mistaken for a team marker |
| Text | `color.text.primary`, `color.text.secondary`, `color.text.muted` and each role's `text` shade | Text is never baked into art (localization 6) |
| Rarity | `rarity.*`, `docs/rarity.md` | No rarity colour, diamond or rarity glow is painted into an asset. The client draws them from `rarityId` |

Rules:
- Effect and glow sprites are authored white or greyscale and tinted at runtime with a role `fill` through vertex colour. One sprite serves every role; there are no per-colour effect textures.
- A coloured element on a role fill uses that role's `on` shade, as UI text does.
- No colour combination is chosen to evoke a flag, party or side (localization 6).
- UI, effect, team, score and role colours come only from the tokens above. Cosmetic illustration art (avatars, tables, frames, tile skins, emoji: skin, hair, materials, table art, shade and highlight tones) uses its own palette per asset, reviewed on the reference board (section 9). It is never dominated by a team or role hue and never evokes a flag or political side (localization 6).

## 3. Shape language

| Shape | Meaning | Used for |
|---|---|---|
| Rounded rectangle | Solid, touchable, the game piece | Tiles, buttons, panels, cards, frames. UI corner radii come from `radius` |
| Circle | Counting and status | Pips, badges, score bubbles, avatar crops |
| Pill | The main action | The primary button (`radius.pill`) |
| Star, ray, wedge | Impact, an event happening now | Bursts, the `block` lock stamp, `rankUp` rays. Only in effects and only for the effect's lifetime |

Rules:
- Big simple masses first: an asset reads as a filled silhouette with no interior lines. If it does not, it is redesigned, not detailed further.
- Detail that vanishes at the smallest size the asset is shown at is removed, not shrunk. The smallest display size per asset class is set with the class spec (#168 to #175).
- Pips are identical circles at a fixed layout per value, so a value reads by pattern, not by counting. Tile proportions and pip layout are #169.
- Corners are rounded; sharp points appear only in impact shapes. Nothing reads as a weapon (localization 6).
- No card suits, chips, reels, roulette or coin piles in shape or silhouette (localization 6, gambling imagery).

## 4. Lighting and outline

| Topic | Rule |
|---|---|
| Key light | One light from above. Cast shadows fall straight down, matching `elevation`, which offsets shadows only on Y |
| Shading | 2D assets: flat cel shading, base, one shade tone, one highlight. No airbrushed gradients, no painted textures, no photographic detail |
| Rim light | Allowed on avatars to separate them from `color.background.base`; one thin band on the side away from the key light |
| Glow | Never baked into a sprite. Glow is the runtime additive sprite tinted by a role fill and scaled by a `glow` level (`glow.soft`, `glow.strong`, `glow.intense`); it costs one overdraw layer |
| Drop shadow | Never baked into a sprite. UI elevation is the 9-slice shadow sprite in `color.shadow` |
| Outer outline | Every avatar, frame, emoji, item icon and tile has a solid outer outline in `color.shadow`, baked into the sprite. One weight per asset class, set with the class spec (#168 to #175), identical across every item of that class |
| Inner lines | Thinner than the outer outline, only where two masses overlap. Their tone belongs to the asset's own palette (section 2) |
| Text outline | Through the font's outline material preset (ui-motion section 3), never a second text copy |
| Post-processing | None in the match scene at any effect level (ui-motion section 4). Art must look finished without bloom, blur or grading |

No runtime outline or rim shader on 2D assets: a baked outline costs nothing per frame, a shader outline costs a pass and overdraw.

## 5. Avatars

### 5.1 Proportions

| Item | Proportion |
|---|---|
| Face | Eyes and mouth carry the emotion and stay readable at seat size |
| Looks | Varied across a set; no face, skin tone or body type stands for a group, region or trait (localization 6, representation) |
| Avatar | Head and shoulders, face centred. Nothing essential in the outer band that a frame (#173) or the rarity glow (`docs/rarity.md`) covers; the band width is set with #168 |

All-ages: no sexualised proportions, poses or outfits, no likeness of a real person, no resemblance to an existing mascot, figure or brand.

## 6. Effect language

Effects are named and budgeted in ui-motion section 7 (`effects.budgets`, `effects.levels`). This section fixes how they look; it never raises a budget.

| Element | Look | Constraint |
|---|---|---|
| Glow | A soft round halo behind its owner, tinted by the event's role fill | Levels only from `glow`; strength by score step comes from `scoreCascade.steps`; one overdraw layer |
| Burst | Short radial spray: small round dots for dust and settle (`tilePlace`), stars and wedges for scoring and wins | Count per level from the effect's budget, for example `effects.budgets.tilePlace` and `effects.budgets.scoreCascade`. Tinted by the role of the event: score bursts by `color.role.score`, team events by the team fill |
| Trail | A tapered streak behind a moving thing: points flying to the team meter, the tile travelling to its end | Part of its effect's particle and draw call budget; it ends when the motion ends (`motion.duration.base` for tile travel) |
| Impact | Squash on settle, ring pulse (`spinnerFormation`), stamp (`block`) | Shake only from `motion.shake.small` or `motion.shake.medium`; full-screen flashes at most `motion.maxFullScreenFlashesPerSecond` |
| Confetti | Simple paper shapes in the winning team fill (`matchWin`) | No coins, cash, chips or reel motion (localization 6) |

Rules:
- An effect is designed at the low level first: it must tell the event with the low particle count (`effects.levels.low.maxParticles` across all live effects) and without animated cosmetics (`effects.levels.low.animatedCosmetics`). Higher levels add particles; they do not change timing or meaning.
- Every effect works with reduce motion on (ui-motion section 6): the information stays, the motion drops.
- Effect sprites are small, tightly meshed and in one shared effects atlas so they batch and keep overdraw within `maxOverdrawLayers` of the level.
- Additive or alpha blending from the shared effect material only; no per-effect material instance.
- No video for in-match effects (ADR 0002).
- A new effect or a change to one reports the measurements in ui-motion section 7 and `docs/performance.md` section 8.

## 7. Static fallback

Every animated asset has a static sprite (ADR 0002, `docs/cosmetics.md` V-2). It is shown on the low effect level, while the animation downloads, after a failed download and when the animation budget (#93) is spent, so it is a finished asset, not a placeholder.

| Rule | Check |
|---|---|
| It is a chosen key frame that carries the item's full identity, not frame 0 by default | Reviewer can name the item from the static alone |
| Same canvas size, pivot and framing as the animation | Swapping static and animated shows no jump |
| A table's static is the flattened composition of all its layers at their rest frame (`docs/cosmetics.md` section 4) | Static and animated tables line up layer for layer |
| It meets every rule in this document, including outline and silhouette | Reviewed together with the animation |
| It passes the cultural review with the animation; switching an item off removes both (localization 6.1) | `culturalReview` is `approved` |

## 8. Do and don't

| Do | Don't |
|---|---|
| Name every UI, effect, team, score and role colour by its token path | Use a colour value that has no token for those |
| Author effect sprites white and tint them at runtime | Ship one effect texture per colour |
| Bake the outer outline of a 2D asset in `color.shadow` | Draw outlines with a runtime shader |
| Leave glow, shadow and rarity marks to the runtime | Paint glow, drop shadows or rarity colour into a sprite |
| Keep text out of images; every word is a localization key | Bake words, numbers in a language, or a language flag into art |
| Test every asset as a silhouette and at its smallest display size | Add detail that disappears at seat or list size |
| Use the score and team hues only for points and teams | Use team or score hues as decoration on cosmetics |
| Design effects at the low level first | Make a look that only works at the high level |
| Make the static fallback a finished key frame | Ship frame 0 or a blank as the fallback |
| Keep expressions and gestures friendly and all-ages | Taunting gestures, weapons, blood, sexualised poses |
| Use abstract game tokens for currency | Coins piled like cash, chips, reels, card suits |
| Create original designs | Copy or imitate another game, brand, mascot or real person |

## 9. Reference board

The board is a set of panels, produced as original art through the workflow of #167 and stored in `assets-src/` under its naming rules. Each panel is drawn on `color.background.base` at the 1080-unit short side of the match canvas (ui-motion section 1), takes UI, effect, team, score and role colours only from the tokens named in this document, shows cosmetic art in its own per-asset palette (section 2), carries no text other than localization-free panel ids, and names no third-party game, studio or product in its files or metadata.

| Panel | Must show | Proves |
|---|---|---|
| B-1 Palette in context | A match scene mock: ground layers, a panel, the primary button, a selection highlight, a score popup, both team meters, an outcome icon. Rarity marks of all five tiers (`docs/rarity.md`) | Section 2 roles on the real ground |
| B-2 Value check | B-1 in greyscale | Pips, legal ends, turn owner and score read without hue |
| B-3 Shape vocabulary | Rounded rectangle, circle, pill and impact shapes side by side, each with one in-game use | Section 3 |
| B-4 Light and outline | One tile lit from above, showing base, shade and highlight and the outer outline in `color.shadow`; one avatar with its baked outline and rim band; both again with glow added at runtime | Section 4 |
| B-5 Proportions | The sample avatar at its display size with the frame band marked, inside its frame | Section 5.1 |
| B-6 Silhouette sheet | The sample avatar's key frames, each with its filled silhouette | Section 3 rules |
| B-7 Effects storyboard | `tilePlace`, `scoreCascade` at steps 1, 3 and 6, `domino` and `block`, each at the low and high level | Section 6 and its budgets |
| B-8 Static fallback | The sample avatar's static sprite over its animated key frame with the shared pivot marked | Section 7 |
| B-9 Do and don't | Paired examples for at least: baked text, baked glow, decoration in a team hue, casino cues, frame 0 fallback | Section 8 |
| B-10 Size test | Avatar, frame, emoji and an item icon at their smallest display size on the match canvas | Readability rules in sections 3 and 5 |

The sample avatar on B-4 to B-8 is the one that #167 takes from concept to game.

A cosmetic's palette is reviewed per asset on the board: shown on both `color.surface.default` and `color.surface.raised` and beside `color.role.teamA` and `color.role.teamB`, so it never reads as a team marker.

## 10. Open items

- Tile face and pip colours have no tokens. Tiles are not a cosmetic type (`docs/cosmetics.md` section 2), so they need tokens added to `docs/design-tokens.json` by owner-approved PR before tile art is final (#169).
- Outline weight, smallest display size and avatar frame band per asset class: #168 to #175.
- The reference board images: #167.
