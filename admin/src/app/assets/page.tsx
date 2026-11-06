import { adminApi } from "@/admin/api";
import { can, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { failedCheckLabels, sizeLabel, statusLabel, type Asset } from "@/assets/assets";
import { utcLabel } from "@/audit/query";
import { serverTranslator } from "@/i18n/server";
import type { Translate } from "@/i18n/translate";
import { AssetUpload, RecheckAsset } from "./AssetUpload";

export const dynamic = "force-dynamic";

export default async function AssetsPage() {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "assets")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const list = await adminApi<{ store_configured: boolean; assets: Asset[] }>("GET", "/admin/assets");
  if (list.status !== 200) {
    return <Notice text={t("assets.load_failed")} />;
  }
  const { store_configured: storeConfigured, assets } = list.body;
  const canUpload = can(admin.permissions, "assets", "create") && storeConfigured;
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <h1 className="text-2xl font-semibold">{t("assets.title")}</h1>
      <p className="text-text-secondary">{t("assets.intro")}</p>
      <p role="alert" className="text-danger-text">
        {t("assets.not_delivered")}
      </p>
      {!storeConfigured && (
        <p role="alert" className="text-danger-text">
          {t("assets.store_off")}
        </p>
      )}
      {canUpload && <AssetUpload />}
      <AssetTable assets={assets} canCheck={canUpload} t={t} />
    </main>
  );
}

function AssetTable({ assets, canCheck, t }: { assets: Asset[]; canCheck: boolean; t: Translate }) {
  if (assets.length === 0) {
    return <p className="text-text-secondary">{t("assets.empty")}</p>;
  }
  return (
    <table className="text-sm">
      <thead className="text-left text-text-muted">
        <tr>
          <th className="py-2 pr-4">{t("assets.preview")}</th>
          <th className="py-2 pr-4">{t("assets.address")}</th>
          <th className="py-2 pr-4">{t("assets.status")}</th>
          <th className="py-2 pr-4">{t("assets.measured")}</th>
          <th className="py-2 pr-4">{t("assets.hash")}</th>
          <th className="py-2 pr-4">{t("assets.uploaded_at")}</th>
        </tr>
      </thead>
      <tbody>
        {assets.map((asset) => (
          <AssetRow key={asset.id} asset={asset} canCheck={canCheck} t={t} />
        ))}
      </tbody>
    </table>
  );
}

function AssetRow({ asset, canCheck, t }: { asset: Asset; canCheck: boolean; t: Translate }) {
  const failed = failedCheckLabels(asset, t);
  return (
    <tr className="border-t border-surface align-top">
      <td className="py-2 pr-4">{asset.view_url ? <img src={asset.view_url} alt={asset.address} className="h-16 w-16 object-contain" /> : "—"}</td>
      <td className="py-2 pr-4 font-mono break-all">{asset.address}</td>
      <td className="py-2 pr-4">
        <div className="flex flex-col gap-1">
          <span>{statusLabel(asset.status, t)}</span>
          {failed.map((label) => (
            <span key={label} className="text-danger-text">
              {label}
            </span>
          ))}
          {canCheck && asset.status === "awaiting_file" && <RecheckAsset id={asset.id} />}
        </div>
      </td>
      <td className="py-2 pr-4 whitespace-nowrap">
        {asset.format && (
          <div>
            {asset.format} {asset.width}×{asset.height}
          </div>
        )}
        {asset.bytes !== null && <div>{sizeLabel(asset.bytes)}</div>}
      </td>
      <td className="py-2 pr-4 font-mono" title={asset.content_hash ?? undefined}>
        {asset.content_hash ? asset.content_hash.slice(0, 12) : "—"}
      </td>
      <td className="py-2 pr-4 whitespace-nowrap">{utcLabel(asset.created_at)}</td>
    </tr>
  );
}

function Notice({ text }: { text: string }) {
  return (
    <main className="mx-auto max-w-3xl px-6 py-24">
      <p className="text-text-secondary">{text}</p>
    </main>
  );
}
