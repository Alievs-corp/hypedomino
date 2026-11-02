import { cookies } from "next/headers";

export const sessionCookie = "hd_admin_session";

export type ApiReply<T> = { status: number; body: T };

export type ApiRefusal = { error?: string; detail?: string };

export async function hasSession(): Promise<boolean> {
  return Boolean((await cookies()).get(sessionCookie)?.value);
}

export async function adminApi<T>(method: string, path: string, body?: unknown): Promise<ApiReply<T & ApiRefusal>> {
  const base = process.env.HD_ADMIN_API_URL;
  if (!base) {
    throw new Error("HD_ADMIN_API_URL is not set");
  }
  const token = (await cookies()).get(sessionCookie)?.value ?? "";
  const response = await fetch(new URL(path, base), {
    method,
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  return { status: response.status, body: (await response.json()) as T & ApiRefusal };
}
