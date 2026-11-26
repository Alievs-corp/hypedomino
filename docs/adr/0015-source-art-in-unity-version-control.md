# 0015. Source art in Unity Version Control, code and file manifests in git

Status: Accepted; migration waits for the Unity Version Control organization (#751 step 1)
Date: 2026-11-27
Issue: #751

## Context
Source art (`assets-src/**`: concepts, layered masters, renders, exports, sound sources) is stored with Git LFS (`.gitattributes`). On 2026-11-16 HEAD held 2192 LFS files, about 1.1 GB; on 2026-11-27 `assets-src` alone held 2164 LFS files, 1,104 MB, and `client/Assets` 36 files, 28 MB. History adds every earlier version of each file.

The organization is on the GitHub Free plan: 10 GiB of LFS storage and 10 GiB of LFS bandwidth a month are included. Storage counts every version ever pushed. Over the quota with a budget of $0, LFS is blocked for the rest of the month; without a payment method, clones get pointer files only. The `assets` CI job, every clone and every artist pull spend the same bandwidth.

Owner decision 2026-11-16: source art is stored the way large game studios store it, binaries in a version control system built for them, with file locking, and code on GitHub. What reaches players does not change: Addressables bundles go to the content bucket and CDN at deploy (`docs/content-delivery.md` section 8, ADR 0002).

## Decision
- Source art binaries live in Unity Version Control (cloud), repository `hypedomino-art` of the owner's Unity Cloud organization. The repository root is the content of `assets-src/`: the server path of `assets-src/<path>` is `/<path>`.
- A binary is every file under `assets-src/` whose extension is in `sourceArtExtensions` of `tools/asset-manifest/asset-manifest.mjs`: the extensions of today's `assets-src` Git LFS rules. Migration moves the latest state only; the git history of those files stays where it is.
- Exclusive checkout (a lock rule on the server) for `*.psd *.blend *.ai *.aep`, the layered formats that cannot be merged.
- Git keeps everything else of an item: `provenance.md`, `work/*.py` scripts, `export/slice.json`, `assets-src/README.md`, and the item's file manifest `files.json` (`docs/asset-workflow.md` 4.3). The manifest lists each binary of the item with its path, SHA-256 and the Unity Version Control changeset of that content, so a git commit pins the exact art it was reviewed with, and a fetched file is checked against the hash.
- `tools/asset-manifest` writes and verifies the manifests; `tools/asset-export pull` fetches the export files of an item with `cm cat` at the listed changeset; the CI `assets` job pulls an item from Unity Version Control when it has `files.json` and from Git LFS otherwise, with the `UVCS_*` names of `docs/environments.md` 2.3. `tools/asset-naming` and `npm run asset:check` keep their rules; the naming check reads the listed paths as well as the files on disk.
- Imported client art (`client/Assets/**` PNG and WAV) stays in Git LFS (Client art below).

## Alternatives considered and why rejected
- Stay on Git LFS and pay for data: the storage of every pushed version keeps growing, artists keep a git client, and the cost is recurring with the catalog. Git LFS locking exists but needs every artist on the command line or a git GUI that supports it.
- Perforce P4: self-hosted is free up to 5 users but needs a server to run, patch and back up; the cloud offer costs about $39 per user a month.
- A separate git repository with LFS for art: the LFS quota is per organization, so it moves the problem without solving it.
- Unity Version Control with the whole `assets-src/` (scripts and provenance too): `provenance.md` is the rights record that must stay with the code for as long as the repository exists (`docs/asset-workflow.md` O-4), and scripts are reviewed in pull requests like code.

## Consequences
- Easier: the LFS quota stops growing with the art; artists get a GUI client with locks; the CI `assets` job fetches only the export files of the items a change touches.
- Harder: two systems. An art change is a Unity Version Control check-in and a git commit of the updated `files.json` (`node tools/asset-manifest/asset-manifest.mjs write --changeset <n> <item>`); a pull request without the manifest update shows no art change. CI needs the `UVCS_*` credentials, and a fork without them cannot check items stored there.
- Unity Version Control pricing (from 2026-03-01, https://unity.com/products/pricing-updates): no seat fee; 25 GB of storage and 100 GB of egress a month per organization are free, and access is blocked above them. CI pulls count as egress; the job pulls export files only, never `work/`.
- Must be verified once the organization exists: the `cm` sign-in of the CI job (`cm configure` with an access token), the lock rule, and a pull of one item end to end (`docs/asset-workflow.md` 4.3, migration steps).

## Client art
`client/Assets/Cosmetics/**/*.png` and `client/Assets/HypeDomino/UI/Audio/Clips/**/*.wav` stay in Git LFS. They are 36 files, 28 MB, and the Unity project must open from a git clone: the `client` CI job, the `Release build` workflow and every developer check out with LFS and would otherwise all need Unity Version Control access and a second checkout before Unity starts. `npm run asset:add` copies every new item's export into `client/Assets/Cosmetics`, so this grows with the catalog; the decision is revisited when the LFS storage of `client/Assets` passes half of the free quota (5 GiB) or LFS bandwidth runs out in a month.

## Git LFS clean-up options
Removing files from HEAD, or from history, does not lower LFS storage: GitHub keeps every pushed LFS object of a repository until the repository is deleted or GitHub Support purges it (https://docs.github.com/en/repositories/working-with-files/managing-large-files/removing-files-from-git-large-file-storage). No history rewrite is part of this decision; the options, for the owner to choose after the migration:

| Option | What it takes | Cost |
|---|---|---|
| A. Leave the objects | Nothing. Storage stays at its level; bandwidth falls as clones and CI no longer fetch `assets-src` | The old objects keep using the storage quota |
| B. Ask GitHub Support to purge the objects of `assets-src` | A support request naming the repository and the paths; Support states whether history must be rewritten first | None if Support purges without a rewrite; otherwise C |
| C. Rewrite history and purge | `git filter-repo` (or `git lfs migrate export`) over all branches, a force push, then B | Every commit hash changes: branch protection forbids force pushes and must be lifted for it, open pull requests and links to commits break, every clone and worktree is cloned again |
| D. Delete and recreate the repository | The only self-service way in the GitHub documentation | Issues, pull requests, stars, Actions history, secrets, environments and branch protection are lost or must be moved by hand |

Recommendation: A at migration, then B once the migration is merged and the storage figure is known; C or D only if Support refuses and the storage is a real cost.
