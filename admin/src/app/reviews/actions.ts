"use server";

import { adminApi } from "@/admin/api";

export type ReviewDecision = "clear" | "confirm" | "permanent-ban";

export type ReviewResult = { ok: true } | { ok: false; error: string; detail: string };

export async function decideReview(id: number, decision: ReviewDecision, reason: string): Promise<ReviewResult> {
  const reply = await adminApi<{ account_id: string }>("POST", `/admin/reviews/${id}/${decision}`, { reason });
  if (reply.status !== 200 && reply.status !== 202) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true };
}
