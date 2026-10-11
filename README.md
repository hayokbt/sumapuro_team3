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
docker compose run --build --rm backend
```

現在のバックエンドは `Hello World` を出力して終了するサンプルです。実行後、コンテナは自動的に削除されます。

## Docker Composeで起動

Docker Composeを使って、フロントエンドとバックエンドをビルドし、MySQLと一緒に起動できます。

```sh
docker compose up --build
```

- フロントエンド: http://localhost:5173
- MySQL: `localhost:3305`（DB: `gomoku_db`、ユーザー: `gomoku_user`）

フロントエンドはViteでビルドした静的ファイルをNginxで配信します。バックエンドはGoバイナリをビルドして実行します。現在のバックエンドは `Hello World` を出力して終了するサンプルのため、コンテナも実行後に終了します。ログは `docker compose logs backend` で確認できます。

停止するには次を実行してください。

```sh
docker compose down
```
