/** リソースサーバー（Go backend）のベース URL。ブラウザからは直接叩かず Next がプロキシする想定 */
export function resourceServerURL(): string {
  return process.env.RESOURCE_SERVER_URL ?? "http://localhost:9090";
}
