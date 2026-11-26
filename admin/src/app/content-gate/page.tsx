import Link from "next/link";
import { adminApi } from "@/admin/api";
import { can, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { checkTitleKey, evidenceKey, reportPath, reportsPath, resultKey, selectedReportId, tally, type GateCheck, type GateReportView, type GateSummary } from "@/contentgate/report";
import { serverTranslator } from "@/i18n/server";
import type { Translate } from "@/i18n/translate";
import { ExportReport } from "./ExportReport";
import { RunGate } from "./RunGate";

export const dynamic = "force-dynamic";

type SearchParams = Record<string, string | string[] | undefined>;

export default async function ContentGatePage({ searchParams }: { searchParams: Promise<SearchParams> }) {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "content_gate_reports")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const params = await searchParams;
  const history = await adminApi<{ reports: GateSummary[] }>("GET", reportsPath);
  const reports = history.status === 200 ? history.body.reports : [];
  const selected = selectedReportId(params.report, reports);
  const shown = selected ? await adminApi<GateReportView>("GET", reportPath(selected)) : null;
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10">
      <h1 className="text-2xl font-semibold">{t("gate.title")}</h1>
      <p className="text-text-secondary">{t("gate.intro")}</p>
      {can(admin.permissions, "content_gate_reports", "create") && <RunGate />}
      {history.status !== 200 && <p role="alert" className="text-danger-text">{t("gate.load_failed")}</p>}
      {history.status === 200 && reports.length === 0 && <p className="text-text-secondary">{t("gate.empty")}</p>}
      {shown && shown.status !== 200 && <p role="alert" className="text-danger-text">{t("gate.load_failed")}</p>}
      {shown && shown.status === 200 && <Report view={shown.body} t={t} />}
      {reports.length > 0 && <History reports={reports} selected={selected} t={t} />}
    </main>
  );
}

function Report({ view, t }: { view: GateReportView; t: Translate }) {
  const { report } = view;
  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-wrap items-baseline justify-between gap-4">
        <h2 className="text-xl font-semibold">
          {t("gate.verdict")}: {report.verdict === "green" ? t("gate.verdict.green") : t("gate.verdict.red")}
        </h2>
        <ExportReport id={report.id} />
      </div>
      <p role="status" className={report.verdict === "green" && !view.stale ? "text-text-secondary" : "text-danger-text"}>
        {t(evidenceKey({ environment: report.environment, verdict: report.verdict, stale: view.stale }))}
      </p>
      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-text-muted">{t("gate.report_id")}</dt>
        <dd className="break-all font-mono">{report.id}</dd>
        <dt className="text-text-muted">{t("gate.environment")}</dt>
        <dd>{report.environment}</dd>
        <dt className="text-text-muted">{t("gate.launch_at")}</dt>
        <dd>{t.when(report.launch_at)}</dd>
        <dt className="text-text-muted">{t("gate.ran_at")}</dt>
        <dd>{t.when(report.ran_at)}</dd>
        <dt className="text-text-muted">{t("gate.runner")}</dt>
        <dd className="break-all">{report.runner}</dd>
        <dt className="text-text-muted">{t("gate.config_version")}</dt>
        <dd>{report.versions.config_version}</dd>
        <dt className="text-text-muted">{t("gate.manifest_version")}</dt>
        <dd>{report.versions.manifest_version}</dd>
        <dt className="text-text-muted">{t("gate.domains")}</dt>
        <dd className="break-words">
          {Object.entries(report.versions.domains)
            .map(([domain, version]) => `${domain} ${version}`)
            .join(" · ")}
        </dd>
        <dt className="text-text-muted">{t("gate.string_bundles")}</dt>
        <dd>
          {Object.entries(report.versions.string_bundles)
            .map(([language, version]) => `${language} ${version}`)
            .join(" · ") || "-"}
        </dd>
      </dl>
      <ol className="flex flex-col gap-3">
        {report.checks.map((check) => (
          <CheckEntry key={check.id} check={check} t={t} />
        ))}
      </ol>
      <Thresholds view={view} t={t} />
      <Counts view={view} t={t} />
    </section>
  );
}

