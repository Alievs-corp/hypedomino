"use client";

import { useState } from "react";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { emptyBundleFilter, filteredBundles, useLabel, usedDomains, type RewardBundle } from "@/rewards/bundles";

const fieldClass = "rounded-md bg-surface px-3 py-1 text-text-primary";
const labelClass = "flex flex-col gap-1 text-sm text-text-secondary";

const domainLabel: Record<string, MessageKey> = {
  shop: "rewards.domain.shop",
  pass: "rewards.domain.pass",
  missions: "rewards.domain.missions",
  rating: "rewards.domain.rating",
  daily_login: "rewards.domain.daily_login",
  promo: "rewards.domain.promo",
  promo_template: "rewards.domain.promo_template",
};

export function BundleList({ bundles }: { bundles: RewardBundle[] }) {
  const t = useTranslator();
  const [filter, setFilter] = useState(emptyBundleFilter);
  const label = (domain: string) => (domainLabel[domain] ? t(domainLabel[domain]) : domain);
  const visible = filteredBundles(bundles, filter);
  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-3">
        <label className={labelClass}>
          {t("rewards.search")}
          <input value={filter.query} onChange={(event) => setFilter({ ...filter, query: event.target.value })} className={fieldClass} />
        </label>
        <label className={labelClass}>
          <span>{t("rewards.used_in")}</span>
          <select aria-label={t("rewards.used_in")} value={filter.domain} onChange={(event) => setFilter({ ...filter, domain: event.target.value })} className={fieldClass}>
            <option value="">{t("rewards.all")}</option>
            {usedDomains(bundles).map((domain) => (
              <option key={domain} value={domain}>
                {label(domain)}
              </option>
            ))}
          </select>
        </label>
        <label className="flex items-center gap-2 text-sm text-text-secondary">
          <input type="checkbox" checked={filter.onlyDiffering} onChange={(event) => setFilter({ ...filter, onlyDiffering: event.target.checked })} />
          {t("rewards.only_differing")}
        </label>
      </div>
      <p className="text-sm text-text-muted">
        {t("rewards.count")}: {visible.length}
      </p>
      <table className="w-full text-sm">
        <thead className="text-left text-text-muted">
          <tr>
            <th className="py-1 pr-3">{t("rewards.id")}</th>
            <th className="py-1 pr-3">{t("rewards.revision")}</th>
            <th className="py-1 pr-3">{t("rewards.contents")}</th>
            <th className="py-1 pr-3">{t("rewards.guide_value")}</th>
            <th className="py-1 pr-3">{t("rewards.used_in")}</th>
          </tr>
        </thead>
        <tbody>
          {visible.map((bundle) => (
            <tr key={`${bundle.id}:${useLabel(bundle.usedBy[0])}`} className="border-t border-surface align-top">
              <td className="py-1 pr-3 font-mono break-all">
                {bundle.id}
                {bundle.differs && <p className="font-sans text-danger-text">{t("rewards.differs")}</p>}
              </td>
              <td className="py-1 pr-3">{bundle.revision}</td>
              <td className="py-1 pr-3">
                <Contents bundle={bundle} />
              </td>
              <td className="py-1 pr-3">{bundle.guideValue}</td>
              <td className="py-1 pr-3">
                <ul className="flex flex-col gap-1">
                  {bundle.usedBy.map((use) => (
                    <li key={`${use.domain}:${useLabel(use)}`}>
                      {label(use.domain)}: <span className="font-mono break-all">{useLabel(use)}</span>
                    </li>
                  ))}
                </ul>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {visible.length === 0 && <p className="text-sm text-text-secondary">{t("rewards.no_match")}</p>}
    </section>
  );
}

function Contents({ bundle }: { bundle: RewardBundle }) {
  const t = useTranslator();
  return (
    <ul className="flex flex-col gap-1">
      {bundle.coins > 0 && (
        <li>
          {t("rewards.coins")}: {bundle.coins}
        </li>
      )}
      {(bundle.crystals ?? 0) > 0 && (
        <li>
          {t("rewards.crystals")}: {bundle.crystals}
        </li>
      )}
      {bundle.passXp > 0 && (
        <li>
          {t("rewards.pass_xp")}: {bundle.passXp}
        </li>
      )}
      {(bundle.cosmetics ?? []).length > 0 && (
        <li>
          {t("rewards.cosmetics")}: <span className="font-mono break-all">{(bundle.cosmetics ?? []).join(", ")}</span>
        </li>
      )}
      {(bundle.medals ?? []).length > 0 && (
        <li>
          {t("rewards.medals")}: <span className="font-mono break-all">{(bundle.medals ?? []).join(", ")}</span>
        </li>
      )}
    </ul>
  );
}
