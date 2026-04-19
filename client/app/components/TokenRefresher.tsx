"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

type Props = {
  /** true のときマウント後に POST /api/oauth/refresh して RSC を再取得 */
  active: boolean;
};

/**
 * アクセストークンの期限が近い／アクセス欠落でリフレッシュのみ残る場合に、
 * バックチャネルで grant_type=refresh_token を走らせる。
 *
 * DevTools に「refresh」と「localhost」が並ぶのは、POST が 303 で / へ飛び
 * fetch が redirect follow で続けて GET しているだけ（リフレッシュは 1 回）。
 */
export function TokenRefresher({ active }: Props) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!active) return;

    (async () => {
      try {
        const res = await fetch("/api/oauth/refresh", {
          method: "POST",
          credentials: "same-origin",
          redirect: "follow",
        });

        if (!res.ok) {
          setError(`リフレッシュに失敗しました (${res.status})`);
        } else if (res.url.includes("oauth_error")) {
          setError(
            "リフレッシュに失敗しました。上のメッセージを確認してください。",
          );
        }

        router.refresh();
      } catch {
        setError("リフレッシュ要求に失敗しました");
      }
    })();
  }, [active, router]);

  if (!error) return null;
  return (
    <div className="alert" role="alert">
      {error}
    </div>
  );
}
