import assert from "node:assert/strict";
import { test } from "node:test";
import { banActive, muteActive, playerApiPath, playerSearch, playersApiPath, searchQuery, signedAmount, validAdjustment, type PlayerAccount } from "./record.ts";

test("the search query is the first trimmed q value", () => {
  assert.equal(searchQuery({ q: "  ayla " }), "ayla");
  assert.equal(searchQuery({ q: ["first", "second"] }), "first");
  assert.equal(searchQuery({}), "");
});

test("query and id are encoded into the admin API paths", () => {
  assert.equal(playersApiPath({ q: "ayla@example.com" }), "/admin/players?q=ayla%40example.com");
  assert.equal(playersApiPath({ q: "a&b=c" }), "/admin/players?q=a%26b%3Dc");
  assert.equal(playersApiPath({ provider: "apple", subject: "001.a&b" }), "/admin/players?provider=apple&subject=001.a%26b");
  assert.equal(playerApiPath("../admin/me"), "/admin/players/..%2Fadmin%2Fme");
});

const account: PlayerAccount = { id: "p", nickname: "n", state: "active", country: "AZ", created_at: "2026-11-01T00:00:00Z", emails: [], banned_until: null, shadowbanned_at: null, muted_until: null, deletion_requested_at: null, deleted_at: null };

test("a ban counts only until its end", () => {
  const now = new Date("2026-11-04T10:00:00Z");
  assert.equal(banActive(account, now), false);
  assert.equal(banActive({ ...account, banned_until: "2026-11-04T12:00:00Z" }, now), true);
  assert.equal(banActive({ ...account, banned_until: "2026-11-04T09:00:00Z" }, now), false);
});

test("a mute counts only until its end", () => {
  const now = new Date("2026-11-04T10:00:00Z");
  assert.equal(muteActive(account, now), false);
  assert.equal(muteActive({ ...account, muted_until: "2026-11-05T10:00:00Z" }, now), true);
  assert.equal(muteActive({ ...account, muted_until: "2026-11-04T09:00:00Z" }, now), false);
});

test("ledger amounts show their sign", () => {
  assert.equal(signedAmount(50), "+50");
  assert.equal(signedAmount(-20), "-20");
  assert.equal(signedAmount(0), "0");
});

test("a coin adjustment is a whole, non-zero number of coins, credit or debit", () => {
  assert.ok(validAdjustment(100));
  assert.ok(validAdjustment(-40));
  assert.ok(!validAdjustment(0));
  assert.ok(!validAdjustment(1.5));
  assert.ok(!validAdjustment(Number.NaN));
  assert.ok(!validAdjustment(2 ** 60));
});

test("a subject searches by auth identity, otherwise q searches the text", () => {
  assert.deepEqual(playerSearch({ provider: "apple", subject: " 001.abc ", q: "ayla" }), { provider: "apple", subject: "001.abc" });
  assert.deepEqual(playerSearch({ provider: "other", subject: "1234" }), { provider: "google", subject: "1234" });
  assert.deepEqual(playerSearch({ q: " ayla ", subject: " " }), { q: "ayla" });
  assert.equal(playerSearch({}), null);
});
