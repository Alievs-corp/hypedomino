import { test } from "node:test";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import { checkTokens, contrast, loadSpec, loadTokens, stepForPoints } from "./design-tokens.mjs";

const repoRoot = new URL("../../", import.meta.url);
const tokens = loadTokens(fileURLToPath(new URL("docs/design-tokens.json", repoRoot)));
const specText = loadSpec(fileURLToPath(new URL("docs/ui-motion.md", repoRoot)));
const raritySpecText = loadSpec(fileURLToPath(new URL("docs/rarity.md", repoRoot)));

function errorsAfter(change) {
  const copy = structuredClone(tokens);
  change(copy);
  return checkTokens(copy, specText, raritySpecText);
}

function assertReported(errors, fragment) {
  assert.ok(errors.some((error) => error.includes(fragment)), errors.join("\n"));
}

test("the committed tokens and spec pass every check", () => {
  assert.deepEqual(checkTokens(tokens, specText, raritySpecText), []);
});

test("contrast matches the WCAG 2 reference values", () => {
  assert.equal(contrast("#FFFFFF", "#000000"), 21);
  assert.equal(contrast("#777777", "#FFFFFF").toFixed(2), "4.48");
});

test("a text role below AA is reported", () => {
  assertReported(errorsAfter((copy) => { copy.color.role.danger.text = "#7A2030"; }), "contrast role.danger.text");
});

test("a fill without non-text contrast against the base is reported", () => {
  assertReported(errorsAfter((copy) => { copy.color.role.teamA.fill = "#1A1640"; }), "contrast role.teamA.fill on background.base");
});

test("a stale contrast table in the spec is reported", () => {
  assertReported(errorsAfter((copy) => { copy.color.text.muted = "#B0A9E6"; }), "spec contrast table is missing or stale");
});

test("a colour missing from the spec is reported", () => {
  assertReported(errorsAfter((copy) => { copy.color.role.teamB.text = "#FFB088"; }), "spec does not list colour #FFB088");
});

test("rarity ordinals with a gap are reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.mythic.ordinal = 6; }), "ordinals must be 1 to 5");
});

test("repeated rarity pip counts are reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.epic.pips = 2; }), "pip counts must be unique");
});

test("a rarity glow weaker than a lower tier is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.legendary.glow = "soft"; }), "legendary.glow must not be weaker");
});

test("a rarity tier with a wrong name key is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.rare.nameKey = "rarity.blue.name"; }), "rare.nameKey");
});

test("a rarity id outside the cosmetics id pattern is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.Ultra = { ...copy.rarity.tiers.mythic, ordinal: 6, pips: 6, nameKey: "rarity.Ultra.name" }; }), "rarity.tiers.Ultra: id must match");
});

test("a rarity fill without non-text contrast on a surface is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.rare.fill = "#3A4BB0"; }), "contrast rarity.tiers.rare.fill on surface.raised");
});

test("a rarity label shade below AA is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.epic.text = "#7A3FB0"; }), "contrast rarity.tiers.epic.text");
});

test("a stale rarity contrast table is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.common.text = "#D0D2E0"; }), "rarity spec contrast table is missing or stale");
});

test("a rarity colour missing from the rarity spec is reported", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.tiers.mythic.on = "#100828"; }), "rarity spec does not list colour #100828");
});

test("a concurrent set over the level budget is reported", () => {
  assertReported(errorsAfter((copy) => { copy.effects.budgets.domino.medium.gpuMs = 2.0; }), "dominoEnd.medium: gpuMs");
});

test("an effect over the level overdraw limit is reported", () => {
  assertReported(errorsAfter((copy) => { copy.effects.budgets.block.low.overdrawLayers = 3; }), "block.low: overdraw");
});

test("a level whose base plus effects exceeds the frame budget is reported", () => {
  assertReported(errorsAfter((copy) => { copy.effects.levels.high.effectCpuMs = 2.5; }), "levels.high: CPU");
});

test("score steps that shrink are reported", () => {
  assertReported(errorsAfter((copy) => { copy.scoreCascade.steps[3].textScale = 1.1; }), "textScale must not decrease");
});

test("a typography style under the minimum size is reported", () => {
  assertReported(errorsAfter((copy) => { copy.typography.scale.caption.size = 24; }), "under minimum");
});

test("score values map to steps with the chain boost capped at the top step", () => {
  assert.deepEqual([5, 10, 15, 20, 25, 30, 45, 80].map((points) => stepForPoints(tokens, points)), [1, 2, 3, 4, 5, 6, 6, 6]);
  assert.equal(stepForPoints(tokens, 0), 0);
  assert.equal(stepForPoints(tokens, 5, 2), 2);
  assert.equal(stepForPoints(tokens, 10, 3), 4);
  assert.equal(stepForPoints(tokens, 10, 7), 4);
  assert.equal(stepForPoints(tokens, 25, 3), 6);
});
