"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { launchInstant } from "@/contentgate/report";
import { useTranslator } from "@/i18n/client";
import { runGate, type RunResult } from "./actions";

export function RunGate() {
  const t = useTranslator();
  const router = useRouter();
  const [launchAt, setLaunchAt] = useState("");
  const [result, setResult] = useState<RunResult | null>(null);
  const [pending, startTransition] = useTransition();
  const instant = launchInstant(launchAt);

  function run() {
    if (!instant) {
      return;
    }
    startTransition(async () => {
      const outcome = await runGate(instant);
      setResult(outcome);
      if (outcome.ok) {
        router.push(`/content-gate?report=${encodeURIComponent(outcome.id)}`);
        router.refresh();
      }
    });
  }

  return (
    <section className="flex flex-col gap-2 rounded-md bg-background-raised p-4">
      <h2 className="text-lg font-semibold">{t("gate.run_title")}</h2>
      <p className="text-sm text-text-secondary">{t("gate.run_intro")}</p>
      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm text-text-secondary">
          {t("gate.launch_at")}
          <input type="datetime-local" value={launchAt} onChange={(event) => setLaunchAt(event.target.value)} className="rounded-md bg-surface px-3 py-2 text-text-primary" />
        </label>
        <button type="button" disabled={pending || !instant} onClick={run} className="rounded-md bg-primary px-4 py-2 font-semibold text-on-primary disabled:opacity-40">
          {t("gate.run")}
        </button>
      </div>
      {result && !result.ok && (
        <p role="status" className="text-text-secondary">
          {t("gate.run_refused")}: {result.error} {result.detail}
        </p>
      )}
    </section>
  );
}
