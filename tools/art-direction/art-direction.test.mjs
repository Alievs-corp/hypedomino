import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { checkArtDirection } from "./art-direction.mjs";

const repoRoot = new URL("../../", import.meta.url);
const tokens = JSON.parse(readFileSync(new URL("docs/design-tokens.json", repoRoot), "utf8"));
const text = readFileSync(new URL("docs/art-direction.md", repoRoot), "utf8");

test("the committed art direction passes", () => {
  assert.deepEqual(checkArtDirection(tokens, text), []);
});

test("a hex colour is reported", () => {
  assert.deepEqual(checkArtDirection(tokens, "Skin is #F2C9A0."), ["colour literal #F2C9A0 found; name a token instead"]);
});

test("a functional colour is reported", () => {
  assert.equal(checkArtDirection(tokens, "Use rgb(10, 20, 30).").length, 1);
});

test("issue numbers are not colours", () => {
  assert.deepEqual(checkArtDirection(tokens, "See #5, #167 and #1234."), []);
});

test("an unknown token path is reported", () => {
  assert.deepEqual(checkArtDirection(tokens, "Tint with `color.role.skin.fill`."), ["unknown token color.role.skin.fill"]);
});

test("existing token paths and non-token code spans pass", () => {
  assert.deepEqual(checkArtDirection(tokens, "`color.shadow`, `glow.soft`, `effects.budgets.tilePlace`, `tilePlace`, `rarityId`"), []);
});
