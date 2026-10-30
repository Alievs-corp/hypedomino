# Rarity

Status: proposal. Every hex value, glow level, size and period here is an initial proposal; the owner approves them by merging the pull request that adds or changes them. Tiers, order and the colour convention are owner decisions of 2026-10-26.

Rarity tiers, their order, colours, the status diamond, the rarity glow, the comparison rule and the table background selection at match start. `docs/design-tokens.json` group `rarity` holds the values; this file explains them. `node tools/design-tokens/design-tokens.mjs` checks the group (ids, ordinals, pips, glow order, contrast) and that the contrast table and colours below match it; CI runs it through `npm run test:tools`.

Items refer to a tier by `rarityId` (`docs/cosmetics.md` section 3). Glow levels, motion rules and effect levels come from `docs/ui-motion.md`; frame budgets from `docs/performance.md`; admin config from ADR 0006.

## 1. Tiers

| `rarityId` | Ordinal | Name key (`en` reference) | `fill` | `on` | `text` | Pips | Glow | Flow |
|---|---|---|---|---|---|---|---|---|
| `common` | 1 | `rarity.common.name` (Common) | `#A3A6BF` | `#120A2C` | `#C9CBDB` | 1 | `glow.none` | No |
| `rare` | 2 | `rarity.rare.name` (Rare) | `#5C78FF` | `#120A2C` | `#A3B3FF` | 2 | `glow.soft` | No |
| `epic` | 3 | `rarity.epic.name` (Epic) | `#B45CFF` | `#120A2C` | `#D3A3FF` | 3 | `glow.strong` | No |
| `legendary` | 4 | `rarity.legendary.name` (Legendary) | `#FFC23D` | `#120A2C` | `#FFD98A` | 4 | `glow.intense` | No |
| `mythic` | 5 | `rarity.mythic.name` (Mythic) | `#FF4530` | `#120A2C` | `#FF9C8F` | 5 | `glow.intense` | Yes |

- Hues follow the genre convention: grey, blue, purple, gold, red. `fill` is the diamond and the glow tint, `on` the pips on the diamond, `text` the tier label on a dark layer, as with the roles in `docs/ui-motion.md` section 2.
- The status colour of an item is its tier's `fill`. The manifest has no colour field.
- Name values in other languages are managed in admin (`docs/localization.md` section 2.1).
- `rare` (hue 230) sits next to `role.teamA` (216), `legendary` (41) next to `role.warning` (39) and `role.score` (50), `mythic` (6) next to `role.danger` (354). Colour is never the only cue: tiers differ by pip pattern, and in a match rarity shows only as a halo behind a seat frame, while team identity comes from seat position and badge, outcomes carry an icon and score yellow appears only in score effects.

## 2. Marks

### 2.1 Status diamond

A square standing on one corner. Its pips use the layout of a domino half (`docs/art-direction.md` section 3: a value reads by pattern), rotated with the square, so tiers never rely on colour alone and read in greyscale:

| Pips | Layout in the diamond |
|---|---|
| 1 | Centre |
| 2 | Top and bottom corner |
| 3 | Vertical line: top, centre, bottom |
| 4 | All four corners |
| 5 | Four corners and centre |
| 6 | Left and right column of three (reserved for a sixth tier) |

- Two white sprites from one atlas: the diamond tinted with `fill`, the pip overlay tinted with `on`, both through vertex colour. One material, so the two batch together.
- Bounding box at least `rarity.diamondMinSize` (72 units, about 4.5 mm on a 6.5-inch phone). Where it does not fit, no diamond is drawn; there is no smaller variant with fewer pips.
- The tier label is the localized name in `caption` style and the tier's `text` shade, next to the diamond, wherever the layout has room (shop and collection detail, reward reveal).
- Nothing about rarity is painted into an asset (`docs/art-direction.md` section 8).

### 2.2 Glow and flow

- The rarity glow is the runtime halo of `docs/ui-motion.md` section 4: one additive sprite from the shared effects atlas, tinted with `fill`, alpha and scale from the tier's `glow` level, one overdraw layer. `glow.none` draws no sprite.
- A tier with flow uses the flow halo instead: the same halo with one bright arc, rotated 360 degrees every `rarity.flowPeriodMs` (2400 ms, `linear`, looping) by a transform tween. Transform only, no material property, no particles.
- Flow stops at the rest angle (arc at the top) when the effect level has `animatedCosmetics` false (low) or reduce motion is on. The halo, its tint and level stay; only the rotation is dropped.

### 2.3 Where marks show

