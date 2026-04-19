import { NextResponse } from "next/server";
import type { OAuthTokenJSON } from "./oauth-token-types";

const secure = process.env.NODE_ENV === "production";

/** トークン交換・リフレッシュ成功時にデモ用 HttpOnly クッキーを付与する */
export function applyDemoTokenCookies(
  res: NextResponse,
  tokens: OAuthTokenJSON,
): void {
  const accessMax = tokens.expires_in ?? 3600;

  res.cookies.set("demo_access_token", tokens.access_token, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    maxAge: accessMax,
    path: "/",
  });

  if (tokens.refresh_token) {
    res.cookies.set("demo_refresh_token", tokens.refresh_token, {
      httpOnly: true,
      sameSite: "lax",
      secure,
      maxAge: 60 * 60 * 24 * 30,
      path: "/",
    });
  } else {
    res.cookies.delete("demo_refresh_token");
  }

  // id_token は認可コード交換時にだけ返ることが多い。レスポンスに無いときは既存クッキーを触らない
  // （リフレッシュで消すと OIDC デモ表示が毎回消えるため）。
  if (tokens.id_token) {
    res.cookies.set("demo_id_token", tokens.id_token, {
      httpOnly: true,
      sameSite: "lax",
      secure,
      maxAge: accessMax,
      path: "/",
    });
  }
}
