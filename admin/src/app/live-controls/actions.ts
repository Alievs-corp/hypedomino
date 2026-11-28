"use server";

import { adminApi } from "@/admin/api";
import type { Incident, LiveControls } from "@/config/liveControls";

export type SwitchResult = { ok: true; version: number; approvalId: number | null; breakGlassId: number | null } | { ok: false; error: string; detail: string; retryAt: string | null };

export async function switchLiveControls(controls: LiveControls, reason: string, breakGlass: Incident | null): Promise<SwitchResult> {
  const body = breakGlass === null ? { controls, reason } : { controls, reason, ref: breakGlass.id, break_glass: breakGlass };
  const reply = await adminApi<{ config_version: number; approval_id?: number; break_glass_id?: number; retry_at?: string }>("PUT", "/admin/live-controls", body);
  if (reply.status !== 200 && reply.status !== 202) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "", retryAt: reply.body.retry_at ?? null };
  }
  return { ok: true, version: reply.body.config_version, approvalId: reply.body.approval_id ?? null, breakGlassId: reply.body.break_glass_id ?? null };
}
