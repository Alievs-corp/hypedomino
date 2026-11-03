import type { MessageKey } from "@/i18n/translate";

const signInRefusals: Readonly<Record<string, MessageKey>> = {
  account_refused: "signin.account_refused",
  not_company_account: "signin.not_company_account",
  token_expired: "signin.retry",
  nonce_rejected: "signin.retry",
  provider_unavailable: "signin.unavailable",
};

export function signInRefusalKey(error: string): MessageKey {
  return signInRefusals[error] ?? "signin.failed";
}
