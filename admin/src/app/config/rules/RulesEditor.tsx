"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { changedKeys, invalidKeys, publishedValues, type RuleKey, type RuleValues } from "@/config/rules";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { publishRules, type PublishResult } from "./actions";

const unitLabels: Readonly<Record<string, MessageKey>> = { seconds: "rules.unit.seconds", timeouts: "rules.unit.timeouts" };

const fieldClass = "w-24 rounded-md bg-surface px-3 py-1 text-text-primary";

export function RulesEditor({ keys, canPublish }: { keys: RuleKey[]; canPublish: boolean }) {
  const t = useTranslator();
  const router = useRouter();
  const [values, setValues] = useState<RuleValues>({});
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<PublishResult | null>(null);
  const [pending, startTransition] = useTransition();
  const invalid = invalidKeys(keys, values);
  const changed = changedKeys(keys, values);

  function publish() {
    startTransition(async () => {
      const outcome = await publishRules(publishedValues(keys, values), reason);
      setResult(outcome);
      if (outcome.ok) {
        setValues({});
        router.refresh();
      }
    });
  }

  return (
    <div className="flex flex-col gap-6">
      <table className="text-sm">
        <thead className="text-left text-text-muted">
          <tr>
            <th className="py-2 pr-4">{t("rules.key")}</th>
            <th className="py-2 pr-4">{t("rules.value")}</th>
            <th className="py-2 pr-4">{t("rules.bounds")}</th>
            <th className="py-2 pr-4">{t("rules.default")}</th>
          </tr>
        </thead>
        <tbody>
          {keys.map((key) => (
            <tr key={key.key} className="border-t border-surface">
              <td className="py-2 pr-4 font-mono">{key.key}</td>
              <td className="py-2 pr-4">
                <input
                  type="number"
                  min={key.min}
                  max={key.max}
                  step={1}
                  readOnly={!canPublish}
                  aria-label={key.key}
                  aria-invalid={invalid.includes(key.key)}
                  value={Number.isNaN(values[key.key]) ? "" : (values[key.key] ?? key.value)}
                  onChange={(event) => setValues((current) => ({ ...current, [key.key]: event.target.valueAsNumber }))}
                  className={fieldClass}
                />{" "}
                <span className="text-text-muted">{unitLabels[key.unit] ? t(unitLabels[key.unit]) : key.unit}</span>
              </td>
              <td className="py-2 pr-4 text-text-secondary">
                {key.min} – {key.max}
              </td>
              <td className="py-2 pr-4 text-text-secondary">{key.default}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {invalid.length > 0 && (
        <p role="alert" className="text-sm text-danger-text">
          {t("rules.out_of_bounds")}: {invalid.join(", ")}
        </p>
      )}
      {canPublish && (
        <section className="flex flex-col gap-3">
          <label className="flex flex-col gap-1 text-sm text-text-secondary">
            {t("rules.reason")}
            <input value={reason} onChange={(event) => setReason(event.target.value)} className="rounded-md bg-surface px-3 py-1 text-text-primary" />
          </label>
          <button
            type="button"
            disabled={pending || changed.length === 0 || invalid.length > 0}
            onClick={publish}
            className="self-start rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40"
          >
            {t("rules.publish")}
          </button>
          {result && (
            <p role="status" className="text-sm text-text-secondary">
              {result.ok ? `${t("rules.published")}: ${result.version}` : `${t("rules.refused")}: ${result.error} ${result.detail}`}
            </p>
          )}
        </section>
      )}
    </div>
  );
}
