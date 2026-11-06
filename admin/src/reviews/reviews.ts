export type ReviewSignal = { kind: string; weight: number; evidence: unknown; at: string };

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
