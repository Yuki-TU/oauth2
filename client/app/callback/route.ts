import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { applyDemoTokenCookies } from "@/lib/apply-token-cookies";
import {
  oauthClientId,
  oauthClientSecret,
  oauthIssuer,
  oauthRedirectURI,
} from "@/lib/oauth-config";
import type { OAuthTokenJSON } from "@/lib/oauth-token-types";

function redirectWithError(request: Request, message: string) {
  const base = new URL("/", request.url);
  base.searchParams.set("oauth_error", message);
  return NextResponse.redirect(base);
}

export async function GET(request: Request) {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state");
  const err = url.searchParams.get("error");
  const errDesc = url.searchParams.get("error_description");

  if (err) {
    const msg = errDesc ? `${err}: ${errDesc}` : err;
    return redirectWithError(request, msg);
  }

  if (!code || !state) {
    return redirectWithError(request, "認可レスポンスに code または state がありません");
  }

  const jar = await cookies();
  const expectedState = jar.get("oauth_state")?.value;
  const codeVerifier = jar.get("oauth_pkce_verifier")?.value;

  if (!codeVerifier || !expectedState || state !== expectedState) {
    return redirectWithError(
      request,
      "state が一致しないか、ログインセッションが切れています。もう一度お試しください。",
    );
  }

  const issuer = oauthIssuer();
  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: oauthRedirectURI(),
    client_id: oauthClientId(),
    client_secret: oauthClientSecret(),
    code_verifier: codeVerifier,
  });

  const tokenRes = await fetch(`${issuer}/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
    cache: "no-store",
  });

  if (!tokenRes.ok) {
    const text = await tokenRes.text();
    return redirectWithError(
      request,
      `トークン交換に失敗しました (${tokenRes.status}): ${text.slice(0, 200)}`,
    );
  }

  const tokens = (await tokenRes.json()) as OAuthTokenJSON;

  const home = new URL("/", request.url);
  const res = NextResponse.redirect(home);

  res.cookies.delete("oauth_pkce_verifier");
  res.cookies.delete("oauth_state");
  res.cookies.delete("oauth_nonce");

  applyDemoTokenCookies(res, tokens);

  return res;
}
