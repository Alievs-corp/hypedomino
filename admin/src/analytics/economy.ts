import type { MessageKey } from "@/i18n/translate";

export type EconomyAlert = "sink_ratio" | "account_daily_earn" | "median_balance";

export type EconomyReport = {
  from: string;
  to: string;
  day: string;
  coinsCreated: number;
  coinsSpent: number;
  sinkRatioPercent: number | null;
  medianBalance: number;
  sourceMix: Record<string, number>;
  accountsOverDailyEarn: number;
  alerts: EconomyAlert[];
};

export type EconomyComparison = { config_version: number; domain: string; published_at: string; before: EconomyReport; after: EconomyReport };

export const alertLabels: Readonly<Record<EconomyAlert, MessageKey>> = {
  sink_ratio: "analytics.alert.sink_ratio",
  account_daily_earn: "analytics.alert.account_daily_earn",
  median_balance: "analytics.alert.median_balance",
};

export function sourceReasons(reports: readonly EconomyReport[]): string[] {
  const totals = new Map<string, number>();
  for (const report of reports) {
    for (const [reason, coins] of Object.entries(report.sourceMix)) {
      totals.set(reason, (totals.get(reason) ?? 0) + coins);
    }
  }
  return [...totals.keys()].sort((a, b) => totals.get(b)! - totals.get(a)! || a.localeCompare(b));
}

export function comparedVersion(value: string | undefined): number | null {
  return value && /^[1-9][0-9]{0,14}$/.test(value) ? Number(value) : null;
}
