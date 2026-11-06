import type { MessageKey, Translate } from "../i18n/translate.ts";

export type AssetStatus = "awaiting_file" | "passed" | "rejected";

export type Asset = {
  id: string;
  item_id: string;
  slot: string;
  address: string;
  status: AssetStatus;
  format: string | null;
  width: number | null;
  height: number | null;
  bytes: number | null;
  content_hash: string | null;
  checks: Record<string, boolean> | null;
  created_at: string;
  view_url?: string;
};

export type AssetRefusal = { ok: false; error: string; detail: string };
export type AssetUploadStart = { ok: true; id: string; uploadUrl: string } | AssetRefusal;
export type AssetCheckResult = { ok: true } | AssetRefusal;

const itemIdPattern = /^(avatar|character|frame|table|emoji|action)\.[a-z0-9_]+$/;

const checkLabels: Record<string, MessageKey> = {
  format: "assets.check.format",
  alpha: "assets.check.alpha",
};

const statusLabels: Record<AssetStatus, MessageKey> = {
  awaiting_file: "assets.status.awaiting_file",
  passed: "assets.status.passed",
  rejected: "assets.status.rejected",
};

export function itemIdValid(id: string): boolean {
  return itemIdPattern.test(id);
}

export function statusLabel(status: AssetStatus, t: Translate): string {
  return t(statusLabels[status]);
}

export function failedCheckLabels(asset: Asset, t: Translate): string[] {
  const failed = Object.entries(asset.checks ?? {}).filter(([, passed]) => !passed);
  return failed.map(([name]) => (name in checkLabels ? t(checkLabels[name]) : name));
}

export function sizeLabel(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
