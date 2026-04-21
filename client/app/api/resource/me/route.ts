import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { resourceServerURL } from "@/lib/resource-server";

/**
 * クッキーのアクセストークンを Authorization に載せてリソースサーバーの GET /api/me を呼ぶ。
 * HttpOnly トークンをブラウザ JS に渡さずに検証付き API を試せる。
 */
export async function GET() {
  const jar = await cookies();
  const access = jar.get("demo_access_token")?.value;
  if (!access) {
    return NextResponse.json(
      { error: "アクセストークンがありません。先にログインしてください。" },
      { status: 401 },
    );
  }

  const base = resourceServerURL().replace(/\/$/, "");
  let res: Response;
  try {
    res = await fetch(`${base}/api/me`, {
      method: "GET",
      headers: { Authorization: `Bearer ${access}` },
      cache: "no-store",
    });
  } catch {
    return NextResponse.json(
      {
        error: `リソースサーバーに接続できませんでした（${base}）。backend を起動しているか確認してください。`,
      },
      { status: 502 },
    );
  }

  const text = await res.text();
  let body: unknown;
  try {
    body = JSON.parse(text) as unknown;
  } catch {
    body = { raw: text };
  }
  return NextResponse.json(body, { status: res.status });
}