| Context | Diamond | Label | Glow |
|---|---|---|---|
| Shop, collection, inventory item card | Yes | Detail view only | Behind the item preview |
| Reward reveal | Yes | Yes | Behind the item |
| Profile and lobby avatar | No | No | Around the frame, from the frame's tier |
| Match seat | No | No | Around the frame, from the frame's tier |
| Match table background | No | No | No; the table art is the whole presentation |

### 2.4 Visual samples

| Tier | Item card on `color.surface.default` |
|---|---|
| `common` | Cool grey diamond with one dark pip in the centre; no halo |
| `rare` | Blue diamond with two pips, top and bottom; faint blue halo (`glow.soft`) |
| `epic` | Purple diamond with a vertical line of three pips; clear purple halo (`glow.strong`) |
| `legendary` | Gold diamond with four corner pips; bright gold halo (`glow.intense`) |
| `mythic` | Red diamond with four corner pips and a centre pip; bright red halo whose bright arc circles the item every 2.4 s; at the low level or with reduce motion the arc rests at the top |

The reference board panel that shows these on the real ground is listed in section 8.

## 3. Contrast

Same thresholds as `docs/ui-motion.md` section 2: `fill` is non-text UI and needs 3:1 on every ground layer (SC 1.4.11), `text` needs 4.5:1 on every ground layer and `on` 4.5:1 on `fill` (SC 1.4.3). Ratios are truncated to two decimals. Generated by `node tools/design-tokens/design-tokens.mjs --rarity-table`; the test fails if this table is stale.

| Foreground | `background.base` | `background.raised` | `surface.default` | `surface.raised` | `on` on `fill` |
|---|---|---|---|---|---|
| `rarity.tiers.common.fill` | 8.03 | 7.31 | 6.41 | 5.47 | 7.92 |
| `rarity.tiers.common.text` | 11.96 | 10.89 | 9.55 | 8.14 | - |
| `rarity.tiers.rare.fill` | 5.12 | 4.66 | 4.09 | 3.48 | 5.05 |
| `rarity.tiers.rare.text` | 9.58 | 8.72 | 7.65 | 6.52 | - |
| `rarity.tiers.epic.fill` | 5.42 | 4.93 | 4.33 | 3.69 | 5.35 |
| `rarity.tiers.epic.text` | 9.58 | 8.72 | 7.65 | 6.52 | - |
| `rarity.tiers.legendary.fill` | 11.95 | 10.88 | 9.54 | 8.13 | 11.79 |
| `rarity.tiers.legendary.text` | 14.25 | 12.97 | 11.38 | 9.70 | - |
| `rarity.tiers.mythic.fill` | 5.63 | 5.13 | 4.50 | 3.83 | 5.55 |
| `rarity.tiers.mythic.text` | 9.54 | 8.68 | 7.61 | 6.49 | - |

## 4. Comparison

- Tiers are ordered by ordinal only: tier A is rarer than tier B if and only if `ordinal(A) > ordinal(B)`. Equal ordinals mean the same tier.
- Ids, names, colours and pip counts carry no order. Code compares ordinals, never ids or strings.
- Ordinals are the integers 1 to N without gaps or repeats, N being the number of tiers.

## 5. Table background selection

At match start the server picks the one table background every client of the match shows.

Input:
- Seats `s_0 .. s_(n-1)` in the server's fixed seat index order, `n` 2 (1v1) or 4 (2v2; North, East, South, West as in `docs/rules.md`).
- `t(s)`: the table item equipped by the player in seat `s`, from the server inventory. Every account always has one: a `default` table exists and replaces a `disabled` one (`docs/cosmetics.md` sections 5 and 6).
- `ord(t)`: the ordinal of `t.rarityId` in the rarity table of the config snapshot the match records (ADR 0006).
- `b`: 32 bytes from the server's cryptographically secure random source (as for shuffling, R-4), drawn once per match for this choice only.

Algorithm:
1. `m = max over s of ord(t(s))`.
2. `C` = the seats with `ord(t(s)) = m`, in seat index order; `k = |C|`.
3. If `k = 1`, `i = 0`. Otherwise `U` = the first 8 bytes of `b` read as an unsigned 64-bit big-endian integer, and `i = U mod k`.
4. The match table is `t(C[i])`.

