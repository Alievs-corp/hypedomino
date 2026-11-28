import assert from "node:assert/strict";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { checkStringKeys, codeKeys, missingTexts, referencedKeys, resourceKeys } from "./string-keys.mjs";

const repoRoot = fileURLToPath(new URL("../../", import.meta.url));

test("every key the client names has an en and az text in the built-in bundle", () => {
  assert.deepEqual(checkStringKeys(repoRoot), []);
});

test("the scan finds keys in runtime code and in shipped resource data", () => {
  const references = referencedKeys(repoRoot);
  assert.equal(references.get("lobby.queue.refused.in_match"), "client/Assets/HypeDomino/Game/Lobby/QueueRefusalKeys.cs");
  assert.equal(references.get("rarity.mythic.name"), "client/Assets/HypeDomino/UI/Rarity/Resources/rarity-defaults.json");
  assert.equal(references.get("cosmetic.avatar.starter.name"), "client/Assets/HypeDomino/UI/Cosmetics/Resources/cosmetics-builtin.json");
  assert.ok(!references.has("color.role.primary.fill"));
  assert.ok([...references.keys()].every((key) => !key.startsWith("test.")));
});

test("only literals of three or more segments count as keys in code", () => {
  const source = 'const string A = "shop.load.offline"; const string B = "wallet.insufficient"; var c = "v1.2"; var d = "Shop.Load.Offline";';
  assert.deepEqual(codeKeys(source), ["shop.load.offline"]);
});

test("resource data names its keys in fields ending in Key", () => {
  const json = '{"id":"avatar.starter","nameKey":"cosmetic.avatar.starter.name","descriptionKey":"medal.first_win.description","staticAddress":"cosmetic/avatar"}';
  assert.deepEqual(resourceKeys(json), ["cosmetic.avatar.starter.name", "medal.first_win.description"]);
});

test("a key without an entry or without one of its required languages is reported", () => {
  const references = new Map([
    ["a.b.c", "x.cs"],
    ["a.b.d", "y.cs"],
    ["a.b.e", "z.json"],
  ]);
  const entries = [
    { key: "a.b.c", en: "C", az: "C" },
    { key: "a.b.d", en: "D", az: "" },
  ];
  assert.deepEqual(missingTexts(references, entries), [
    "a.b.d (y.cs) has no az text in client/Assets/HypeDomino/UI/Localization/Resources/builtin-strings.json",
    "a.b.e (z.json) has no en or az text in client/Assets/HypeDomino/UI/Localization/Resources/builtin-strings.json",
  ]);
});
