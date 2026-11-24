import { adminApi } from "@/admin/api";
import { canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import {
  alertLabels,
  barPercent,
  comparedVersion,
  mixReasons,
  monitorsCsv,
  monitorsCsvFileName,
  type Mix,
  type EconomyComparison,
  type EconomyReport,
  type WindowReport,
} from "@/analytics/economy";
import { serverTranslator } from "@/i18n/server";
import type { Translate } from "@/i18n/translate";

export const dynamic = "force-dynamic";

type Column = WindowReport & { label: string };

export default async function AnalyticsPage({ searchParams }: { searchParams: Promise<{ version?: string }> }) {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "dashboards")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const version = comparedVersion((await searchParams).version);
  const now = await adminApi<EconomyReport>("GET", "/admin/economy/monitors");
  if (now.status !== 200) {
    return <Notice text={t("analytics.load_failed")} />;
  }
  const comparison = version ? await adminApi<EconomyComparison>("GET", `/admin/economy/monitors/compare?version=${version}`) : null;
  const columns: Column[] = [{ window: "now", label: t("analytics.now"), report: now.body }];
  if (comparison?.status === 200) {
    columns.push({ window: "before", label: t("analytics.before"), report: comparison.body.before }, { window: "after", label: t("analytics.after"), report: comparison.body.after });
  }
  const exportedVersion = comparison?.status === 200 ? comparison.body.config_version : null;
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold">{t("analytics.title")}</h1>
        <p className="text-sm text-text-secondary">{t("analytics.explain")}</p>
      </div>
      <Alerts t={t} report={now.body} />
      <form method="get" className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm text-text-secondary">
          {t("analytics.compare_version")}
          <input name="version" inputMode="numeric" defaultValue={version ?? ""} className="w-32 rounded-md bg-surface px-3 py-1 text-text-primary" />
        </label>
        <button type="submit" className="rounded-md bg-primary px-4 py-2 text-on-primary">
          {t("analytics.compare")}
        </button>
      </form>
      {comparison && comparison.status !== 200 && <p className="text-sm text-danger-text">{t("analytics.compare_failed")}</p>}
      {comparison?.status === 200 && (
        <p className="text-sm text-text-secondary">
          {t("analytics.compared")} <span className="font-mono">{comparison.body.domain}</span> v{comparison.body.config_version}, {t.when(comparison.body.published_at)}
        </p>
      )}
      <Monitors t={t} columns={columns} />
      <MixChart t={t} columns={columns} mix="sourceMix" title={t("analytics.source_mix")} empty={t("analytics.no_sources")} barClass="bg-primary" />
      <MixChart t={t} columns={columns} mix="sinkMix" title={t("analytics.sink_mix")} empty={t("analytics.no_sinks")} barClass="bg-text-muted" />
      <a
        href={`data:text/csv;charset=utf-8,${encodeURIComponent(monitorsCsv(columns))}`}
        download={monitorsCsvFileName(now.body, exportedVersion)}
        className="self-start rounded-md bg-surface px-3 py-1 text-text-secondary"
      >
        {t("analytics.export_csv")}
      </a>
    </main>
  );
}

function Alerts({ t, report }: { t: Translate; report: EconomyReport }) {
  if (report.alerts.length === 0) {
    return <p className="text-sm text-text-secondary">{t("analytics.no_alerts")}</p>;
  }
  return (
    <ul className="flex flex-col gap-1 text-sm font-semibold text-danger-text">
      {report.alerts.map((alert) => (
        <li key={alert}>{t(alertLabels[alert])}</li>
      ))}
    </ul>
  );
}

function Monitors({ t, columns }: { t: Translate; columns: Column[] }) {
  const rows: [string, (report: EconomyReport) => string][] = [
    [t("analytics.window"), (report) => `${t.when(report.from)} – ${t.when(report.to)}`],
    [t("analytics.coins_created"), (report) => t.number(report.coinsCreated)],
    [t("analytics.coins_spent"), (report) => t.number(report.coinsSpent)],
    [t("analytics.sink_ratio"), (report) => (report.sinkRatioPercent === null ? "–" : `${t.number(report.sinkRatioPercent)}%`)],
    [t("analytics.median_balance"), (report) => t.number(report.medianBalance)],
    [t("analytics.day"), (report) => t.when(report.day)],
    [t("analytics.accounts_over_daily_earn"), (report) => t.number(report.accountsOverDailyEarn)],
  ];
  return (
    <table className="w-full text-left text-sm">
      <thead className="text-text-muted">
        <tr>
          <th className="py-2">{t("analytics.monitor")}</th>
          {columns.map((column) => (
            <th key={column.label} className="py-2">
              {column.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map(([label, value]) => (
          <tr key={label}>
            <td className="py-1 text-text-secondary">{label}</td>
            {columns.map((column) => (
              <td key={column.label} className="py-1">
                {value(column.report)}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function MixChart({ t, columns, mix, title, empty, barClass }: { t: Translate; columns: Column[]; mix: Mix; title: string; empty: string; barClass: string }) {
  const reasons = mixReasons(
    columns.map((column) => column.report),
    mix,
  );
  const max = Math.max(0, ...columns.flatMap((column) => Object.values(column.report[mix])));
  return (
    <section className="flex flex-col gap-3">
      <h2 className="font-semibold">{title}</h2>
      {reasons.length === 0 && <p className="text-sm text-text-secondary">{empty}</p>}
      {reasons.map((reason) => (
        <div key={reason} className="flex flex-col gap-1 text-sm">
          <span className="font-mono">{reason}</span>
          {columns.map((column) => {
            const coins = column.report[mix][reason] ?? 0;
            return (
              <div key={column.label} className="flex items-center gap-3">
                {columns.length > 1 && <span className="w-28 shrink-0 text-text-muted">{column.label}</span>}
                <div className="h-3 flex-1 rounded-sm bg-surface">
                  <div className={`h-full rounded-sm ${barClass}`} style={{ width: `${barPercent(coins, max)}%` }} />
                </div>
                <span className="w-20 shrink-0 text-right tabular-nums">{t.number(coins)}</span>
              </div>
            );
          })}
        </div>
      ))}
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
