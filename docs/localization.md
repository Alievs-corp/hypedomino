# Localization and content policy

Languages, text delivery, text rules, nickname policy and the cultural content checklist. Rule and penalty meaning stays in `docs/rules.md` and `docs/penalties.md`; this document only fixes how their text is shown. Every name in `CamelCase` code font below is admin config (ADR 0006); any value shown is a proposed default, not a constant.

## 1. Launch languages

| Code | Language | Script | Notes |
|---|---|---|---|
| `az` | Azerbaijani | Latin | Primary market. Source language for new keys. |
| `en` | English | Latin | Final fallback. Always complete. |
| `tr` | Turkish | Latin | Dotted/dotless i casing (4.6). |
| `ru` | Russian | Cyrillic | Three plural forms (4.2). |

Adding a language is an admin and bundle change plus a glyph check (4.5); it needs no client release unless the font set lacks the glyphs.

### 1.1 Selection

| Step | Rule |
|---|---|
| First launch | Walk the device locale list in order and take the first entry whose language subtag maps to a launch language. |
| Mapping | `az*` (any script or region) to `az`; `tr*` to `tr`; `ru*` to `ru`; `en*` to `en`. Other languages are skipped. If no entry maps, use `en`. |
| Persistence | The choice is stored as a player preference and synced with the account. It is a display setting only; the server never trusts it for game logic. |
| Change | Settings lists the four languages by their own name (Azərbaycanca, English, Türkçe, Русский). It applies immediately, without restart. |
| Later launches | The device locale is no longer consulted. |

### 1.2 Fallback order

A key is looked up in the selected language, then along its chain, ending in `en`.

| Selected | Chain | Why |
|---|---|---|
| `az` | `az`, `tr`, `en` | Azerbaijani and Turkish share Latin script and are largely mutually understood, so a Turkish string is a better stand-in than English for most `az` players. |
| `tr` | `tr`, `en` | The relation is not symmetric in practice: a Turkish player gets the neutral language rather than a text they did not choose. |
| `ru` | `ru`, `en` | No related launch language. |
| `en` | `en` | Final fallback. |

The chains are admin config (`LocaleFallbackChains`) and `en` is always last. The owner may change them without a release.

## 2. Text system

### 2.1 Keys and values

| Item | Rule |
|---|---|
| Key format | Stable dot-separated lowercase ids, `area.screen.item`, for example `match.result.win`. Characters: `a-z`, `0-9`, `.`, `_`. |
| Stability | A key is never renamed or reused for a different meaning. A retired key is marked deprecated in admin and removed once no supported client references it. |
| Value | A UTF-8 string in the ICU MessageFormat subset (4.3). |
| Ownership | Values are managed in admin as key-value per language, with the `en` reference and a per-language status (missing, draft, approved). Only approved values enter a published bundle. |
| Context | Each key has a translator note: where it appears, max length, what each placeholder is. |

Max length counts the visible characters of the longest text the value can produce: literal characters after NFC, combining marks not counted, placeholders and `#` not counted, and for a plural or select the longest branch.

