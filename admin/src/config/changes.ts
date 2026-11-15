export type FieldChange = { field: string; before: string; after: string };

export type EntryChange = { id: string; added: boolean; fields: FieldChange[] };

export type CatalogChanges = { settings: FieldChange[]; entries: EntryChange[] };

type Entry = { id: string };

const absent = "—";

function shown(value: unknown): string {
  return typeof value === "string" ? value : JSON.stringify(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function flatten(value: unknown, path: string, into: Map<string, string>): Map<string, string> {
  if (!isRecord(value)) {
    into.set(path, shown(value));
    return into;
  }
  for (const [key, inner] of Object.entries(value)) {
    flatten(inner, path ? `${path}.${key}` : key, into);
  }
  return into;
}

export function fieldChanges(before: unknown, after: unknown): FieldChange[] {
  const older = flatten(before, "", new Map());
  const newer = flatten(after, "", new Map());
  const fields = [...new Set([...older.keys(), ...newer.keys()])];
  return fields
    .filter((field) => older.get(field) !== newer.get(field))
    .map((field) => ({ field, before: older.get(field) ?? absent, after: newer.get(field) ?? absent }));
}

export function entryChanges(before: readonly Entry[], after: readonly Entry[]): EntryChange[] {
  const previous = new Map(before.map((entry) => [entry.id, entry]));
  return after.flatMap((entry): EntryChange[] => {
    const older = previous.get(entry.id);
    if (!older) {
      return [{ id: entry.id, added: true, fields: [] }];
    }
    const fields = fieldChanges(older, entry);
    return fields.length > 0 ? [{ id: entry.id, added: false, fields }] : [];
  });
}

export function changeText(change: FieldChange): string {
  return `${change.field} ${change.before} → ${change.after}`;
}
