# Push notifications

Scope: device tokens, notification categories and the server's send path (#132). Campaign definitions and scheduling live in the admin panel (`docs/admin.md` 5.13, #146). Personal data rows: `docs/privacy.md` D-11 and D-22.

## 1. Categories

| Category | Covers | Default | Legal basis |
|---|---|---|---|
| `gameplay` | Rewards ready, friend invites, season reminders, SEV1 notices (`docs/launch-runbook.md`) | On | Contract (D-22) |
| `promotional` | Promotional and event notifications | Off | `promotional_push` consent (`docs/privacy.md` 6.2) |

- **PN-1** A notification is sent only when the account's setting for its category is on. A category that is off sends nothing to any device.
- **PN-2** The OS permission is asked at the first contextual moment, never at first launch (`docs/privacy.md` 6.1). Without it the client does not register a token.
- **PN-3** Turning `promotional` on is the `promotional_push` consent; the client shows it with the other consent purposes and in Settings, Privacy (#157).

## 2. Devices

- **PN-4** A device registers its FCM (Android) or APNs (iOS) token after sign-in and whenever the OS rotates it. A token belongs to one account; registering it again under another account moves it.
- **PN-5** The client removes its token before logout. A token that FCM or APNs reports as invalid is deleted at once (D-22 retention).
- **PN-6** Payloads carry game string keys (`titleKey`, `bodyKey`), never resolved text; the device shows them in the player's language (`docs/localization.md`).

## 3. API

Player HTTP API with the access token (`server/internal/push`):

| Route | Effect |
|---|---|
| `PUT /push/devices` `{token, platform}` | Registers the token; `platform` is `android` or `ios` |
| `DELETE /push/devices` `{token}` | Removes the player's own token |
| `GET /push/settings` | `{gameplay, promotional}` |
| `PUT /push/settings` `{gameplay, promotional}` | Changes the categories given, keeps the others, returns all of them |

Errors: `push.token_required`, `push.unknown_platform`, `push.malformed` (400).

## 4. Sending

- **PN-7** The server sends through one sender per platform behind a single interface; tests use a fake. FCM and APNs credentials come only from the environment (`docs/environments.md`).
- **PN-8** Owner setup before real sends: a Firebase project with an FCM service account for the Android app, and an APNs authentication key (key id, team id, bundle id) for the iOS app.
