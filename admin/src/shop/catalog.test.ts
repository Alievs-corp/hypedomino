import assert from "node:assert/strict";
import { test } from "node:test";
import {
  allowedCurrencies,
  periodStart,
  rotationCandidates,
  rotationFits,
  toggledOffer,
  withRotation,
  withoutRotation,
  type Rotation,
  type RotationRules,
  changedOffers,
  windowEndsBeforeStart,
  windowMinute,
  withWindow,
  chestCandidates,
  crystalPacks,
  listsChest,
  oddsTotal,
  permanentMythics,
  priceOf,
  replaceOffer,
  withPackCrystals,
  withPoolItem,
  withPrice,
  withRegionalNote,
  type CosmeticItem,
  type Offer,
  type ShopCatalog,
} from "./catalog.ts";

function offer(fields: Partial<Offer>): Offer {
  return { id: "offer.x", kind: "cosmetic", section: "event", window: { start: "2026-01-01T00:00:00Z" }, nameKey: "k", ...fields };
}

test("each offer offers only the currencies its kind and section allow", () => {
  const cases: [Partial<Offer>, string[]][] = [
    [{ kind: "chest", section: "chests" }, ["coins"]],
    [{ kind: "crystals", section: "crystals" }, ["product"]],
    [{ kind: "bundle", section: "featured" }, ["crystals", "product"]],
    [{ kind: "cosmetic", section: "daily_rotation" }, ["coins", "crystals"]],
    [{ kind: "cosmetic", section: "event" }, ["coins", "crystals"]],
    [{ kind: "pass_levels", section: "pass" }, ["product", "coins", "crystals"]],
  ];
  for (const [fields, want] of cases) {
    assert.deepEqual(allowedCurrencies(offer(fields)), want, JSON.stringify(fields));
  }
});

test("a new price replaces every other price and keeps only its own reference price", () => {
  const coins = offer({ priceCoins: 400, referencePriceCoins: 500, referencePriceCrystals: 70 });
  assert.deepEqual(priceOf(coins), { currency: "coins", amount: 400, productId: "" });
  const crystals = withPrice(coins, { currency: "crystals", amount: 60, productId: "" });
  assert.equal(crystals.priceCrystals, 60);
  assert.equal(crystals.priceCoins, undefined);
  assert.equal(crystals.referencePriceCoins, undefined);
  assert.equal(crystals.referencePriceCrystals, 70);
  const product = withPrice(crystals, { currency: "product", amount: 0, productId: "offer_starter" });
  assert.deepEqual(priceOf(product), { currency: "product", amount: 0, productId: "offer_starter" });
  assert.equal(product.priceCrystals, undefined);
  assert.equal(product.referencePriceCrystals, undefined);
  assert.deepEqual(priceOf(withPrice(product, { currency: "coins", amount: 300, productId: "" })), { currency: "coins", amount: 300, productId: "" });
});

test("a regional override note belongs to a store product price only", () => {
  const product = withRegionalNote(offer({ productId: "offer_starter" }), "TR override approved 2026-10-30");
  assert.equal(product.regionalOverrideNote, "TR override approved 2026-10-30");
  assert.equal(withPrice(product, { currency: "product", amount: 0, productId: "offer_bundle" }).regionalOverrideNote, "TR override approved 2026-10-30");
  assert.equal("regionalOverrideNote" in withPrice(product, { currency: "crystals", amount: 60, productId: "" }), false);
  assert.equal("regionalOverrideNote" in withRegionalNote(product, "  "), false);
});

test("crystal packs list in tier order and change only their crystals", () => {
  const tier2 = offer({ id: "offer.crystals_tier2", kind: "crystals", section: "crystals", productId: "crystals_tier2", bundle: { id: "bundle.crystals_tier2", revision: 1, crystals: 550 } });
  const tier1 = offer({ id: "offer.crystals_tier1", kind: "crystals", section: "crystals", productId: "crystals_tier1", bundle: { id: "bundle.crystals_tier1", revision: 1, crystals: 100 } });
  const catalog: ShopCatalog = { schema: 1, offers: [tier2, offer({}), tier1], rotations: [] };
  assert.deepEqual(crystalPacks(catalog).map((pack) => pack.productId), ["crystals_tier1", "crystals_tier2"]);
  const changed = replaceOffer(catalog, withPackCrystals(tier2, 600));
  assert.equal(changed.offers[0].bundle?.crystals, 600);
  assert.equal(changed.offers[0].bundle?.id, "bundle.crystals_tier2");
  assert.equal(catalog.offers[0].bundle?.crystals, 550);
});

