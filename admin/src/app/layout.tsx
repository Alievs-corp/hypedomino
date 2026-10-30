import type { Metadata } from "next";
import type { ReactNode } from "react";
import { LocaleProvider } from "@/i18n/client";
import { LocaleSwitcher } from "@/i18n/LocaleSwitcher";
import { currentLocale, serverTranslator } from "@/i18n/server";
import "./globals.css";

export async function generateMetadata(): Promise<Metadata> {
  const t = await serverTranslator();
  return {
    title: t("app.title"),
    robots: { index: false, follow: false },
  };
}

export default async function RootLayout({ children }: { children: ReactNode }) {
  const locale = await currentLocale();
  return (
    <html lang={locale}>
      <body className="min-h-screen antialiased">
        <LocaleProvider locale={locale}>
          <header className="flex justify-end px-6 py-4">
            <LocaleSwitcher />
          </header>
          {children}
        </LocaleProvider>
      </body>
    </html>
  );
}
