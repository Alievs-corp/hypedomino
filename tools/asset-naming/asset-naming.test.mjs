import { test } from "node:test";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import { checkAssetPaths, listFiles } from "./asset-naming.mjs";

const character = "cosmetic/character/comet_fox";
const validTree = [
  "README.md",
  `${character}/provenance.md`,
  `${character}/concept/20261026-01.png`,
  `${character}/work/comet_fox.psb`,
  `${character}/work/comet_fox_head.ai`,
  `${character}/export/static.png`,
  `${character}/export/clip/idle/0001.png`,
  `${character}/export/clip/win/0001.png`,
  `${character}/export/model/comet_fox.fbx`,
  `${character}/export/model_hq/comet_fox.fbx`,
  "cosmetic/avatar/comet_fox/provenance.md",
  "cosmetic/avatar/comet_fox/export/static.png",
  "cosmetic/avatar/comet_fox/export/anim/sheet.png",
  "cosmetic/table/night_garden/provenance.md",
  "cosmetic/table/night_garden/export/static.png",
  "cosmetic/table/night_garden/export/layer/back/0001.png",
  "medal/first_win/provenance.md",
  "medal/first_win/export/static.png",
  "board/b_5/provenance.md",
  "board/b_5/work/b_5.psd",
];

function assertReported(paths, fragment) {
  const errors = checkAssetPaths(paths);
  assert.ok(errors.some((error) => error.includes(fragment)), `expected "${fragment}" in:\n${errors.join("\n")}`);
}

test("a tree that follows the convention passes", () => {
  assert.deepEqual(checkAssetPaths(validTree), []);
});

test("the repository assets-src passes", () => {
  const directory = fileURLToPath(new URL("../../assets-src/", import.meta.url));
  assert.deepEqual(checkAssetPaths(listFiles(directory)), []);
});

test("capitals, spaces and unknown classes are rejected", () => {
  assertReported([`${character}/work/Comet Fox.psb`, `${character}/provenance.md`], "invalid name Comet Fox.psb");
  assertReported(["props/lamp/provenance.md"], "unknown asset class props");
  assertReported(["cosmetic/hat/top_hat/provenance.md"], "unknown cosmetic type hat");
  assertReported(["board/b_11/provenance.md"], "invalid slug b_11");
});

test("an item folder without provenance is rejected", () => {
  assertReported([`${character}/concept/20261026-01.png`], "missing provenance.md");
});

test("an export without the static fallback is rejected", () => {
  assertReported([`${character}/provenance.md`, `${character}/export/clip/idle/0001.png`], "missing export/static.png");
});

test("export folders must match the cosmetic type and its address slots", () => {
  assertReported([`${character}/provenance.md`, `${character}/export/static.png`, `${character}/export/anim/0001.png`], "export/anim is not allowed for character");
  assertReported([`${character}/provenance.md`, `${character}/export/static.png`, `${character}/export/clip/dance/0001.png`], "unknown clip dance");
  assertReported([`${character}/provenance.md`, `${character}/export/static.png`, `${character}/export/model`], "export/model needs files inside a folder");
  assertReported(["cosmetic/table/t/provenance.md", "cosmetic/table/t/export/static.png", "cosmetic/table/t/export/layer/0001.png"], "invalid layer name 0001.png");
  assertReported(["cosmetic/emoji/e/provenance.md", "cosmetic/emoji/e/export/static.png", "cosmetic/emoji/e/export/anim"], "needs files inside a folder");
});

test("concept and work names follow the convention", () => {
  assertReported([`${character}/provenance.md`, `${character}/concept/final_v2.png`], "must be concept/<YYYYMMDD>-<nn>");
  assertReported([`${character}/provenance.md`, `${character}/work/other_fox.psd`], "must be work/<slug>[_<part>].<ext>");
  assertReported([`${character}/provenance.md`, `${character}/notes.txt`], "notes.txt is not allowed in an item folder");
});
