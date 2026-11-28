import assert from "node:assert/strict";
import { test } from "node:test";
import { badFields, clearsPolicyVersion, configFromFields, fieldsFromConfig, maxLinkLength, validLink } from "./config.ts";

test("an empty field leaves its key out of the snapshot", () => {
  const fields = fieldsFromConfig({ PrivacyPolicyVersion: "2026-11-01", ConsentRecordRetentionYears: 5 });
  assert.deepEqual(fields, { PrivacyPolicyVersion: "2026-11-01", TermsVersion: "", PrivacyPolicyUrl: "", TermsOfUseUrl: "", ConsentRecordRetentionYears: "5" });
  assert.deepEqual(configFromFields(fields), { PrivacyPolicyVersion: "2026-11-01", ConsentRecordRetentionYears: 5 });
  assert.deepEqual(configFromFields(fieldsFromConfig(null)), {});
});

test("versions follow the server format", () => {
  const valid = ["v", "2026-11-01", "v2.1_rc", "abcdefghijklmnopqrstuvwxyz012345"];
  for (const version of valid) {
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), PrivacyPolicyVersion: version, TermsVersion: version }), [], version);
  }
  const invalid = ["two words", "-leading", ".leading", "abcdefghijklmnopqrstuvwxyz0123456", "v/2", "versión"];
  for (const version of invalid) {
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), PrivacyPolicyVersion: version, TermsVersion: version }), ["PrivacyPolicyVersion", "TermsVersion"], version);
  }
});

test("legal links are https URLs with a host, like on the server", () => {
  const longest = "https://a.example/" + "x".repeat(maxLinkLength - 18);
  for (const link of ["https://hypedomino.example/privacy", "https://hypedomino.example/terms?lang=az#top", longest]) {
    assert.equal(validLink(link), true, link);
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), PrivacyPolicyUrl: link, TermsOfUseUrl: link }), [], link);
  }
  const invalid = [
    "http://hypedomino.example/privacy",
    "hypedomino.example/privacy",
    "https://user:pass@hypedomino.example/privacy",
    "https://hypedomino.example/a b",
    "javascript:alert(1)",
    longest + "x",
  ];
  for (const link of invalid) {
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), PrivacyPolicyUrl: link, TermsOfUseUrl: link }), ["PrivacyPolicyUrl", "TermsOfUseUrl"], link);
  }
  const fields = { ...fieldsFromConfig(null), PrivacyPolicyUrl: "https://hypedomino.example/privacy" };
  assert.deepEqual(configFromFields(fields), { PrivacyPolicyUrl: "https://hypedomino.example/privacy" });
  assert.deepEqual(fieldsFromConfig(configFromFields(fields)), fields);
});

test("the consent record retention is a whole number of years from 1 to 10", () => {
  for (const years of ["1", "3", "10"]) {
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), ConsentRecordRetentionYears: years }), [], years);
  }
  for (const years of ["0", "11", "2.5", "-1", "three"]) {
    assert.deepEqual(badFields({ ...fieldsFromConfig(null), ConsentRecordRetentionYears: years }), ["ConsentRecordRetentionYears"], years);
  }
});

test("only clearing a published privacy policy version stops consent recording", () => {
  const published = { PrivacyPolicyVersion: "2026-11-01", TermsVersion: "2026-10-01" };
  assert.equal(clearsPolicyVersion(published, { TermsVersion: "2026-10-01" }), true);
  assert.equal(clearsPolicyVersion(published, { PrivacyPolicyVersion: "2026-12-01" }), false);
  assert.equal(clearsPolicyVersion({ TermsVersion: "2026-10-01" }, {}), false);
  assert.equal(clearsPolicyVersion(null, {}), false);
});
