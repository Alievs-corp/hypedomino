import type { MessageKey } from "../i18n/translate.ts";

export type BundleUse = { domain: string; object: string; detail?: string };

export type RewardBundle = {
  id: string;
  revision: number;
  coins: number;
  crystals?: number;
  cosmetics: string[] | null;
  medals: string[] | null;
  passXp: number;
  guideValue: number;
  usedBy: BundleUse[];
  differs: boolean;
  editable: boolean;
};

export type BundleEdit = { coins: number; crystals: number; cosmetics: string[]; medals: string[]; passXp: number };

export type BundleFilter = { query: string; domain: string; onlyDiffering: boolean };

export const emptyBundleFilter: BundleFilter = { query: "", domain: "", onlyDiffering: false };

function searchable(bundle: RewardBundle): string[] {
  return [bundle.id, ...(bundle.cosmetics ?? []), ...(bundle.medals ?? []), ...bundle.usedBy.map((use) => use.object)];
}

export function filteredBundles(bundles: RewardBundle[], filter: BundleFilter): RewardBundle[] {
  const query = filter.query.trim().toLowerCase();
  return bundles.filter(
    (bundle) =>
      (!filter.onlyDiffering || bundle.differs) &&
      (!filter.domain || bundle.usedBy.some((use) => use.domain === filter.domain)) &&
      (!query || searchable(bundle).some((value) => value.toLowerCase().includes(query))),
  );
}

export function usedDomains(bundles: RewardBundle[]): string[] {
  return [...new Set(bundles.flatMap((bundle) => bundle.usedBy.map((use) => use.domain)))].sort();
}

export function useLabel(use: BundleUse): string {
  return use.detail ? `${use.object} ${use.detail}` : use.object;
}

export function editOf(bundle: RewardBundle): BundleEdit {
  return { coins: bundle.coins, crystals: bundle.crystals ?? 0, cosmetics: bundle.cosmetics ?? [], medals: bundle.medals ?? [], passXp: bundle.passXp };
}

export function parsedIds(text: string): string[] {
  return [...new Set(text.split(",").map((id) => id.trim()).filter((id) => id !== ""))];
}

function sameList(a: string[], b: string[]): boolean {
  return [...a].sort().join() === [...b].sort().join();
}

export function bundleEditChanged(bundle: RewardBundle, edit: BundleEdit): boolean {
  const before = editOf(bundle);
  return before.coins !== edit.coins || before.crystals !== edit.crystals || before.passXp !== edit.passXp || !sameList(before.cosmetics, edit.cosmetics) || !sameList(before.medals, edit.medals);
}

export function bundleEditIssues(edit: BundleEdit): MessageKey[] {
  const found: MessageKey[] = [];
  if (![edit.coins, edit.crystals, edit.passXp].every((value) => Number.isInteger(value) && value >= 0)) found.push("rewards.issue.number");
  if (edit.coins === 0 && edit.crystals === 0 && edit.passXp === 0 && edit.cosmetics.length === 0 && edit.medals.length === 0) found.push("rewards.issue.empty");
  return found;
}
