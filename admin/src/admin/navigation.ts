import type { MessageKey } from "@/i18n/translate";

export type Permissions = Readonly<Record<string, readonly string[] | undefined>>;

export type Screen = { href: string; labelKey: MessageKey; object: string };

export const screens: readonly Screen[] = [
  { href: "/shop", labelKey: "nav.shop", object: "shop_offers" },
  { href: "/players", labelKey: "nav.players", object: "players" },
  { href: "/config/rules", labelKey: "nav.rules", object: "game_config" },
  { href: "/audit", labelKey: "nav.audit", object: "audit_log" },
];

export function can(permissions: Permissions, object: string, action: string): boolean {
  return permissions[object]?.includes(action) ?? false;
}

export function canView(permissions: Permissions, object: string): boolean {
  return can(permissions, object, "view");
}

export function canRollBack(permissions: Permissions, object: string): boolean {
  return can(permissions, object, "publish") && (can(permissions, "snapshots", "rollback") || can(permissions, "snapshots", "rollback_own_domains"));
}

export function visibleScreens(permissions: Permissions, all: readonly Screen[] = screens): Screen[] {
  return all.filter((screen) => canView(permissions, screen.object));
}
