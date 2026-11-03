import { googleClientId } from "@/admin/google";
import { currentAdmin } from "@/admin/session";
import { serverTranslator } from "@/i18n/server";
import { GoogleSignIn } from "./auth/GoogleSignIn";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  return (
    <main className="mx-auto flex max-w-3xl flex-col justify-center gap-6 px-6 py-24">
      <h1 className="text-3xl font-semibold break-words">{t("app.title")}</h1>
      {admin ? (
        <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
          <dt className="text-text-muted">{t("home.signed_in_as")}</dt>
          <dd className="break-words">{admin.account.display_name}</dd>
          <dt className="text-text-muted">{t("home.roles")}</dt>
          <dd className="break-words">{admin.account.roles.join(", ")}</dd>
        </dl>
      ) : (
        <>
          <p className="text-text-secondary">{t("auth.required")}</p>
          <GoogleSignIn clientId={googleClientId()} />
        </>
      )}
    </main>
  );
}
