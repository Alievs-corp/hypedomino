"use client";

import Script from "next/script";
import { useRef, useState } from "react";
import { signInRefusalKey } from "@/admin/refusals";
import { useTranslator } from "@/i18n/client";
import type { MessageKey } from "@/i18n/translate";
import { completeSignIn, issueNonce } from "./actions";

type GoogleIdentity = {
  accounts: {
    id: {
      initialize(options: { client_id: string; nonce: string; ux_mode: "popup"; callback: (response: { credential: string }) => void }): void;
      renderButton(parent: HTMLElement, options: Record<string, string>): void;
    };
  };
};

declare global {
  interface Window {
    google?: GoogleIdentity;
  }
}

export function GoogleSignIn({ clientId }: { clientId: string }) {
  const t = useTranslator();
  const button = useRef<HTMLDivElement>(null);
  const [problem, setProblem] = useState<MessageKey | null>(null);

  async function prepare() {
    const nonce = await issueNonce();
    const identity = window.google?.accounts.id;
    if (!nonce || !identity || !button.current) {
      setProblem("signin.unavailable");
      return;
    }
    identity.initialize({
      client_id: clientId,
      nonce,
      ux_mode: "popup",
      callback: async ({ credential }) => {
        const refusal = await completeSignIn(credential, nonce);
        setProblem(signInRefusalKey(refusal.error));
        await prepare();
      },
    });
    identity.renderButton(button.current, { type: "standard", theme: "filled_black", size: "large", text: "signin_with" });
  }

  return (
    <div className="flex flex-col items-start gap-4">
      <Script src="https://accounts.google.com/gsi/client" strategy="afterInteractive" onReady={() => void prepare()} />
      <div ref={button} />
      {problem && (
        <p role="alert" className="text-danger-text">
          {t(problem)}
        </p>
      )}
    </div>
  );
}
