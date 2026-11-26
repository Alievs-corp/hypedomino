"use server";

import { adminApi, adminApiResponse, type ApiRefusal } from "@/admin/api";
import { parseReply } from "@/admin/forwarding";
import { exportPath, reportsPath, type GateReportView } from "@/contentgate/report";

export type RunResult = { ok: true; id: string } | { ok: false; error: string; detail: string };

export type ExportResult = { ok: true; json: string } | { ok: false; error: string; detail: string };

export async function runGate(launchAt: string): Promise<RunResult> {
  const reply = await adminApi<GateReportView>("POST", reportsPath, { launch_at: launchAt });
  if (reply.status !== 201) {
    return { ok: false, error: reply.body.error ?? String(reply.status), detail: reply.body.detail ?? "" };
  }
  return { ok: true, id: reply.body.report.id };
}

export async function exportReport(id: string): Promise<ExportResult> {
  const response = await adminApiResponse("GET", exportPath(id));
  const text = await response.text();
  if (response.status !== 200) {
    const refusal = parseReply<ApiRefusal>(text);
    return { ok: false, error: refusal.error ?? String(response.status), detail: refusal.detail ?? "" };
  }
  return { ok: true, json: text };
}
