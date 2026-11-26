"use server";

import { adminApi } from "@/admin/api";
import type { ConfigResult } from "@/config/result";
import type { PrivacyConfig } from "@/privacy/config";

export async function publishPrivacy(catalog: PrivacyConfig, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ config_version: number }>("PUT", "/admin/config/privacy", { catalog, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}

export async function rollbackPrivacy(version: number, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ config_version: number }>("POST", "/admin/config/privacy/rollback", { version, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}
