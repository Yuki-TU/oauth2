/** Base64URL セグメントを UTF-8 文字列へ（Edge / ブラウザでも Buffer 不要） */
function decodeBase64UrlUtf8(segment: string): string {
  const b64 = segment.replace(/-/g, "+").replace(/_/g, "/");
  const pad = b64.length % 4 === 0 ? "" : "=".repeat(4 - (b64.length % 4));
  const binary = atob(b64 + pad);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return new TextDecoder("utf-8").decode(bytes);
}

/** JWT ペイロードを検証せずに表示用にデコードする（デモ用） */
export function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split(".");
    if (parts.length < 2) return null;
    const json = decodeBase64UrlUtf8(parts[1]);
    return JSON.parse(json) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function claimsExpMs(claims: Record<string, unknown> | null): number | null {
  if (!claims || claims.exp == null) return null;
  const n = Number(claims.exp);
  if (!Number.isFinite(n)) return null;
  // JWT の exp は秒。数値以外（文字列の Unix 秒など）も Number で吸収する。
  return n * 1000;
}

/**
 * exp が現在から skewMs 以内なら true（期限切れ間近＝リフレッシュ対象）。
 * exp が無い・壊れている JWTは更新対象とみなす。
 */
export function jwtExpWithin(
  claims: Record<string, unknown> | null,
  skewMs: number,
): boolean {
  const expMs = claimsExpMs(claims);
  if (expMs === null) return true;
  return expMs < Date.now() + skewMs;
}
