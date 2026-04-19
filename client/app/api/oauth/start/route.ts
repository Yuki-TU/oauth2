import { NextResponse } from "next/server";
import {
  oauthClientId,
  oauthIssuer,
  oauthRedirectURI,
} from "@/lib/oauth-config";
import {
  codeChallengeS256,
  generateCodeVerifier,
  generateState,
} from "@/lib/pkce";

const cookieOpts = {
  httpOnly: true as const,
  sameSite: "lax" as const,
  secure: process.env.NODE_ENV === "production",
  maxAge: 600,
  path: "/",
};

export async function GET() {
  const issuer = oauthIssuer();
  const clientId = oauthClientId();
  const redirectUri = oauthRedirectURI();

  const codeVerifier = generateCodeVerifier();
  const codeChallenge = codeChallengeS256(codeVerifier);
  const state = generateState();
  const nonce = generateState();

  const authorize = new URL("/authorize", issuer);
  authorize.searchParams.set("client_id", clientId);
  authorize.searchParams.set("redirect_uri", redirectUri);
  authorize.searchParams.set("response_type", "code");
  authorize.searchParams.set("scope", "read write openid profile");
  authorize.searchParams.set("state", state);
  authorize.searchParams.set("nonce", nonce);
  authorize.searchParams.set("code_challenge", codeChallenge);
  authorize.searchParams.set("code_challenge_method", "S256");

  const res = NextResponse.redirect(authorize.toString());
  res.cookies.set("oauth_pkce_verifier", codeVerifier, cookieOpts);
  res.cookies.set("oauth_state", state, cookieOpts);
  res.cookies.set("oauth_nonce", nonce, cookieOpts);
  return res;
}
