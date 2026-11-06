"use server";

import { adminApi } from "@/admin/api";
import type { LiveControls } from "@/config/liveControls";

export type SwitchResult = { ok: true; version: number; approvalId: number | null } | { ok: false; error: string; detail: string };

export async function switchLiveControls(controls: LiveControls, reason: string): Promise<SwitchResult> {
  const reply = await adminApi<{ config_version: number; approval_id?: number }>("PUT", "/admin/live-controls", { controls, reason });
  if (reply.status !== 200 && reply.status !== 202) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version, approvalId: reply.body.approval_id ?? null };
}
