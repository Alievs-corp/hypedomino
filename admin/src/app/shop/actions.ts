"use server";

import { adminApi } from "@/admin/api";
import type { ConfigResult } from "@/config/result";
import type { ShopCatalog } from "@/shop/catalog";

export type PublishResult = ConfigResult;

export async function publishCatalog(catalog: ShopCatalog, reason: string): Promise<PublishResult> {
  const reply = await adminApi<{ config_version: number }>("PUT", "/admin/config/shop", { catalog, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}

export type PreviewSection = { section: string; ends_at?: string; offers: string[] };
export type PreviewResult = { ok: true; at: string; sections: PreviewSection[] } | { ok: false; error: string; detail: string };

export async function previewCatalog(catalog: ShopCatalog, at: string | undefined): Promise<PreviewResult> {
  const reply = await adminApi<{ at: string; sections: PreviewSection[] }>("POST", "/admin/config/shop/preview", { catalog, at });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, at: reply.body.at, sections: reply.body.sections };
}

export async function rollbackCatalog(version: number, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ config_version: number }>("POST", "/admin/config/shop/rollback", { version, reason });
  if (reply.status !== 200) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, version: reply.body.config_version };
}
