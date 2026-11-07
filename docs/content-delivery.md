# Content delivery

How the client gets art that is not in the install package: where it lives on the CDN, how the client learns which content is live, what it does offline, how it downloads and retries, and when it frees memory. The decision behind it is ADR 0002 (Addressables, CDN, ASTC). Addresses, groups and labels of cosmetics are `docs/cosmetics.md` 8. The disk cache and its limits are #160. This document is #68.

## 1. Install-time and remote content

| Id | Rule |
|---|---|
| CD-1 | The install package holds the boot path, the match scene essentials, the default tile set and the `cosmetic-builtin` group (`docs/cosmetics.md` 8.2). With these alone the game starts, plays a match and shows every screen. |
| CD-2 | Every other Addressables group is remote: built with the catalog, uploaded to the CDN and fetched on demand. Install-time groups never point at the CDN. |
| CD-3 | Remote files are immutable. A bundle file name carries the hash of its content, so a new build of an item is a new file and an unchanged item keeps its file across catalog versions. |

## 2. Where content lives

The server announces two values in an accepted `ServerHello` (`proto/README.md`): `catalog_version`, the active catalog, and `content_base_url`, the CDN root of the environment (`HD_CATALOG_VERSION`, `HD_CONTENT_BASE_URL` in `docs/environments.md` 2.1). The client never builds a CDN host from a constant.

| Path under `content_base_url` | Content |
|---|---|
| `addressables/<platform>/catalog_<catalog_version>.bin` | The catalog of one version |
| `addressables/<platform>/<bundle file>` | Every remote bundle of every version, by its content-hashed name (CD-3) |

`<platform>` is `Android` or `iOS`. `content_base_url` is an `https` URL without a trailing slash, query or fragment. Language bundles share the same root (`docs/localization.md` 2.2); their path is #214's.

## 3. Catalog version

| Id | Rule |
|---|---|
| CD-4 | Start never waits for the network. Addressables starts with the install-time catalog, then loads the last remote catalog the device stored, if any, from disk (CD-6). |
| CD-5 | After an accepted `ServerHello` with a non-empty `catalog_version` and `content_base_url`, the client compares the version with the stored one. A different version is downloaded, stored and loaded. The same version does nothing. An empty value keeps what is loaded. |
| CD-6 | The client stores one remote catalog: the last one it loaded, with its version. A new one replaces it only after it downloaded completely, so a failed download keeps the previous one. |
| CD-7 | A new catalog is loaded outside a match only, like language bundles (`docs/localization.md` 2.2). Received during a match, it is loaded when the match ends. Art already on screen keeps its loaded handle; new loads resolve against the new catalog. |
| CD-8 | A rollback is the server announcing an older version (`docs/launch-runbook.md` 5.4). The client treats it as any other different version. |
| CD-9 | A new item published with a new catalog shows without a store update: its address resolves once the catalog is loaded (#68 acceptance). |

## 4. Download

| Id | Rule |
|---|---|
| CD-10 | A remote asset downloads when a view first needs it. Nothing is prefetched in bulk; a screen that must show content before it opens (the match loading screen, #85) asks for its addresses and shows their progress. |
| CD-11 | Progress is downloaded bytes over the total download size of the requested addresses. Addresses already on disk count as done. |
| CD-12 | A failed catalog or bundle download is retried after each delay of `content.retryDelaysMs` in `docs/design-tokens.json`, then fails. A request that takes longer than `content.timeoutSeconds` without progress counts as failed. |
| CD-13 | A failed cosmetic shows its static fallback, or the built-in static of its type when the static itself failed (`docs/cosmetics.md` 8.3). It is asked for again the next time a screen opens, never in a loop. Input never waits for a download. |
| CD-14 | A failed catalog download keeps the stored catalog and is tried again at the next accepted `ServerHello`. |

## 5. Releasing memory

| Id | Rule |
|---|---|
| CD-15 | Every load is paired with exactly one release. Views do not load directly; they acquire from a shared cache that counts references (`CosmeticArtCache`, `docs/cosmetics.md` 8.3) and release when hidden or reused. |
| CD-16 | The asset is released when its last reference goes. Addressables then unloads the bundle when no loaded asset needs it. |
| CD-17 | A screen releases everything it acquired when it closes. A match releases its seats' art when the match scene unloads. |
| CD-18 | Install-time content follows the same rules; only its download step is missing. |

## 6. Not here

- The disk cache, its size, the clear button and the removal of old versions: #160.
- Uploading a built catalog and its bundles to the CDN and activating a version: #135.
- Install size measurement: `docs/performance.md` 3, in #68's last pull request.