test("the chest pool editor offers only live common and rare items", () => {
  const items: CosmeticItem[] = [
    { id: "avatar.paper_kite", type: "avatar", rarity: "common", availability: "live", sources: [{ type: "chest", ref: "offer.chest_avatar" }] },
    { id: "avatar.moon_owl", type: "avatar", rarity: "rare", availability: "live", sources: [] },
    { id: "frame.season_1_top", type: "frame", rarity: "epic", availability: "live", sources: [] },
    { id: "avatar.old_logo", type: "avatar", rarity: "rare", availability: "retired", sources: [] },
  ];
  assert.deepEqual(chestCandidates(items).map((item) => item.id), ["avatar.paper_kite", "avatar.moon_owl"]);
  const chest = offer({ id: "offer.chest_avatar", kind: "chest", section: "chests", priceCoins: 500, bundle: { id: "bundle.chest_avatar", revision: 1, cosmetics: ["avatar.paper_kite"] } });
  assert.equal(listsChest(items[0], chest), true);
  assert.equal(listsChest(items[1], chest), false);
  const added = withPoolItem(chest, "avatar.moon_owl", true);
  assert.deepEqual(added.bundle?.cosmetics, ["avatar.paper_kite", "avatar.moon_owl"]);
  assert.deepEqual(withPoolItem(withPoolItem(added, "avatar.moon_owl", true), "avatar.paper_kite", false).bundle?.cosmetics, ["avatar.moon_owl"]);
  assert.equal(oddsTotal({ rarityPercent: { common: 85, rare: 15 }, rarePityCount: 10 }), 100);
  assert.equal(oddsTotal({ rarityPercent: { common: 85 }, rarePityCount: 10 }), 85);
});

test("a window instant shows as its UTC minute and an empty end stays empty", () => {
  assert.equal(windowMinute("2026-11-04T09:30:00Z"), "2026-11-04T09:30");
  assert.equal(windowMinute("2026-11-04T13:30:00+04:00"), "2026-11-04T09:30");
  assert.equal(windowMinute(undefined), "");
  assert.equal(windowMinute("not a time"), "");
});

test("editing a window sets UTC instants and an empty end removes the end", () => {
  const start = offer({ window: { start: "2026-01-01T00:00:00Z", end: "2026-02-01T00:00:00Z" } });
  assert.deepEqual(withWindow(start, "2026-11-10T08:00", "2026-11-17T08:00").window, { start: "2026-11-10T08:00:00Z", end: "2026-11-17T08:00:00Z" });
  assert.deepEqual(withWindow(start, "2026-11-10T08:00", "").window, { start: "2026-11-10T08:00:00Z" });
  assert.deepEqual(withWindow(start, "", "").window, { start: "2026-01-01T00:00:00Z" });
});

test("a window must end after it starts", () => {
  assert.equal(windowEndsBeforeStart(offer({ window: { start: "2026-11-10T08:00:00Z" } })), false);
  assert.equal(windowEndsBeforeStart(offer({ window: { start: "2026-11-10T08:00:00Z", end: "2026-11-10T09:00:00Z" } })), false);
  assert.equal(windowEndsBeforeStart(offer({ window: { start: "2026-11-10T08:00:00Z", end: "2026-11-10T08:00:00Z" } })), true);
});

