import type { AuditEntry } from "@/audit/query";
import { breakGlassRevertOf, type BreakGlassRevert } from "@/config/liveControls";
import type { Translate } from "@/i18n/translate";

export function RevertedBreakGlass({ entries, t }: { entries: readonly AuditEntry[]; t: Translate }) {
  const reverts = entries.map(breakGlassRevertOf).filter((revert): revert is BreakGlassRevert => revert !== null);
  if (reverts.length === 0) {
    return <p className="text-sm text-text-secondary">{t("live.break_glass_reverted_none")}</p>;
  }
  return (
    <ul className="flex flex-col gap-2 text-sm">
      {reverts.map((revert) => (
        <li key={revert.id} className="rounded-md bg-surface p-3">
          <span className="font-mono">{revert.incident}</span> · {t("live.break_glass_restored_as")} {revert.version} · {revert.reason} · {revert.actor} · {t.when(revert.at)}
        </li>
      ))}
    </ul>
  );
}
