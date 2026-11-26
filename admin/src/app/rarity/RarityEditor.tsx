"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import type { ConfigResult } from "@/config/result";
import { useTranslator } from "@/i18n/client";
import { badFields, fieldsFromTable, shades, tableFromFields, type RarityTable, type TierFields } from "@/rarity/table";
import { publishRarity } from "./actions";

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const cellClass = "w-24 rounded-md bg-surface px-2 py-1 font-mono text-text-primary";
const labelClass = "flex flex-col gap-1 text-sm text-text-secondary";

export function RarityEditor({ published, glowLevels, canPublish }: { published: RarityTable; glowLevels: string[]; canPublish: boolean }) {
  const t = useTranslator();
  const router = useRouter();
  const [fields, setFields] = useState(fieldsFromTable(published));
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<ConfigResult | null>(null);
  const [pending, startTransition] = useTransition();
  const bad = badFields(fields, glowLevels);
  const changed = JSON.stringify(tableFromFields(fields)) !== JSON.stringify(tableFromFields(fieldsFromTable(published)));
  const ready = !pending && bad.length === 0 && changed && reason.trim() !== "";
  const setTier = (at: number, change: Partial<TierFields>) => setFields({ ...fields, tiers: fields.tiers.map((tier, index) => (index === at ? { ...tier, ...change } : tier)) });
  return (
    <section className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2">
        {(["diamondMinSize", "flowPeriodMs"] as const).map((name) => (
          <label key={name} className={labelClass}>
            <span className="font-mono">{name}</span>
            <input inputMode="numeric" readOnly={!canPublish} value={fields[name]} onChange={(event) => setFields({ ...fields, [name]: event.target.value.trim() })} className={fieldClass} />
          </label>
        ))}
      </div>
      <div className="overflow-x-auto">
        <table className="text-sm">
          <thead>
            <tr className="text-left text-text-secondary">
              {["id", "ordinal", ...shades, "pips", "glow", "flow"].map((column) => (
                <th key={column} className="py-2 pr-3 font-mono font-normal">
                  {column}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {fields.tiers.map((tier, at) => (
              <tr key={tier.id} className="border-t border-surface">
                <td className="py-2 pr-3 font-mono">{tier.id}</td>
                <td className="py-2 pr-3">
                  <input inputMode="numeric" readOnly={!canPublish} value={tier.ordinal} onChange={(event) => setTier(at, { ordinal: event.target.value.trim() })} className={cellClass} />
                </td>
                {shades.map((shade) => (
                  <td key={shade} className="py-2 pr-3">
                    <span className="flex items-center gap-2">
                      <span aria-hidden className="inline-block h-4 w-4 rounded-sm border border-surface" style={{ backgroundColor: tier[shade] }} />
                      <input readOnly={!canPublish} value={tier[shade]} onChange={(event) => setTier(at, { [shade]: event.target.value.trim() })} className={cellClass} />
                    </span>
                  </td>
                ))}
                <td className="py-2 pr-3">
                  <input inputMode="numeric" readOnly={!canPublish} value={tier.pips} onChange={(event) => setTier(at, { pips: event.target.value.trim() })} className={cellClass} />
                </td>
                <td className="py-2 pr-3">
                  <select disabled={!canPublish} value={tier.glow} onChange={(event) => setTier(at, { glow: event.target.value })} className={fieldClass}>
                    {glowLevels.map((level) => (
                      <option key={level} value={level}>
                        {level}
                      </option>
                    ))}
                  </select>
                </td>
                <td className="py-2 pr-3">
                  <input type="checkbox" disabled={!canPublish} checked={tier.flow} onChange={(event) => setTier(at, { flow: event.target.checked })} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {bad.length > 0 && (
        <p role="alert" className="text-sm text-danger-text">
          {t("rarity.bad_fields")}: {bad.join(", ")}
        </p>
      )}
      {canPublish && (
        <div className="flex flex-col gap-3">
          <label className={labelClass}>
            {t("rarity.reason")}
            <input value={reason} onChange={(event) => setReason(event.target.value)} className={fieldClass} />
          </label>
          <button
            type="button"
            disabled={!ready}
            onClick={() =>
              startTransition(async () => {
                const outcome = await publishRarity(tableFromFields(fields), reason);
                setResult(outcome);
                if (outcome.ok) {
                  setReason("");
                  router.refresh();
                }
              })
            }
            className="self-start rounded-md bg-primary px-4 py-2 text-on-primary disabled:opacity-40"
          >
            {t("rarity.publish")}
          </button>
          {result && !result.ok && (
            <p role="status" className="text-sm text-text-secondary">
              {t("rarity.refused")}: {result.error} {result.detail}
            </p>
          )}
        </div>
      )}
    </section>
  );
}
