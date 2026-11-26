"use server";

import { adminApi } from "@/admin/api";
import type { KeyValues } from "@/config/keys";
import type { ConfigResult } from "@/config/result";

export async function rollbackAntiCheat(version: number, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ config_version: number }>("POST", "/admin/config/anticheat/rollback", { version, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}

export async function publishAntiCheat(values: KeyValues, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ config_version: number }>("PUT", "/admin/config/anticheat", { values, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}
