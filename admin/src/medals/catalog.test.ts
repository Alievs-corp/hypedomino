import assert from "node:assert/strict";
import test from "node:test";
import {
  addMedal,
  changedMedals,
  emptyMedalCatalog,
  medalIssues,
  newMedal,
  replaceMedal,
  validMedalSlug,
  validShowcaseSlots,
  withNewSource,
  withoutSource,
  withSource,
  type MedalCatalog,
  type MedalRules,
} from "./catalog.ts";

const rules: MedalRules = { rarities: ["common", "rare"], source_types: ["mission", "rank_reward"], availability: ["live", "retired", "disabled"] };

const catalog: MedalCatalog = {
  schema: 1,
  profileMedalSlots: 3,
  medals: [
    {
      id: "medal.wins_50",
      revision: 1,
      nameKey: "medal.wins_50.name",
      descriptionKey: "medal.wins_50.description",
      rarityId: "common",
      staticAddress: "medal/wins_50/static",
      sources: [{ type: "mission", ref: "mission.win_50" }],
      availability: "live",
    },
  ],
};

test("a new medal takes the server's first rarity, source type and availability and one source to fill", () => {
  assert.deepEqual(newMedal("first_cup", rules), {
    id: "medal.first_cup",
    revision: 0,
    nameKey: "medal.first_cup.name",
    descriptionKey: "medal.first_cup.description",
    rarityId: "common",
    staticAddress: "medal/first_cup/static",
    sources: [{ type: "mission", ref: "" }],
    availability: "live",
  });
  assert.deepEqual(emptyMedalCatalog(), { schema: 1, profileMedalSlots: 0, medals: [] });
  assert.ok(validMedalSlug("wins_5") && !validMedalSlug("Wins-5") && !validMedalSlug(""));
});

test("medals are added once and replaced by id", () => {
  assert.equal(addMedal(catalog, newMedal("wins_50", rules)), null);
  assert.equal(addMedal(catalog, newMedal("cup", rules))?.medals.length, 2);
  assert.equal(replaceMedal(catalog, { ...catalog.medals[0], rarityId: "rare" }).medals[0].rarityId, "rare");
});

test("sources are edited, added and removed by position", () => {
  const added = withNewSource(catalog.medals[0], rules);
  assert.deepEqual(added.sources[1], { type: "mission", ref: "" });
  const edited = withSource(added, 1, { type: "rank_reward", ref: "rank.season_1.gold" });
  assert.deepEqual(edited.sources.map((source) => source.type), ["mission", "rank_reward"]);
  assert.deepEqual(withoutSource(edited, 0).sources, [{ type: "rank_reward", ref: "rank.season_1.gold" }]);
});

test("issues name what the server would refuse or what cannot be meant", () => {
  assert.deepEqual(medalIssues(catalog.medals[0]), []);
  assert.deepEqual(medalIssues({ ...catalog.medals[0], sources: [] }), ["medals.issue.no_source"]);
  assert.deepEqual(medalIssues(newMedal("cup", rules)), ["medals.issue.empty_ref"]);
  const twice = withSource(withNewSource(catalog.medals[0], rules), 1, { type: "mission", ref: " mission.win_50 " });
  assert.deepEqual(medalIssues(twice), ["medals.issue.duplicate_source"]);
  assert.ok(validShowcaseSlots(0) && validShowcaseSlots(3) && !validShowcaseSlots(-1) && !validShowcaseSlots(1.5));
});

test("history lists the medals a version added or changed", () => {
  const next = replaceMedal(addMedal(catalog, newMedal("cup", rules))!, { ...catalog.medals[0], availability: "retired" });
  assert.deepEqual(changedMedals(catalog, next), ["medal.wins_50", "medal.cup"]);
  assert.deepEqual(changedMedals(null, catalog), ["medal.wins_50"]);
});
