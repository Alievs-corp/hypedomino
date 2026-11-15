import assert from "node:assert/strict";
import { test } from "node:test";
import { bundlesOf, chainError, chainFromText, placeholdersFromText, placeholdersText, problemsText, valueOf, visibleStrings, type GameString } from "./strings.ts";

function entry(key: string, fields: Partial<GameString> = {}): GameString {
  return { key, note: "", max_length: 40, placeholders: [], deprecated: false, values: {}, ...fields };
}

const strings = [
  entry("match.result.win", { note: "Shown after a win", values: { en: { text: "You won", status: "approved" } } }),
  entry("shop.offer.title", { values: { az: { text: "Təklif", status: "draft" } } }),
  entry("quick.hello"),
];

test("search matches the key, the note and any language's text, ignoring case", () => {
  assert.deepEqual(visibleStrings(strings, "RESULT", null).map((s) => s.key), ["match.result.win"]);
  assert.deepEqual(visibleStrings(strings, "after a", null).map((s) => s.key), ["match.result.win"]);
  assert.deepEqual(visibleStrings(strings, "təklif", null).map((s) => s.key), ["shop.offer.title"]);
  assert.equal(visibleStrings(strings, "  ", null).length, 3);
});

test("the missing filter keeps only the keys the report lists for that language", () => {
  assert.deepEqual(visibleStrings(strings, "", ["shop.offer.title", "quick.hello"]).map((s) => s.key), ["shop.offer.title", "quick.hello"]);
  assert.deepEqual(visibleStrings(strings, "quick", ["shop.offer.title", "quick.hello"]).map((s) => s.key), ["quick.hello"]);
});

test("a language without a value reads as missing", () => {
  assert.deepEqual(valueOf(strings[2], "ru"), { text: "", status: "missing" });
  assert.deepEqual(valueOf(strings[0], "en"), { text: "You won", status: "approved" });
});

test("placeholders round-trip through name:type text", () => {
  const placeholders = placeholdersFromText(" name:string , count : integer,,");
  assert.deepEqual(placeholders, [{ name: "name", type: "string" }, { name: "count", type: "integer" }]);
  assert.equal(placeholdersText(placeholders), "name:string, count:integer");
  assert.deepEqual(placeholdersFromText("bare"), [{ name: "bare", type: "" }]);
});

test("bundles of one language come newest first", () => {
  const bundles = [
    { language: "en", version: 1, hash: "a", by: "x", at: "" },
    { language: "az", version: 3, hash: "b", by: "x", at: "" },
    { language: "en", version: 2, hash: "c", by: "x", at: "" },
  ];
  assert.deepEqual(bundlesOf(bundles, "en").map((b) => b.version), [2, 1]);
});

test("blocked problems read as key: error", () => {
  assert.equal(problemsText([{ key: "a.b", error: "too long" }, { key: "c.d", error: "bad placeholder" }]), "a.b: too long; c.d: bad placeholder");
});

const languages = ["az", "en", "ru", "tr"];

test("a chain reads from comma separated language codes", () => {
  assert.deepEqual(chainFromText(" az, tr ,,en "), ["az", "tr", "en"]);
  assert.deepEqual(chainFromText("  "), []);
});

test("a chain starts with its language, ends with en and repeats no known language", () => {
  assert.equal(chainError("az", ["az", "tr", "en"], languages), null);
  assert.equal(chainError("en", ["en"], languages), null);
  assert.equal(chainError("az", [], languages), "start");
  assert.equal(chainError("az", ["tr", "az", "en"], languages), "start");
  assert.equal(chainError("tr", ["tr", "az"], languages), "end");
  assert.equal(chainError("tr", ["tr", "de", "en"], languages), "unknown");
  assert.equal(chainError("az", ["az", "tr", "tr", "en"], languages), "repeat");
});
