export const versionFields = ["PrivacyPolicyVersion", "TermsVersion"] as const;
export const retentionField = "ConsentRecordRetentionYears";
export const maxRetentionYears = 10;

type VersionField = (typeof versionFields)[number];

export type PrivacyConfig = Partial<Record<VersionField, string> & Record<typeof retentionField, number>>;
export type PrivacyFields = Record<VersionField | typeof retentionField, string>;

const versionPattern = /^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$/;
const wholeNumber = /^\d+$/;

export function fieldsFromConfig(config: PrivacyConfig | null): PrivacyFields {
  const retention = config?.[retentionField];
  return {
    PrivacyPolicyVersion: config?.PrivacyPolicyVersion ?? "",
    TermsVersion: config?.TermsVersion ?? "",
    [retentionField]: retention === undefined ? "" : String(retention),
  };
}

export function badFields(fields: PrivacyFields): string[] {
  const bad: string[] = versionFields.filter((name) => fields[name] !== "" && !versionPattern.test(fields[name]));
  const retention = fields[retentionField];
  if (retention !== "" && !(wholeNumber.test(retention) && Number(retention) >= 1 && Number(retention) <= maxRetentionYears)) {
    bad.push(retentionField);
  }
  return bad;
}

export function configFromFields(fields: PrivacyFields): PrivacyConfig {
  const config: PrivacyConfig = {};
  for (const name of versionFields) {
    if (fields[name] !== "") {
      config[name] = fields[name];
    }
  }
  if (fields[retentionField] !== "") {
    config[retentionField] = Number(fields[retentionField]);
  }
  return config;
}

export function clearsPolicyVersion(published: PrivacyConfig | null, next: PrivacyConfig): boolean {
  return published?.PrivacyPolicyVersion !== undefined && next.PrivacyPolicyVersion === undefined;
}
