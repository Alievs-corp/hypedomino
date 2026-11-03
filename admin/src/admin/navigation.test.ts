import assert from "node:assert/strict";
import { test } from "node:test";
import { adminApiHeaders, parseReply } from "./forwarding.ts";
import { canView, visibleScreens, type Screen } from "./navigation.ts";
import { signInRefusalKey } from "./refusals.ts";

const shop: Screen = { href: "/shop", labelKey: "nav.shop", object: "shop_offers" };
const audit: Screen = { href: "/audit", labelKey: "nav.shop", object: "audit_log" };

test("an account without permissions sees no screen", () => {
  assert.deepEqual(visibleScreens({}, [shop, audit]), []);
});

test("a screen shows only with view on its object", () => {
  assert.deepEqual(visibleScreens({ shop_offers: ["view", "publish"] }, [shop, audit]), [shop]);
  assert.deepEqual(visibleScreens({ shop_offers: ["view"], audit_log: ["view"] }, [shop, audit]), [shop, audit]);
});

test("other actions without view do not show a screen", () => {
  assert.equal(canView({ shop_offers: ["publish", "edit"] }, "shop_offers"), false);
  assert.equal(canView({ audit_log: ["view"] }, "shop_offers"), false);
});

function incoming(values: Record<string, string>) {
  return { get: (name: string) => values[name] ?? null };
}

test("admin API calls carry the token, the browser address and the user agent", () => {
  const sent = adminApiHeaders(incoming({ "x-forwarded-for": "203.0.113.7", "user-agent": "Browser/1.0" }), "token-1");
  assert.deepEqual(sent, {
    "Content-Type": "application/json",
    Authorization: "Bearer token-1",
    "X-Forwarded-For": "203.0.113.7",
    "User-Agent": "Browser/1.0",
  });
});

test("a sign-in call without a session sends no authorization", () => {
  assert.deepEqual(adminApiHeaders(incoming({}), undefined), { "Content-Type": "application/json" });
});

test("an empty reply body reads as an empty object", () => {
  assert.deepEqual(parseReply(""), {});
  assert.deepEqual(parseReply('{"error":"forbidden"}'), { error: "forbidden" });
});

test("sign-in refusals map to a message and unknown codes to the generic one", () => {
  assert.equal(signInRefusalKey("account_refused"), "signin.account_refused");
  assert.equal(signInRefusalKey("not_company_account"), "signin.not_company_account");
  assert.equal(signInRefusalKey("nonce_rejected"), "signin.retry");
  assert.equal(signInRefusalKey("token_expired"), "signin.retry");
  assert.equal(signInRefusalKey("provider_unavailable"), "signin.unavailable");
  assert.equal(signInRefusalKey("500"), "signin.failed");
});
