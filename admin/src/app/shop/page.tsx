import { adminApi, hasSession } from "@/admin/api";
import { serverTranslator } from "@/i18n/server";
import type { CosmeticItem, ShopCatalog } from "@/shop/catalog";
import { ShopEditor } from "./ShopEditor";

export const dynamic = "force-dynamic";

export default async function ShopPage() {
  const t = await serverTranslator();
  if (!(await hasSession())) {
    return <Notice text={t("auth.required")} />;
  }
  const [shop, items] = await Promise.all([
    adminApi<{ config_version: number; catalog: ShopCatalog | null }>("GET", "/admin/config/shop"),
    adminApi<{ items: CosmeticItem[] }>("GET", "/admin/cosmetics/items"),
  ]);
  if (shop.status !== 200 || items.status !== 200) {
    return <Notice text={t("shop.load_failed")} />;
  }
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <h1 className="text-2xl font-semibold">{t("shop.title")}</h1>
      {shop.body.catalog === null ? (
        <p className="text-text-secondary">{t("shop.empty")}</p>
      ) : (
        <ShopEditor initialCatalog={shop.body.catalog} version={shop.body.config_version} items={items.body.items} />
      )}
    </main>
  );
}

function Notice({ text }: { text: string }) {
  return (
    <main className="mx-auto max-w-3xl px-6 py-24">
      <p className="text-text-secondary">{text}</p>
    </main>
  );
}
