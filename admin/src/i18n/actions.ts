"use server";

import { cookies } from "next/headers";
import { isLocale, localeCookie } from "./locales";

const oneYearInSeconds = 365 * 24 * 60 * 60;

export async function chooseLocale(locale: string) {
  if (!isLocale(locale)) {
    return;
  }
  (await cookies()).set(localeCookie, locale, {
    httpOnly: true,
    sameSite: "strict",
    secure: process.env.NODE_ENV === "production",
    maxAge: oneYearInSeconds,
  });
}
