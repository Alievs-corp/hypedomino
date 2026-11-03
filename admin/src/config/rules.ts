export type RuleKey = {
  key: string;
  type: string;
  unit: string;
  min: number;
  max: number;
  default: number;
  value: number;
  set: boolean;
};

export type RulesConfig = { config_version: number; keys: RuleKey[] };

export type RulesVersion = { config_version: number; published_at: string; values: Readonly<Record<string, number>> };

export type RuleValues = Readonly<Record<string, number>>;

function valueOf(key: RuleKey, values: RuleValues): number {
  return values[key.key] ?? key.value;
}

export function withinBounds(key: RuleKey, value: number): boolean {
  return Number.isInteger(value) && value >= key.min && value <= key.max;
}

export function invalidKeys(keys: readonly RuleKey[], values: RuleValues): string[] {
  return keys.filter((key) => !withinBounds(key, valueOf(key, values))).map((key) => key.key);
}

export function changedKeys(keys: readonly RuleKey[], values: RuleValues): string[] {
  return keys.filter((key) => valueOf(key, values) !== key.value).map((key) => key.key);
}

export function publishedValues(keys: readonly RuleKey[], values: RuleValues): RuleValues {
  return Object.fromEntries(keys.filter((key) => valueOf(key, values) !== key.default).map((key) => [key.key, valueOf(key, values)]));
}
