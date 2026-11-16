import type { SearchParams } from "../audit/query.ts";

export type ReviewSignal ={ kind: string; weight: number; evidence: unknown; at: string };

export type Review = {
  id: number;
  account_id: string;
  nickname: string;
  cause: string;
  opened_at: string;
  shadowbanned_at: string | null;
  signals: ReviewSignal[];
};

export function signalWeight(signals: readonly ReviewSignal[]): number {
  return signals.reduce((total, signal) => total + signal.weight, 0);
}

export function evidenceSummary(evidence: unknown): string {
  if (evidence === null || typeof evidence !== "object" || Array.isArray(evidence)) {
    return JSON.stringify(evidence) ?? "";
  }
  return Object.entries(evidence)
    .map(([key, value]) => `${key}: ${typeof value === "string" ? value : JSON.stringify(value)}`)
    .join("; ");
}

export const queuePageSize = 50;

export const reviewCauses = ["score", "hard_signal"] as const;

export const reportReasons = ["offensive_nickname", "offensive_room_name", "impersonation", "other"] as const;

export type ReviewQuery = { cause: (typeof reviewCauses)[number] | ""; account: string; after: string };

export type ReportQuery = { reason: (typeof reportReasons)[number] | ""; account: string; after: string };

const reviewIdPattern = /^[1-9][0-9]{0,17}$/;

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function single(value: string | string[] | undefined): string {
  return (Array.isArray(value) ? value[0] : value)?.trim() ?? "";
}

function known<T extends string>(values: readonly T[], value: string): T | "" {
  return values.find((candidate) => candidate === value) ?? "";
}

export function readReviewQuery(params: SearchParams): ReviewQuery {
  const after = single(params.after);
  return { cause: known(reviewCauses, single(params.cause)), account: single(params.account), after: reviewIdPattern.test(after) ? after : "" };
}

export function readReportQuery(params: SearchParams): ReportQuery {
  const after = single(params.after);
  return { reason: known(reportReasons, single(params.reason)), account: single(params.account), after: uuidPattern.test(after) ? after : "" };
}

function searchOf(query: ReviewQuery | ReportQuery): URLSearchParams {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value) {
      search.set(key, value);
    }
  }
  return search;
}

export function queueApiPath(path: "/admin/reviews" | "/admin/reports", query: ReviewQuery | ReportQuery): string {
  const search = searchOf(query);
  search.set("limit", String(queuePageSize));
  return `${path}?${search}`;
}

export function nextQueueHref(screen: "/reviews" | "/reports", query: ReviewQuery | ReportQuery, cursors: readonly string[]): string | null {
  if (cursors.length < queuePageSize) {
    return null;
  }
  return `${screen}?${searchOf({ ...query, after: cursors[cursors.length - 1] })}`;
}
