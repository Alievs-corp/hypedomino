import assert from "node:assert/strict";
import { test } from "node:test";
import { changeSummary, objectHref, ownProposal, type Proposal } from "./proposals.ts";

const grant: Proposal = {
  id: 7,
  kind: "player_grant",
  object_type: "players",
  object_id: "11111111-2222-3333-4444-555555555555",
  change: { coins: 5000, request_id: "r-1", proposer_id: "admin-1" },
  reason: "event compensation",
  proposer_id: "admin-1",
  proposer_roles: ["support"],
  proposed_at: "2026-11-04T09:00:00Z",
  expires_at: "2026-11-05T09:00:00Z",
  state: "pending",
};

test("a proposal is the admin's own only when they proposed it", () => {
  assert.equal(ownProposal(grant, "admin-1"), true);
  assert.equal(ownProposal(grant, "admin-2"), false);
});

test("player proposals link to the player record and others link nowhere", () => {
  assert.equal(objectHref(grant), "/players/11111111-2222-3333-4444-555555555555");
  assert.equal(objectHref({ ...grant, object_type: "admin_accounts" }), null);
});

test("the change summary shows what changes and hides bookkeeping ids", () => {
  assert.equal(changeSummary(grant.change), "coins: 5000");
  assert.equal(changeSummary({ roles: ["super_admin", "content"] }), "roles: super_admin, content");
  assert.equal(changeSummary("raw"), '"raw"');
});
