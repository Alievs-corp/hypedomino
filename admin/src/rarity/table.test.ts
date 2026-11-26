import assert from "node:assert/strict";
import { test } from "node:test";
import { badFields, fieldsFromTable, tableFromFields, type RarityTable } from "./table.ts";

const glow = ["none", "soft", "strong", "intense"];

const table: RarityTable = {
  diamondMinSize: 72,
  flowPeriodMs: 2400,
  tiers: {
    rare: { ordinal: 2, nameKey: "rarity.rare.name", fill: "#5C78FF", on: "#0B0605", text: "#A3B3FF", pips: 2, glow: "soft", flow: false },
    common: { ordinal: 1, nameKey: "rarity.common.name", fill: "#A3A6BF", on: "#0B0605", text: "#C9CBDB", pips: 1, glow: "none", flow: false },
  },
};

test("tiers are listed by ordinal and round-trip with their name keys", () => {
  const fields = fieldsFromTable(table);
  assert.deepEqual(fields.tiers.map((tier) => tier.id), ["common", "rare"]);
  assert.equal(fields.diamondMinSize, "72");
  assert.deepEqual(tableFromFields(fields), table);
  assert.deepEqual(badFields(fields, glow), []);
});

test("an edited tier keeps the name key of its id", () => {
  const fields = fieldsFromTable(table);
  fields.tiers[1] = { ...fields.tiers[1], fill: "#6070FF", glow: "strong", flow: true };
  assert.deepEqual(tableFromFields(fields).tiers.rare, { ordinal: 2, nameKey: "rarity.rare.name", fill: "#6070FF", on: "#0B0605", text: "#A3B3FF", pips: 2, glow: "strong", flow: true });
});

test("broken numbers, colours, pips and glow levels are flagged", () => {
  const fields = fieldsFromTable(table);
  fields.diamondMinSize = "0";
  fields.flowPeriodMs = "2.5";
  fields.tiers[0] = { ...fields.tiers[0], ordinal: "x", fill: "grey", text: "#FFF", pips: "7" };
  fields.tiers[1] = { ...fields.tiers[1], glow: "blinding" };
  assert.deepEqual(badFields(fields, glow), ["diamondMinSize", "flowPeriodMs", "common.ordinal", "common.fill", "common.text", "common.pips", "rare.glow"]);
});
