import { NextResponse } from "next/server";

export async function POST(request: Request) {
  const home = new URL("/", request.url);
  const res = NextResponse.redirect(home, 303);
  res.cookies.delete("demo_access_token");
  res.cookies.delete("demo_refresh_token");
  res.cookies.delete("demo_id_token");
  res.cookies.delete("oauth_pkce_verifier");
  res.cookies.delete("oauth_state");
  res.cookies.delete("oauth_nonce");
  return res;
}
