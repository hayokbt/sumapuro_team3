# gomokutaisen sumapuro_team3

## コンテナ名

| 対象 | Composeサービス名 | コンテナ名 |
|---|---|---|
| フロントエンド | `frontend` | `sumapuro_team3-frontend-1` |
| バックエンド | `backend` | `sumapuro_team3-backend-1` |
| MySQL | `db` | `gomoku-db` |

フロントエンドとバックエンドのコンテナ名は、Composeのプロジェクト名が `sumapuro_team3` の場合です。プロジェクト名を変更すると、コンテナ名の先頭部分も変わります。

## 個別に起動

### フロントエンド

```sh
docker compose up --build frontend
```

起動後、 http://localhost:5173 にアクセスしてください。

### バックエンド

```sh
docker compose up --build backend
```

起動後、GETリクエストで http://localhost:8080 にアクセスすると `Hello World!` を返します。ログは `docker compose logs backend` で確認できます。

## Docker Composeで起動

Docker Composeを使って、フロントエンドとバックエンドをビルドし、MySQLと一緒に起動できます。

```sh
docker compose up --build
```

- フロントエンドのログイン画面: http://localhost:5173/login
- アカウント作成: http://localhost:5173/account/create
- アカウント復旧: http://localhost:5173/account/recovery
- パスワード再設定: http://localhost:5173/account/reset
- バックエンド: http://localhost:8080（GETで `Hello World!` を返します）
- MySQL: `localhost:3305`（DB: `gomoku_db`、ユーザー: `gomoku_user`）

フロントエンドはViteでビルドした静的ファイルをNginxで配信します。バックエンドはGoバイナリをビルドしてHTTPサーバーを実行します。ログは `docker compose logs backend` で確認できます。

停止するには次を実行してください。

```sh
docker compose down
```
