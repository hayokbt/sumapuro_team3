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

ログインでは、ブラウザーからNode.jsのBFFへアカウント名とパスワードを送り、BFFがGoバックエンドの `POST /login` に中継します。GoバックエンドはMySQLに保存されたパスワードハッシュと照合し、成功時はフロントエンドが `/home` へ移動します。認証に失敗した場合は、日本語のエラーメッセージを表示します。

アカウント作成では、ブラウザーからNode.jsのBFFへリクエストを送り、BFFがGoバックエンドの `POST /account/create` に中継します。Mailpitから届く6桁コードを入力すると、BFF経由の `POST /account/verify` で確認し、正しければアカウントを作成します。受信メールはMailpitのWeb UIで確認できます。

アカウント復旧では、登録済みメールアドレスに届く6桁コードを入力して確認すると、Goバックエンドが一時的なリセットトークンを発行します。新しいアカウント名とパスワードを送ると、Goバックエンドがトークンを検証し、MySQLのアカウント情報を一度だけ更新します。

復旧APIはBFFの `POST /api/account/recovery/request`、`POST /api/account/recovery/verify`、`POST /api/account/recovery/reset` からGoバックエンドへ中継されます。確認コードとリセットトークンは有効期限付きで、確認コードは入力回数を制限し、リセットトークンはDBにハッシュを保存して一度だけ利用します。

フロントエンドのNode.jsサーバーはViteでビルドした静的ファイルを配信し、登録・ログイン・アカウント復旧APIをバックエンドへ中継します。`frontend/server.mjs` が起動とリクエストの振り分け、`frontend/server/api-proxy.mjs` がAPI中継、`frontend/server/static-files.mjs` が静的ファイル配信を担当します。バックエンドはGoバイナリをビルドし、MySQLの準備完了後にHTTPサーバーを実行します。ログは `docker compose logs backend` で確認できます。

Dockerを使わず開発する場合は、MySQLとMailpitを起動したうえで、バックエンド、BFF、Viteをそれぞれ別のターミナルから起動します。BFF用にはフロントエンドを一度ビルドしてから、次を実行してください。

```sh
cd frontend
npm run build
BACKEND_URL=http://localhost:8080 PORT=3000 npm run bff
```

別のターミナルで `cd frontend && npm run dev` を実行すると、Viteは `/api` リクエストをこのBFFへ中継します。

## バックエンドの構成

バックエンドは `backend/internal` の各パッケージに責務を分けています。

| ディレクトリ | 責務 | 機能での例 |
|---|---|---|
| `domain` | 登録・復旧ルール、確認コードの有効期限・試行回数、層をまたいで使うエラーとインターフェース | 登録入力の検証、確認コードの正誤判定 |
| `usecase/accountregistration`、`usecase/authentication`、`usecase/accountrecovery` | domainのルールと各インターフェースを組み合わせて登録・認証・復旧を進行 | パスワードハッシュ化・照合、確認コード発行、アカウント情報更新 |
| `interface/httpapi` | `router.go`でルートとハンドラーを対応付け、`accountcreate`・`accountverify`・`login`・`accountrecovery`で各API処理、`response`でJSON入出力とエラー応答を処理 | `POST /account/create`、`POST /account/verify`、`POST /login`、`POST /account/recovery/*` |
| `infrastructure` | インターフェースの具体的な実装と外部サービス接続 | MySQLリポジトリ、SMTP送信、PBKDF2 |

`cmd/server` はこれらの実装を組み立ててサーバーを起動する場所です。ルートごとの処理や登録ルールは置かず、依存関係の配線に限定しています。

停止するには次を実行してください。

```sh
docker compose down
```

## JWTによるログイン状態管理

Goが署名付きJWTを発行し、`gomoku_auth` HttpOnly Cookieに保存します。
JWTは30分で期限切れとなり、期限切れ後は再ログインが必要です。
JWTにはユーザーID・セッションID・認証バージョン・発行者・利用先・期限を含め、
APIでは署名検証に加えてMySQLのセッション状態と認証バージョンを確認します。

Docker Composeでは、次のコマンドだけで起動できます。

```sh
docker compose up --build
```

`JWT_SECRET`が未設定の場合、バックエンドはランダムな署名鍵を自動生成し、
専用の`auth_data`ボリューム内に権限0600で保存します。
コンテナを再作成しても同じ鍵を利用します。ボリュームを削除すると鍵が変わり、
既存のJWTは利用できなくなります。

自分で鍵を管理する場合は、32バイト以上の`JWT_SECRET`を指定してください。
環境変数の鍵は保存ファイルより優先します。

```sh
export JWT_SECRET="$(openssl rand -hex 32)"
```

DockerなしでGoを起動する場合は、`JWT_SECRET`を設定するか、
`JWT_SECRET_FILE=/保存先/jwt-secret`を指定して自動生成・保存します。
再起動後もログインを維持するには同じ鍵を使ってください。
短すぎる鍵や読み取れない鍵ファイルでは起動エラーになります。
DB・Mailpitのみの起動には鍵の設定は不要です。

- `COOKIE_SECURE`：ローカルHTTPでは`false`、HTTPS環境では`true`を指定。
- `APP_ORIGINS`：ブラウザーの許可元をカンマ区切りで指定。既定値は`http://localhost:5173,http://localhost:3000`。
- CookieはHttpOnly・SameSite=Lax・Path=/。ブラウザーの更新リクエストは許可Originと`X-Requested-With: gomoku-web`ヘッダーで検証します。Originがない非ブラウザークライアントでも確認ヘッダーが必要です。BFFはOriginを改変せず中継します。
- `GET /api/auth/me`：認証済みならユーザーID、未認証なら401。
- `POST /api/logout`：現在のセッションをDBで失効し、Cookieを削除。
- アカウント復旧：アカウント更新・復旧トークン消費・認証バージョン加算を同じトランザクションで実施。全端末の旧JWTを次の認証チェックから拒否。
- `/home`：homeの下にログイン／ログアウト状態を表示。フォーカス復帰時と30秒ごとに状態を再確認します。画面自体は未認証でも表示でき、保護するAPIには`middleware.Authentication`を適用します。

既存DBにはバックエンド起動時に`users.auth_version`を追加し、`auth_sessions`を作成します。
既存データは保持します。期限切れ・失効済みのセッションレコードは自動削除しないため、長期運用では定期削除を設定してください。

### 認証のテスト

```sh
cd backend
go test ./...
# localhost:3305のMySQLとlocalhost:1025/8025のMailpitを利用
AUTH_INTEGRATION=1 go test ./internal/interface/httpapi -run TestAuthenticationIntegration -v
```

結合テストは専用の一時ユーザーでログイン、現在端末のログアウト、メールによる復旧、
全端末失効、再ログイン、復旧トークンの再利用拒否、CSRF拒否を確認します。
テストユーザーと取得したテストメールは終了時に削除します。

```sh
cd frontend
npm run build
npm run lint
node --test server/api-proxy.test.mjs
```
