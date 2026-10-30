# フロントエンド

React・TypeScript・Viteで画面を構築し、Node.jsのBFFがAPIの中継とビルド済みファイルの配信を行います。

## 開発

このディレクトリで実行します。

```sh
npm ci
npm run dev
```

APIを利用するには、バックエンドとBFFの起動も必要です。Dockerでの起動、JWTの設定、DB・Mailpitの使い方は[プロジェクトのREADME](../README.md)を参照してください。

## 確認

```sh
npm run build
npm run lint
node --test server/api-proxy.test.mjs
```

## 構成

- `src/router`：URLと画面の対応
- `src/controller`：画面の状態管理とAPI呼び出し
- `src/ui`：フォーム・状態表示
- `src/api`：BFFへのリクエスト
- `src/App.css`：画面のスタイル
- `server.mjs`・`server/`：BFFと静的ファイル配信
