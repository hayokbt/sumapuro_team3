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

バックエンドの起動時にMySQLへ接続し、必要なテーブルを作成します。起動後、GETリクエストで http://localhost:8080 にアクセスすると `This is go backend!!` を返します。ログは `docker compose logs backend` で確認できます。

Dockerを使わずにバックエンドを実行する場合は、先にMySQLを起動してから次を実行してください。

### データベース
デフォルトでは `localhost:3305` のMySQLへ接続します。

```sh
docker compose up -d db mailpit
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

アカウント作成では、ブラウザーからNode.jsのBFFへリクエストを送り、BFFがGoバックエンドの `POST /account/create` に中継します。Mailpitから届く6桁コードを入力すると、BFF経由の `POST /account/verify` で確認し、正しければアカウントを作成します。受信メールはMailpitのWeb UIで確認できます。

フロントエンドのNode.jsサーバーはViteでビルドした静的ファイルを配信し、登録APIだけをバックエンドへ中継します。`frontend/server.mjs` が起動とリクエストの振り分け、`frontend/server/api-proxy.mjs` がAPI中継、`frontend/server/static-files.mjs` が静的ファイル配信を担当します。バックエンドはGoバイナリをビルドし、MySQLの準備完了後にHTTPサーバーを実行します。ログは `docker compose logs backend` で確認できます。

Dockerを使わず開発する場合は、MySQLとMailpitを起動したうえで、バックエンド、BFF、Viteをそれぞれ別のターミナルから起動します。BFF用にはフロントエンドを一度ビルドしてから、次を実行してください。

```sh
cd frontend
npm run build
BACKEND_URL=http://localhost:8080 PORT=3000 npm run bff
```

別のターミナルで `cd frontend && npm run dev` を実行すると、Viteは `/api` リクエストをこのBFFへ中継します。

## バックエンドの構成

バックエンドは `backend/internal` の各パッケージに責務を分けています。

| ディレクトリ | 責務 | 登録機能での例 |
|---|---|---|
| `domain` | 登録ルール、確認コードの有効期限・試行回数、層をまたいで使うエラーとインターフェース | 登録入力の検証、確認コードの正誤判定 |
| `usecase/accountregistration` | domainのルールと各インターフェースを組み合わせて登録処理を進行 | パスワードハッシュ化、確認コード発行、確認メール送信 |
| `interface/httpapi` | `router.go`でルートとハンドラーを対応付け、`accountcreate`・`accountverify`で各API処理、`response`でJSON入出力とエラー応答を処理 | `POST /account/create`、`POST /account/verify` |
| `infrastructure` | インターフェースの具体的な実装と外部サービス接続 | MySQLリポジトリ、SMTP送信、PBKDF2 |

`cmd/server` はこれらの実装を組み立ててサーバーを起動する場所です。ルートごとの処理や登録ルールは置かず、依存関係の配線に限定しています。

停止するには次を実行してください。

```sh
docker compose down
```
