import { changeText, type CatalogChanges } from "@/config/changes";
import type { Translate } from "@/i18n/translate";

export function CatalogChangeList({ changes, t }: { changes: CatalogChanges; t: Translate }) {
  if (changes.settings.length === 0 && changes.entries.length === 0) {
    return <>{t("config.no_changes")}</>;
  }
  return (
    <ul className="flex flex-col gap-1">
      {changes.settings.map((change) => (
        <li key={change.field}>{changeText(change)}</li>
      ))}
      {changes.entries.map((entry) => (
        <li key={entry.id}>
          {entry.id}: {entry.added ? t("config.added") : entry.fields.map(changeText).join(", ")}
        </li>
      ))}
    </ul>
  );
}
