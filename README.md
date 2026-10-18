# gomokutaisen sumapuro_team3

## コンテナ名

| 対象 | Composeサービス名 | コンテナ名 |
|---|---|---|
| フロントエンド | `frontend` | `sumapuro_team3-frontend-1` |
| バックエンド | `backend` | `sumapuro_team3-backend-1` |
| MySQL | `db` | `gomoku-db` |
| Mailpit | `mailpit` | `sumapuro_team3-mailpit-1` |

フロントエンドとバックエンドのコンテナ名は、Composeのプロジェクト名が `sumapuro_team3` の場合です。
プロジェクト名を変更すると、コンテナ名の先頭部分も変わります。

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

バックエンドの起動時にMySQLへ接続し、`users` と `user_scores` テーブルを作成します。起動後、GETリクエストで http://localhost:8080 にアクセスすると `This is go backend!!` を返します。ログは `docker compose logs backend` で確認できます。

Dockerを使わずにバックエンドを実行する場合は、先にMySQLを起動してから次を実行してください。

### データベース
デフォルトでは `localhost:3305` のMySQLへ接続します。

```sh
docker compose up -d db
cd backend
go run ./cmd/server
```

MySQLへの接続とテーブル作成を確認するには、MySQLを起動した状態で `backend` ディレクトリから次を実行してください。

```sh
go run ./test
```

直接操作したい場合は、Dockerでdbを起動してから、ターミナルで次を実行してください。
```docker compose exec db mysql -u gomoku_user -p gomoku_db
```

パスワードを聞かれたら、"gomoku_password"と打ってください。

### Mailpit
メールをローカルで確認するには、次のコマンドでMailpitを起動してください。

```sh
docker compose up -d mailpit
```

- Web UI: http://localhost:8025
- SMTP: `localhost:1025`

Composeネットワーク内のコンテナからSMTP接続する場合は、ホストに `mailpit`、ポートに `1025` を指定してください。

## Docker Composeで起動

Docker Composeを使って、フロントエンドとバックエンドをビルドし、MySQLとMailpitと一緒に起動できます。

```sh
docker compose up --build
```

- フロントエンドのログイン画面: http://localhost:5173/login
- アカウント作成: http://localhost:5173/account/create
- アカウント復旧: http://localhost:5173/account/recovery
- パスワード再設定: http://localhost:5173/account/reset
- バックエンド: http://localhost:8080（GETで `This is go backend!!` を返します）
- MySQL: `localhost:3305`（DB: `gomoku_db`、ユーザー: `gomoku_user`）
- Mailpit: http://localhost:8025（SMTP: `localhost:1025`）

フロントエンドはViteでビルドした静的ファイルをNginxで配信します。バックエンドはGoバイナリをビルドし、MySQLの準備完了後にHTTPサーバーを実行します。ログは `docker compose logs backend` で確認できます。

停止するには次を実行してください。

```sh
docker compose down
```
