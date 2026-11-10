"use server";

import { adminApi, adminApiResponse, adminApiText, type ApiRefusal } from "@/admin/api";
import { parseReply } from "@/admin/forwarding";
import type { ConfigResult } from "@/config/result";
import { problemsText, type Placeholder, type Problem, type Status } from "@/strings/strings";

export type Refused = { ok: false; error: string; detail: string };

export type SavedResult = { ok: true } | Refused;

export type ExportResult = { ok: true; text: string; filename: string } | Refused;

export type ImportResult = { ok: true; changed: number } | Refused;

export type KeyFields = { note: string; max_length: number; placeholders: Placeholder[]; deprecated: boolean };

type Blocked = ApiRefusal & { problems?: Problem[] };

type Active = { config_version: number; bundles: Record<string, number>; fallback_chains: Record<string, string[]> };

function refused(status: number, body: Blocked): Refused {
  return { ok: false, error: body.error ?? String(status), detail: body.problems ? problemsText(body.problems) : (body.detail ?? "") };
}

function keyPath(key: string): string {
  return `/admin/strings/${encodeURIComponent(key)}`;
}

function transferPath(language: string, format: string): string {
  return `?language=${encodeURIComponent(language)}&format=${encodeURIComponent(format)}`;
}

export async function saveKey(key: string, fields: KeyFields, reason: string): Promise<SavedResult> {
  const reply = await adminApi<object>("PUT", keyPath(key), { ...fields, reason });
  return reply.status === 200 ? { ok: true } : refused(reply.status, reply.body);
}

export async function saveValue(key: string, language: string, text: string, status: Status): Promise<SavedResult> {
  const reply = await adminApi<object>("PUT", `${keyPath(key)}/values/${encodeURIComponent(language)}`, { text, status });
  return reply.status === 200 ? { ok: true } : refused(reply.status, reply.body);
}

export async function exportStrings(language: string, format: string): Promise<ExportResult> {
  const response = await adminApiResponse("GET", `/admin/strings/export${transferPath(language, format)}`);
  const text = await response.text();
  if (response.status !== 200) {
    return refused(response.status, parseReply<Blocked>(text));
  }
  return { ok: true, text, filename: `strings-${language}.${format}` };
}

export async function importStrings(language: string, format: string, text: string): Promise<ImportResult> {
  const reply = await adminApiText<{ changed: number }>("POST", `/admin/strings/import${transferPath(language, format)}`, text);
  return reply.status === 200 ? { ok: true, changed: reply.body.changed } : refused(reply.status, reply.body as Blocked);
}

export async function publishBundle(language: string, reason: string): Promise<ConfigResult> {
  const reply = await adminApi<{ version: number }>("POST", `/admin/strings/bundles/${encodeURIComponent(language)}`, { reason });
  return reply.status === 200 ? { ok: true, version: reply.body.version } : refused(reply.status, reply.body as Blocked);
}

export async function activateBundle(language: string, version: number, reason: string): Promise<ConfigResult> {
  const active = await adminApi<Active>("GET", "/admin/strings/active");
  if (active.status !== 200) {
    return refused(active.status, active.body);
  }
  const bundles = { ...active.body.bundles, [language]: version };
  const reply = await adminApi<Active>("PUT", "/admin/strings/active", { bundles, fallback_chains: active.body.fallback_chains, reason });
  return reply.status === 200 ? { ok: true, version: reply.body.config_version } : refused(reply.status, reply.body);
}
