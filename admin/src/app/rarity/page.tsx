import { adminApi } from "@/admin/api";
import { googleClientId } from "@/admin/google";
import { can, canRollBack, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { serverTranslator } from "@/i18n/server";
import type { RarityTable } from "@/rarity/table";
import { Rollback } from "../config/Rollback";
import { rollbackRarity } from "./actions";
import { RarityEditor } from "./RarityEditor";

export const dynamic = "force-dynamic";

type RarityVersion = { config_version: number; published_at: string; catalog: RarityTable };

export default async function RarityPage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "rarity_table")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const [config, history] = await Promise.all([
    adminApi<{ config_version: number; catalog: RarityTable | null; defaults: RarityTable; glow: string[] }>("GET", "/admin/config/rarity"),
    adminApi<{ versions: RarityVersion[] }>("GET", "/admin/config/rarity/history"),
  ]);
  if (config.status !== 200 || history.status !== 200) {
    return <Notice text={t("rarity.load_failed")} />;
  }
  const rollback = canRollBack(admin.permissions, "rarity_table") ? googleClientId() : null;
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold">{t("rarity.title")}</h1>
        <p className="text-sm text-text-muted">
          {t("rarity.version")}: {config.body.config_version}
        </p>
        <p className="text-sm text-text-secondary">{t("rarity.intro")}</p>
      </div>
      <RarityEditor
        key={config.body.config_version}
        published={config.body.catalog ?? config.body.defaults}
        glowLevels={config.body.glow}
        canPublish={can(admin.permissions, "rarity_table", "publish")}
      />
      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">{t("rarity.history")}</h2>
        <table className="text-sm">
          <tbody>
            {history.body.versions.map((version, index) => (
              <tr key={version.config_version} className="border-t border-surface align-top">
                <td className="py-2 pr-4">{version.config_version}</td>
                <td className="whitespace-nowrap py-2 pr-4">{t.when(version.published_at)}</td>
                <td className="py-2 pr-4 font-mono">{JSON.stringify(version.catalog)}</td>
                {rollback && <td className="py-2 pr-4">{index > 0 && <Rollback version={version.config_version} clientId={rollback} action={rollbackRarity} />}</td>}
              </tr>
            ))}
          </tbody>
        </table>
      </section>
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
