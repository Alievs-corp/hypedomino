import { adminApi } from "@/admin/api";
import { googleClientId } from "@/admin/google";
import { can, canRollBack, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { utcLabel } from "@/audit/query";
import type { RulesConfig, RulesVersion } from "@/config/rules";
import { serverTranslator } from "@/i18n/server";
import type { Translate } from "@/i18n/translate";
import { RulesEditor } from "./RulesEditor";
import { Rollback } from "../Rollback";
import { rollbackRules } from "./actions";

export const dynamic = "force-dynamic";

export default async function RulesPage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "game_config")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const [config, history] = await Promise.all([
    adminApi<RulesConfig>("GET", "/admin/config/rules"),
    adminApi<{ versions: RulesVersion[] }>("GET", "/admin/config/rules/history"),
  ]);
  if (config.status !== 200 || history.status !== 200) {
    return <Notice text={t("rules.load_failed")} />;
  }
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold">{t("rules.title")}</h1>
        <p className="text-sm text-text-muted">
          {t("rules.version")}: {config.body.config_version}
        </p>
        <p className="text-sm text-text-secondary">{t("rules.applies_to_new_matches")}</p>
      </div>
      <RulesEditor keys={config.body.keys} canPublish={can(admin.permissions, "game_config", "publish")} />
      <History versions={history.body.versions} rollback={canRollBack(admin.permissions, "game_config") ? googleClientId() : null} t={t} />
    </main>
  );
}

function History({ versions, rollback, t }: { versions: RulesVersion[]; rollback: string | null; t: Translate }) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{t("rules.history")}</h2>
      {versions.length === 0 ? (
        <p className="text-text-secondary">{t("rules.history_empty")}</p>
      ) : (
        <table className="text-sm">
          <thead className="text-left text-text-muted">
            <tr>
              <th className="py-2 pr-4">{t("rules.version")}</th>
              <th className="py-2 pr-4">{t("rules.published_at")}</th>
              <th className="py-2 pr-4">{t("rules.values")}</th>
              {rollback && <th className="py-2 pr-4" />}
            </tr>
          </thead>
          <tbody>
            {versions.map((version, index) => (
              <tr key={version.config_version} className="border-t border-surface align-top">
                <td className="py-2 pr-4">{version.config_version}</td>
                <td className="py-2 pr-4 whitespace-nowrap">{utcLabel(version.published_at)}</td>
                <td className="py-2 pr-4 font-mono">
                  {Object.entries(version.values)
                    .map(([key, value]) => `${key}=${value}`)
                    .join(", ") || t("rules.all_defaults")}
                </td>
                {rollback && <td className="py-2 pr-4">{index > 0 && <Rollback version={version.config_version} clientId={rollback} action={rollbackRules} />}</td>}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function Notice({ text }: { text: string }) {
  return (
    <main className="mx-auto max-w-3xl px-6 py-24">
      <p className="text-text-secondary">{text}</p>
    </main>
  );
}
