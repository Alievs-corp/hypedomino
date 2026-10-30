import { serverTranslator } from "@/i18n/server";

export default async function HomePage() {
  const t = await serverTranslator();
  return (
    <main className="mx-auto flex max-w-3xl flex-col justify-center px-6 py-24">
      <h1 className="text-3xl font-semibold break-words">{t("app.title")}</h1>
    </main>
  );
}