Implemented in `server/internal/gamestrings` (#214): the key format, entries with note, max length and typed placeholders (the param types of section 3), value checks for the 4.3 subset with exactly the declared placeholders, plural only on `integer` or `decimal`, select only on `string`, every 4.2 category plus `other`, `=N` branches allowed, and the max length above; a bundle per language built from approved, non-deprecated values that is blocked as a whole when any of them fails a check, encoded as JSON with a SHA-256 content hash; and the missing-value report per language with percentage and key list.

Storage (migration `00011_game_strings`): entries and their values per language are stored with who changed them and when. A value may be saved as a draft that still fails its checks, so it can be fixed in the editor; it is approved only if it passes them. Publishing a language builds its bundle from the stored entries, refuses it with the list of failing keys, and otherwise stores it as the next version of that language with its content and hash. Publishing unchanged content returns the existing version instead of a new one. Stored bundles are append-only, so every earlier version stays available for rollback (2.2). The editor, import and export, the audit log entries, upload to the CDN and the active version switch follow in #214.

### 2.2 Delivery

| Item | Rule |
|---|---|
| Bundle | One versioned bundle per language, published from admin and served from the CDN as an immutable content-hashed file (ADR 0002). The active version per language comes from the server after the handshake (ADR 0004); URLs come from server config. |
| Built-in bundle | The install package ships a fallback bundle for all four languages, built from the approved set at build time. The client works offline and on first start with it alone. |
| Update | The client downloads the active bundle for the selected language and `en` in the background and swaps it in at the next screen load, never mid-match. A failed download keeps the previous bundle. |
| Rollback | Switching the active bundle version in admin. |
| Hot path | Bundles load as one table per language. Strings for a live screen are formatted when the screen opens or a value changes, never per frame (zero GC alloc budget in `CLAUDE.md`). |

### 2.3 Missing key behaviour

| Step | Condition | Behaviour |
|---|---|---|
| 1 | Key present in the selected language (downloaded bundle, else built-in) | Use it. |
| 2 | Missing | Try the next language in the chain (1.2). |
| 3 | Missing in the whole chain including `en` | Use `common.text.unavailable` from the built-in `en` bundle. If that is missing too, show an empty-state glyph, not words. |
| Logging | Any lookup that reaches step 2 or later | Log once per key per session (key, language, bundle version) and send it to telemetry so admin can list untranslated keys. |
| Release builds | | A raw key is never shown to a player. |
| Debug builds | | Show `[!key]` so testers see the gap. |

A value that fails to parse (bad placeholder syntax) is treated as missing for that lookup only.

## 3. Server-originated text

The server never sends rendered text for display. It sends `key` plus typed parameters and the client renders in the player's current language, so a language change re-renders stored notices.

| Rule | Detail |
|---|---|
| Shape | `{ key, params: { name: typed value } }`. Param types: integer, decimal, string, timestamp (UTC instant), duration (seconds), item id. |
| Items and currencies | Sent as ids. The client resolves the display name from the localized catalog, not from the notice. |
| Numbers | Sent raw. Formatting per locale is the client's job (4.4). |
| Applies to | Rewards, system notices, penalty notices and appeal outcomes (`docs/penalties.md`), error and rejection reasons, rank changes. |
| Errors | Responses carry a stable reason code that maps to a key, never prose. |
| Player-supplied strings | Nicknames and custom room names pass through as parameters, are not translated, and are inserted by placeholder (4.7). |
| Unknown key (newer server, older client) | Handled by 2.3 with the generic `notice.generic` key, never the raw key. |
| Server logs | English; this section covers player-facing text only. The admin panel UI has its own languages (`docs/admin.md` section 10). |

## 4. Text rules

### 4.1 Text expansion budget

Layouts are designed against the longest language. Admin enforces a max length per key.

| Language | Typical length vs `en` | UI design budget |
|---|---|---|
| `en` | 100% | Reference |
| `az` | about 110% | 130% |
| `tr` | about 115% | 130% |
| `ru` | about 120% to 130% | 150% |

Short labels (buttons, tabs, badges) expand worst: design them to 200% or allow wrap or shrink within admin-set limits. Every screen is reviewed in `ru` and `tr` before it ships.

### 4.2 Plurals

Plural form is chosen by CLDR category in the message, never by code testing for `1`.

| Language | Categories | Note |
|---|---|---|
| `en` | `one`, `other` | |
| `ru` | `one`, `few`, `many`, `other` | `one`: 1, 21, 31 (not 11). `few`: 2-4, 22-24 (not 12-14). `many`: 0, 5-20, 25-30. `other`: fractions. |
| `az` | `one`, `other` | After a number the noun stays singular (`5 oyun`); write count messages with the singular noun in `other`. |
| `tr` | `one`, `other` | Same: no plural agreement after a number. |

Admin rejects a plural message that lacks a category its language needs plus `other`.

### 4.3 Placeholders

ICU MessageFormat subset: `{name}`, `{count, plural, one {...} few {...} many {...} other {...}}`, `{kind, select, ...}`.

- Names are descriptive (`{opponentName}`, `{coins}`) and identical across languages for one key.
- Translators may reorder placeholders; code never assumes order.
- A value uses exactly the placeholders declared for its key; admin rejects missing or extra ones.
- Placeholders carry typed values; the formatter, not the translator, formats numbers, dates and durations.

### 4.4 Number, date and duration formatting

| Item | Rule |
|---|---|
| Locale source | The selected game language, not the device locale, so one screen is internally consistent. |
| Numbers | Locale grouping and decimal separators (for example `1 234,5` in `ru`, `1.234,5` in `az` and `tr`, `1,234.5` in `en`). Digits are always Latin. |
| Dates and times | Locale order, 24-hour clock. Timestamps arrive in UTC and display in the device time zone. Time logic never uses the device clock. |
| Durations | Built from localized unit keys with plural rules (4.2), not concatenated fragments. |
| Compact numbers | Localized compact units only where a layout key allows them. |

### 4.5 Font glyph coverage

| Requirement | Detail |
|---|---|
| Latin | `a-z` plus `ə Ə ğ Ğ ı İ ş Ş ç Ç ö Ö ü Ü` (az, tr). |
| Cyrillic | The Russian alphabet including `ё Ё`. |
| Shared | Digits, punctuation, currency signs, UI symbols. |
| Display fonts | Every stylised or effect font contains all of the above, or has a per-script fallback font. A missing glyph (tofu box) blocks release. |
| Atlas | Dynamic atlases for text; static atlases only for fixed numeric displays. Font data ships in the install package or an Addressables group within the budget in `CLAUDE.md`. |
| Check | A CI test checks font assets against these sets and against every character in approved bundle values. |

### 4.6 Casing: the Turkish i trap

| Language | Pairs |
|---|---|
| `tr`, `az` | `i`/`İ` and `ı`/`I` |
| `en`, `ru` | `i`/`I` |

- Every case change of display text uses the selected language's culture (`ToUpper(culture)`, `ToLower(culture)`). Invariant casing is only for identifiers and keys.
- A bare `ToUpper()` or `ToLower()` on display text is a review and lint item: in `tr` and `az` it turns `i` into `I`.
- Headlines that must be uppercase are stored as their own approved value rather than uppercased at runtime.
- Case-insensitive comparison of player text uses the normalisation in 5.2, not display casing.

### 4.7 No sentence concatenation

A sentence is one key with placeholders. Code never joins fragments (`name + " " + verb + count`): word order, case endings and plurals differ per language. Lists use a localized list-pattern key. `az`, `tr` and `ru` inflect nouns and names, so keys that insert a player name are phrased without inflecting it (a label followed by the name).

## 5. Nickname policy

The server alone validates, normalises and stores nicknames. The client may pre-check for responsiveness but never decides.

Implemented in `server/internal/profile` (#59): 5.1 (`NicknameAllowedSet` is fixed to the sets below, not configurable yet), 5.2 steps 1 to 4 and uniqueness on the step 4 form, 5.4, the format and taken codes of 5.5 with `nickname.invalid_separators`, `nickname.mixed_scripts`, `nickname.digits_only` and `nickname.cooldown` added, the registration default, and the frequency and history rows of 5.6. The word filter, reserved names and `nickname.not_allowed` are #60, the rate limit is #184, the rename price is #101. The admin names map to the server variables `HD_NICKNAME_MIN_LENGTH`, `HD_NICKNAME_MAX_LENGTH`, `HD_NICKNAME_ALLOW_MIXED_SCRIPT`, `HD_NICKNAME_RENAME_COOLDOWN` and `HD_NICKNAME_RELEASE` until admin config serves them.

### 5.1 Format

| Item | Rule | Proposed default (admin config) |
|---|---|---|
| Length | Grapheme clusters, after normalisation | min 3, max 16 (`NicknameMinLength`, `NicknameMaxLength`) |
| Allowed characters | Letters of Latin (including `ə ğ ı İ ş ç ö ü`) and Cyrillic, digits, and `_`, `.`, `-`, single space | `NicknameAllowedSet` |
| Not allowed | Emoji, symbols, control and format characters (zero-width, bidi overrides), more than one combining mark per letter, leading or trailing separators, more than two identical separators in a row | fixed |
| Mixed scripts | Latin and Cyrillic letters may not be mixed within one word | `NicknameAllowMixedScript` = off |
| Digits only | Rejected | fixed |

### 5.2 Normalisation before filtering and uniqueness

Applied in order. The stored display form is the player's text after step 1; the later steps produce comparison forms only.

| Step | Operation |
|---|---|
| 1 | Unicode NFKC. |
| 2 | Strip zero-width and format characters, collapse whitespace, trim. |
| 3 | Full Unicode case folding, locale independent, then `İ`, `I`, `ı` and `i̇` (i with combining dot) all map to `i`. Standard folding alone keeps `İ` apart from `I`, so this explicit step stops Turkish casing from dodging uniqueness. |
| 4 | Confusable folding to a canonical letter from a maintained table. Includes Latin/Cyrillic look-alikes (`а`/`a`, `е`/`e`, `о`/`o`, `р`/`p`, `с`/`c`, `у`/`y`, `х`/`x`, `і`/`i`) and the az/tr set (`ə`/`a`, `ı`/`i`, `ö`/`o`, `ü`/`u`, `ş`/`s`, `ç`/`c`, `ğ`/`g`). |
| 5 | Leetspeak folding (`0`->`o`, `1`->`i`, `3`->`e`, `4`->`a`, `5`->`s`, `7`->`t`, `@`->`a`, `$`->`s`). |
| 6 | A second form with separators (`_ . - space`) removed, so split spellings still match the word filter. |

The word filter and reserved names match on the step 5 and 6 forms. Uniqueness uses the step 4 form (leetspeak digits are legitimate in names, so they are not folded for uniqueness).

### 5.3 Filtering and reserved names

| Item | Rule |
|---|---|
| Word filter | One list per language (`az`, `en`, `tr`, `ru`) plus a language-independent list, maintained in admin with entry, match type (substring or whole word) and a list version. All lists apply to every nickname regardless of the player's language. |
| Categories | Profanity, sexual content, hate and slurs, violence and self-harm, drugs, personal data patterns (phone, email, URL), impersonation. The lists are owned by the owner or moderators; this document does not contain the words. |
| Reserved names | Staff and system identities (for example `admin`, `moderator`, `support`, `system`, the game name, and localized equivalents) matched after folding, including as a substring. Only staff accounts may hold them. Admin config `ReservedNames`. |

### 5.4 Uniqueness

Unique across the game on the step 4 comparison form. Check and write are one atomic server operation so concurrent requests cannot both win. `Ayşe` and `Ayse` are both stored as typed but cannot coexist. A released nickname (rename or deleted account) becomes available after `NicknameReleaseDays`.

### 5.5 Rejection

| Case | Result |
|---|---|
| Format violation | Reason code (`nickname.invalid_length`, `nickname.invalid_characters`, ...), shown as a localized message. |
| Taken | `nickname.taken`; the server may offer suggestions. |
| Filter or reserved hit | One generic `nickname.not_allowed`. The matched word or list is never revealed. |
| Repeated attempts | Rate-limited per account and device (`NicknameAttemptsPerHour`). |
| Registration | Cannot continue without an accepted nickname. A server-generated default is offered as a one-tap option. |

### 5.6 Renaming

| Rule | Detail |
|---|---|
| Frequency | One free rename at setup. Later renames are limited by `NicknameRenameCooldownDays` and may cost currency (price is economy config, not set here). |
| Moderator rename | After a confirmed report a moderator forces a rename. The player cannot enter ranked matches or custom rooms until they pick a valid name. A forced rename does not start the cooldown. |
| Rule changes | When the filter or confusables table changes, stored names are revalidated lazily at login; new violations force a rename. |
| History | Previous nicknames are kept per account for moderation and never shown to players. |

### 5.7 Scope: no free-text chat

In-game communication is emoji and quick messages only (Epic 13, #149). There is no free-text chat, so the filter applies to exactly two player-authored strings.

| String | Rules |
|---|---|
| Nickname | 5.1 to 5.6. |
| Custom room name | Same normalisation (5.2) and word filter (5.3). Own limit `RoomNameMaxLength` (proposed default 24). No uniqueness. A rejected name is replaced by a generated default. |
| Emoji and quick messages | A fixed catalog of keys (section 2). Only ids travel over the wire, never player text. |

Any future feature that accepts free text must update this section and add moderation before it ships.

### 5.8 Reports

| Step | Rule |
|---|---|
| Entry | A report action on a player profile and on the match end screen with fixed reasons: offensive nickname, offensive room name, impersonation, other. No free-text comment. |
| Record | Reporter, target, reason, the name text at report time, match id. Repeat reports by one reporter on one target merge. |
| Review | Reports go to a queue in the admin panel. A moderator approves the name, forces a rename, or adds a pattern to the word filter. Automatic action thresholds are admin config (`ReportAutoRenameThreshold`). |
| Abuse | Reporters with many rejected reports are rate-limited. |
| Feedback | The reporter gets a generic localized acknowledgement, not the outcome. The reported player gets a localized notice (reason key plus the rename requirement). |

## 6. Cultural content checklist

Applies to every art asset, animation, sound, store screenshot and text before it ships. Launch markets are Azerbaijan, Turkey and Russian-speaking regions other than Russia, which is not a store country at launch (`docs/privacy.md` 9.3); content must also pass the global store review.

| Area | Rule |
|---|---|
| Colours | UI, effect, team, score and role colours come only from colour tokens. Cosmetic illustration art (avatars, characters, tables, frames, emoji) may use its own palette, but is never dominated by a team or role hue. No colour used to evoke a national flag, party or political side. A status colour is never the only signal; add shape or label. |
| Symbols | No national or party emblems, no military insignia, no symbols tied to banned organisations. |
| Flags and maps | No flags as decoration and no flag standing in for a language (the language list uses names). No maps in launch content. Any later use needs owner review for disputed territory. |
| Religion | No religious symbols, scriptures, places of worship or clergy in assets or cosmetic themes. Holiday events are reviewed per region by the owner. |
| Gestures | No avatar or emote with a hand or body sign that is offensive in any launch market. The emote list is reviewed for this. |
| Gambling imagery | Dominoes with coins must not read as a casino. Avoid slot and roulette cues, card suits, chips, cash piles, spinning-reel reward presentation and the words jackpot, bet, wager. Currency is abstract game tokens. Random rewards follow store disclosure rules where they apply. |
| Alcohol, tobacco, drugs | None in assets, avatars, items or text. |
| Violence and fear | Cartoon impact effects only. No weapons, blood or real tragedies. |
| Sexual content | No sexualised characters, poses or outfits. Avatars are all-ages. |
| Representation | No ethnic, regional, gender or disability stereotypes in characters or names. AI-generated people are checked for artefacts and for resemblance to real people. |
| Likeness and brands | No real persons, brands, third-party game names or look-alike styles. The art source (including AI-generated) is recorded in the manifest. |
| Text in art | No text baked into images; every word is a key (section 2). Numerals and icons read the same in all four languages. |
| Motion and sound | Flashing stays within the photosensitivity limits (admin config) and a reduce-effects setting exists. No sudden loud sounds. |
| Wording | A word acceptable in one launch language but offensive in another is not used. A native speaker reviews each language's text. |

### 6.1 Review step

| Step | Who | When |
|---|---|---|
| 1. Self-check against section 6 | Asset author | Before upload to admin |
| 2. Native-speaker review of text values (`az`, `tr`, `ru`, `en`) | Reviewer per language | Before a value is set to approved |
| 3. Cultural review of art and sound | Owner or delegate | Before the manifest entry can go live; recorded as approved in the manifest |
| 4. Store-facing set (icon, screenshots, descriptions, trailer) | Owner | Before each store submission |

The pipeline does not publish an asset without an approved cultural review. A shipped asset found in violation is switched off from admin, its static fallback included, and fixed before it returns.

## 7. Age rating targets

A proposal for the owner to confirm. The stores' questionnaires decide the final rating from the real content at submission.

| System | Proposed target | Condition |
|---|---|---|
| IARC (Google Play and others) | Lowest tier (all ages) | No simulated gambling; no free-text chat; purchases and random items disclosed |
| PEGI | 3 | Same; in-app purchase and random-item notices may apply |
| App Store | 4+ | No gambling imagery; no unrestricted user text; purchase and random-reward answers given honestly |
| Turkey | Follow the local classification shown on the store listing | Checked at submission |

Features that would raise the target and need an owner decision before they are built: free-text chat, voice chat, real-money prizes, casino-style reward presentation. Nicknames and room names are user-generated text, so the store answer is "moderated user content" and the report flow (5.8) must exist before submission.
