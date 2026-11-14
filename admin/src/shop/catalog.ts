export type Bundle = {
  id: string;
  revision: number;
  coins?: number;
  crystals?: number;
  cosmetics?: string[];
  medals?: string[];
  passXp?: number;
};

export type Offer = {
  id: string;
  kind: string;
  section: string;
  priceCoins?: number;
  priceCrystals?: number;
  productId?: string;
  referencePriceCoins?: number;
  referencePriceCrystals?: number;
  regionalOverrideNote?: string;
  bundle?: Bundle;
  window: { start: string; end?: string };
  limitPerAccount?: number;
  nameKey: string;
  [field: string]: unknown;
};

export type ChestRules = {
  rarityPercent: Record<string, number>;
  rarePityCount: number;
};

export type RotationPeriod = "daily" | "weekly";

export type Rotation = { period: RotationPeriod; start: string; offers: string[] };

export type RotationRules = { daily_slots: number; weekly_slots: number; reset_hour_utc: number };

export type ShopCatalog = {
  schema: number;
  offers: Offer[];
  rotations: Rotation[];
  chests?: ChestRules;
  [field: string]: unknown;
};

export type CosmeticItem = {
  id: string;
  type: string;
  rarity: string;
  availability: string;
  sources: { type: string; ref?: string }[];
};

export type PriceCurrency = "coins" | "crystals" | "product";

export type Price = { currency: PriceCurrency; amount: number; productId: string };

export const chestRarities = ["common", "rare"] as const;

export const crystalPackProducts = ["crystals_tier1", "crystals_tier2", "crystals_tier3", "crystals_tier4", "crystals_tier5"];

const inGame: PriceCurrency[] = ["coins", "crystals"];

export function allowedCurrencies(offer: Offer): PriceCurrency[] {
  switch (true) {
    case offer.kind === "chest":
      return ["coins"];
    case offer.kind === "crystals":
      return ["product"];
    case offer.section === "featured":
      return ["crystals", "product"];
    case offer.section === "pass":
      return ["product", ...inGame];
    default:
      return inGame;
  }
}

export function priceOf(offer: Offer): Price {
  if (offer.productId) {
    return { currency: "product", amount: 0, productId: offer.productId };
  }
  if (offer.priceCrystals) {
    return { currency: "crystals", amount: offer.priceCrystals, productId: "" };
  }
  return { currency: "coins", amount: offer.priceCoins ?? 0, productId: "" };
}

export function withPrice(offer: Offer, price: Price): Offer {
  const { priceCoins, priceCrystals, productId, referencePriceCoins, referencePriceCrystals, regionalOverrideNote, ...rest } = offer;
  switch (price.currency) {
    case "product":
      return { ...rest, productId: price.productId, ...(regionalOverrideNote ? { regionalOverrideNote } : {}) };
    case "crystals":
      return { ...rest, priceCrystals: price.amount, ...(referencePriceCrystals ? { referencePriceCrystals } : {}) };
    case "coins":
      return { ...rest, priceCoins: price.amount, ...(referencePriceCoins ? { referencePriceCoins } : {}) };
  }
}

export function withRegionalNote(offer: Offer, note: string): Offer {
  const { regionalOverrideNote, ...rest } = offer;
  return note.trim() ? { ...rest, regionalOverrideNote: note } : rest;
}

export function replaceOffer(catalog: ShopCatalog, offer: Offer): ShopCatalog {
  return { ...catalog, offers: catalog.offers.map((current) => (current.id === offer.id ? offer : current)) };
}

export function crystalPacks(catalog: ShopCatalog): Offer[] {
  return catalog.offers
    .filter((offer) => offer.kind === "crystals")
    .toSorted((a, b) => crystalPackProducts.indexOf(a.productId ?? "") - crystalPackProducts.indexOf(b.productId ?? ""));
}

export function withPackCrystals(pack: Offer, crystals: number): Offer {
  return { ...pack, bundle: { ...(pack.bundle ?? { id: "", revision: 1 }), crystals } };
}

export function chestCandidates(items: CosmeticItem[]): CosmeticItem[] {
  return items.filter((item) => item.availability === "live" && (chestRarities as readonly string[]).includes(item.rarity));
}

export function listsChest(item: CosmeticItem, chest: Offer): boolean {
  return item.sources.some((source) => source.type === "chest" && source.ref === chest.id);
}

