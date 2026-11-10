export type Status = "missing" | "draft" | "approved";

export type Placeholder = { name: string; type: string };

export type StringValue = { text: string; status: Status };

export type GameString = {
  key: string;
  note: string;
  max_length: number;
  placeholders: Placeholder[];
  deprecated: boolean;
  values: Partial<Record<string, StringValue>>;
};

export type Coverage = { language: string; keys: number; approved: number; percent: number; missing: string[] };

export type Bundle = { language: string; version: number; hash: string; by: string; at: string };

export type Problem = { key: string; error: string };

export const editableStatuses = ["draft", "approved"] as const;

export const transferFormats = ["csv", "json"] as const;

export function valueOf(entry: GameString, language: string): StringValue {
  return entry.values[language] ?? { text: "", status: "missing" };
}

export function visibleStrings(strings: readonly GameString[], search: string, missing: readonly string[] | null): GameString[] {
  const needle = search.trim().toLowerCase();
  const wanted = missing === null ? null : new Set(missing);
  return strings.filter((entry) => (wanted === null || wanted.has(entry.key)) && (needle === "" || matches(entry, needle)));
}

function matches(entry: GameString, needle: string): boolean {
  return [entry.key, entry.note, ...Object.values(entry.values).map((value) => value?.text ?? "")].some((text) => text.toLowerCase().includes(needle));
}

export function placeholdersFromText(text: string): Placeholder[] {
  return text
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part !== "")
    .map((part) => {
      const [name, type = ""] = part.split(":").map((piece) => piece.trim());
      return { name, type };
    });
}

export function placeholdersText(placeholders: readonly Placeholder[]): string {
  return placeholders.map((placeholder) => `${placeholder.name}:${placeholder.type}`).join(", ");
}

export function bundlesOf(bundles: readonly Bundle[], language: string): Bundle[] {
  return bundles.filter((bundle) => bundle.language === language).sort((a, b) => b.version - a.version);
}

export function problemsText(problems: readonly Problem[]): string {
  return problems.map((problem) => `${problem.key}: ${problem.error}`).join("; ");
}
