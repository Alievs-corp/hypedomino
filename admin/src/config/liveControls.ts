export type LiveControls = {
  minimum_app_version: string;
  build_allowlist_enforced: boolean;
  build_allowlist: string[];
  purchases_disabled: boolean;
  ranked_disabled: boolean;
};

export type LiveControlsView = { config_version: number; controls: LiveControls | null };

export const unpublishedControls: LiveControls = {
  minimum_app_version: "",
  build_allowlist_enforced: false,
  build_allowlist: [],
  purchases_disabled: false,
  ranked_disabled: false,
};

export function buildsFromText(text: string): string[] {
  const lines = text.split("\n").map((line) => line.trim());
  return [...new Set(lines.filter((line) => line !== ""))];
}

export function validVersion(version: string): boolean {
  return /^\d+\.\d+\.\d+$/.test(version);
}

export function emptyEnforcedAllowlist(controls: LiveControls): boolean {
  return controls.build_allowlist_enforced && controls.build_allowlist.length === 0;
}

export function sameControls(left: LiveControls, right: LiveControls): boolean {
  return (
    left.minimum_app_version === right.minimum_app_version &&
    left.build_allowlist_enforced === right.build_allowlist_enforced &&
    left.purchases_disabled === right.purchases_disabled &&
    left.ranked_disabled === right.ranked_disabled &&
    left.build_allowlist.join("\n") === right.build_allowlist.join("\n")
  );
}
