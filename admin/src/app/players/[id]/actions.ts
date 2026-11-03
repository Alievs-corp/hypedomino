"use server";

import { adminApi } from "@/admin/api";
import { playerApiPath } from "@/players/record";

export type PlayerActionResult = { ok: true; proposed: boolean; repeated: boolean } | { ok: false; error: string; detail: string };

async function act(method: string, path: string, body: Record<string, unknown>): Promise<PlayerActionResult> {
  const reply = await adminApi<{ repeated?: boolean }>(method, path, body);
  if (reply.status !== 200 && reply.status !== 202) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, proposed: reply.status === 202, repeated: reply.body.repeated ?? false };
}

export async function adjustCoins(id: string, amount: number, requestId: string, reason: string): Promise<PlayerActionResult> {
  return act("POST", `${playerApiPath(id)}/coins`, { amount, request_id: requestId, reason });
}

export async function banTemporarily(id: string, reason: string): Promise<PlayerActionResult> {
  return act("POST", `${playerApiPath(id)}/temporary-ban`, { reason });
}

export async function liftTemporaryBan(id: string, reason: string): Promise<PlayerActionResult> {
  return act("DELETE", `${playerApiPath(id)}/temporary-ban`, { reason });
}

export async function forceRename(id: string, reason: string): Promise<PlayerActionResult> {
  return act("POST", `${playerApiPath(id)}/force-rename`, { reason });
}
