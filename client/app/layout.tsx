import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "OAuth2 デモクライアント",
  description: "認可コードフロー（PKCE）の Next.js デモ",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>
        {process.env.NODE_ENV === "development" ? (
          <script
            dangerouslySetInnerHTML={{
              __html: `(function(){if(typeof navigator==="undefined"||!("serviceWorker"in navigator))return;navigator.serviceWorker.getRegistrations().then(function(rs){rs.forEach(function(r){r.unregister();});});})();`,
            }}
          />
        ) : null}
        {children}
      </body>
    </html>
  );
}
