import assert from "node:assert/strict";
import test from "node:test";
import {
  addMedal,
  emptyMedalCatalog,
  medalChanges,
  medalIssues,
  medalTextKeys,
  newMedal,
  replaceMedal,
  validMedalSlug,
  validShowcaseSlots,
  withNewSource,
  withoutSource,
  withSource,
  withReview,
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
      culturalReview: "approved",
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
    culturalReview: "",
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
  assert.deepEqual(medalIssues(newMedal("cup", rules)), ["medals.issue.empty_ref", "medals.issue.review"]);
  const twice = withSource(withNewSource(catalog.medals[0], rules), 1, { type: "mission", ref: " mission.win_50 " });
  assert.deepEqual(medalIssues(twice), ["medals.issue.duplicate_source"]);
  assert.ok(validShowcaseSlots(0) && validShowcaseSlots(3) && !validShowcaseSlots(-1) && !validShowcaseSlots(1.5));
});

test("history lists each field a version changed, per medal and in the showcase slots", () => {
  const next = { ...replaceMedal(addMedal(catalog, newMedal("cup", rules))!, { ...catalog.medals[0], availability: "retired" }), profileMedalSlots: 4 };
  assert.deepEqual(medalChanges(catalog, next), {
    settings: [{ field: "profileMedalSlots", before: "3", after: "4" }],
    entries: [
      { id: "medal.wins_50", change: "changed", fields: [{ field: "availability", before: "live", after: "retired" }] },
      { id: "medal.cup", change: "added", fields: [] },
    ],
  });
  assert.deepEqual(medalChanges(null, catalog), { settings: [], entries: [{ id: "medal.wins_50", change: "added", fields: [] }] });
});

test("the keys checked before a medal publish are each medal's name and description", () => {
  const second = { ...newMedal("first_cup", rules), revision: 1 };
  assert.deepEqual(medalTextKeys({ ...catalog, medals: [...catalog.medals, second] }), [
    "medal.wins_50.name",
    "medal.wins_50.description",
    "medal.first_cup.name",
    "medal.first_cup.description",
  ]);
  assert.deepEqual(medalTextKeys(emptyMedalCatalog()), []);
});

test("only a live medal needs an approved cultural review", () => {
  const medal = catalog.medals[0];
  const cases: { availability: string; approved: boolean; issue: boolean }[] = [
    { availability: "live", approved: true, issue: false },
    { availability: "live", approved: false, issue: true },
    { availability: "retired", approved: false, issue: false },
    { availability: "disabled", approved: false, issue: false },
  ];
  for (const { availability, approved, issue } of cases) {
    const reviewed = withReview({ ...medal, availability }, approved);
    assert.equal(reviewed.culturalReview, approved ? "approved" : "");
    assert.equal(medalIssues(reviewed).includes("medals.issue.review"), issue);
  }
});
