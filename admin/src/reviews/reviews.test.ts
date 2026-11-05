import assert from "node:assert/strict";
import { test } from "node:test";
import { evidenceSummary, signalWeight } from "./reviews.ts";

test("the listed signals add up to their weight", () => {
  assert.equal(signalWeight([]), 0);
  const signal = { kind: "refund", evidence: {}, at: "2026-11-05T00:00:00Z" };
  assert.equal(signalWeight([{ ...signal, weight: 3 }, { ...signal, weight: 20 }]), 23);
});

test("evidence reads as its fields", () => {
  assert.equal(evidenceSummary({ reason: "token_stale", count: 5 }), "reason: token_stale; count: 5");
  assert.equal(evidenceSummary({}), "");
  assert.equal(evidenceSummary(null), "null");
  assert.equal(evidenceSummary(["a"]), '["a"]');
});
