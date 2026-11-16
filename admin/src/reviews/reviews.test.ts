import assert from "node:assert/strict";
import { test } from "node:test";
import { evidenceSummary, nextQueueHref, queueApiPath, queuePageSize, readReportQuery, readReviewQuery, signalWeight } from "./reviews.ts";

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

test("the queue queries keep a known cause or reason, the account search and a well-formed cursor", () => {
  const account = "5f2c9d3e-6d4a-4f4e-9d11-1a2b3c4d5e6f";
  assert.deepEqual(readReviewQuery({ cause: "score", account: " 5f2c ", after: "42" }), { cause: "score", account: "5f2c", after: "42" });
  assert.deepEqual(readReviewQuery({ cause: "toString", after: "0" }), { cause: "", account: "", after: "" });
  assert.deepEqual(readReportQuery({ reason: ["impersonation", "other"], after: account }), { reason: "impersonation", account: "", after: account });
  assert.deepEqual(readReportQuery({ reason: "spam", after: "42" }), { reason: "", account: "", after: "" });
});

test("the API path always asks for one page and the next link follows the last entry", () => {
  const query = { cause: "hard_signal", account: "", after: "" } as const;
  assert.equal(queueApiPath("/admin/reviews", query), `/admin/reviews?cause=hard_signal&limit=${queuePageSize}`);
  const cursors = Array.from({ length: queuePageSize }, (_, at) => String(at + 1));
  assert.equal(nextQueueHref("/reviews", query, cursors), `/reviews?cause=hard_signal&after=${queuePageSize}`);
  assert.equal(nextQueueHref("/reports", { reason: "", account: "ab", after: "" }, cursors.slice(1)), null);
});
