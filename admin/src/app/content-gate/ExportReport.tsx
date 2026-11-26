"use client";

import { useState, useTransition } from "react";
import { exportFileName } from "@/contentgate/report";
import { useTranslator } from "@/i18n/client";
import { exportReport, type ExportResult } from "./actions";

function download(id: string, json: string) {
  const link = document.createElement("a");
  link.href = URL.createObjectURL(new Blob([json], { type: "application/json" }));
  link.download = exportFileName(id);
  link.click();
  URL.revokeObjectURL(link.href);
}

export function ExportReport({ id }: { id: string }) {
  const t = useTranslator();
  const [result, setResult] = useState<ExportResult | null>(null);
  const [pending, startTransition] = useTransition();

  function send() {
    startTransition(async () => {
      const outcome = await exportReport(id);
      setResult(outcome);
      if (outcome.ok) {
        download(id, outcome.json);
      }
    });
  }

  return (
    <div className="flex flex-col gap-2 self-start">
      <button type="button" disabled={pending} onClick={send} className="rounded-md bg-surface px-4 py-2 whitespace-nowrap text-text-secondary disabled:opacity-40">
        {t("gate.export")}
      </button>
      {result && !result.ok && (
        <p role="status" className="text-text-secondary">
          {t("gate.export_refused")}: {result.error} {result.detail}
        </p>
      )}
    </div>
  );
}
