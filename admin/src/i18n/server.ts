import { cookies } from "next/headers";
import { defaultLocale, isLocale, localeCookie, type Locale } from "./locales";
import { translator } from "./translate";

export async function currentLocale(): Promise<Locale> {
  const stored = (await cookies()).get(localeCookie)?.value;
  return isLocale(stored) ? stored : defaultLocale;
}

export async function serverTranslator() {
  return translator(await currentLocale());
}
