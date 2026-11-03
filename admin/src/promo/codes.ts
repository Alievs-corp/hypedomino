import type { Bundle } from "../shop/catalog.ts";

export type PromoKind = "single_use" | "multi_use" | "targeted";

export const promoKinds: readonly PromoKind[] = ["single_use", "multi_use", "targeted"];

export type PromoCode = {
  id: string;
  campaign: string;
  kind: PromoKind;
  bundle: Bundle;
  start: string;
  end: string;
  cap: number;
  state: string;
  creator: string;
  approver?: string;
  batch?: string;
  redemptions: number;
};

export type CampaignUsage = {
  campaign: string;
  codes: number;
  active: number;
  pending: number;
  disabled: number;
  cap: number;
  redemptions: number;
  start: string;
  end: string;
};

export type PromoDraft = {
  id: string;
  campaign: string;
  kind: PromoKind;
  bundleId: string;
  coins: number;
  cosmetics: string;
  passXp: number;
  startMinute: string;
  endMinute: string;
  cap: number;
  code: string;
  count: number;
};

export const emptyDraft: PromoDraft = {
  id: "promo.",
  campaign: "campaign.",
  kind: "multi_use",
  bundleId: "bundle.",
  coins: 0,
  cosmetics: "",
  passXp: 0,
  startMinute: "",
  endMinute: "",
  cap: 1,
  code: "",
  count: 100,
};

export type PromoRequest = {
  id: string;
  campaign: string;
  kind: PromoKind;
  bundle: Bundle;
  start: string;
  end: string;
  cap?: number;
  code?: string;
  count?: number;
};

export function listOf(text: string): string[] {
  return [...new Set(text.split(/[\s,;]+/).filter(Boolean))];
}

export function bundleOf(draft: PromoDraft): Bundle {
  const bundle: Bundle = { id: draft.bundleId.trim(), revision: 1 };
  const cosmetics = listOf(draft.cosmetics);
  if (draft.coins > 0) {
    bundle.coins = draft.coins;
  }
  if (cosmetics.length > 0) {
    bundle.cosmetics = cosmetics;
  }
  if (draft.passXp > 0) {
    bundle.passXp = draft.passXp;
  }
  return bundle;
}

export function promoRequest(draft: PromoDraft, batch: boolean): PromoRequest {
  const request: PromoRequest = {
    id: draft.id.trim(),
    campaign: draft.campaign.trim(),
    kind: batch ? "single_use" : draft.kind,
    bundle: bundleOf(draft),
    start: `${draft.startMinute}:00Z`,
    end: `${draft.endMinute}:00Z`,
  };
  if (batch) {
    return { ...request, count: draft.count };
  }
  if (draft.kind === "multi_use") {
    request.cap = draft.cap;
    if (draft.code.trim()) {
      request.code = draft.code.trim();
    }
  }
  return request;
}

export function draftReady(draft: PromoDraft, batch: boolean): boolean {
  const bundle = bundleOf(draft);
  const grants = (bundle.coins ?? 0) > 0 || (bundle.cosmetics?.length ?? 0) > 0 || (bundle.passXp ?? 0) > 0;
  const windowSet = draft.startMinute !== "" && draft.endMinute !== "" && draft.endMinute > draft.startMinute;
  if (!grants || !windowSet) {
    return false;
  }
  if (batch) {
    return Number.isInteger(draft.count) && draft.count >= 1 && draft.count <= maxBatchCount;
  }
  if (draft.kind === "multi_use") {
    return Number.isInteger(draft.cap) && draft.cap >= 1;
  }
  return true;
}

export const maxBatchCount = 10000;

export function shownState(code: PromoCode, now: Date): string {
  return code.state === "active" && new Date(code.end) <= now ? "expired" : code.state;
}

export function bundleSummary(bundle: Bundle): string {
  const parts: string[] = [];
  if (bundle.coins) {
    parts.push(`${bundle.coins} coins`);
  }
  if (bundle.cosmetics?.length) {
    parts.push(bundle.cosmetics.join(", "));
  }
  if (bundle.passXp) {
    parts.push(`${bundle.passXp} pass XP`);
  }
  return parts.join(" + ");
}

export function promoApiPath(campaign: string): string {
  return campaign ? `/admin/promo-codes?campaign=${encodeURIComponent(campaign)}` : "/admin/promo-codes";
}

export function disablePath(id: string, batch: boolean): string {
  return `/admin/${batch ? "promo-batches" : "promo-codes"}/${encodeURIComponent(id)}/disable`;
}

type PromoChange = { batch?: boolean; count?: number; code?: PromoCode };

export function promoChangeSummary(change: unknown): string | null {
  const promo = change as PromoChange | null;
  if (!promo || typeof promo !== "object" || !promo.code || typeof promo.code !== "object") {
    return null;
  }
  const { code } = promo;
  const codes = promo.batch ? `${promo.count} single_use codes` : code.kind;
  return `${codes}; ${code.campaign}; ${bundleSummary(code.bundle)}; ${code.start} to ${code.end}`;
}

export function batchLeads(codes: readonly PromoCode[]): Set<string> {
  const seen = new Set<string>();
  const leads = new Set<string>();
  for (const code of codes) {
    if (code.batch && !seen.has(code.batch)) {
      seen.add(code.batch);
      leads.add(code.id);
    }
  }
  return leads;
}

export type Redeemed = { account_id: string; redeemed_at: string };

export type PromoDetail = { code: PromoCode; personal?: { accounts: string[]; redeemed: Redeemed[] } };

export function detailPath(id: string): string {
  return `/admin/promo-codes/${encodeURIComponent(id)}`;
}

export function targetsPath(id: string): string {
  return `${detailPath(id)}/accounts`;
}

export function keepsRedeemers(accounts: readonly string[], redeemed: readonly Redeemed[]): boolean {
  return redeemed.every((entry) => accounts.includes(entry.account_id));
}
