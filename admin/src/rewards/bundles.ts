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
};

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