export function withPoolItem(chest: Offer, itemId: string, inPool: boolean): Offer {
  const pool = (chest.bundle?.cosmetics ?? []).filter((id) => id !== itemId);
  const cosmetics = inPool ? [...pool, itemId] : pool;
  return { ...chest, bundle: { ...(chest.bundle ?? { id: "", revision: 1 }), cosmetics } };
}

export function chestRulesOf(catalog: ShopCatalog): ChestRules {
  return catalog.chests ?? { rarityPercent: { common: 0, rare: 0 }, rarePityCount: 0 };
}

export function oddsTotal(rules: ChestRules): number {
  return chestRarities.reduce((total, rarity) => total + (rules.rarityPercent[rarity] ?? 0), 0);
}

export function windowMinute(instant: string | undefined): string {
  const date = instant ? new Date(instant) : null;
  return date && !Number.isNaN(date.getTime()) ? date.toISOString().slice(0, 16) : "";
}

export function withWindow(offer: Offer, startMinute: string, endMinute: string): Offer {
  const start = startMinute ? `${startMinute}:00Z` : offer.window.start;
  return { ...offer, window: endMinute ? { start, end: `${endMinute}:00Z` } : { start } };
}

export function windowEndsBeforeStart(offer: Offer): boolean {
  return offer.window.end !== undefined && new Date(offer.window.end) <= new Date(offer.window.start);
}

export function changedOffers(previous: ShopCatalog | null, next: ShopCatalog): string[] {
  const before = new Map((previous?.offers ?? []).map((offer) => [offer.id, JSON.stringify(offer)]));
  return next.offers.filter((offer) => before.get(offer.id) !== JSON.stringify(offer)).map((offer) => offer.id);
}

const rotationSections: Record<RotationPeriod, string> = { daily: "daily_rotation", weekly: "weekly_rotation" };

const windowedSections = ["featured", "event"];

export function timeLimited(offer: Offer): boolean {
  return Object.values(rotationSections).includes(offer.section) || (windowedSections.includes(offer.section) && offer.window.end !== undefined);
}

export function permanentMythics(offer: Offer, items: CosmeticItem[]): string[] {
  if (offer.kind === "chest" || timeLimited(offer)) {
    return [];
  }
  const mythics = new Set(items.filter((item) => item.rarity === "mythic").map((item) => item.id));
  return (offer.bundle?.cosmetics ?? []).filter((id) => mythics.has(id));
}

export function rotationCandidates(catalog: ShopCatalog, period: RotationPeriod): string[] {
  return catalog.offers.filter((offer) => offer.section === rotationSections[period]).map((offer) => offer.id);
}

export function slotsOf(rules: RotationRules, period: RotationPeriod): number {
  return period === "weekly" ? rules.weekly_slots : rules.daily_slots;
}

export function periodStart(period: RotationPeriod, day: string, rules: RotationRules): string | null {
  const start = new Date(`${day}T${String(rules.reset_hour_utc).padStart(2, "0")}:00:00Z`);
  if (Number.isNaN(start.getTime()) || (period === "weekly" && start.getUTCDay() !== 1)) {
    return null;
  }
  return start.toISOString().replace(".000Z", "Z");
}

function sameSlot(a: Rotation, period: RotationPeriod, start: string): boolean {
  return a.period === period && new Date(a.start).getTime() === new Date(start).getTime();
}

export function withRotation(catalog: ShopCatalog, rotation: Rotation): ShopCatalog {
  const others = catalog.rotations.filter((current) => !sameSlot(current, rotation.period, rotation.start));
  const rotations = [...others, rotation].toSorted((a, b) => a.period.localeCompare(b.period) || new Date(a.start).getTime() - new Date(b.start).getTime());
  return { ...catalog, rotations };
}

export function withoutRotation(catalog: ShopCatalog, rotation: Rotation): ShopCatalog {
  return { ...catalog, rotations: catalog.rotations.filter((current) => !sameSlot(current, rotation.period, rotation.start)) };
}

export function toggledOffer(rotation: Rotation, offerId: string): Rotation {
  const offers = rotation.offers.includes(offerId) ? rotation.offers.filter((id) => id !== offerId) : [...rotation.offers, offerId];
  return { ...rotation, offers };
}

export function rotationFits(rotation: Rotation, rules: RotationRules): boolean {
  return rotation.offers.length >= 1 && rotation.offers.length <= slotsOf(rules, rotation.period);
}
