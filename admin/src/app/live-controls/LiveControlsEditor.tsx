"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { needsStepUp } from "@/admin/refusals";
import { StepUpPrompt } from "@/app/auth/StepUpPrompt";
import { buildsFromText, emptyEnforcedAllowlist, sameControls, validVersion, type LiveControls } from "@/config/liveControls";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { switchLiveControls, type SwitchResult } from "./actions";

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";

type Toggle = "purchases_disabled" | "ranked_disabled" | "build_allowlist_enforced";

const toggles: readonly { key: Toggle; label: MessageKey }[] = [
  { key: "purchases_disabled", label: "live.purchases_disabled" },
  { key: "ranked_disabled", label: "live.ranked_disabled" },
  { key: "build_allowlist_enforced", label: "live.build_allowlist_enforced" },
];

export function LiveControlsEditor({ published, clientId }: { published: LiveControls; clientId: string | null }) {
  const t = useTranslator();
  const router = useRouter();
  const [controls, setControls] = useState(published);
  const [builds, setBuilds] = useState(published.build_allowlist.join("\n"));
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<SwitchResult | null>(null);
  const [pending, startTransition] = useTransition();
  const next = { ...controls, build_allowlist: buildsFromText(builds) };
  const versionValid = validVersion(next.minimum_app_version);
  const allowlistEmpty = emptyEnforcedAllowlist(next);
  const stepUp = result !== null && !result.ok && needsStepUp(result.error);
  const canSwitch = clientId !== null;

  function send() {
    startTransition(async () => {
      const outcome = await switchLiveControls(next, reason);
      setResult(outcome);
      if (outcome.ok) {
        setReason("");
        router.refresh();
      }
    });
  }

  return (
    <div className="flex flex-col gap-6">
      {toggles.map(({ key, label }) => (
        <label key={key} className="flex items-center gap-2 text-sm">
          <input type="checkbox" disabled={!canSwitch} checked={controls[key]} onChange={(event) => setControls({ ...controls, [key]: event.target.checked })} />
          {t(label)}
        </label>
      ))}
      <label className="flex flex-col gap-1 text-sm text-text-secondary">
        {t("live.minimum_app_version")}
        <input
          readOnly={!canSwitch}
          aria-invalid={!versionValid}
          value={controls.minimum_app_version}
          onChange={(event) => setControls({ ...controls, minimum_app_version: event.target.value.trim() })}
          className={`w-40 font-mono ${fieldClass}`}
        />
      </label>
      <p className="text-sm text-text-secondary">{t("live.minimum_warning")}</p>
      <label className="flex flex-col gap-1 text-sm text-text-secondary">
        {t("live.build_allowlist")}
        <textarea readOnly={!canSwitch} rows={5} value={builds} onChange={(event) => setBuilds(event.target.value)} className={`font-mono ${fieldClass}`} />
      </label>
      {!versionValid && (
        <p role="alert" className="text-sm text-danger-text">
          {t("live.bad_version")}
        </p>
      )}
      {allowlistEmpty && (
        <p role="alert" className="text-sm text-danger-text">
          {t("live.empty_allowlist")}
        </p>
      )}
      {canSwitch && (
        <section className="flex flex-col gap-3">
          <label className="flex flex-col gap-1 text-sm text-text-secondary">
            {t("live.reason")}
            <input value={reason} onChange={(event) => setReason(event.target.value)} className={fieldClass} />
          </label>
          <button
            type="button"
            disabled={pending || !versionValid || allowlistEmpty || reason.trim() === "" || sameControls(next, published)}
            onClick={send}
            className="self-start rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40"
          >
            {t("live.switch")}
          </button>
          {stepUp && <StepUpPrompt clientId={clientId} onConfirmed={send} />}
          {result && !stepUp && <Outcome result={result} />}
        </section>
      )}
    </div>
  );
}

function Outcome({ result }: { result: SwitchResult }) {
  const t = useTranslator();
  if (!result.ok) {
    return (
      <p role="status" className="text-sm text-text-secondary">
        {t("live.refused")}: {result.error} {result.detail}
      </p>
    );
  }
  return (
    <p role="status" className="text-sm text-text-secondary">
      {result.approvalId === null ? `${t("live.switched")}: ${result.version}` : `${t("live.proposed")}: ${result.approvalId}`}
    </p>
  );
}
