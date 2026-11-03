import assert from "node:assert/strict";
import { test } from "node:test";
import { changedKeys, invalidKeys, publishedValues, withinBounds, type RuleKey } from "./rules.ts";

const turn: RuleKey = { key: "TurnSeconds", type: "integer", unit: "seconds", min: 10, max: 60, default: 20, value: 25, set: true };
const afk: RuleKey = { key: "AfkConsecutiveTimeouts", type: "integer", unit: "timeouts", min: 1, max: 10, default: 3, value: 3, set: false };
const keys = [turn, afk];

test("a value is valid only as an integer inside the key's bounds", () => {
  assert.ok(withinBounds(turn, 10));
  assert.ok(withinBounds(turn, 60));
  assert.ok(!withinBounds(turn, 9));
  assert.ok(!withinBounds(turn, 61));
  assert.ok(!withinBounds(turn, 20.5));
  assert.ok(!withinBounds(turn, Number.NaN));
});

test("untouched keys keep their current value and are neither changed nor invalid", () => {
  assert.deepEqual(changedKeys(keys, {}), []);
  assert.deepEqual(invalidKeys(keys, {}), []);
  assert.deepEqual(changedKeys(keys, { TurnSeconds: 25 }), []);
});

test("edits are reported as changed and out-of-bounds edits as invalid", () => {
  assert.deepEqual(changedKeys(keys, { TurnSeconds: 30, AfkConsecutiveTimeouts: 11 }), ["TurnSeconds", "AfkConsecutiveTimeouts"]);
  assert.deepEqual(invalidKeys(keys, { TurnSeconds: 30, AfkConsecutiveTimeouts: 11 }), ["AfkConsecutiveTimeouts"]);
});

test("a publish carries every value that differs from the default and leaves defaults out", () => {
  assert.deepEqual(publishedValues(keys, {}), { TurnSeconds: 25 });
  assert.deepEqual(publishedValues(keys, { TurnSeconds: 20 }), {});
  assert.deepEqual(publishedValues(keys, { AfkConsecutiveTimeouts: 5 }), { TurnSeconds: 25, AfkConsecutiveTimeouts: 5 });
});