test("a mythic item is flagged unless its offer rotates or its featured or event window ends", () => {
  const items: CosmeticItem[] = [
    { id: "table.night_market", type: "table", rarity: "mythic", availability: "live", sources: [] },
    { id: "avatar.moon_owl", type: "avatar", rarity: "rare", availability: "live", sources: [] },
  ];
  const bundle = { id: "bundle.night", revision: 1, cosmetics: ["avatar.moon_owl", "table.night_market"] };
  const end = "2027-01-01T00:00:00Z";
  const cases: [Partial<Offer>, string[]][] = [
    [{ section: "weekly_rotation" }, []],
    [{ section: "daily_rotation" }, []],
    [{ section: "featured", window: { start: "2026-01-01T00:00:00Z", end } }, []],
    [{ section: "event", window: { start: "2026-01-01T00:00:00Z", end } }, []],
    [{ section: "featured" }, ["table.night_market"]],
    [{ section: "event" }, ["table.night_market"]],
    [{ kind: "pass", section: "pass", window: { start: "2026-01-01T00:00:00Z", end } }, ["table.night_market"]],
  ];
  for (const [fields, want] of cases) {
    assert.deepEqual(permanentMythics(offer({ kind: "bundle", bundle, ...fields }), items), want, JSON.stringify(fields));
  }
  assert.deepEqual(permanentMythics(offer({ section: "featured", bundle: { id: "bundle.owl", revision: 1, cosmetics: ["avatar.moon_owl"] } }), items), []);
});

test("a version lists the offers it added or changed against the one before", () => {
  const older: ShopCatalog = { schema: 1, offers: [offer({ id: "offer.a" }), offer({ id: "offer.b" })], rotations: [] };
  const newer: ShopCatalog = { schema: 1, offers: [offer({ id: "offer.a" }), offer({ id: "offer.b", priceCoins: 900 }), offer({ id: "offer.c" })], rotations: [] };
  assert.deepEqual(changedOffers(older, newer), ["offer.b", "offer.c"]);
  assert.deepEqual(changedOffers(null, older), ["offer.a", "offer.b"]);
});

const rules: RotationRules = { daily_slots: 2, weekly_slots: 1, reset_hour_utc: 4 };

test("a rotation starts at the reset hour, and a weekly one only on a Monday", () => {
  assert.equal(periodStart("daily", "2026-11-04", rules), "2026-11-04T04:00:00Z");
  assert.equal(periodStart("weekly", "2026-11-09", rules), "2026-11-09T04:00:00Z");
  assert.equal(periodStart("weekly", "2026-11-04", rules), null);
  assert.equal(periodStart("daily", "", rules), null);
});

test("only offers of the period's rotation section can fill its slots", () => {
  const catalog: ShopCatalog = { schema: 1, offers: [offer({ id: "offer.d", section: "daily_rotation" }), offer({ id: "offer.w", section: "weekly_rotation" }), offer({ id: "offer.e" })], rotations: [] };
  assert.deepEqual(rotationCandidates(catalog, "daily"), ["offer.d"]);
  assert.deepEqual(rotationCandidates(catalog, "weekly"), ["offer.w"]);
});

test("a rotation fills between one offer and the period's slots, in the order picked", () => {
  const empty: Rotation = { period: "daily", start: "2026-11-04T04:00:00Z", offers: [] };
  assert.equal(rotationFits(empty, rules), false);
  const two = toggledOffer(toggledOffer(empty, "offer.b"), "offer.a");
  assert.deepEqual(two.offers, ["offer.b", "offer.a"]);
  assert.equal(rotationFits(two, rules), true);
  assert.equal(rotationFits(toggledOffer(two, "offer.c"), rules), false);
  assert.deepEqual(toggledOffer(two, "offer.b").offers, ["offer.a"]);
});

test("one rotation per period start, kept in date order, and removable", () => {
  const base: ShopCatalog = { schema: 1, offers: [], rotations: [] };
  const later: Rotation = { period: "daily", start: "2026-11-05T04:00:00Z", offers: ["offer.a"] };
  const earlier: Rotation = { period: "daily", start: "2026-11-04T04:00:00.000Z", offers: ["offer.b"] };
  const both = withRotation(withRotation(base, later), earlier);
  assert.deepEqual(both.rotations.map((rotation) => rotation.start), [earlier.start, later.start]);
  const replaced = withRotation(both, { ...earlier, start: "2026-11-04T04:00:00Z", offers: ["offer.c"] });
  assert.equal(replaced.rotations.length, 2);
  assert.deepEqual(withoutRotation(replaced, later).rotations.map((rotation) => rotation.offers), [["offer.c"]]);
});
