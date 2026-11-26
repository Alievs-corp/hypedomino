# Legal document drafts

> **Draft for legal review. Not legal advice. Not published.** Nothing in this folder may be linked from the game, a store listing or a web page until counsel has reviewed it (`docs/privacy.md` L-28), every placeholder below is filled, and the owner has approved it on #195.

Player-facing drafts of the privacy policy, the terms of use, the purchase and refund terms, the promo code terms and the consent texts (#195; `docs/privacy.md` checklist items L-10 and L-11). They restate, in player language, what the repository already decides and builds. They add no rule of their own: where a draft and a spec disagree, the spec wins and the draft is corrected in the same PR.

## 1. Files

| Document | English master | Azerbaijani | Turkish, Russian | Covers |
|---|---|---|---|---|
| Privacy policy | [privacy-policy.en.md](privacy-policy.en.md) | [privacy-policy.az.md](privacy-policy.az.md) | Translation work (section 4) | GDPR Art. 13, KVKK Art. 10 notice, `docs/privacy.md` sections 2 to 8, L-10 |
| Terms of use | [terms-of-use.en.md](terms-of-use.en.md) | [terms-of-use.az.md](terms-of-use.az.md) | Translation work | Account, conduct, virtual items, sanctions and appeals, L-10 |
| Purchase and refund terms | [purchase-terms.en.md](purchase-terms.en.md) | [purchase-terms.az.md](purchase-terms.az.md) | Translation work | `docs/economy.md` EP-1 to EP-8, sections 7 and 9 |
| Promo code terms | [promo-code-terms.en.md](promo-code-terms.en.md) | [promo-code-terms.az.md](promo-code-terms.az.md) | Translation work | `docs/economy.md` section 8 |
| Consent texts | [consent-texts.en.md](consent-texts.en.md) | [consent-texts.az.md](consent-texts.az.md) | Translation work | `docs/privacy.md` 6.2, KVKK explicit consent separate from the notice, AZ-4 statement, L-11 |

The English file is the master. A change is made there first and carried to every translation in the same PR; a translation never states something the master does not.

## 2. Drafting rules

- Every statement traces to a decided fact: a row, rule or decision in `docs/privacy.md`, `docs/economy.md`, `docs/anti-cheat.md`, `docs/penalties.md`, `docs/localization.md` or `docs/push.md`. Section 6 lists the facts that the code does not deliver yet.
- Numbers that are admin config (retention days, limits, durations) are written only where the player needs them to understand a right or an obligation, with the value of `docs/privacy.md` section 4 at the time of writing. Before publication each is checked against the published production config (section 6, item V-1).
- No third-party game or company terms are named, quoted or adapted. Platform names (Google, Apple, Hetzner, Sentry) appear only as recipients of data, as `docs/privacy.md` section 2 names them.
- Placeholders are written `[[LP-n: what goes here]]` and listed in section 3. A published document holds none; `grep -rn --exclude=README.md "\[\[LP-\|\[\[COUNSEL" docs/legal` must print nothing before publication.
- Questions for counsel are written `[[COUNSEL: question]]` inside the text where the answer would go.

## 3. Placeholders (owner decisions and inputs)

| Id | Placeholder | Blocks | Source |
|---|---|---|---|
| LP-1 | Controller and operator: legal name, legal form, registered address, registration number | Every document | `docs/privacy.md` OQ-11, L-1; `docs/store-listing.md` SF-9, ST-1 |
| LP-2 | Privacy contact (email and postal address) and support email | Privacy policy, terms | L-1; `docs/launch-runbook.md` OI-13 |
| LP-3 | Data protection officer, if one is appointed | Privacy policy | OQ-11 (not researched) |
| LP-4 | EU representative (name, address, contact), needed only if the controller has no EU establishment | Privacy policy | EU-1, L-3 |
| LP-5 | UK representative, needed only if the controller has no UK establishment | Privacy policy | EU-1, L-3 |
| LP-6 | Turkish data controller representative and VERBIS registration, needed for a controller resident outside Turkey | Privacy policy (Turkey section) | TR-3, L-4 |
| LP-7 | Azerbaijan: whether the controller is an Azerbaijani entity; information system registration (AZ-2), licence answer (AZ-3) and the consent statement (AZ-4) | Privacy policy (Azerbaijan section), consent texts | AZ-2 to AZ-4, L-29 |
| LP-8 | Digital consent ages for Turkey and Azerbaijan; whether a birth month equal to the current month counts as having reached the age | Privacy policy (children), consent texts | OQ-14; #852 (age gate questions from #957) |
| LP-9 | Supervisory authority the player is pointed to first (depends on LP-1) | Privacy policy | GDPR Art. 13(2)(d), 77 |
| LP-10 | Governing law, courts and consumer dispute bodies (depends on LP-1) | Terms, purchase terms | Counsel |
| LP-11 | EU trader details shown on the store listing (address, phone, email) | Terms, purchase terms | `docs/store-listing.md` DC-10, ST-2 |
| LP-12 | Effective date and version of each document; the `PrivacyPolicyVersion` and `TermsVersion` values that go with them | Every document | `docs/privacy.md` section 11 |
| LP-13 | Published URLs: privacy policy, terms, account deletion web page, FAQ | Every document | L-27; `docs/store-listing.md` SF-11 to SF-13; #852 (#157, #957) |
| LP-14 | Name of the object storage and content delivery provider | Privacy policy | `docs/privacy.md` section 2 names the role, not the vendor |
| LP-15 | Rows for support tickets and their retention (the in-game support form and the support email) | Privacy policy | `docs/closed-beta.md` CE-11, TP-4; P-2 requires the row before the form ships |

## 4. Translation work

| Language | State | Needed by |
|---|---|---|
| `en` | Master draft | L-10 |
| `az` | Draft translation of the master, awaiting counsel and a native reviewer | L-10 |
| `tr` | Not started. Starts after counsel review of the master, so the KVKK wording is translated once; the Turkish notice (aydınlatma) and the explicit consent texts need a Turkish-qualified reviewer | L-10, L-11, before any player in Turkey (`docs/closed-beta.md` CE-10) |
| `ru` | Not started. Starts after counsel review of the master | L-10 |

## 5. Counsel questions

Collected from the `[[COUNSEL: ...]]` markers in the drafts.

| Id | Question | Where |
|---|---|---|
| CQ-1 | EU withdrawal right for digital content supplied at once: is the stores' checkout enough, or does the game need its own express consent and acknowledgement step before a purchase? No such step exists in the client today | Purchase terms 4 |
| CQ-2 | Wording of liability limits and warranty disclaimers that holds for consumers in the EU, UK, Turkey and Azerbaijan | Terms 13 |
| CQ-3 | What a player is owed for unused crystals and the current pass season if the service shuts down | Terms 12, purchase terms 8 |
| CQ-4 | Capacity of 13 to 17 year olds to accept the terms and to buy, per launch market, and whether a parent's agreement must be stated | Terms 2, purchase terms 3 |
| CQ-5 | Transfer mechanism, if any, for push delivery (FCM, APNs) and crash reporting processed outside the EEA | Privacy policy 9 |
| CQ-6 | Whether an in-app statement without an enhanced e-signature is valid consent under Azerbaijani law (AZ-4 residual risk) | Privacy policy 17, consent texts 4 |
| CQ-7 | Whether the promo code terms may exclude promo grants from consumer remedies, since nothing is paid for them | Promo code terms 7 |
| CQ-8 | Advance notice before a new version of the terms takes effect; today the game shows the new version at the next launch and continuing accepts it | Terms 14 |

## 6. Facts to verify before publication

The drafts describe the game as it will run at launch. These statements depend on work the repository does not deliver yet; each is either built or removed from the drafts before publication.

| Id | Statement in the drafts | Missing piece | Source |
|---|---|---|---|
| V-1 | Retention periods and other config values as written | Production config published with these values | `docs/privacy.md` section 4 |
| V-2 | Age gate before sign-in; no account below the minimum age | Age gate (step 2) not built | `docs/privacy.md` 6.3, L-14, #73 |
| V-3 | Consent toggles in Settings, Privacy | Client toggles not built | `docs/privacy.md` 6.3, #157 |
| V-4 | Sign in with Apple tokens revoked at deletion; deletion request sent to Sentry; ban-evasion hash | Not built | `docs/privacy.md` 8.2, L-17, L-18 |
| V-5 | Export from Settings, Account | Settings, Account screen not built | `docs/privacy.md` 8.3 |
| V-6 | Account deletion web page | Page and URL not live | `docs/privacy.md` 8.2, LP-13 |
| V-7 | Links to these documents from the first-launch notice and Settings | URLs not configured | `docs/privacy.md` 6.3, L-27 |
| V-8 | Purchases granted through store billing | Client store SDK not wired (`IStoreBilling` has no SDK yet) | `docs/economy.md` 7.5 |
| V-9 | Purchase consumption information is never sent to Apple | Holds while OQ-13 stays as decided | `docs/privacy.md` 12.2 OQ-13 |

## 7. Not covered

- Data processing agreements with processors (#195 "DPA", L-8): these are the processors' contracts, signed by the owner; nothing to draft here.
- Retention and destruction policy (L-5), record of processing (L-6), DPIA and legitimate interest assessments (L-7): internal documents owned on #187.
- Web pages (FAQ, account deletion page): `docs/privacy.md` has no row for their cookies or logs yet; the privacy policy covers them once a row exists.
- Closed beta tester data outside the game (sign-up form, invitation lists): rows are added before CB-2 (`docs/closed-beta.md` CE-11), then the policy gains a section for them.
