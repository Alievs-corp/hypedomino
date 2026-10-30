import type { Locale } from "./locales";
import { az } from "./messages/az";
import { en } from "./messages/en";

export type MessageKey = keyof typeof en;

export type Messages = Record<MessageKey, string>;

const messagesByLocale: Record<Locale, Messages> = { az, en };

export type Translate = (key: MessageKey) => string;

export function translator(locale: Locale): Translate {
  const messages = messagesByLocale[locale];
  return (key) => messages[key];
}
