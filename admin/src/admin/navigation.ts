import type { MessageKey } from "@/i18n/translate";

export type Permissions = Readonly<Record<string, readonly string[] | undefined>>;

export type Screen = { href: string; labelKey: MessageKey; object: string };

export const screens: readonly Screen[] = [
  { href: "/shop", labelKey: "nav.shop", object: "shop_offers" },
  { href: "/audit", labelKey: "nav.audit", object: "audit_log" },
];

export function canView(permissions: Permissions, object: string): boolean {
  return permissions[object]?.includes("view") ?? false;
}

export function visibleScreens(permissions: Permissions, all: readonly Screen[] = screens): Screen[] {
  return all.filter((screen) => canView(permissions, screen.object));
}
