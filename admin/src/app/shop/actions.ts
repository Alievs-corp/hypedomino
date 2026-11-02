"use server";

import { adminApi } from "@/admin/api";
import type { ShopCatalog } from "@/shop/catalog";

export type PublishResult = { ok: true; version: number } | { ok: false; error: string; detail: string };

export async function publishCatalog(catalog: ShopCatalog, reason: string): Promise<PublishResult> {
  const reply = await adminApi<{ config_version: number }>("PUT", "/admin/config/shop", { catalog, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}
