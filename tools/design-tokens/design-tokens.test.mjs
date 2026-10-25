import { test } from "node:test";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import { checkTokens, contrast, loadSpec, loadTokens, stepForPoints } from "./design-tokens.mjs";

const repoRoot = new URL("../../", import.meta.url);
const tokens = loadTokens(fileURLToPath(new URL("docs/design-tokens.json", repoRoot)));
const specText = loadSpec(fileURLToPath(new URL("docs/ui-motion.md", repoRoot)));

function errorsAfter(change, spec = specText) {
  const copy = structuredClone(tokens);
  change(copy);
  return checkTokens(copy, spec);
}

function assertReported(errors, fragment) {
  assert.ok(errors.some((error) => error.includes(fragment)), errors.join("\n"));
}

test("the committed tokens and spec pass every check", () => {
  assert.deepEqual(checkTokens(tokens, specText), []);
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

test("rarity values are rejected until #5 defines them", () => {
  assertReported(errorsAfter((copy) => { copy.rarity.common = "#AAAAAA"; }), "rarity");
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
