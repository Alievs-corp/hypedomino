# Content delivery

How the client gets art that is not in the install package: where it lives on the CDN, how the client learns which content is live, what it does offline, how it downloads and retries, and when it frees memory. The decision behind it is ADR 0002 (Addressables, CDN, ASTC). Addresses, groups and labels of cosmetics are `docs/cosmetics.md` 8. This document is #68; its section 7, the disk cache, is #160.

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

`<platform>` is `Android` or `iOS`. `content_base_url` is an `https` URL without a trailing slash, query or fragment. `catalog_version` is 1 to 64 letters, digits, `.`, `_` or `-`, starting with a letter or digit, because it is part of a file name and a URL; the server refuses any other value at start. Language bundles share the same root (`docs/localization.md` 2.2); their path is #214's.

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
| CD-12 | A failed catalog download is retried after 2 s, 8 s and 30 s, then fails (CD-14). A failed bundle download is retried twice at once by Addressables, then fails (CD-13). A catalog request that has not finished in 20 s, or a bundle request with no progress for 20 s, counts as failed. These are engineering limits, not design values, so they are constants in the client. |
| CD-13 | A failed cosmetic shows its static fallback, or the built-in static of its type when the static itself failed (`docs/cosmetics.md` 8.3). It is asked for again the next time a screen opens, never in a loop. Input never waits for a download. |
| CD-14 | A failed catalog download keeps the stored catalog and is tried again at the next accepted `ServerHello`. |

## 5. Releasing memory

| Id | Rule |
|---|---|
| CD-15 | Every load is paired with exactly one release. Views do not load directly; they acquire from a shared cache that counts references (`CosmeticArtCache`, `docs/cosmetics.md` 8.3) and release when hidden or reused. |
| CD-16 | The asset is released when its last reference goes. Addressables then unloads the bundle when no loaded asset needs it. |
| CD-17 | A screen releases everything it acquired when it closes. A match releases its seats' art when the match scene unloads. |
| CD-18 | Install-time content follows the same rules; only its download step is missing. |

## 6. Release build

| Id | Rule |
|---|---|
| CD-19 | A release build is made for one catalog version. It names the catalog `catalog_<catalog_version>.bin` and fails when the version does not match 2 or the catalog file is missing after the build. |
| CD-20 | One build makes both the install package and the remote folder `client/ServerData/<platform>`. They are one release candidate: the package's install-time catalog and the remote folder come from the same content. |
| CD-21 | Every release candidate's install size is measured and gated as `docs/performance.md` 3 says. The remote folder is listed beside it, not counted. |

On Android the `Release build` workflow (`.github/workflows/release-build.yml`) runs `HypeDomino.Editor.ReleaseBuild.BuildAndroid` for a given catalog version. It keeps the AAB, the remote folder and the size report as artifacts.

## 7. Disk cache

Remote bundles are kept in Unity's AssetBundle cache under the app's data folder; install-time bundles are part of the install and never in it. This section is #160.

| Id | Rule |
|---|---|
| CD-22 | The cache holds at most 1,000,000,000 bytes. When a new bundle does not fit, Unity first removes the bundles unused the longest. |
| CD-23 | A bundle unused for 30 days is removed. |
| CD-24 | After a remote catalog loads, at start or after a new version (CD-5), every cached bundle that no loaded catalog references is removed, so old versions of an item do not stay on disk. Nothing is removed when the stored catalog failed to load. |
| CD-25 | The cache size shown to the player is the cached bundles plus the stored catalog files. |
| CD-26 | The player may clear the cache outside a match only. A clear removes every cached bundle and keeps the stored catalog. A bundle in use blocks it, so the clear runs at the next start instead, before any content loads, and the player is told so. Art needed afterwards downloads again the first time a view asks for it (CD-10). |

The limit and the 30 days were set by the owner on 2026-11-08. They protect the device's storage and are not design values, so they are constants in the client.

## 8. Not here

- Uploading a built catalog and its bundles to the CDN and activating a version: #135.
- The iOS release build and its App Thinning Size Report: with the iOS pipeline.
- The settings row that shows the cache size and the clear button: #157 (`ContentBoot.Cache`).
