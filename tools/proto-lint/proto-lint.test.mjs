import assert from "node:assert/strict";
import { test } from "node:test";
import { forbiddenTokens, isClientMessage, lintSources, parseProto } from "./proto-lint.mjs";

const clean = `syntax = "proto3";

package hypedomino.v1;

message Tile {
  uint32 low = 1;
  uint32 high = 2;
}

message PlaceTileIntent {
  Tile tile = 1;
  BoardEnd end = 2;
}

message PassIntent {}

message WalletState {
  int64 balance = 1;
}
`;

test("parses messages, nested messages, oneofs and empty messages", () => {
  const source = `message Envelope {
  uint64 sequence = 1;
  oneof payload {
    ClientHello client_hello = 10;
  }
  message Inner {
    repeated string names = 1;
    map<string, int32> counts = 2;
  }
}

enum Platform {
  PLATFORM_UNSPECIFIED = 0;
}

message Empty {}
`;
  const messages = parseProto(source);
  assert.deepEqual([...messages.keys()], ["Envelope", "Envelope.Inner", "Empty"]);
  assert.deepEqual(
    messages.get("Envelope").fields.map((field) => [field.type, field.name, field.line]),
    [
      ["uint64", "sequence", 2],
      ["ClientHello", "client_hello", 4],
    ],
  );
  assert.deepEqual(
    messages.get("Envelope.Inner").fields.map((field) => [field.type, field.name]),
    [
      ["string", "names"],
      ["map<string, int32>", "counts"],
    ],
  );
  assert.deepEqual(messages.get("Empty").fields, []);
});

test("client messages are intents and the client hello", () => {
  assert.equal(isClientMessage("PlaceTileIntent"), true);
  assert.equal(isClientMessage("ClientHello"), true);
  assert.equal(isClientMessage("ServerHello"), false);
  assert.equal(isClientMessage("MatchSnapshot"), false);
});

test("forbidden tokens match whole snake_case parts only", () => {
  assert.deepEqual(forbiddenTokens("coin_balance"), ["coin", "balance"]);
  assert.deepEqual(forbiddenTokens("ranked"), []);
  assert.deepEqual(forbiddenTokens("help"), []);
  assert.deepEqual(forbiddenTokens("level_count"), []);
});

test("clean contracts pass and server messages may carry values", () => {
  assert.deepEqual(lintSources([{ path: "clean.proto", source: clean }]), []);
});

test("a value field on an intent is reported with its location", () => {
  const source = `message BuyOfferIntent {
  string offer_id = 1;
  uint32 price = 2;
}
`;
  const violations = lintSources([{ path: "shop.proto", source }]);
  assert.deepEqual(violations, [
    { path: "shop.proto", line: 3, message: "BuyOfferIntent", field: "price", token: "price", root: "BuyOfferIntent" },
  ]);
});

test("a value field reached through a nested type is reported against the intent", () => {
  const source = `message Claim {
  string id = 1;
  uint32 reward_coins = 2;
}

message ClaimRewardIntent {
  Claim claim = 1;
}
`;
  const violations = lintSources([{ path: "claim.proto", source }]);
  assert.deepEqual(
    violations.map(({ message, field, token, root }) => ({ message, field, token, root })),
    [
      { message: "Claim", field: "reward_coins", token: "reward", root: "ClaimRewardIntent" },
      { message: "Claim", field: "reward_coins", token: "coins", root: "ClaimRewardIntent" },
    ],
  );
});

test("types are resolved across files and through package prefixes", () => {
  const tile = `message Tile {
  uint32 score = 1;
}
`;
  const intent = `message PlaceTileIntent {
  hypedomino.v1.Tile tile = 1;
}
`;
  const violations = lintSources([
    { path: "tile.proto", source: tile },
    { path: "intent.proto", source: intent },
  ]);
  assert.equal(violations.length, 1);
  assert.equal(violations[0].path, "tile.proto");
  assert.equal(violations[0].root, "PlaceTileIntent");
});

test("recursive types terminate", () => {
  const source = `message Node {
  Node next = 1;
}

message WalkIntent {
  Node start = 1;
}
`;
  assert.deepEqual(lintSources([{ path: "node.proto", source }]), []);
});
