import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { applyDemoTokenCookies } from "@/lib/apply-token-cookies";
import {
  oauthClientId,
  oauthClientSecret,
  oauthIssuer,
} from "@/lib/oauth-config";
import type { OAuthTokenJSON } from "@/lib/oauth-token-types";

function redirectHomeWithError(request: Request, message: string) {
  const u = new URL("/", request.url);
  u.searchParams.set("oauth_error", message);
  return NextResponse.redirect(u, 303);
}

/** grant_type=refresh_token でアクセストークンを再取得し、クッキーを更新する */
export async function POST(request: Request) {
  const jar = await cookies();
  const refresh = jar.get("demo_refresh_token")?.value;
  if (!refresh) {
    return redirectHomeWithError(request, "リフレッシュトークンがありません。再度ログインしてください。");
  }

  const issuer = oauthIssuer();
  const body = new URLSearchParams({
    grant_type: "refresh_token",
    refresh_token: refresh,
    client_id: oauthClientId(),
    client_secret: oauthClientSecret(),
  });

  let tokenRes: Response;
  try {
    tokenRes = await fetch(`${issuer}/token`, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body.toString(),
      cache: "no-store",
    });
  } catch {
    return redirectHomeWithError(
      request,
      `認可サーバーに接続できませんでした（${issuer}/token）。起動しているか確認してください。`,
    );
  }

  if (!tokenRes.ok) {
    const text = await tokenRes.text();
    const res = redirectHomeWithError(
      request,
      `リフレッシュに失敗しました (${tokenRes.status}): ${text.slice(0, 180)}`,
    );
    // 無効なリフレッシュならクッキーを掃除して再ログインへ誘導
    res.cookies.delete("demo_access_token");
    res.cookies.delete("demo_refresh_token");
    res.cookies.delete("demo_id_token");
    return res;
  }

  const tokens = (await tokenRes.json()) as OAuthTokenJSON;
  const home = new URL("/", request.url);
  const res = NextResponse.redirect(home, 303);
  applyDemoTokenCookies(res, tokens);
  return res;
}
