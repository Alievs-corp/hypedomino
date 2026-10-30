import { readFileSync } from "node:fs";
import { join } from "node:path";
import { pathToFileURL } from "node:url";

export const locales = ["az", "en", "tr", "ru"];

const captionIds = ["S-1", "S-2", "S-3", "S-4", "S-5", "S-6", "S-7", "S-8"].map((shot) => `caption.${shot}`);

export const fields = [
  { id: "play.title", limit: 30, short: true },
  { id: "play.short", limit: 80, short: true },
  { id: "apple.name", limit: 30, short: true },
  { id: "apple.subtitle", limit: 30, short: true },
  { id: "apple.keywords", limit: 100, short: true, keywords: true },
  { id: "apple.promo", limit: 170 },
  { id: "description", limit: 4000 },
  ...captionIds.map((id) => ({ id, limit: 32, short: true })),
];

const fieldsById = new Map(fields.map((field) => [field.id, field]));
const variantPattern = /^(.+)@([A-Z])$/;

export const forbiddenTerms = [
  "casino*", "gambl*", "bet", "bets", "betting", "wager*", "jackpot*", "slot", "slots", "slot machine*", "roulette*", "poker*", "lotter*", "loot box*",
  "real money", "win money", "earn money", "cash prize*", "prize money", "cash out",
  "kazino*", "qumar*", "mərc*", "rulet*", "lotereya*", "cekpot*", "real pul*", "pul qazan*", "pul uduş*",
  "kumar*", "bahis*", "piyango*", "gerçek para*", "para kazan*", "nakit ödül*", "para ödül*",
  "казино*", "азартн*", "ставк*", "джекпот*", "рулетк*", "лотере*", "слот*", "покер*", "реальн* деньг*", "реальн* денег", "выигр* деньг*", "заработ* деньг*", "денежн* приз*",
  "usd", "eur", "euro*", "dollar*", "azn", "manat*", "lira", "liras", "рубл*", "доллар*",
];

export const promotionalTerms = [
  "free", "best", "top", "new", "sale", "download*", "install now", "play now", "#1", "№1", "no 1",
  "pulsuz", "ən yaxşı", "yeni", "endirim*", "yüklə*",
  "ücretsiz", "bedava", "en iyi", "indirim*", "indir",
  "бесплатн*", "лучш*", "нов*", "скидк*", "скача*",
];

const currencySymbols = /[$€£₼₺₽]/u;
const pictographic = /\p{Extended_Pictographic}/u;
const placeholder = /[<>{}]/;

export function fold(text) {
  return text.toLowerCase().replace(/̇/g, "").replace(/ı/g, "i");
}

function escapeRegex(text) {
  return text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function termPattern(term) {
  const body = fold(term).split("*").map((part) => escapeRegex(part).replace(/ /g, "\\s+")).join("[\\p{L}\\p{N}]*");
  return new RegExp(`(?<![\\p{L}\\p{N}])${body}(?![\\p{L}\\p{N}])`, "u");
}

const forbiddenPatterns = forbiddenTerms.map((term) => [term, termPattern(term)]);
const promotionalPatterns = promotionalTerms.map((term) => [term, termPattern(term)]);

export function length(text) {
  return [...text].length;
}

export function parseListing(source) {
  const text = source.replace(/\r\n/g, "\n");
  const values = new Map();
  const errors = [];
  const blockPattern = /^## (.+)\n\n```text\n([\s\S]*?)\n```$/gm;
  for (const match of text.matchAll(blockPattern)) {
    const id = match[1].trim();
    if (values.has(id)) errors.push(`${id}: defined twice`);
    values.set(id, match[2]);
  }
  for (const heading of text.matchAll(/^## (.+)$/gm)) {
    const id = heading[1].trim();
    if (!values.has(id)) errors.push(`${id}: heading without a text block`);
  }
  return { values, errors };
}

function baseField(id) {
  const variant = id.match(variantPattern);
  return fieldsById.get(variant ? variant[1] : id);
}

function matches(patterns, value) {
  const folded = fold(value);
  return patterns.filter(([, pattern]) => pattern.test(folded)).map(([term]) => term);
}

function keywordErrors(value, values) {
  const errors = [];
  const keywords = value.split(",");
  const seen = new Set();
  const nameWords = new Set(fold(`${values.get("apple.name") ?? ""} ${values.get("apple.subtitle") ?? ""}`).split(/[^\p{L}\p{N}]+/u).filter(Boolean));
  for (const keyword of keywords) {
    if (keyword.trim() === "") { errors.push("empty keyword"); continue; }
    if (keyword !== keyword.trim()) errors.push(`keyword "${keyword}" has surrounding spaces`);
    const folded = fold(keyword.trim());
    if (seen.has(folded)) errors.push(`keyword "${keyword.trim()}" repeated`);
    seen.add(folded);
    if (nameWords.has(folded)) errors.push(`keyword "${keyword.trim()}" repeats a word of the name or subtitle`);
  }
  return errors;
}

function valueErrors(field, value, values) {
  const errors = [];
  if (value.trim() === "") errors.push("empty");
  if (value !== value.trim()) errors.push("leading or trailing whitespace");
  const size = length(value);
  if (size > field.limit) errors.push(`${size} characters, limit ${field.limit}`);
  for (const term of matches(forbiddenPatterns, value)) errors.push(`forbidden term "${term}"`);
  if (field.short) for (const term of matches(promotionalPatterns, value)) errors.push(`promotional term "${term}"`);
  if (currencySymbols.test(value)) errors.push("currency symbol");
  if (pictographic.test(value)) errors.push("emoji");
  if (placeholder.test(value)) errors.push("unresolved placeholder or markup");
  if (field.keywords) errors.push(...keywordErrors(value, values));
  return errors;
}

export function checkListing(values) {
  const errors = [];
  for (const field of fields) {
    if (!values.has(field.id)) errors.push(`${field.id}: missing`);
  }
  for (const [id, value] of values) {
    const field = baseField(id);
    if (!field) { errors.push(`${id}: unknown field`); continue; }
    for (const error of valueErrors(field, value, values)) errors.push(`${id}: ${error}`);
  }
  return errors;
}

export function checkListingDirectory(directory) {
  return locales.flatMap((locale) => {
    const file = `listing.${locale}.md`;
    const { values, errors } = parseListing(readFileSync(join(directory, file), "utf8"));
    return [...errors, ...checkListing(values)].map((error) => `${file}: ${error}`);
  });
}

function main(args) {
  const directory = args[0] ?? "docs/store";
  const errors = checkListingDirectory(directory);
  for (const error of errors) console.error(error);
  if (errors.length > 0) process.exit(1);
  console.log(`store listing ok (${locales.length} locales, ${fields.length} fields each)`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main(process.argv.slice(2));
