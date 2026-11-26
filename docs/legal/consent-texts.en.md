# HypeDomino Consent Texts

> **Draft for legal review. Not legal advice. Not published.** English master of the consent texts (#195, `docs/privacy.md` L-11). Source: `docs/privacy.md` 6.1 to 6.3, 7.4 and 12.2 AZ-4; `docs/push.md` PN-3. Placeholders `[[LP-n]]` and counsel questions are listed in [README.md](README.md).

Version: [[LP-12: version, the same `PrivacyPolicyVersion` as the policy these texts go with]]

These are the texts of the consent step (first launch, step 4 of `docs/privacy.md` 6.1) and of Settings, Privacy. Each purpose has its own text and its own switch, shown separately from the privacy notice and the terms of use, so that consent is specific and never bundled (KVKK explicit consent, GDPR Art. 7(2)). Every switch starts off. Accept and decline have the same weight on the screen. The texts become game string keys when they are approved (`docs/localization.md` section 2); no key exists yet.

## 1. Step introduction

> **Your choices**
>
> You can allow the uses below. Each one is optional and off until you turn it on. Saying no changes nothing in the game or in your rewards. You can change your mind at any time in Settings, Privacy.

## 2. Purposes

### 2.1 Analytics (`analytics`)

> **Help us improve the game**
>
> Allow HypeDomino to send events about how you use the game (for example screens opened, matches played, missions completed) to our own servers in Germany. Events carry a random analytics id, never your account id, email, nickname or IP address, and only we can link that id to your account. We use them to measure how many players come back and to improve the game. Raw events are deleted after 395 days. If you turn this off, sending stops at once and your analytics id is replaced. Details: privacy policy, section 5.

### 2.2 Diagnostics (`diagnostics`)

> **Send crash reports**
>
> Allow HypeDomino to send a report when the app crashes or hits an error: what went wrong in the code, your device model, operating system and app version. Reports carry no account id or analytics id and are stored by our processor Sentry in its EU data region for 90 days. We use them only to find and fix problems. If you turn this off, no further report is sent. Details: privacy policy, section 5.

### 2.3 Promotional notifications (`promotional_push`)

> **Event and offer notifications**
>
> Allow HypeDomino to send you notifications about events and offers. Game notifications, such as rewards ready and friend invites, have their own setting and do not need this. If you turn this off, no promotional notification is sent. Details: privacy policy, section 5.

## 3. Withdrawal confirmation

Shown in Settings, Privacy after a switch is turned off.

> Turned off. This applies from now on; what was sent before stays under the privacy policy until its retention period ends.

## 4. Azerbaijan consent statement (AZ-4)

[[LP-7: only if the controller is an Azerbaijani entity. Shown at sign-in to players whose account country is Azerbaijan, separately from the texts above, and recorded in the consent records. COUNSEL: CQ-6]]

Draft, with the content `docs/privacy.md` AZ-4 requires (Law No. 998-IIIQ Art. 8.2):

> **Consent to the processing of personal data**
>
> I agree that [[LP-1]] processes my personal data listed in section 3 of the privacy policy (account and sign-in data, device and connection data, game, purchase and fair play data) to provide the HypeDomino game, keep matches fair and fulfil legal obligations, by collecting, recording, storing, using, transferring and deleting it, and that it is transferred to and stored in Germany. This consent is valid while my account exists. I can withdraw it at any time by deleting my account in the game; data is then kept only as section 10 of the privacy policy states.
>
> [I agree] [Cancel]

## 5. Not shown

- Players below the digital consent age of their country (`docs/privacy.md` 7.4, [[LP-8]]) do not see sections 1 and 2; every purpose stays off.
- Nothing here is shown before the age gate. A player below the minimum account age never reaches this step.