function CheckEntry({ check, t }: { check: GateCheck; t: Translate }) {
  const title = checkTitleKey(check.id);
  return (
    <li className="rounded-md bg-background-raised p-4">
      <div className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <span className="font-mono text-sm text-text-muted">{check.id}</span>
        <span className="font-semibold">{title ? t(title) : check.id}</span>
        <span className={check.result === "fail" ? "text-danger-text" : "text-text-secondary"}>{t(resultKey(check.result))}</span>
        {check.count !== undefined && (
          <span className="text-sm text-text-secondary">
            {t("gate.ready_count")}: {t.number(check.count)}
            {check.minimum !== undefined && ` / ${t.number(check.minimum)}`}
          </span>
        )}
      </div>
      {check.failures.length > 0 && (
        <details className="mt-2" open={check.failures.length <= 10}>
          <summary className="cursor-pointer text-sm text-text-secondary">
            {t("gate.failures")} ({t.number(check.failures.length)})
          </summary>
          <ul className="mt-2 flex flex-col gap-1 text-sm">
            {check.failures.map((failure, at) => (
              <li key={at} className="break-words">
                <span className="font-mono">{failure.rule}</span> <span className="break-all">{failure.object}</span>
                {failure.detail && <span className="text-text-secondary">: {failure.detail}</span>}
              </li>
            ))}
          </ul>
        </details>
      )}
      {check.unchecked.length > 0 && (
        <details className="mt-2">
          <summary className="cursor-pointer text-sm text-text-secondary">
            {t("gate.unchecked")} ({t.number(check.unchecked.length)})
          </summary>
          <ul className="mt-2 flex flex-col gap-1 text-sm">
            {check.unchecked.map((gap) => (
              <li key={gap.rule} className="break-words">
                <span className="font-mono">{gap.rule}</span> <span className="text-text-secondary">{gap.reason}</span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </li>
  );
}

function Thresholds({ view, t }: { view: GateReportView; t: Translate }) {
  return (
    <section className="flex flex-col gap-2">
      <h3 className="font-semibold">{t("gate.thresholds")}</h3>
      <table className="text-left text-sm">
        <thead className="text-text-muted">
          <tr>
            <th className="pr-4 font-normal">{t("gate.threshold_key")}</th>
            <th className="pr-4 font-normal">{t("gate.threshold_value")}</th>
            <th className="pr-4 font-normal">{t("gate.threshold_default")}</th>
            <th className="font-normal">{t("gate.config_version")}</th>
          </tr>
        </thead>
        <tbody>
          {view.report.thresholds.map((threshold) => (
            <tr key={threshold.key} className={threshold.value < threshold.default ? "text-danger-text" : undefined}>
              <td className="pr-4 font-mono">{threshold.key}</td>
              <td className="pr-4">{t.number(threshold.value)}</td>
              <td className="pr-4">{t.number(threshold.default)}</td>
              <td>{threshold.config_version ?? t("gate.repo_default")}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function Counts({ view, t }: { view: GateReportView; t: Translate }) {
  const { ready, live_not_ready: notReady } = view.report.info;
  const categories = [...new Set([...Object.keys(ready), ...Object.keys(notReady)])].sort();
  if (categories.length === 0) {
    return null;
  }
  return (
    <section className="flex flex-col gap-2">
      <h3 className="font-semibold">{t("gate.info")}</h3>
      <ul className="flex flex-col gap-1 text-sm">
        {categories.map((category) => (
          <li key={category}>
            <span className="font-mono">{category}</span>{" "}
            <span className="text-text-secondary">
              {Object.entries(ready[category] ?? {})
                .map(([rarity, count]) => `${rarity} ${t.number(count)}`)
                .join(" · ") || "-"}
              {" · "}
              {t("gate.live_not_ready")}: {t.number(notReady[category] ?? 0)}
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function History({ reports, selected, t }: { reports: readonly GateSummary[]; selected: string | null; t: Translate }) {
  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">{t("gate.history")}</h2>
      <ol className="flex flex-col gap-2">
        {reports.map((summary) => {
          const counts = tally(summary.results);
          return (
            <li key={summary.id} className={summary.id === selected ? "rounded-md bg-surface p-3" : "rounded-md bg-background-raised p-3"}>
              <Link href={`/content-gate?report=${encodeURIComponent(summary.id)}`} className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
                <time dateTime={summary.ran_at} className="text-sm text-text-secondary">
                  {t.when(summary.ran_at)}
                </time>
                <span className="font-semibold">{summary.verdict === "green" ? t("gate.verdict.green") : t("gate.verdict.red")}</span>
                <span className="text-sm">{summary.environment}</span>
                <span className="text-sm text-text-secondary">
                  {t("gate.launch_at")}: {t.when(summary.launch_at)}
                </span>
                <span className="text-sm text-text-secondary">
                  {t("gate.result.pass")} {t.number(counts.pass)} · {t("gate.result.fail")} {t.number(counts.fail)} · {t("gate.result.unchecked")} {t.number(counts.unchecked)}
                </span>
                {summary.stale && <span className="text-sm text-danger-text">{t("gate.stale")}</span>}
              </Link>
            </li>
          );
        })}
      </ol>
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
