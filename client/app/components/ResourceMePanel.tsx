"use client";

import { useState } from "react";

type Phase = "idle" | "loading" | "done";

type Props = {
  /** UI 表示用（サーバー側の RESOURCE_SERVER_URL と揃える） */
  resourceOrigin: string;
};

export function ResourceMePanel({ resourceOrigin }: Props) {
  const [phase, setPhase] = useState<Phase>("idle");
  const [status, setStatus] = useState<number | null>(null);
  const [body, setBody] = useState<string>("");

  async function callMe() {
    setPhase("loading");
    setStatus(null);
    setBody("");
    try {
      const r = await fetch("/api/resource/me", { cache: "no-store" });
      const j = (await r.json()) as unknown;
      setStatus(r.status);
      setBody(JSON.stringify(j, null, 2));
      setPhase("done");
    } catch {
      setStatus(0);
      setBody(JSON.stringify({ error: "このアプリへの fetch に失敗しました" }, null, 2));
      setPhase("done");
    }
  }

  const base = resourceOrigin.replace(/\/$/, "");

  return (
    <section className="demo-strip" style={{ marginTop: "1.5rem" }}>
      <h2 style={{ fontSize: "1rem", margin: "0 0 0.5rem" }}>
        リソースサーバー（検証付き API）
        <span className="demo-badge">デモ</span>
      </h2>
      <p>
        別プロセス <code>{base}</code> の <code>GET /api/me</code> を、同じオリジンの{" "}
        <code>/api/resource/me</code> がプロキシします。HttpOnly のアクセストークンはブラウザの JS からは読めません。
      </p>
      <div className="row" style={{ marginTop: 0 }}>
        <button className="btn" type="button" onClick={callMe} disabled={phase === "loading"}>
          {phase === "loading" ? "呼び出し中…" : "リソースサーバーに「自分」を問い合わせる"}
        </button>
        <a
          className="btn btn-ghost"
          href="/api/resource/me"
          target="_blank"
          rel="noopener noreferrer"
        >
          新しいタブで JSON を開く
        </a>
      </div>
      {status !== null ? (
        <p style={{ marginTop: "0.75rem", marginBottom: "0.35rem", fontSize: "0.9rem" }}>
          HTTP <strong style={{ color: "var(--text)" }}>{status}</strong>
        </p>
      ) : null}
      {body ? (
        <pre style={{ marginTop: "0.25rem", maxHeight: "22rem", overflow: "auto" }}>{body}</pre>
      ) : null}
    </section>
  );
}
