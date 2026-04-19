import { cookies } from "next/headers";
import Link from "next/link";
import { decodeJwtPayload } from "@/lib/jwt-decode";
import { oauthIssuer, oauthIssuerPortLabel } from "@/lib/oauth-config";

type Props = {
  searchParams: Promise<{ oauth_error?: string }>;
};

export default async function Home({ searchParams }: Props) {
  const params = await searchParams;
  const jar = await cookies();
  const access = jar.get("demo_access_token")?.value;
  const idToken = jar.get("demo_id_token")?.value;

  const accessClaims = access ? decodeJwtPayload(access) : null;
  const idClaims = idToken ? decodeJwtPayload(idToken) : null;

  const issuer = oauthIssuer();
  const issuerOrigin = issuer.replace(/\/$/, "");

  return (
    <main>
      <div className="card">
        <h1>
          OAuth2 デモクライアント
          <span className="port-badge port-badge--client">:3000</span>
        </h1>
        <p>
          この画面はクライアント（Next.js）です。
          <strong>ログイン</strong>
          を押すと、認可サーバー
          <span className="port-badge port-badge--auth">
            {oauthIssuerPortLabel()}
          </span>
          側で認可コードフロー（PKCE）が進みます。
        </p>

        {params.oauth_error ? (
          <div className="alert" role="alert">
            {params.oauth_error}
          </div>
        ) : null}

        {!access ? (
          <>
            <div className="flow">
              <strong style={{ color: "var(--text)" }}>想定フロー</strong>
              <ol>
                <li>
                  下の <strong style={{ color: "var(--text)" }}>ログイン</strong>{" "}
                  で、このアプリが{" "}
                  <code>
                    {issuerOrigin}/authorize
                  </code>{" "}
                  へリダイレクトします。
                </li>
                <li>
                  未ログインなら、同じ認可サーバー上の{" "}
                  <code>{issuerOrigin}/login</code> で ID / パスワードを入力します。
                </li>
                <li>
                  認可が終わると、ブラウザはこのアプリの{" "}
                  <code>http://localhost:3000/callback</code> に戻り、トークンを受け取ります。
                </li>
              </ol>
            </div>
            <div className="row">
              <Link className="btn" href="/api/oauth/start">
                ログイン（{issuerOrigin} で認可）
              </Link>
            </div>
          </>
        ) : (
          <>
            <p>
              認可サーバー（<code>{issuerOrigin}</code>
              ）でのフローが完了し、アクセストークンを HttpOnly クッキーに保存しています。
            </p>
            <div className="row">
              <form action="/api/oauth/logout" method="post">
                <button className="btn btn-ghost" type="submit">
                  ログアウト
                </button>
              </form>
            </div>
            {accessClaims ? (
              <>
                <h2 style={{ marginTop: "1.5rem", fontSize: "1rem" }}>
                  アクセストークン JWT（検証なし・表示のみ）
                </h2>
                <pre>{JSON.stringify(accessClaims, null, 2)}</pre>
              </>
            ) : null}
            {idClaims ? (
              <>
                <h2 style={{ marginTop: "1.25rem", fontSize: "1rem" }}>
                  ID トークン JWT（検証なし・表示のみ）
                </h2>
                <pre>{JSON.stringify(idClaims, null, 2)}</pre>
              </>
            ) : null}
          </>
        )}
      </div>
    </main>
  );
}
