import { adminApi } from "@/admin/api";
import { canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { serverTranslator } from "@/i18n/server";
import type { RewardBundle } from "@/rewards/bundles";
import { BundleList } from "./BundleList";

export const dynamic = "force-dynamic";

export default async function RewardsPage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "reward_bundles")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const reply = await adminApi<{ bundles: RewardBundle[]; promo_truncated: boolean }>("GET", "/admin/reward-bundles");
  if (reply.status !== 200) {
    return <Notice text={t("rewards.load_failed")} />;
  }
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-10">
      <h1 className="text-2xl font-semibold">{t("rewards.title")}</h1>
      <p className="text-text-secondary">{t("rewards.intro")}</p>
      {reply.body.promo_truncated && <p className="text-sm text-text-muted">{t("rewards.promo_truncated")}</p>}
      <BundleList bundles={reply.body.bundles} />
    </main>
  );
}

function Notice({ text }: { text: string }) {
  return (
    <main className="mx-auto max-w-3xl px-6 py-24">
      <p className="text-text-secondary">{text}</p>
    </main>
  );
}
