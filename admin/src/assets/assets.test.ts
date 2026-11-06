import assert from "node:assert/strict";
import test from "node:test";
import type { Translate } from "../i18n/translate.ts";
import { failedCheckLabels, itemIdValid, sizeLabel, statusLabel, type Asset } from "./assets.ts";

const t = ((key: string) => `<${key}>`) as Translate;

const asset: Asset = {
  id: "a-1",
  item_id: "avatar.fox",
  slot: "static",
  address: "cosmetic/avatar/fox/static",
  status: "rejected",
  format: null,
  width: null,
  height: null,
  bytes: 19,
  content_hash: "abc",
  checks: { format: false, alpha: false },
  created_at: "2026-11-06T10:00:00Z",
};

test("an item id names a known type and a slug", () => {
  for (const id of ["avatar.fox", "character.ayla_2", "frame.gold", "table.night", "emoji.wow", "action.dance"]) {
    assert.equal(itemIdValid(id), true, id);
  }
  for (const id of ["", "fox", "pet.fox", "avatar.", "avatar.Fox", "../avatar.fox", "avatar.fox/static", "avatar.fox\n"]) {
    assert.equal(itemIdValid(id), false, id);
  }
});

test("only failed checks are named, and an unknown check still shows", () => {
  assert.deepEqual(failedCheckLabels(asset, t), ["<assets.check.format>", "<assets.check.alpha>"]);
  assert.deepEqual(failedCheckLabels({ ...asset, checks: { format: true, alpha: false } }, t), ["<assets.check.alpha>"]);
  assert.deepEqual(failedCheckLabels({ ...asset, checks: { atlas: false } }, t), ["atlas"]);
  assert.deepEqual(failedCheckLabels({ ...asset, checks: { format: true, alpha: true } }, t), []);
  assert.deepEqual(failedCheckLabels({ ...asset, checks: null }, t), []);
});

test("every status has its own label", () => {
  const labels = (["awaiting_file", "passed", "rejected"] as const).map((status) => statusLabel(status, t));
  assert.equal(new Set(labels).size, 3);
});

test("sizes read in the nearest unit", () => {
  assert.equal(sizeLabel(0), "0 B");
  assert.equal(sizeLabel(1023), "1023 B");
  assert.equal(sizeLabel(1536), "1.5 KB");
  assert.equal(sizeLabel(1572864), "1.5 MB");
});
