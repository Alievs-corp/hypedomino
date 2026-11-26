# HypeDomino Privacy Policy

> **Draft for legal review. Not legal advice. Not published.** English master of the privacy policy (#195, `docs/privacy.md` L-10). Source of every statement: `docs/privacy.md`. Placeholders `[[LP-n]]` and counsel questions are listed in [README.md](README.md).

Version: [[LP-12: version and effective date]]

This policy explains what personal data the HypeDomino game processes, why, on which legal basis, where it is stored, how long it is kept, who receives it, and how you use your rights. It applies to the HypeDomino app for Android and iOS and to the account deletion web page [[LP-13: deletion page URL]].

## 1. Who we are

The controller of your personal data is [[LP-1: legal name, legal form, registered address, registration number]] ("we", "us").

| Contact | Details |
|---|---|
| Privacy questions and requests | [[LP-2: privacy email and postal address]] |
| Data protection officer | [[LP-3: name and contact, or remove this row if none is appointed]] |
| EU representative | [[LP-4: name, address, contact, or remove this row if the controller is established in the EU]] |
| UK representative | [[LP-5: name, address, contact, or remove this row if the controller is established in the UK]] |
| Data controller representative in Turkey | [[LP-6: name, address, contact]] |

## 2. The short version

- We need your account, match and device data to run online play, keep matches fair and protect your account.
- Product analytics, crash reports and promotional notifications are off until you turn them on. Turning them off is one tap in Settings, Privacy, and changes nothing in gameplay or rewards.
- We show no ads, do not read your device's advertising id and do not track you across other apps.
- We do not sell your data. Our service providers process it only on our instructions.
- Your data is stored in the European Union: our servers are in Germany.
- You can download your data and delete your account in the game at any time.
- Accounts are for players aged 13 and older.

## 3. Data we process and why

Legal bases: "contract" means the data is needed to provide the game under the terms of use (GDPR Art. 6(1)(b); KVKK Art. 5(2)(c)). "Legal obligation" is GDPR Art. 6(1)(c) and KVKK Art. 5(2)(ç). "Legitimate interest" is GDPR Art. 6(1)(f) and KVKK Art. 5(2)(f), and for evidence in disputes also KVKK Art. 5(2)(e); the interests are named in each row. "Consent" is GDPR Art. 6(1)(a) and KVKK Art. 5(1) explicit consent, given in the game and withdrawable at any time (section 5).

| Data | Where it comes from | What we use it for | Legal basis |
|---|---|---|---|
| Account id, a random internal number | We create it | Linking everything below to your account; it also opens your public profile | Contract |
| Sign-in identity: the provider (Google or Apple) and its account identifier | The sign-in provider | Signing you in and finding your account | Contract |
| Email address (Apple may give a relay address) | The sign-in provider | Support and account notices | Contract |
| Session data: a hashed refresh token, the device the session is bound to, a resume token | We create it | Keeping you signed in and reconnecting you to a match | Contract |
| Install id, created by the app on first launch | Your device | Binding sessions to a device; detecting many accounts on one device | Contract; legitimate interest in preventing fraud and keeping matches fair |
| Device and app details: device model, operating system version, app version, build and protocol version | Your device | Checking that your app version can play, compatibility, fair play checks, support | Contract; legitimate interest in game integrity |
| IP address | Your connection | Delivering data to you, limiting request rates, finding your country if you did not choose one, fair play checks (we keep only a keyed hash of it for this) | Contract; legitimate interest in security and fair play |
| Account country | Your choice in your profile, otherwise derived from your IP address | Country and region leaderboards; age rules of your country | Contract |
| Nickname, earlier nicknames, names you give your custom rooms | You | Your identity in the game; moderation of names | Contract; legitimate interest in moderating names |
| Profile and progress: equipped cosmetics, level and XP of the current pass season, statistics, rank, rank points and season results | We compute it from your play | Your profile, progression and ranked play | Contract |
| Settings: language, sound, graphics, vibration, notification categories | You | Your settings | Contract |
| Match records: participants, seats, results, scores, events, the game configuration version | We record them | Running matches, results, ranking and rewards | Contract |
| Replay records: the shuffle seed, the deal and the moves of a match | We record them | Resolving disputes and reviewing collusion | Legitimate interest in fair play and handling disputes |
| Emoji and quick messages you send in a match (only the id of a fixed message) | You | Showing them to the other players in the room; they are not stored | Contract |
| Wallet and items: coin and crystal ledger, inventory, pass progress, missions, medals, daily rewards, promo code redemptions | We record them | Granting and keeping what you earned or bought | Contract |
| Purchases: store, product, transaction id, purchase token or receipt, purchase time, the account token bound to the purchase, refund and chargeback notices | Google Play or the App Store | Granting and restoring purchases, handling refunds, preventing payment fraud, keeping our books | Contract; legal obligation (bookkeeping); legitimate interest in preventing fraud |
| Offenses and penalties for idling or leaving matches | We record them | Fair play in ranked and casual matches (section 7) | Contract |
| Anti-cheat data: signals with their evidence, a risk score, device integrity results from Google Play Integrity or Apple App Attest, sanctions, appeals | We, Google, Apple | Preventing cheating and fraud (section 7) | Legitimate interest in game integrity and fraud prevention |
| Reports and moderation: who reported whom, the reason, the reported name, the match | Players and our moderators | Moderating names | Legitimate interest in a safe game |
| Friends, blocks, invites, online status | You and us | Friends, invites, blocking | Contract |
| Push notification token and notification categories | Google Firebase Cloud Messaging or Apple Push Notification service | Game notifications (contract); promotional notifications only with your consent | Contract; consent for promotional notifications |
| Consent records and your age gate answer (birth year and month only) | You | Proving what you consented to and applying the age rules | Legal obligation |
| Analytics events, keyed by a separate random analytics id | The app and our server | Understanding how the game is played, measuring retention and improving the game | Consent |
| Crash reports and app error reports: stack trace, device model, operating system, app version | The app | Finding and fixing crashes | Consent |
| Server logs and metrics | Our servers | Running the service and investigating security problems | Legitimate interest in operating and securing the service |
| Records of staff actions on your account | Our admin tools | Accountability of our staff | Legitimate interest in accountable support and moderation |
| Deletion records (section 13) | We create them | Re-applying your deletion after a backup restore; keeping bans effective | Legitimate interest in honouring deletions and preventing ban evasion |
| Data export files (section 14) | We create them | Giving you your data | Legal obligation |
| Database backups | We create them | Recovering from a disaster | Legitimate interest in not losing player data |

[[LP-15: support tickets from the in-game support form and the support email, with their purpose, legal basis and retention, once `docs/privacy.md` section 3 has the row]]

Most of this data is needed to play online: without it there is no account and no match. The data marked consent is optional.

## 4. What we do not collect

Your real name, phone number, postal address, precise location, contacts, photos, audio, files, payment card or bank data, the advertising id of your device, the list of your installed apps, your browsing history, and free text messages. There is no free text or voice chat in the game. We do not request your name or photo from Google or Apple; sign-in asks only for an identifier and your email address. We never read hardware identifiers such as IMEI, serial number or MAC address. We process no special categories of data (health, religion, ethnicity, biometrics, political opinion). The nickname filter rejects names that look like personal data.

## 5. Your consent choices

| Choice | What it allows | Default |
|---|---|---|
| Analytics | Sending analytics events about how you play, keyed by a random analytics id that only our server can link to your account | Off |
| Diagnostics | Sending crash reports and app error reports | Off |
| Promotional notifications | Promotional and event notifications | Off |

- You are asked once, after the privacy notice and before sign-in. Accepting and refusing are equally easy, nothing is pre-selected, and refusing changes nothing in gameplay or rewards.
- You can change each choice at any time in Settings, Privacy. Withdrawing takes effect at once: the app stops sending, and for analytics your analytics id is replaced, so later events cannot be linked to earlier ones. Withdrawal does not affect processing done before it.
- Nothing is sent for analytics or diagnostics before you turn it on; events from before your consent are dropped, not kept.
- Analytics events never carry your account id, email, nickname, IP address or install id. Analytics is ours: events go only to our own servers, and no third-party analytics service receives them.
- Players below the digital consent age of their country (section 11) are not offered these choices, and they stay off.
- Game notifications (rewards ready, friend invites, season reminders, service notices) are part of the game and have their own on and off setting. Your device asks for notification permission the first time a notification is useful, never at first launch.
- We keep a record of every choice: the choice, the policy version you saw, the time, where you made it and the app version.

## 6. What other players see

| Who | What |
|---|---|
| Everyone | Your nickname, equipped cosmetics, level, statistics on your profile, rank, rank points and leaderboard position, ranked wins on a leaderboard, the country board you appear on (which shows your account country), and the emblems and medals you choose to show on your profile. Your random account id opens your profile from a leaderboard |
| Players in your room | Your custom room name, seats, emoji and quick messages |
| Your friends | Whether you are online |
| Nobody | Your email, sign-in identity, hidden rating, install id, IP address, purchases, offenses, anti-cheat data and reports |

After your account is deleted, past matches and season results show a neutral placeholder instead of your nickname.

## 7. Fair play, anti-cheat and automated decisions

The server decides every move, score and reward, so a modified app gains nothing. To keep matches fair we record signals such as invalid requests, modified or unverified apps, device integrity results, clock manipulation, collusion and many accounts on one device. Each signal has a weight; the weights add up to a risk score that decreases over time.

| Measure | How it is decided | What it means for you |
|---|---|---|
| Review | Automatically, when the risk score reaches a review level or on strong evidence | A member of our staff looks at your account |
| Matchmaking with other flagged accounts and removal from public leaderboards | Automatically from the risk score, applied in periodic waves, or by staff | Ranked matches only with other flagged accounts; you are not notified, and your data export shows it |
| Temporary ban | Automatically from the risk score, applied in periodic waves, or by staff | No sign-in for a limited time |
| Ranked ban | By staff | No ranked matches or custom rooms for a limited time; sign-in and casual matches stay open |
| Permanent ban | Only by staff, from the review queue | No sign-in |
| Penalties for idling or leaving matches | Automatically, by fixed rules (warning, queue delay, matchmaking ban, review) | As described in the terms of use |

- A temporary ban from the risk score is an automated decision with a significant effect on you (GDPR Art. 22). We take it because reviewing every ban wave by hand is not possible at the volume of an online game, and it is needed to provide fair matches under the terms (Art. 22(2)(a)). You can appeal every sanction from the game; a member of our staff reviews the appeal, considers your view and can reverse the decision. Accounts of players under 18 never receive a temporary ban from the risk score; they go to staff review instead.
- When you are banned, the game tells you the reason category (for example cheating, collusion, idling or leaving, offensive name, payment fraud, other), never the signals or the score.
- Your data export and any access request show every sanction with its reason category, start and end, and the categories and times of the signals that led to it. We do not disclose the risk score, signal weights, thresholds, which check fired or raw evidence that would show how detection works, because that would let cheaters avoid it and harm other players (GDPR Art. 15(4)). When we withhold something we tell you what and why, and that you may complain to a supervisory authority.

## 8. Who receives your data

| Recipient | Role | What it receives |
|---|---|---|
| Google (Sign-In, Play Billing, Play Integrity, Firebase Cloud Messaging) | Independent controller for your Google account and payments; our processor for delivering push notifications | Sign-in token exchange, purchase verification, device integrity tokens, push token and notification content |
| Apple (Sign in with Apple, App Store, App Attest, Apple Push Notification service) | Independent controller for your Apple account and payments; our processor for delivering push notifications | The same as Google |
| Hetzner, Germany | Our hosting processor | Everything stored on our servers |
| [[LP-14: object storage and content delivery provider]] | Our processor | IP address and request details when your app downloads game content; data export files |
| Sentry, EU data region | Our processor | Crash reports, only with your diagnostics consent |

- Google Play and the App Store sell real-money items to you as merchant of record. We never receive your card or bank data.
- We do not sell personal data, do not share it for advertising and use no advertising or attribution services.
- Our staff see personal data only as their role needs: support and administrators see account data; analysts see pseudonymous analytics and figures without ids. Every staff action on an account is logged.

## 9. Where your data is stored

Our servers, databases, logs, analytics store and backups are in Germany, in the European Union. Crash reports stay in Sentry's EU data region. Google and Apple process the data they receive under their own terms and privacy policies. [[COUNSEL: CQ-5, transfer mechanism for push delivery and crash reporting outside the EEA, if any]]

For players in Turkey, each of our processors outside Turkey is covered by the standard contract of the Turkish Personal Data Protection Board, notified to the Authority as the law requires (KVKK Art. 9).

## 10. How long we keep data

Periods counted from an event run from that event. Expired data is deleted, or anonymized where the table says so, by a daily job.

| Data | Kept for |
|---|---|
| Account, sign-in identity, email, country, nickname, profile, progress, wallet, items, settings, friends | As long as your account exists |
| Session tokens | Access tokens 15 minutes; refresh tokens 30 days from last use |
| Install id and device details | 180 days after their last use |
| Server logs, including IP addresses in them | 30 days |
| Keyed hash of the IP address for fair play checks | 90 days |
| Match records, rank history | As long as your account exists, then anonymized |
| Replay records | 90 days, longer while an open case or appeal refers to them |
| Offenses for idling or leaving | Until the offense expires, then 90 days |
| Anti-cheat data | 365 days after the case closes or the sanction ends |
| Reports | 365 days after they are resolved |
| Purchase records | 10 years from the end of the year of the purchase, for bookkeeping; after account deletion without contact data |
| Consent records and age gate answer | As long as your account exists, then 3 years from the end of the year of deletion, as proof |
| Raw analytics events | 395 days; aggregate figures without any id are kept |
| Crash reports | 90 days |
| Records of staff actions | 730 days |
| Data export files | 72 hours |
| Backups, deletion records | 35 days |
| Ban evasion hash of a deleted, permanently banned account | 1095 days after deletion; for a temporary ban until the ban would have ended |

We never delete an account because it is inactive.

## 11. Children

- Accounts are for players aged 13 and older. Before sign-in the game asks for your birth year and month, without suggesting an answer. Below 13 no account is created, and the game explains why.
- We store only the birth year and month, never the day, so the rules can follow your country.
- Below the digital consent age of your country the consent choices of section 5 are not offered and stay off. That age is 16 in the EU unless a member state sets a lower age between 13 and 15. [[LP-8: digital consent age for Turkey and Azerbaijan, and how a birth month equal to the current month is counted]]
- The game sells nothing outside Google Play and the App Store, so their family and parental controls apply to every purchase.

## 12. Your rights

| Right | How to use it |
|---|---|
| Information | This policy, the privacy notice at first launch, section 6 |
| Access and portability | Download your data in the game (section 14) |
| Rectification | Change your nickname, country, language and settings in the game. Your email follows your Google or Apple account at your next sign-in. Correcting inaccurate data is free |
| Erasure | Delete your account in the game or on the web page (section 13) |
| Restriction and objection | Change your consent choices in Settings, Privacy. To object to processing based on legitimate interest, contact support |
| Human review of automated decisions | Appeal any sanction in the game (section 7) |
| Withdraw consent | Settings, Privacy, at any time |
| Complaint | You may complain to a supervisory authority, in particular in the EU country where you live or work. [[LP-9: supervisory authority we point to first, with contact]] |

- We verify requests through your signed-in session, or on the web deletion page by signing in with the same Google or Apple account. We do not accept identity from an email alone; if you write to us, we point you to these flows.
- We answer within 30 days, free of charge.

## 13. Deleting your account

1. Choose Settings, Account, Delete account in the game, or use the web page [[LP-13: deletion page URL]]. You sign in again to confirm. Deletion is not possible while you are in a match.
2. Your account becomes pending deletion at once: you are signed out on every device, removed from matchmaking, friends lists and current leaderboards, and your push tokens are deleted.
3. For 30 days you can cancel by signing in and choosing cancel; your account comes back unchanged.
4. After 30 days we delete your account data. Your coins, crystals, items and pass progress are lost. Deleting an account does not refund purchases.
5. What remains, without your nickname or contact data and linked only to a random replacement id: match results and season results as they appear to other players (with a placeholder name), purchase records and consent records for the periods in section 10, and anti-cheat, offense, report and replay records until their own periods end. If your account was banned, a keyed hash of your sign-in identity keeps the ban effective for its remaining time, or 1095 days for a permanent ban; a new account with the same sign-in identity inherits the ban.
6. We revoke your Sign in with Apple authorization, delete your raw analytics events and ask Sentry to delete crash reports linked to you. Backups rotate out within 35 days; if we ever restore one, your deletion is applied again before the restored data is used.

## 14. Downloading your data

Choose Settings, Account in the game to request a file with your data. We prepare one JSON file and give you a download link valid for 72 hours. You can request a new file once every 7 days.

The file contains your account id, email, country, nickname and its history, profile, settings, your own match records, rank state, wallet and items, purchases (without purchase tokens or receipts), offenses and sanctions with their reason categories and signal categories, friends and blocks by nickname, consent records and your analytics events. It never contains other players' account ids or hidden tiles, who reported you, the risk score, detection details, receipts, tokens or logs.

## 15. Security

Data travels encrypted (TLS with certificate pinning) and is encrypted at rest. Session tokens on your device are kept in the platform's secure storage. Staff access is limited by role and every action is logged. Logs never contain tokens, receipts, emails or raw request bodies. If a breach affects your data, we notify the competent authority within 72 hours and inform you where the law requires it.

## 16. Players in Turkey (KVKK Art. 10)

- Controller: [[LP-1]]. Data controller representative in Turkey: [[LP-6]].
- Purposes and legal bases: section 3, with the KVKK articles named there.
- Recipients and the purpose of each transfer, including transfers abroad: sections 8 and 9.
- Method of collection: automatically through the app and our servers, and from you, Google and Apple, as section 3 lists per item.
- Your rights under KVKK Art. 11: to learn whether your data is processed, to request information about it, to learn the purpose and whether it is used accordingly, to know the recipients in Turkey and abroad, to request correction, deletion or destruction and the notification of these to recipients, to object to a result against you that arises only from automated analysis, and to claim compensation for damage from unlawful processing. Send applications to [[LP-2]] or through the representative; we answer within 30 days, free of charge.
- Explicit consent is asked separately from this notice and from the terms of use, for each optional purpose (section 5).

## 17. Players in Azerbaijan

[[LP-7: if the controller is an Azerbaijani entity: the registration of the information system, the separate consent statement for processing and for the transfer to Germany shown at sign-in, its validity and withdrawal (withdrawing it means deleting the account). If not, remove this section or replace it with what counsel requires of a foreign controller. COUNSEL: CQ-6]]

## 18. Changes to this policy

Each version of this policy has a number. If a new version changes what data we process, why, or who receives it, the game shows the notice and your consent choices again at your next launch. Editorial changes are announced with a notice in the game.

## 19. Contact

[[LP-1]], [[LP-2]].
