/** 認可サーバー /token の JSON レスポンス（利用フィールドのみ） */
export type OAuthTokenJSON = {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
};
