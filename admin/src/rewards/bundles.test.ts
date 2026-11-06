import assert from "node:assert/strict";
import test from "node:test";
import { emptyBundleFilter, filteredBundles, useLabel, usedDomains, type RewardBundle } from "./bundles.ts";

const daily: RewardBundle = {
  id: "bundle.daily_login_day_1",
  revision: 1,
  coins: 30,
  cosmetics: null,
  medals: null,
  passXp: 0,
  guideValue: 30,
  usedBy: [{ domain: "daily_login", object: "day.1" }],
  differs: false,
};

const wins: RewardBundle = {
  id: "bundle.wins",
  revision: 2,
  coins: 100,
  cosmetics: ["avatar.fox"],
  medals: ["medal.winner"],
  passXp: 0,
  guideValue: 1600,
  usedBy: [{ domain: "missions", object: "mission.first" }, { domain: "pass", object: "season_1", detail: "free.3" }],
  differs: true,
};

const ids = (change: object) => filteredBundles([daily, wins], { ...emptyBundleFilter, ...change }).map((bundle) => bundle.id);

test("the list searches ids, items, medals and the objects that grant a bundle", () => {
  assert.deepEqual(ids({}), ["bundle.daily_login_day_1", "bundle.wins"]);
  assert.deepEqual(ids({ query: " AVATAR.fox " }), ["bundle.wins"]);
  assert.deepEqual(ids({ query: "medal.winner" }), ["bundle.wins"]);
  assert.deepEqual(ids({ query: "day.1" }), ["bundle.daily_login_day_1"]);
  assert.deepEqual(ids({ query: "nothing" }), []);
});

test("the list narrows to a domain or to ids holding differing content", () => {
  assert.deepEqual(ids({ domain: "pass" }), ["bundle.wins"]);
  assert.deepEqual(ids({ domain: "daily_login" }), ["bundle.daily_login_day_1"]);
  assert.deepEqual(ids({ onlyDiffering: true }), ["bundle.wins"]);
});

test("the domain filter offers each domain in use once", () => {
  assert.deepEqual(usedDomains([daily, wins]), ["daily_login", "missions", "pass"]);
});

test("a use shows its object and its place inside it", () => {
  assert.equal(useLabel(wins.usedBy[0]), "mission.first");
  assert.equal(useLabel(wins.usedBy[1]), "season_1 free.3");
});
