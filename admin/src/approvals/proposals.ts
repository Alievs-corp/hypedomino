import { promoChangeSummary } from "../promo/codes.ts";

export type Proposal = {
  id: number;
  kind: string;
  object_type: string;
  object_id: string;
  change: unknown;
  reason: string;
  proposer_id: string;
  proposer_roles: string[];
  proposed_at: string;
  expires_at: string;
  state: string;
};

export function ownProposal(proposal: Proposal, accountId: string): boolean {
  return proposal.proposer_id === accountId;
}

export function objectHref(proposal: Proposal): string | null {
  if (proposal.object_type === "promo_codes") {
    return "/promo";
  }
  return proposal.object_type === "players" ? `/players/${encodeURIComponent(proposal.object_id)}` : null;
}

export function changeSummary(change: unknown): string {
  const promo = promoChangeSummary(change);
  if (promo !== null) {
    return promo;
  }
  if (change === null || typeof change !== "object" || Array.isArray(change)) {
    return JSON.stringify(change);
  }
  return Object.entries(change)
    .filter(([key]) => key !== "proposer_id" && key !== "request_id")
    .map(([key, value]) => `${key}: ${Array.isArray(value) ? value.join(", ") : JSON.stringify(value)}`)
    .join("; ");
}
