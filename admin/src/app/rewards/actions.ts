"use server";

import { adminApi } from "@/admin/api";
import type { ConfigResult } from "@/config/result";
import type { BundleEdit } from "@/rewards/bundles";

export async function editRewardBundle(id: string, revision: number, edit: BundleEdit, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ bundle: { revision: number } }>("PUT", `/admin/reward-bundles/${encodeURIComponent(id)}`, { revision, ...edit, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.bundle.revision };
}
