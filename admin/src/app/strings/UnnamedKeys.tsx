"use client";

import Link from "next/link";
import { useTranslator } from "@/i18n/client";
import { keyHref, unnamedKeys } from "@/strings/strings";

export function UnnamedKeys({ error, detail }: { error: string; detail: string }) {
  const t = useTranslator();
  const keys = unnamedKeys(error, detail);
  if (keys.length === 0) {
    return null;
  }
  return (
    <div className="flex flex-col gap-1 text-sm text-text-secondary">
      <p>{t("strings.unnamed")}</p>
      <ul className="flex flex-col gap-1">
        {keys.map((key) => (
          <li key={key}>
            <Link href={keyHref(key)} prefetch={false} className="font-mono text-primary underline">
              {key}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
