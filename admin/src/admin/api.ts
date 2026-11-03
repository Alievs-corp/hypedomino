import { cookies, headers } from "next/headers";
import { adminApiHeaders, parseReply } from "./forwarding";

export const sessionCookie = "hd_admin_session";

export type ApiReply<T> = { status: number; body: T };

export type ApiRefusal = { error?: string; detail?: string };

export async function sessionToken(): Promise<string | undefined> {
  return (await cookies()).get(sessionCookie)?.value || undefined;
}

export async function hasSession(): Promise<boolean> {
  return Boolean(await sessionToken());
}

export async function adminApi<T>(method: string, path: string, body?: unknown): Promise<ApiReply<T & ApiRefusal>> {
  return callAdminApi<T>(method, path, body, await sessionToken());
}

export async function anonymousAdminApi<T>(method: string, path: string, body?: unknown): Promise<ApiReply<T & ApiRefusal>> {
  return callAdminApi<T>(method, path, body, undefined);
}

async function callAdminApi<T>(method: string, path: string, body: unknown, token: string | undefined): Promise<ApiReply<T & ApiRefusal>> {
  const base = process.env.HD_ADMIN_API_URL;
  if (!base) {
    throw new Error("HD_ADMIN_API_URL is not set");
  }
  const response = await fetch(new URL(path, base), {
    method,
    cache: "no-store",
    headers: adminApiHeaders(await headers(), token),
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  return { status: response.status, body: parseReply<T & ApiRefusal>(await response.text()) };
}
