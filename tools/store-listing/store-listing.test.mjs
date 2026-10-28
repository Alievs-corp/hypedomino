import { test } from "node:test";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import { checkListing, checkListingDirectory, fields, fold, length, parseListing } from "./store-listing.mjs";

const validValues = () => new Map(fields.map((field) => [field.id, "Domino tiles"]));

function assertReported(values, fragment) {
  const errors = checkListing(values);
  assert.ok(errors.some((error) => error.includes(fragment)), `expected "${fragment}" in:\n${errors.join("\n")}`);
}

function withValue(id, value) {
  const values = validValues();
  values.set(id, value);
  return values;
}

test("the committed listings pass", () => {
  const directory = fileURLToPath(new URL("../../docs/store/", import.meta.url));
  assert.deepEqual(checkListingDirectory(directory), []);
});

test("a listing with every field in limit passes", () => {
  assert.deepEqual(checkListing(validValues()), []);
});

test("fenced blocks are parsed exactly, with CRLF line ends", () => {
  const source = "# Title\r\n\r\nIntro.\r\n\r\n## play.title\r\n\r\n```text\r\nHypeDomino\r\n```\r\n\r\n## description\r\n\r\n```text\r\nLine one.\r\n\r\nLine two.\r\n```\r\n";
  const { values, errors } = parseListing(source);
  assert.deepEqual(errors, []);
  assert.equal(values.get("play.title"), "HypeDomino");
  assert.equal(values.get("description"), "Line one.\n\nLine two.");
});

test("a heading without a text block and a repeated field are reported", () => {
  const { errors } = parseListing("## play.title\n\nHypeDomino\n\n## apple.name\n\n```text\nA\n```\n\n## apple.name\n\n```text\nB\n```\n");
  assert.ok(errors.includes("play.title: heading without a text block"));
  assert.ok(errors.includes("apple.name: defined twice"));
});

test("limits count characters, not bytes", () => {
  assert.equal(length("Домино"), 6);
  assert.deepEqual(checkListing(withValue("play.title", "Ə".repeat(30))), []);
  assertReported(withValue("play.title", "Ə".repeat(31)), "31 characters, limit 30");
  assertReported(withValue("apple.keywords", "a,".repeat(50) + "b"), "limit 100");
});

test("missing and unknown fields are reported", () => {
  const values = validValues();
  values.delete("apple.subtitle");
  values.set("play.tagline", "Domino");
  assertReported(values, "apple.subtitle: missing");
  assertReported(values, "play.tagline: unknown field");
});

test("experiment variants of a known field are checked against its limit", () => {
  assert.deepEqual(checkListing(withValue("play.short@B", "Domino with friends")), []);
  assertReported(withValue("play.short@B", "x".repeat(81)), "play.short@B: 81 characters, limit 80");
  assertReported(withValue("play.short@b", "Domino"), "play.short@b: unknown field");
});

test("gambling and real-money terms are reported in every language", () => {
  assertReported(withValue("description", "A casino night."), 'forbidden term "casino*"');
  assertReported(withValue("description", "Win real   money."), 'forbidden term "real money"');
  assertReported(withValue("description", "Bahis oyna."), 'forbidden term "bahis*"');
  assertReported(withValue("description", "Real pulla oyna."), 'forbidden term "real pul*"');
  assertReported(withValue("description", "Реальные деньги за победу."), 'forbidden term "реальн* деньг*"');
  assertReported(withValue("description", "Азартная игра."), 'forbidden term "азартн*"');
});

test("Turkish dotted and dotless capitals cannot dodge the term list", () => {
  assert.equal(fold("KUMARHANE"), "kumarhane");
  assert.equal(fold("CASİNO"), "casino");
  assert.equal(fold("casıno"), "casino");
  assertReported(withValue("description", "CASİNO"), 'forbidden term "casino*"');
  assertReported(withValue("description", "GERÇEK PARA"), 'forbidden term "gerçek para*"');
});

test("terms match whole words only", () => {
  assert.deepEqual(checkListing(withValue("description", "Better tiles between friends; topla daşları; ставь кость; ортадан.")), []);
  assertReported(withValue("description", "Place a bet."), 'forbidden term "bet"');
});

test("prices and currencies are reported", () => {
  assertReported(withValue("description", "Only $4.99"), "currency symbol");
  assertReported(withValue("description", "5 manat"), 'forbidden term "manat*"');
  assertReported(withValue("description", "299 рублей"), 'forbidden term "рубл*"');
});

test("promotional terms are reported in short fields only", () => {
  assertReported(withValue("play.title", "Best Domino"), 'promotional term "best"');
  assertReported(withValue("apple.subtitle", "Pulsuz domino"), 'promotional term "pulsuz"');
  assertReported(withValue("caption.S-1", "#1 domino"), 'promotional term "#1"');
  assertReported(withValue("play.short", "Бесплатное домино"), 'promotional term "бесплатн*"');
  assert.deepEqual(checkListing(withValue("description", "Free and premium tracks.")), []);
});

test("emoji, placeholders and surrounding whitespace are reported", () => {
  assertReported(withValue("play.title", "Domino 🎲"), "emoji");
  assertReported(withValue("description", "Support: <support-url>"), "unresolved placeholder");
  assertReported(withValue("play.short", " Domino"), "leading or trailing whitespace");
  assertReported(withValue("caption.S-2", ""), "empty");
});

test("keywords are checked for spacing, repeats and words already in the name", () => {
  const values = validValues();
  values.set("apple.name", "HypeDomino: Online Dominoes");
  values.set("apple.subtitle", "Ranked 1v1 and 2v2 matches");
  values.set("apple.keywords", "board game, tiles,Tiles,,online");
  assertReported(values, 'keyword " tiles" has surrounding spaces');
  assertReported(values, 'keyword "Tiles" repeated');
  assertReported(values, "empty keyword");
  assertReported(values, 'keyword "online" repeats a word of the name or subtitle');
});
