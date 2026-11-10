"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { needsStepUp } from "@/admin/refusals";
import { StepUpPrompt } from "@/app/auth/StepUpPrompt";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { emptyGrant, grantReady, grantRequest, grantSourceTypes, type GrantDraft } from "@/players/grant";
import { validAdjustment } from "@/players/record";
import { ladders, penaltyRunning, type PenaltyView } from "@/players/penalties";
import { adjustCoins, banTemporarily, forceRename, grantReward, liftPenalty, liftShadowban, liftTemporaryBan, mute, removeOffense, shadowban, unmute, type PlayerActionResult } from "./actions";

type Pending = { action: () => Promise<PlayerActionResult>; onApplied?: () => void };

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const buttonClass = "rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40";
const labelClass = "flex flex-col gap-1 text-sm text-text-secondary";

function outcomeKey(result: PlayerActionResult): MessageKey {
  if (!result.ok) {
    return "actions.refused";
  }
  if (result.proposed) {
    return "actions.proposed";
  }
  return result.repeated ? "actions.repeated" : "actions.done";
}

export function PlayerActions({ id, clientId, banned, shadowbanned, muted, renameRequired, penalties }: { id: string; clientId: string; banned: boolean; shadowbanned: boolean; muted: boolean; renameRequired: boolean; penalties: PenaltyView | null }) {
  const t = useTranslator();
  const router = useRouter();
  const [reason, setReason] = useState("");
  const [amount, setAmount] = useState(Number.NaN);
  const [requestId, setRequestId] = useState(() => crypto.randomUUID());
  const [grant, setGrant] = useState<GrantDraft>(emptyGrant);
  const [grantId, setGrantId] = useState(() => crypto.randomUUID());
  const [last, setLast] = useState<Pending | null>(null);
  const [result, setResult] = useState<PlayerActionResult | null>(null);
  const [pending, startTransition] = useTransition();
  const stepUp = result !== null && !result.ok && needsStepUp(result.error);
  const noReason = reason.trim() === "";

  function run(next: Pending) {
    setLast(next);
    const { action, onApplied } = next;
    startTransition(async () => {
      const outcome = await action();
      setResult(outcome);
      if (outcome.ok) {
        onApplied?.();
        router.refresh();
      }
    });
  }

  const resetCoins = () => {
    setAmount(Number.NaN);
    setRequestId(crypto.randomUUID());
  };

  const resetGrant = () => {
    setGrant(emptyGrant);
    setGrantId(crypto.randomUUID());
  };

  const editGrant = <K extends keyof GrantDraft>(key: K, value: GrantDraft[K]) => setGrant((current) => ({ ...current, [key]: value }));

  return (
    <section className="flex flex-col gap-4 rounded-md border border-surface p-4">
      <h2 className="text-lg font-semibold">{t("actions.title")}</h2>
      <label className="flex flex-col gap-1 text-sm text-text-secondary">
        {t("actions.reason")}
        <input value={reason} onChange={(event) => setReason(event.target.value)} className={fieldClass} />
      </label>
      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm text-text-secondary">
          {t("actions.coins_amount")}
          <input
            type="number"
            step={1}
            value={Number.isNaN(amount) ? "" : amount}
            onChange={(event) => setAmount(event.target.valueAsNumber)}
            className={`${fieldClass} w-32`}
          />
        </label>
        <button
          type="button"
          disabled={pending || noReason || !validAdjustment(amount)}
          onClick={() => run({ action: () => adjustCoins(id, amount, requestId, reason), onApplied: resetCoins })}
          className={buttonClass}
        >
          {t("actions.adjust_coins")}
        </button>
      </div>
      <fieldset className="flex flex-col gap-3 rounded-md border border-surface p-3">
        <legend className="px-1 text-sm font-semibold">{t("actions.grant")}</legend>
        <p className="text-sm text-text-muted">{t("actions.grant_hint")}</p>
        <div className="grid gap-3 sm:grid-cols-2">
          <label className={labelClass}>
            {t("actions.grant_source_type")}
            <select value={grant.sourceType} onChange={(event) => editGrant("sourceType", event.target.value as GrantDraft["sourceType"])} className={fieldClass}>
              {grantSourceTypes.map((type) => (
                <option key={type} value={type}>
                  {type}
                </option>
              ))}
            </select>
          </label>
          <label className={labelClass}>
            {t("actions.grant_source_ref")}
            <input value={grant.sourceRef} onChange={(event) => editGrant("sourceRef", event.target.value)} className={fieldClass} />
          </label>
          <label className={labelClass}>
            {t("promo.bundle_id")}
            <input value={grant.bundleId} onChange={(event) => editGrant("bundleId", event.target.value)} className={fieldClass} />
          </label>
          <label className={labelClass}>
            {t("promo.coins")}
            <input type="number" min={0} step={1} value={Number.isNaN(grant.coins) ? "" : grant.coins} onChange={(event) => editGrant("coins", event.target.valueAsNumber)} className={fieldClass} />
          </label>
          <label className={labelClass}>
            {t("promo.pass_xp")}
            <input type="number" min={0} step={1} value={Number.isNaN(grant.passXp) ? "" : grant.passXp} onChange={(event) => editGrant("passXp", event.target.valueAsNumber)} className={fieldClass} />
          </label>
          <label className={labelClass}>
            {t("actions.grant_medals")}
            <input value={grant.medals} onChange={(event) => editGrant("medals", event.target.value)} className={fieldClass} />
          </label>
          <label className={`${labelClass} sm:col-span-2`}>
            {t("actions.grant_cosmetics")}
            <input value={grant.cosmetics} onChange={(event) => editGrant("cosmetics", event.target.value)} className={fieldClass} />
          </label>
        </div>
        <button
          type="button"
          disabled={pending || noReason || !grantReady(grant)}
          onClick={() => run({ action: () => grantReward(id, grantRequest(grant), grantId, reason), onApplied: resetGrant })}
          className={`${buttonClass} self-start`}
        >
          {t("actions.grant_apply")}
        </button>
      </fieldset>
      <div className="flex flex-wrap gap-3">
        {banned ? (
          <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => liftTemporaryBan(id, reason) })} className={buttonClass}>
            {t("actions.lift_ban")}
          </button>
        ) : (
          <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => banTemporarily(id, reason) })} className={buttonClass}>
            {t("actions.ban")}
          </button>
        )}
        <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => (shadowbanned ? liftShadowban(id, reason) : shadowban(id, reason)) })} className={buttonClass}>
          {t(shadowbanned ? "actions.lift_shadowban" : "actions.shadowban")}
        </button>
        <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => (muted ? unmute(id, reason) : mute(id, reason)) })} className={buttonClass}>
          {t(muted ? "actions.unmute" : "actions.mute")}
        </button>
        {!renameRequired && (
          <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => forceRename(id, reason) })} className={buttonClass}>
            {t("actions.force_rename")}
          </button>
        )}
        {penalties && penaltyRunning(penalties) && (
          <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => liftPenalty(id, reason) })} className={buttonClass}>
            {t("actions.lift_penalty")}
          </button>
        )}
        {penalties &&
          ladders.flatMap((queue) =>
            penalties[queue].offenses.map((offense) => (
              <button key={`${queue}:${offense.match_id}`} type="button" disabled={pending || noReason} onClick={() => run({ action: () => removeOffense(id, queue, offense.match_id, reason) })} className={buttonClass}>
                {t("actions.remove_offense")} {queue} {offense.kind}
              </button>
            )),
          )}
      </div>
      {stepUp && last && <StepUpPrompt clientId={clientId} onConfirmed={() => run(last)} />}
      {result && !stepUp && (
        <p role="status" className="text-sm text-text-secondary">
          {t(outcomeKey(result))}
          {!result.ok && `: ${result.error} ${result.detail}`}
        </p>
      )}
    </section>
  );
}
