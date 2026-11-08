import Link from "next/link";
import type { ReactNode } from "react";
import { adminApi } from "@/admin/api";
import { googleClientId } from "@/admin/google";
import { can, canView } from "@/admin/navigation";
import { currentAdmin } from "@/admin/session";
import { utcLabel } from "@/audit/query";
import { serverTranslator } from "@/i18n/server";
import { matchHref } from "@/matches/replay";
import type { MessageKey, Translate } from "@/i18n/translate";
import { banActive, playerApiPath, signedAmount, type PlayerRecord } from "@/players/record";
import { Notice } from "../Notice";
import { PlayerActions } from "./PlayerActions";

export const dynamic = "force-dynamic";

type Column<Row> = { label: MessageKey; cell: (row: Row) => ReactNode };

const when = (instant: string | null) => (instant ? utcLabel(instant) : "");

export default async function PlayerPage({ params }: { params: Promise<{ id: string }> }) {
  const t = await serverTranslator();
  const admin = await currentAdmin();
  if (!admin) {
    return <Notice text={t("auth.required")} />;
  }
  if (!canView(admin.permissions, "players")) {
    return <Notice text={t("auth.forbidden")} />;
  }
  const reply = await adminApi<PlayerRecord>("GET", playerApiPath((await params).id));
  if (reply.status === 404) {
    return <Notice text={t("player.not_found")} />;
  }
  if (reply.status !== 200) {
    return <Notice text={t("player.load_failed")} />;
  }
  const record = reply.body;
  const { account, profile } = record;
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10">
      <Link href="/players" className="self-start text-sm text-text-secondary underline">
        {t("player.back")}
      </Link>
      <h1 className="text-2xl font-semibold break-words">{account.nickname || account.id}</h1>
      <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
        <Fact label={t("player.id")} value={<span className="font-mono">{account.id}</span>} />
        <Fact label={t("players.state")} value={account.state} />
        <Fact label={t("players.country")} value={account.country} />
        <Fact label={t("players.created_at")} value={utcLabel(account.created_at)} />
        <Fact label={t("player.emails")} value={account.emails.join(", ")} />
        {account.identities && <Fact label={t("player.identities")} value={account.identities.map((identity) => `${identity.provider}:${identity.subject}`).join(", ")} />}
        <Fact label={t("player.banned_until")} value={banActive(account, new Date()) ? when(account.banned_until) : t("player.not_banned")} />
        <Fact label={t("player.shadowbanned_at")} value={account.shadowbanned_at ? when(account.shadowbanned_at) : t("player.not_banned")} />
        {account.deletion_requested_at && <Fact label={t("player.deletion_requested_at")} value={when(account.deletion_requested_at)} />}
        {account.deleted_at && <Fact label={t("player.deleted_at")} value={when(account.deleted_at)} />}
        <Fact label={t("player.coins")} value={record.coins} />
        <Fact label={t("player.crystals")} value={record.crystals} />
        {profile && <Fact label={t("player.level")} value={`${profile.level} (${profile.xp} XP)`} />}
        {profile?.rename_required && <Fact label={t("player.rename_required")} value={t("player.yes")} />}
      </dl>
      {can(admin.permissions, "players", "act_on_player") && account.state !== "deleted" && (
        <PlayerActions id={account.id} clientId={googleClientId()} banned={banActive(account, new Date())} shadowbanned={account.shadowbanned_at !== null} renameRequired={profile?.rename_required ?? false} />
      )}
      <Rows title="player.nickname_history" t={t} rows={record.nickname_history} columns={[
        { label: "players.nickname", cell: (row) => row.nickname },
        { label: "player.replaced_at", cell: (row) => when(row.replaced_at) },
      ]} />
      <Rows title="player.ledger" t={t} rows={record.ledger} columns={[
        { label: "player.at", cell: (row) => when(row.at) },
        { label: "player.currency", cell: (row) => row.currency },
        { label: "player.amount", cell: (row) => signedAmount(row.amount) },
        { label: "player.balance_after", cell: (row) => row.balance_after },
        { label: "player.reason", cell: (row) => row.reason },
        { label: "player.source", cell: (row) => `${row.source_type} ${row.source_ref}` },
      ]} />
      <Rows title="player.inventory" t={t} rows={record.inventory} columns={[
        { label: "player.item", cell: (row) => row.item_id },
        { label: "player.source", cell: (row) => `${row.source_type} ${row.source_ref}` },
        { label: "player.granted_at", cell: (row) => when(row.granted_at) },
        { label: "player.revoked_at", cell: (row) => when(row.revoked_at) },
      ]} />
      <Rows title="player.grants" t={t} rows={record.grants} columns={[
        { label: "player.at", cell: (row) => when(row.at) },
        { label: "player.bundle", cell: (row) => `${row.bundle_id} r${row.revision}` },
        { label: "player.source", cell: (row) => `${row.source_type} ${row.source_ref}` },
        { label: "player.revoked_at", cell: (row) => when(row.revoked_at) },
      ]} />
      <Rows title="player.purchases" t={t} rows={record.purchases} columns={[
        { label: "player.at", cell: (row) => when(row.purchased_at) },
        { label: "player.product", cell: (row) => `${row.platform} ${row.product_id} ${row.offer_id}` },
        { label: "players.state", cell: (row) => `${row.state} (${row.environment})` },
        { label: "player.refunded_at", cell: (row) => when(row.refunded_at) },
      ]} />
      <Rows title="player.matches" t={t} rows={record.matches} columns={[
        { label: "player.at", cell: (row) => when(row.started_at) },
        { label: "player.mode", cell: (row) => `${row.mode} ${row.kind}` },
        { label: "player.result", cell: (row) => (row.cancelled ? t("player.cancelled") : `${row.result} ${row.score_a}:${row.score_b}`) },
        { label: "match.open_replay", cell: (row) => <Link href={matchHref(row.match_id)} prefetch={false} className="text-primary underline">{t("match.open_replay")}</Link> },
      ]} />
    </main>
  );
}

function Fact({ label, value }: { label: string; value: ReactNode }) {
  return (
    <>
      <dt className="text-text-muted">{label}</dt>
      <dd className="break-words">{value}</dd>
    </>
  );
}

function Rows<Row>({ title, rows, columns, t }: { title: MessageKey; rows: Row[] | null; columns: Column<Row>[]; t: Translate }) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{t(title)}</h2>
      {!rows || rows.length === 0 ? (
        <p className="text-sm text-text-secondary">{t("player.empty")}</p>
      ) : (
        <table className="text-sm">
          <thead className="text-left text-text-muted">
            <tr>
              {columns.map((column) => (
                <th key={column.label} className="py-2 pr-4">
                  {t(column.label)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, index) => (
              <tr key={index} className="border-t border-surface align-top">
                {columns.map((column) => (
                  <td key={column.label} className="py-2 pr-4 break-words">
                    {column.cell(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
