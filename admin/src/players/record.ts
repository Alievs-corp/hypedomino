export type PlayerSummary = { id: string; nickname: string; state: string; country: string; created_at: string };

export type PlayerAccount = PlayerSummary & {
  emails: string[];
  identities?: { provider: string; subject: string }[];
  banned_until: string | null;
  deletion_requested_at: string | null;
  deleted_at: string | null;
};

export type PlayerProfile = { rename_required: boolean; level: number; xp: number };

export type PlayerRecord = {
  account: PlayerAccount;
  profile: PlayerProfile | null;
  nickname_history: { nickname: string; replaced_at: string }[];
  inventory: { item_id: string; source_type: string; source_ref: string; granted_at: string; revoked_at: string | null }[];
  coins: number;
  crystals: number;
  ledger: { id: number; currency: string; amount: number; reason: string; source_type: string; source_ref: string; config_version: number; balance_after: number; at: string }[];
  grants: { id: number; bundle_id: string; revision: number; source_type: string; source_ref: string; config_version: number; at: string; revoked_at: string | null }[];
  purchases: { platform: string; transaction_id: string; offer_id: string; product_id: string; environment: string; state: string; purchased_at: string; refunded_at: string | null }[];
  matches: { match_id: string; started_at: string; ended_at: string | null; mode: string; kind: string; team: number; result: string; score_a: number; score_b: number; cancelled: boolean }[];
};

export type SearchParams = Readonly<Record<string, string | string[] | undefined>>;

export function searchQuery(params: SearchParams): string {
  const value = params.q;
  return (Array.isArray(value) ? value[0] : value)?.trim() ?? "";
}

export function playersApiPath(query: string): string {
  return `/admin/players?${new URLSearchParams({ q: query })}`;
}

export function playerApiPath(id: string): string {
  return `/admin/players/${encodeURIComponent(id)}`;
}

export function banActive(account: PlayerAccount, now: Date): boolean {
  return account.banned_until !== null && new Date(account.banned_until) > now;
}

export function signedAmount(amount: number): string {
  return amount > 0 ? `+${amount}` : String(amount);
}

export function validAdjustment(amount: number): boolean {
  return Number.isSafeInteger(amount) && amount !== 0;
}
