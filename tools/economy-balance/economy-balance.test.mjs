import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { checkEconomy, parseBalance, renderBalance } from "./economy-balance.mjs";

const repoRoot = new URL("../../", import.meta.url);
const csvText = readFileSync(new URL("docs/economy/balance.csv", repoRoot), "utf8");
const docText = readFileSync(new URL("docs/economy.md", repoRoot), "utf8");

const withValue = (key, value) => csvText.split("\n").map((line) => (line.startsWith(`${key},`) ? line.replace(/^([^,]+),[^,]+,/, `$1,${value},`) : line)).join("\n");

function assertReported(errors, fragment) {
  assert.ok(errors.some((error) => error.includes(fragment)), `expected "${fragment}" in:\n${errors.join("\n")}`);
}

test("the committed balance sheet is valid and matches docs/economy.md", () => {
  assert.deepEqual(checkEconomy(csvText, docText), []);
});

test("the earn rate per mode follows match length, win rate and rewards", () => {
  const rendered = renderBalance(parseBalance(csvText).rows);
  assert.match(rendered, /\| Ranked 1v1 \| 30 \| 10 \| 5\.45 \| 109 \| 3\.7 \|/);
});

test("a changed value makes the document stale", () => {
  assertReported(checkEconomy(withValue("MatchCoinsRanked1v1", "35;10"), docText), "balance sheet is stale");
});

test("a key the document does not name is reported", () => {
  assertReported(checkEconomy(`${csvText}UnusedKey,1,coins,config,10\n`, docText), "does not name UnusedKey");
});

test("a duplicate key is rejected", () => {
  assertReported(parseBalance(`${csvText}FirstWinBonusCoins,5,coins,config,2.2\n`).errors, "duplicate key FirstWinBonusCoins");
});

test("a non-numeric value is rejected", () => {
  assertReported(parseBalance(withValue("RenameCoinPrice", "free")).errors, "is not a number list");
});

test("a list of the wrong length is rejected", () => {
  assertReported(parseBalance(withValue("GuidePriceCoins", "400;1500;3500;12000")).errors, "GuidePriceCoins has 4 values, expected 5");
});

test("an unknown kind is rejected", () => {
  assertReported(parseBalance(csvText.replace("PromoLockMinutes,60,minutes,config", "PromoLockMinutes,60,minutes,guess")).errors, "is not config or model");
});

test("a wrong header is rejected", () => {
  assert.deepEqual(parseBalance(csvText.replace("key,value", "name,value")).errors, ["header must be key,value,unit,kind,section"]);
});
