"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import type { ConfigResult } from "@/config/result";
import { useTranslator } from "@/i18n/client";
import { badFields, clearsPolicyVersion, configFromFields, fieldsFromConfig, linkFields, retentionField, versionFields, type PrivacyConfig } from "@/privacy/config";
import { publishPrivacy } from "./actions";

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const labelClass = "flex flex-col gap-1 text-sm text-text-secondary";

export function PrivacyEditor({ published, canPublish }: { published: PrivacyConfig | null; canPublish: boolean }) {
  const t = useTranslator();
  const router = useRouter();
  const [fields, setFields] = useState(fieldsFromConfig(published));
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<ConfigResult | null>(null);
  const [pending, startTransition] = useTransition();
  const bad = badFields(fields);
  const next = configFromFields(fields);
  const changed = JSON.stringify(next) !== JSON.stringify(configFromFields(fieldsFromConfig(published)));
  const ready = !pending && bad.length === 0 && changed && reason.trim() !== "";
  const set = (name: keyof typeof fields, value: string) => setFields({ ...fields, [name]: value.trim() });
  return (
    <section className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2">
        {versionFields.map((name) => (
          <label key={name} className={labelClass}>
            <span className="font-mono">{name}</span>
            <input readOnly={!canPublish} placeholder={t("privacy.not_published")} value={fields[name]} onChange={(event) => set(name, event.target.value)} className={fieldClass} />
          </label>
        ))}
        <label className={labelClass}>
          <span className="font-mono">{retentionField}</span>
          <input inputMode="numeric" readOnly={!canPublish} placeholder={t("privacy.retention_default")} value={fields[retentionField]} onChange={(event) => set(retentionField, event.target.value)} className={fieldClass} />
        </label>
        {linkFields.map((name) => (
          <label key={name} className={`${labelClass} sm:col-span-2`}>
            <span className="font-mono">{name}</span>
            <input type="url" readOnly={!canPublish} placeholder={t("privacy.link_hidden")} value={fields[name]} onChange={(event) => set(name, event.target.value)} className={fieldClass} />
          </label>
        ))}
      </div>
      {bad.length > 0 && (
        <p role="alert" className="text-sm text-danger-text">
          {t("privacy.bad_fields")}: {bad.join(", ")}
        </p>
      )}
      {clearsPolicyVersion(published, next) && (
        <p role="alert" className="text-sm text-danger-text">
          {t("privacy.policy_version_cleared")}
        </p>
      )}
      {canPublish && (
        <div className="flex flex-col gap-3">
          <label className={labelClass}>
            {t("privacy.reason")}
            <input value={reason} onChange={(event) => setReason(event.target.value)} className={fieldClass} />
          </label>
          <button
            type="button"
            disabled={!ready}
            onClick={() =>
              startTransition(async () => {
                const outcome = await publishPrivacy(next, reason);
                setResult(outcome);
                if (outcome.ok) {
                  setReason("");
                  router.refresh();
                }
              })
            }
            className="self-start rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40"
          >
            {t("privacy.publish")}
          </button>
          {result && !result.ok && (
            <p role="status" className="text-sm text-text-secondary">
              {t("privacy.refused")}: {result.error} {result.detail}
            </p>
          )}
        </div>
      )}
    </section>
  );
}
