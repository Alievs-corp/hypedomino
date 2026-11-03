"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { needsStepUp } from "@/admin/refusals";
import { StepUpPrompt } from "@/app/auth/StepUpPrompt";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { validAdjustment } from "@/players/record";
import { adjustCoins, banTemporarily, forceRename, liftTemporaryBan, type PlayerActionResult } from "./actions";

type Pending = { action: () => Promise<PlayerActionResult>; onApplied?: () => void };

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const buttonClass = "rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40";

function outcomeKey(result: PlayerActionResult): MessageKey {
  if (!result.ok) {
    return "actions.refused";
  }
  if (result.proposed) {
    return "actions.proposed";
  }
  return result.repeated ? "actions.repeated" : "actions.done";
}

export function PlayerActions({ id, clientId, banned, renameRequired }: { id: string; clientId: string; banned: boolean; renameRequired: boolean }) {
  const t = useTranslator();
  const router = useRouter();
  const [reason, setReason] = useState("");
  const [amount, setAmount] = useState(Number.NaN);
  const [requestId, setRequestId] = useState(() => crypto.randomUUID());
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
        {!renameRequired && (
          <button type="button" disabled={pending || noReason} onClick={() => run({ action: () => forceRename(id, reason) })} className={buttonClass}>
            {t("actions.force_rename")}
          </button>
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
