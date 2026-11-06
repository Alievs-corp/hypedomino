"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import type { ConfigResult } from "@/config/result";
import { useTranslator } from "@/i18n/client";
import { bundleEditChanged, bundleEditIssues, editOf, parsedIds, useLabel, type RewardBundle } from "@/rewards/bundles";
import { editRewardBundle } from "./actions";

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const labelClass = "flex flex-col gap-1 text-sm text-text-secondary";

type Props = { bundle: RewardBundle; domainLabel: (domain: string) => string };

export function BundleEditor({ bundle, domainLabel }: Props) {
  const t = useTranslator();
  const router = useRouter();
  const [edit, setEdit] = useState(editOf(bundle));
  const [cosmetics, setCosmetics] = useState((bundle.cosmetics ?? []).join(", "));
  const [medals, setMedals] = useState((bundle.medals ?? []).join(", "));
  const [reason, setReason] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [result, setResult] = useState<ConfigResult | null>(null);
  const [pending, startTransition] = useTransition();
  const next = { ...edit, cosmetics: parsedIds(cosmetics), medals: parsedIds(medals) };
  const issues = bundleEditIssues(next);
  const number = (value: string) => (value === "" ? Number.NaN : Number(value));
  const ready = bundleEditChanged(bundle, next) && issues.length === 0 && reason.trim() !== "" && confirmed && !pending;
  return (
    <section className="flex flex-col gap-3 rounded-md border border-surface p-3">
      <h2 className="font-mono text-lg font-semibold break-all">{bundle.id}</h2>
      <div className="flex flex-wrap gap-3">
        <label className={labelClass}>
          {t("rewards.coins")}
          <input type="number" min={0} value={Number.isNaN(edit.coins) ? "" : edit.coins} onChange={(event) => setEdit({ ...edit, coins: number(event.target.value) })} className={fieldClass} />
        </label>
        <label className={labelClass}>
          {t("rewards.crystals")}
          <input type="number" min={0} value={Number.isNaN(edit.crystals) ? "" : edit.crystals} onChange={(event) => setEdit({ ...edit, crystals: number(event.target.value) })} className={fieldClass} />
        </label>
        <label className={labelClass}>
          {t("rewards.pass_xp")}
          <input type="number" min={0} value={Number.isNaN(edit.passXp) ? "" : edit.passXp} onChange={(event) => setEdit({ ...edit, passXp: number(event.target.value) })} className={fieldClass} />
        </label>
      </div>
      <label className={labelClass}>
        {t("rewards.cosmetics")}
        <input value={cosmetics} onChange={(event) => setCosmetics(event.target.value)} className={`${fieldClass} font-mono`} />
      </label>
      <label className={labelClass}>
        {t("rewards.medals")}
        <input value={medals} onChange={(event) => setMedals(event.target.value)} className={`${fieldClass} font-mono`} />
      </label>
      {issues.length > 0 && (
        <ul role="alert" className="flex flex-col gap-1 text-sm text-danger-text">
          {issues.map((issue) => (
            <li key={issue}>{t(issue)}</li>
          ))}
        </ul>
      )}
      <div role="alert" className="flex flex-col gap-1 text-sm text-danger-text">
        <p>{t("rewards.edit_warning")}</p>
        <ul className="flex flex-col gap-1">
          {bundle.usedBy.map((use) => (
            <li key={`${use.domain}:${useLabel(use)}`}>
              {domainLabel(use.domain)}: <span className="font-mono break-all">{useLabel(use)}</span>
            </li>
          ))}
        </ul>
      </div>
      <label className={labelClass}>
        {t("rewards.reason")}
        <input value={reason} onChange={(event) => setReason(event.target.value)} className={fieldClass} />
      </label>
      <label className="flex items-center gap-2 text-sm text-text-secondary">
        <input type="checkbox" checked={confirmed} onChange={(event) => setConfirmed(event.target.checked)} />
        {t("rewards.edit_confirm")}
      </label>
      <button
        type="button"
        disabled={!ready}
        onClick={() =>
          startTransition(async () => {
            const outcome = await editRewardBundle(bundle.id, bundle.revision, next, reason);
            setResult(outcome);
            if (outcome.ok) {
              router.refresh();
            }
          })
        }
        className="self-start rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40"
      >
        {t("rewards.publish")}
      </button>
      {result && !result.ok && (
        <p role="status" className="text-sm text-text-secondary">
          {t("rewards.refused")}: {result.error} {result.detail}
        </p>
      )}
    </section>
  );
}
