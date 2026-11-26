# assets-src

Source art: Blender, Adobe and AI-generated files. Never shipped directly; exported assets go through the pipeline into `client/`.

Heavy binary formats are stored with Git LFS (see `.gitattributes`) until they move to Unity Version Control, where an item's `files.json` pins each binary by hash and changeset (ADR 0015, `docs/asset-workflow.md` 4.3). Run `git lfs install` once before committing here. Naming, folders and workflow: `docs/asset-workflow.md`; check with `node tools/asset-naming/asset-naming.mjs`.
