"use server";

import { adminApi } from "@/admin/api";
import type { RuleValues } from "@/config/rules";

export type RollbackResult = { ok: true; version: number } | { ok: false; error: string; detail: string };

export async function rollbackRules(version: number, reason: string): Promise<RollbackResult> {
  const reply = await adminApi<{ config_version: number }>("POST", "/admin/config/rules/rollback", { version, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}

export type PublishResult = { ok: true; version: number } | { ok: false; error: string; detail: string };

export async function publishRules(values: RuleValues, reason: string): Promise<PublishResult> {
  const reply = await adminApi<{ config_version: number }>("PUT", "/admin/config/rules", { values, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}
