export function oauthIssuer(): string {
  return process.env.OAUTH_ISSUER ?? "http://localhost:8080";
}

export function oauthClientId(): string {
  return process.env.OAUTH_CLIENT_ID ?? "oauth2_demo_client";
}

export function oauthClientSecret(): string {
  return process.env.OAUTH_CLIENT_SECRET ?? "demo_client_secret_12345";
}

export function oauthRedirectURI(): string {
  return (
    process.env.OAUTH_REDIRECT_URI ?? "http://localhost:3000/callback"
  );
}

/** UI 用（例: `:8080`）。認可サーバーの URL からポートを推測する */
export function oauthIssuerPortLabel(): string {
  try {
    const u = new URL(oauthIssuer());
    if (u.port) return `:${u.port}`;
    return u.protocol === "https:" ? ":443" : ":80";
  } catch {
    return "";
  }
}
