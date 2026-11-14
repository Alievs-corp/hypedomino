export type ConfigKey = {
  key: string;
  type: string;
  unit: string;
  min: number;
  max?: number;
  default: number;
  value: number;
  set: boolean;
};

export type KeyedConfig = { config_version: number; keys: ConfigKey[] };

export type KeyValues = Readonly<Record<string, number>>;

export type KeyedVersion = { config_version: number; published_at: string; values: KeyValues };

function valueOf(key: ConfigKey, values: KeyValues): number {
  return values[key.key] ?? key.value;
}

export function withinBounds(key: ConfigKey, value: number): boolean {
  return Number.isInteger(value) && value >= key.min && (key.max === undefined || value <= key.max);
}

export function invalidKeys(keys: readonly ConfigKey[], values: KeyValues): string[] {
  return keys.filter((key) => !withinBounds(key, valueOf(key, values))).map((key) => key.key);
}

export function changedKeys(keys: readonly ConfigKey[], values: KeyValues): string[] {
  return keys.filter((key) => valueOf(key, values) !== key.value).map((key) => key.key);
}

export function publishedValues(keys: readonly ConfigKey[], values: KeyValues): KeyValues {
  return Object.fromEntries(keys.filter((key) => valueOf(key, values) !== key.default).map((key) => [key.key, valueOf(key, values)]));
}
