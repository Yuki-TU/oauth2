import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import {
  oauthClientId,
  oauthClientSecret,
  oauthIssuer,
  oauthRedirectURI,
} from "@/lib/oauth-config";

function redirectWithError(request: Request, message: string) {
  const base = new URL("/", request.url);
  base.searchParams.set("oauth_error", message);
  return NextResponse.redirect(base);
}

type TokenResponse = {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
};

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

  const tokens = (await tokenRes.json()) as TokenResponse;

  const home = new URL("/", request.url);
  const res = NextResponse.redirect(home);

  res.cookies.delete("oauth_pkce_verifier");
  res.cookies.delete("oauth_state");
  res.cookies.delete("oauth_nonce");

  res.cookies.set("demo_access_token", tokens.access_token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    maxAge: tokens.expires_in ?? 3600,
    path: "/",
  });

  if (tokens.refresh_token) {
    res.cookies.set("demo_refresh_token", tokens.refresh_token, {
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      maxAge: 60 * 60 * 24 * 30,
      path: "/",
    });
  }

  if (tokens.id_token) {
    res.cookies.set("demo_id_token", tokens.id_token, {
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      maxAge: tokens.expires_in ?? 3600,
      path: "/",
    });
  }

  return res;
}
