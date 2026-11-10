import assert from "node:assert/strict";
import { test } from "node:test";
import { comparedVersion, sourceReasons, type EconomyReport } from "./economy.ts";

function report(sourceMix: Record<string, number>): EconomyReport {
  return { from: "", to: "", day: "", coinsCreated: 0, coinsSpent: 0, sinkRatioPercent: null, medianBalance: 0, sourceMix, accountsOverDailyEarn: 0, alerts: [] };
}

test("source reasons are ordered by coins across every report, then by name", () => {
  const reasons = sourceReasons([report({ mission: 50, match_reward: 400 }), report({ daily_login: 300, mission: 300, first_win: 350 })]);
  assert.deepEqual(reasons, ["match_reward", "first_win", "mission", "daily_login"]);
});

test("only a positive whole version is compared", () => {
  assert.equal(comparedVersion("12"), 12);
  assert.equal(comparedVersion(undefined), null);
  assert.equal(comparedVersion(""), null);
  assert.equal(comparedVersion("0"), null);
  assert.equal(comparedVersion("-3"), null);
  assert.equal(comparedVersion("2.5"), null);
});
