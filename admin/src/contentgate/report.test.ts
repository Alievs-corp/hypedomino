import assert from "node:assert/strict";
import { test } from "node:test";
import { en } from "../i18n/messages/en.ts";
import { checkTitleKey, evidenceKey, exportFileName, exportPath, launchInstant, reportPath, resultKey, selectedReportId, tally, type GateSummary } from "./report.ts";

function summary(id: string): GateSummary {
  return { id, environment: "production", launch_at: "2027-01-15T12:00:00Z", ran_at: "2027-01-12T09:00:00Z", runner: "admin-1", verdict: "red", stale: false, results: [] };
}

test("every check of the gate has a title and unknown ids have none", () => {
  for (const id of ["CG-2", "CG-3", "CG-4", "CG-5", "CG-6", "CG-7", "CG-8", "CG-9", "CG-10", "CG-11", "CG-12", "CG-13"]) {
    const key = checkTitleKey(id);
    assert.ok(key && en[key], id);
  }
  assert.equal(checkTitleKey("CG-1"), null);
});

test("each result has its own message", () => {
  assert.equal(resultKey("pass"), "gate.result.pass");
  assert.equal(resultKey("fail"), "gate.result.fail");
  assert.equal(resultKey("unchecked"), "gate.result.unchecked");
});

test("report paths escape the id and the export is named after it", () => {
  assert.equal(reportPath("a/b"), "/admin/content-gate/reports/a%2Fb");
  assert.equal(exportPath("r1"), "/admin/content-gate/reports/r1/export");
  assert.equal(exportFileName("r1"), "content-gate-r1.json");
});

test("the launch time is read as a UTC minute and anything else is refused", () => {
  assert.equal(launchInstant("2027-01-15T12:00"), "2027-01-15T12:00:00Z");
  assert.equal(launchInstant(""), null);
  assert.equal(launchInstant("2027-01-15T12:00:00+04:00"), null);
});

test("only a green, current production report is evidence", () => {
  assert.equal(evidenceKey({ environment: "production", verdict: "green", stale: false }), "gate.evidence.green");
  assert.equal(evidenceKey({ environment: "production", verdict: "green", stale: true }), "gate.evidence.stale");
  assert.equal(evidenceKey({ environment: "production", verdict: "red", stale: false }), "gate.evidence.red");
  assert.equal(evidenceKey({ environment: "staging", verdict: "green", stale: false }), "gate.evidence.rehearsal");
});

test("the requested report is shown when it is in the history, otherwise the latest", () => {
  const history = [summary("new"), summary("old")];
  assert.equal(selectedReportId("old", history), "old");
  assert.equal(selectedReportId(["old", "new"], history), "old");
  assert.equal(selectedReportId("gone", history), "new");
  assert.equal(selectedReportId(undefined, history), "new");
  assert.equal(selectedReportId("old", []), null);
});

test("results are counted per outcome", () => {
  assert.deepEqual(tally([{ result: "pass" }, { result: "fail" }, { result: "unchecked" }, { result: "unchecked" }]), { pass: 1, fail: 1, unchecked: 2 });
});
