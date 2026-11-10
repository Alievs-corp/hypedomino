import Link from "next/link";
import { adminApi } from "@/admin/api";
import { googleClientId } from "@/admin/google";
import { can, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { serverTranslator } from "@/i18n/server";
import type { Translate } from "@/i18n/translate";
import { evidenceSummary, signalWeight, type Review } from "@/reviews/reviews";
import { Notice } from "../players/Notice";
import { ReviewDecision } from "./ReviewDecision";

export const dynamic = "force-dynamic";

export default async function ReviewsPage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "queues")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const reply = await adminApi<{ reviews: Review[] }>("GET", "/admin/reviews");
  if (reply.status !== 200) {
    return <Notice text={t("reviews.load_failed")} />;
  }
  const clientId = can(admin.permissions, "queues", "act_on_queue") ? googleClientId() : null;
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold">{t("reviews.title")}</h1>
        <p className="text-sm text-text-secondary">{t("reviews.explain")}</p>
      </div>
      {reply.body.reviews.length === 0 ? <p className="text-text-secondary">{t("reviews.none")}</p> : reply.body.reviews.map((review) => <ReviewEntry key={review.id} review={review} clientId={clientId} t={t} />)}
    </main>
  );
}

function ReviewEntry({ review, clientId, t }: { review: Review; clientId: string | null; t: Translate }) {
  return (
    <section className="flex flex-col gap-3 border-t border-surface pt-4 text-sm">
      <div className="flex flex-wrap gap-x-6 gap-y-1">
        <Link href={`/players/${encodeURIComponent(review.account_id)}`} prefetch={false} className="text-primary underline">
          {review.nickname || review.account_id}
        </Link>
        <span>
          {t("reviews.cause")}: <span className="font-mono">{review.cause}</span>
        </span>
        <span>
          {t("reviews.opened_at")}: {t.when(review.opened_at)}
        </span>
        <span>
          {t("reviews.total_weight")}: {signalWeight(review.signals)}
        </span>
        <span>
          {t("reviews.shadowban")}: {review.shadowbanned_at ? t.when(review.shadowbanned_at) : t("reviews.no_shadowban")}
        </span>
      </div>
      <table>
        <thead className="text-left text-text-muted">
          <tr>
            <th className="py-1 pr-4">{t("reviews.signal")}</th>
            <th className="py-1 pr-4">{t("reviews.weight")}</th>
            <th className="py-1 pr-4">{t("reviews.at")}</th>
            <th className="py-1 pr-4">{t("reviews.evidence")}</th>
          </tr>
        </thead>
        <tbody>
          {review.signals.map((signal) => (
            <tr key={`${signal.kind}-${signal.at}`} className="border-t border-surface align-top">
              <td className="py-1 pr-4 font-mono">{signal.kind}</td>
              <td className="py-1 pr-4">{signal.weight}</td>
              <td className="py-1 pr-4 whitespace-nowrap">{t.when(signal.at)}</td>
              <td className="py-1 pr-4">{evidenceSummary(signal.evidence)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {clientId && <ReviewDecision id={review.id} clientId={clientId} />}
    </section>
  );
}
