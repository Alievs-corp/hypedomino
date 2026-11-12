# tile.sema

The sample item of `docs/asset-workflow.md` section 6: the sky tile back skin of the approved prototype (`table-v2`, owner decision 2026-11-12, board panel B-4). Every line below is either read from a file or marked as not recorded. Recorded on 2026-11-12.

- artSource: ai_assisted
- Authors: concept generated in an agent-driven ChatGPT browser session for the owner on 2026-11-12; no person has done W-3 or W-4 yet
- Generator: ChatGPT, image model `gpt-image`, as named by the C2PA manifest embedded in `concept/20261112-01.png` (see below). Account holder and plan: not recorded (RC-1 open)
- Terms: not recorded (RC-1 open, O-2)
- Canonical concept: concept/20261112-01.png
- Inputs uploaded to the generator: not recorded
- Prompt: not recorded. The prompt was typed into the ChatGPT conversation but not copied here; the account holder can copy it from that conversation's history (generated 2026-11-12 11:29:26 UTC)
- Similarity check: not done (RC-3 open, O-7)
- Human work: none yet. The two `work/` files are mechanical: a downscale and a 3D render (below). W-3 clean-up has not started (RC-5 open)
- Self-check (docs/localization.md 6.1 step 1): not done
- Cultural review: not done. Point for the reviewer: the central emblem is a crescent with a gem, surrounded by four-point stars. `docs/localization.md` section 6 (Symbols, Religion) forbids national emblems and religious symbols; whether this crescent reads as one in the four launch markets is the reviewer's call

## Embedded C2PA manifest

`concept/20261112-01.png` carries a C2PA manifest in a `caBX` PNG chunk. Read as text on 2026-11-12; the signature was not validated with a C2PA tool.

| Field | Value |
|---|---|
| Action | `c2pa.created`, 2026-11-12T11:29:26Z |
| Software agent | `ChatGPT`, version `gpt-image` |
| Digital source type | `trainedAlgorithmicMedia` (IPTC) |
| Further actions | `c2pa.converted`, `c2pa.watermarked.unbound`, same time |
| Claim generator | `OpenAI Media Service API` |
| Signer | OpenAI OpCo, LLC, certificate issued by SSL.com C2PA ICA R1 2025 |

## Files

| File | sha256 | What it is | How it was made |
|---|---|---|---|
| concept/20261112-01.png | `f6fb7abb6f8148d76c89be1b3a099f6afe342911f06b7c7b86876dec62673ef3` | 887 x 1774 RGB, the generator's download, unchanged | Byte copy of prototype `gen/hd-back-sema.png` |
| work/sema_texture.png | `0605711fe48cdf668c9aa619187267bb5f930e1e88d7185dd8f6808224c72dc0` | 400 x 800 RGB texture for the render | Lanczos downscale of the concept, no other change; identical pixel for pixel to a Pillow 12.1 `LANCZOS` resize of the concept. Byte copy of prototype `gen/skin-sema.png` |
| work/sema_render.png | `f091c3f9b6c01c8d5db78c68f3634c901bd3aee4e2016f0c87d17ad536e994ee` | 300 x 560 RGBA render of a tile with the skin on its back | Blender Cycles, prototype script `render_tiles.py` (sha256 `16cdd236056576dc42438a2179db3e792f874f315cc9c582eac9b9f5475340ec`), mode `back`, texture `work/sema_texture.png`: 96 samples, denoised, orthographic camera, transparent film, the modelled ivory tile body. File metadata: rendered 2026-11-12 15:30:32 local time. Byte copy of prototype `out/back_sema.png` |

Reproduced on 2026-11-12 with Blender 5.2.2 LTS (`blender -b --factory-startup -P render_tiles.py -- <out> back work/sema_texture.png sema`): the new render matches `work/sema_render.png` with a mean channel difference of 0 and a maximum of 9 of 255 (denoiser noise).

Not imported: prototype `gen/preview-sema.jpg`, a 300 x 600 JPEG of the same image (no other use), and `web/img/back_sema.png`, the render cropped to 264 x 514 for the prototype page.

| Date | File | Prompt or edit |
|---|---|---|
| 2026-11-12 | concept/20261112-01.png | ChatGPT `gpt-image`; prompt not recorded |
| 2026-11-12 | work/sema_texture.png | Lanczos downscale of the concept to 400 x 800 |
| 2026-11-12 | work/sema_render.png | `render_tiles.py` mode `back` with `work/sema_texture.png` |