Rules:
- Highest rarity wins. Ties are uniform over seats: each tied seat has probability `1/k` (modulo bias under 2^-62). Two tied seats with the same item give that item probability `2/k`; the result is the same table either way.
- The server computes the choice and sends only the table item id in the match start state; clients render it and never compute it. `b` stays on the server (`docs/anti-cheat.md` NT-4) and is stored in the replay record with the moves (`docs/privacy.md` D-13), so a replay reproduces the choice.
- The table is fixed for the whole match: hands, reconnects and equip changes during the match do not change it.
- Rendering follows `docs/cosmetics.md` section 8.3: animated if allowed and loaded, otherwise its static, otherwise the built-in default table static.
- Implementation and the uniformity test: `server/internal/cosmetics` (`PickTable`, `Catalog.Table`, `NewDraw`) with the worked examples below and a uniformity test over 60 000 draws per tie size (#98). The draw is `replay.Record.TableDraw`, the chosen id is `room.Options.Table` and `MatchView.table`; ordinals come from `HD_RARITY_ORDINALS` until admin config serves the rarity table.

Worked examples (`U` in hex):

| Seats and tiers | `m` | `C` | `U` | Result |
|---|---|---|---|---|
| N `legendary`, E `common`, S `rare`, W `epic` | 4 | N | unused | N's table |
| N `epic`, E `mythic`, S `common`, W `mythic` | 5 | E, W | `0000000000000003` | 3 mod 2 = 1, W's table |
| N, E, S, W all `rare` | 2 | N, E, S, W | `FFFFFFFFFFFFFFFF` | (2^64 - 1) mod 4 = 3, W's table |
| 1v1: both `common` (`table.starter`) | 1 | seat 0, seat 1 | `000000000000000A` | 10 mod 2 = 0, seat 0's table, `table.starter` |

## 6. Cost and effect levels

Rarity adds no effect to `effects.budgets`; seat halos are part of the base scene and are measured with it.

| Item | Budget |
|---|---|
| Halo, any tier other than `glow.none` | 1 sprite, 1 overdraw layer, 0 particles |
| Flow | 1 transform tween from the pool of `maxTweens` (`docs/ui-motion.md` section 6, rule 3) |
| Diamond | 2 sprites in the item card's batch |
| Match scene, all seats | At most one halo per seat (2 or 4), all from the shared effects atlas and material on the seats' animated sub-canvas: 1 draw call, 1 overdraw layer, at most 4 tweens (0 on the low level). Counted in `baseSceneBatches` and base scene CPU and GPU time |

| Effect level / setting | Glow | Flow |
|---|---|---|
| low (`animatedCosmetics` false) | Static halo at the tier's level | Rest angle |
| medium, high | Static halo at the tier's level | Rotates |
| Reduce motion on, any level | Static halo at the tier's level | Rest angle |

The worst case for the replay harness (`docs/performance.md` section 2) is four seats with `mythic` frames. A PR that changes a halo, the flow or the diamond reports the measurements of `docs/ui-motion.md` section 7.

## 7. Admin config

The rarity table is admin config (ADR 0006): published, versioned and rolled back like any config, with `docs/design-tokens.json` holding the defaults. Each match records the config version it used. The client gets the active table with its config and ships the defaults built in, as for the cosmetics manifest (`docs/cosmetics.md` section 7.3).

| Field | Rule |
|---|---|
| Tier id | Key of `rarity.tiers`, `^[a-z][a-z0-9_]*$` (the `rarityId` pattern of `docs/cosmetics/schema.json`). Stable: never renamed, never reused, never removed while an item of any published manifest references it (`docs/cosmetics.md` V-14) |
| `ordinal` | Integers 1 to N without gaps or repeats. Changing them changes table selection for matches started after the publish |
| `nameKey` | `rarity.<id>.name` |
| `fill`, `on`, `text` | `#RRGGBB`, with the section 3 contrast on all four ground layers |
| `pips` | 1 to 6, unique across tiers. A seventh tier needs a new mark design and a spec change |
| `glow` | A level of the `glow` tokens, never weaker than the glow of a lower tier |
| `flow` | Boolean |
| `diamondMinSize` | Positive integer, UI units |
| `flowPeriodMs` | Positive integer, ms |

A publish that breaks a rule is rejected (ADR 0006). The checker in `tools/design-tokens/` enforces the same rules on the defaults. A client that meets a `rarityId` missing from its table draws no rarity mark for that item.

## 8. Server authority and open items

- The client never sends a rarity, ordinal or table choice. Rarity is display data on the client; the server uses it only for table selection.
- Prices, drop chances and rewards are not derived from rarity here; they belong to #7 and #9.
- Open: the reference board panel with the five item cards and the four-mythic-seat mock (`docs/art-direction.md` section 9, #167).
- Reward reveal for pass, mission and medal grants: `docs/pass-missions-medals.md` 6.1. Open: the shop purchase reveal (#7).
