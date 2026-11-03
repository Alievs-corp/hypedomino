import assert from "node:assert/strict";
import { test } from "node:test";
import {
  allowedCurrencies,
  changedOffers,
  windowEndsBeforeStart,
  windowMinute,
  withWindow,
  chestCandidates,
  crystalPacks,
  listsChest,
  oddsTotal,
  priceOf,
  replaceOffer,
  withPackCrystals,
  withPoolItem,
  withPrice,
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

test("a version lists the offers it added or changed against the one before", () => {
  const older: ShopCatalog = { schema: 1, offers: [offer({ id: "offer.a" }), offer({ id: "offer.b" })], rotations: [] };
  const newer: ShopCatalog = { schema: 1, offers: [offer({ id: "offer.a" }), offer({ id: "offer.b", priceCoins: 900 }), offer({ id: "offer.c" })], rotations: [] };
  assert.deepEqual(changedOffers(older, newer), ["offer.b", "offer.c"]);
  assert.deepEqual(changedOffers(null, older), ["offer.a", "offer.b"]);
});
