import type { MessageKey } from "../i18n/translate.ts";

export const reportsPath = "/admin/content-gate/reports";

export type CheckResult = "pass" | "fail" | "unchecked";

export type GateFailure = { object: string; rule: string; detail?: string };

export type GateGap = { rule: string; reason: string };

export type GateCheck = {
  id: string;
  result: CheckResult;
  count?: number;
  minimum?: number;
  failures: GateFailure[];
  unchecked: GateGap[];
};

export type GateThreshold = { key: string; value: number; default: number; config_version: number | null };

export type GateReport = {
  id: string;
  environment: string;
  launch_at: string;
  ran_at: string;
  runner: string;
  verdict: "green" | "red";
  versions: { config_version: number; domains: Record<string, number>; manifest_version: number; string_bundles: Record<string, number> };
  thresholds: GateThreshold[];
  checks: GateCheck[];
  info: { ready: Record<string, Record<string, number>>; live_not_ready: Record<string, number> };
};

export type GateReportView = { report: GateReport; stale: boolean };

export type GateSummary = {
  id: string;
  environment: string;
  launch_at: string;
  ran_at: string;
  runner: string;
  verdict: "green" | "red";
  stale: boolean;
  results: { id: string; result: CheckResult }[];
};

const checkTitles: Readonly<Record<string, MessageKey>> = {
  "CG-2": "gate.check.cg2",
  "CG-3": "gate.check.cg3",
  "CG-4": "gate.check.cg4",
  "CG-5": "gate.check.cg5",
  "CG-6": "gate.check.cg6",
  "CG-7": "gate.check.cg7",
  "CG-8": "gate.check.cg8",
  "CG-9": "gate.check.cg9",
  "CG-10": "gate.check.cg10",
  "CG-11": "gate.check.cg11",
  "CG-12": "gate.check.cg12",
  "CG-13": "gate.check.cg13",
};

const resultKeys: Readonly<Record<CheckResult, MessageKey>> = {
  pass: "gate.result.pass",
  fail: "gate.result.fail",
  unchecked: "gate.result.unchecked",
};

export function checkTitleKey(id: string): MessageKey | null {
  return checkTitles[id] ?? null;
}

export function resultKey(result: CheckResult): MessageKey {
  return resultKeys[result];
}

export function reportPath(id: string): string {
  return `${reportsPath}/${encodeURIComponent(id)}`;
}

export function exportPath(id: string): string {
  return `${reportPath(id)}/export`;
}

export function exportFileName(id: string): string {
  return `content-gate-${id}.json`;
}

export function launchInstant(minute: string): string | null {
  return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(minute) ? `${minute}:00Z` : null;
}

export function evidenceKey(view: { environment: string; verdict: string; stale: boolean }): MessageKey {
  if (view.environment !== "production") {
    return "gate.evidence.rehearsal";
  }
  if (view.stale) {
    return "gate.evidence.stale";
  }
  return view.verdict === "green" ? "gate.evidence.green" : "gate.evidence.red";
}

export function selectedReportId(requested: string | string[] | undefined, history: readonly GateSummary[]): string | null {
  const wanted = Array.isArray(requested) ? requested[0] : requested;
  if (wanted && history.some((summary) => summary.id === wanted)) {
    return wanted;
  }
  return history[0]?.id ?? null;
}

export function tally(results: readonly { result: CheckResult }[]): Record<CheckResult, number> {
  const counts: Record<CheckResult, number> = { pass: 0, fail: 0, unchecked: 0 };
  for (const entry of results) {
    counts[entry.result] += 1;
  }
  return counts;
}
