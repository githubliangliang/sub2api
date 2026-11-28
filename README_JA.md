<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API（SQLite 版）

[![Go](https://img.shields.io/badge/Go-1.27.4-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D.svg)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-Only-003B57.svg)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**Sub2API SQLite 版** — 軽量で **SQLite 専用**の Sub2API ビルド（[githubliangliang/sub2api](https://github.com/githubliangliang/sub2api)）。  
すべてのデータを単一のローカル SQLite ファイル（`modernc.org/sqlite`）に保存し、プロセス内組み込み Redis にも対応。PostgreSQL や外部 Redis は不要です。

リポジトリ：**https://github.com/githubliangliang/sub2api** · リリース：**https://github.com/githubliangliang/sub2api/releases** · [llms.txt](./llms.txt)

[English](README.md) | [中文](README_CH.md) | [日本語](README_JA.md)

</div>

## Sub2API SQLite 版について

本リポジトリは、外部データベース不要の単一ノード運用に向けた **SQLite 専用版 Sub2API** を提供します。

| 特徴 | 内容 |
|------|------|
| **SQLite 専用** | 純 Go 製 SQLite ドライバ（`modernc.org/sqlite`、CGO 不要）を使用。実行時はローカルの SQLite ファイルのみを開き、`database.driver` / `DATABASE_DRIVER` に `postgres` を指定しても無視されます |
| **SQLite 向けセットアップウィザード** | Web、CLI（`--setup`）、`AUTO_SETUP` では SQLite ファイルのパスのみを設定し、PostgreSQL の host / port / user / ssl は不要です |
| **ネイティブ SQLite バックアップ** | 管理画面のバックアップは `VACUUM INTO` で整合性のある `*.db.gz` スナップショットを作成。コンテナイメージに `pg_dump` / `psql` は含まれません |
| **任意の組み込み Redis** | `redis.enabled: false` / `REDIS_ENABLED=false` でプロセス内組み込み Redis を使用（単一インスタンスのみ）。Web ウィザードのステップ 2 でも「外部 Redis を使用」は初期状態でオフです |
| **メニューの非表示設定** | `hidden_menu_keys` または管理画面「メニュー管理」（`/admin/menu`）でサイドバー項目を個別に非表示にできます |
| **利用量記録の補強** | SQLite 向けに `usage_logs` の冪等性インデックスと `usage_billing_dedup` テーブルを整備し、成功したリクエストの利用量欠落を防ぎます |
| **決済ページの整理** | 廃止された `/admin/payment/providers` へのフロントエンドリクエストを削除しています |

> **注意**：本バージョンは PostgreSQL および複数インスタンスによるスケールアウトには**対応していません**。`backend/migrations/*.sql` はすべて SQLite 方言です。適用済みの `SELECT 1` no-op マイグレーションは削除しないでください（チェックサムは変更不可）。

関連ドキュメント：

| 内容 | ドキュメント |
|------|--------------|
| SQLite リファクタリング詳細 | [REFACTOR.md](./REFACTOR.md) |
| 開発規約と注意点 | [DEV_GUIDE.md](./DEV_GUIDE.md) |
| 廃止されたページ | [REMOVED_PAGES.md](./REMOVED_PAGES.md) |
| ネイティブ systemd / リモート運用ガイド | [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) |
| 上流変更の取り込み（リポジトリ全体の merge **不可**） | [docs/upstream-sync/README.md](./docs/upstream-sync/README.md) |

---

## インストール

環境に合わせていずれかの方法を選んでください。いずれも PostgreSQL や外部 Redis は不要です。

### 方法 1：ビルド済みバイナリ（最速）

[GitHub Releases](https://github.com/githubliangliang/sub2api/releases) から OS・アーキテクチャ（Linux / macOS / Windows）に合うアーカイブをダウンロードして展開し、起動します。

```bash
# 直接起動（初回は Web セットアップウィザードが起動します）
./sub2api
```

- ブラウザーで `http://localhost:8080` を開き、ウィザードで SQLite ファイルパスを設定します（外部 Redis をオフのままにすると組み込み Redis を使用）。
- ターミナルで対話的に初期化したい場合は `./sub2api --setup` を実行します。
- `config.yaml` と SQLite ファイルの保存先を指定したい場合は `DATA_DIR` を設定します：

```bash
DATA_DIR=./data ./sub2api
```

### 方法 2：Docker（単一コンテナ）

公開イメージをそのまま起動し、SQLite データと設定用のディレクトリを 1 つマウントします。

```bash
docker run -d \
  --name sub2api \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e AUTO_SETUP=true \
  -e DATABASE_PATH=/app/data/sub2api.db \
  -e REDIS_ENABLED=false \
  -e RUN_MODE=standard \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  ghcr.io/githubliangliang/sub2api:latest
```

または **Docker Compose**（ローカルリポジトリからビルド）を使用します：

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api/deploy

cp .env.sqlite.example .env.sqlite
# 推奨：ADMIN_EMAIL、ADMIN_PASSWORD、JWT_SECRET（openssl rand -hex 32）を設定

docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite logs -f
```

よく使う Docker Compose コマンド：

```bash
# 停止
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite down

# コード更新後に再ビルド
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# 状態確認
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite ps
```

- アクセス URL：`http://localhost:8080`（ポートは `.env.sqlite` の `SERVER_PORT`）。
- データ保存先：`deploy/data/` に `sub2api.db`、`config.yaml`、`logs/` を保存。
- 主な環境変数（全項目は `deploy/.env.sqlite.example` を参照）：

| 変数 | 説明 |
|------|------|
| `RUN_MODE` | `standard`（デフォルト、推奨）または `simple`。違いは [`run_mode`](#run_modestandard推奨) を参照 |
| `DATABASE_PATH` | SQLite データベースパス。デフォルト `/app/data/sub2api.db`（`DATABASE_DRIVER` は無視） |
| `REDIS_ENABLED` | `false` でプロセス内組み込み Redis を使用 |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | 初回 `AUTO_SETUP` で作成する管理者（空欄時は自動生成され起動ログに出力） |
| `JWT_SECRET` | セッション署名キー。コンテナ再作成時のログアウトを防ぐため固定を推奨 |
| `TOTP_ENCRYPTION_KEY` | 2FA 利用時は 64 文字の 16 進数キーを固定必須（未固定だと再起動で既存 2FA が無効化） |
| `SERVER_PORT` / `BIND_HOST` | ホスト側ポート設定（デフォルト `8080`） |

> 注：`deploy/docker-compose.yml` や `deploy/docker-compose.local.yml` は上流版の残存ファイルです。本バージョンは常に SQLite のみを開くため、`deploy/docker-compose.sqlite.yml` を使用してください。

### 方法 3：ソースからビルド

Go `1.27.4`（`backend/go.mod` 参照）、Node.js `20+`（CI は Node 24）、**pnpm**（npm 不可）が必要です。

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api

# 1. フロントエンドをビルド
cd frontend
pnpm install
pnpm run build
cd ..

# 2. フロントエンド組み込みの単一バイナリをビルド（純 Go、CGO 不要）
cd backend
CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o ../sub2api ./cmd/server
cd ..

# 3. 起動
DATA_DIR=./data ./sub2api
```

> 別プラットフォーム向けにクロスコンパイルする場合は、`go build` の前に `GOOS=linux GOARCH=amd64`（または `arm64`、`darwin`、`windows`）を付けてください。

### 設定ファイルとシークレット

セットアップウィザードの代わりに初回起動前から `config.yaml` で設定したい場合は、`deploy/config.personal.sqlite.yaml` を `$DATA_DIR/config.yaml` にコピーして編集します。

必須の調整項目：

```yaml
database:
  driver: "sqlite"
  path: "./data/sub2api.db"

redis:
  enabled: false              # 組み込み Redis を使用（単一インスタンス）

jwt:
  secret: "openssl rand -hex 32 で生成"

totp:
  encryption_key: "openssl rand -hex 32 で生成"  # 64 文字の 16 進数（32 バイト）が必須

default:
  admin_email: "you@example.com"
  admin_password: "強力なパスワード"
  user_balance: 100000        # standard モードでは > 0 が必須（0 だと初回 API 呼び出しで 403）
```

> **`totp.encryption_key` は固定してください**：空欄のままにすると起動のたびにランダムなキーが生成され、既存の 2FA 設定がすべて無効になります。形式は 32 バイトの 64 文字 16 進数である必要があります（形式不正時は起動失敗）。

Linux の常駐サービスとして運用する場合の systemd ユニット（`deploy/sub2api-sqlite.service`）やデプロイスクリプト（`deploy/deploy-remote.sh`）については [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) を参照してください。

---

## `run_mode`：`standard`（推奨）

有効な値は `standard` と `simple` のみです（その他の値は `config.NormalizeRunMode` により `standard` に正規化されます）。デフォルトは `standard` です。

| 動作 | `standard`（推奨） | `simple` |
|------|-------------------|----------|
| 残高 / サブスクリプション確認 | 有効：残高 ≤ 0 または `billing.minimum_balance_reserve`（既定 `0.000001`）未満 → `403 INSUFFICIENT_BALANCE` | スキップ |
| 利用量と課金 | `usage_logs` に記録し、トークン実コストを残高から減算 | `usage_logs` 記録のみ（コストはログ出力されるが残高は不変） |
| グループ別スケジューリング | グループ分離：グループ紐付け済みキーはそのグループのアカウントのみ使用。未紐付けキーは**どのグループにも属さないアカウント**のみ使用 | グループを無視し、プラットフォームの全利用可能アカウントを使用 |
| 起動時の副作用 | なし | デフォルトグループ（`anthropic-default`、`openai-default` 等）を自動作成し、同時実行数 5 の管理者を一度だけ 30 に引き上げ |
| サイドバーとルート | 全画面表示。不要な項目は `/admin/menu` で個別に非表示化可能 | ユーザー、グループ、チャネル、サブスクリプション、監査ログを非表示にし、一部ページをリダイレクト |
| `/v1/sub2api/billing` | 利用可能 | 404 |

### `standard` を推奨する理由

- **グループはアカウント分離とキー振り分けの唯一の手段です**（モデルマッピング、グループ上限、キーのグループ紐付け）。`simple` では UI 上で設定保存できても、スケジューリング時には無視されます。
- 利用量統計、監査ログ、ユーザー管理をそのまま利用できます。メニューが多い場合は `simple` に下げず、`/admin/menu`（`hidden_menu_keys`）で個別に非表示にしてください。

### `standard` モードでの初期残高設定

`standard` モードでは API 呼び出しに残高が必要です：

- **初回起動前に `config.yaml` で設定**：`default.user_balance` を正の値（例：`100000`）にします。将来ユーザー登録を開放する場合は `0` に戻し、管理者だけ手動加算してください。
- **Web ウィザードまたは Docker `AUTO_SETUP` で構築した場合**（初期残高は `0`）：ログイン後、**管理画面 → ユーザー管理**（`/admin/users`）から自分の残高を増やしてください（`POST /admin/users/:id/balance`）。

---

## バックアップとアップグレード

- **停止時ファイルバックアップ**：プロセスを停止し、データディレクトリ全体（`sub2api.db`、`sub2api.db-wal`、`sub2api.db-shm`、`config.yaml`）をコピーします。`jwt.secret` と `totp.encryption_key` を失うと全セッションと 2FA 設定が無効になるため、`config.yaml` も必ず含めてください。
- **オンライン S3 バックアップ**：管理画面で S3 を設定すると、SQLite `VACUUM INTO` で整合性のある `*.db.gz` スナップショットを作成できます。復元時は稼働中の `.db` を置き換え、古い `-wal` / `-shm` を削除します。
- **アップグレード**：`sub2api` バイナリを差し替える（または Docker イメージを更新・再ビルドする）だけで、再起動時に SQLite マイグレーションが自動適用されます。

---

## トラブルシューティング

| 症状 | 原因と対処 |
|------|------------|
| API リクエストで `403 INSUFFICIENT_BALANCE` | `standard` モードで残高が `0` です。`/admin/users` で残高を追加するか、初回起動前に `default.user_balance` を設定してください |
| アカウントが選択されない / 利用可能アカウントなし | `standard` はグループで分離します。グループ紐付け済みキーはそのグループのアカウントのみ、未紐付けキーは**どのグループにも属さないアカウント**のみを使用します |
| `config.yaml` の `admin_password` を変えてもログインパスワードが変わらない | `default.admin_*` はユーザー数 0 の初回作成時のみ参照されます。UI のプロフィール画面、または `users` テーブルから変更してください |
| 再起動後にすべての 2FA が無効になる | `totp.encryption_key` 未設定のため起動ごとにランダム生成されています。64 文字の 16 進数キーを固定し `config.yaml` と一緒にバックアップしてください |
| `default.admin_password must be changed from the example value` で起動終了 | テンプレートの `CHANGE_ME_STRONG_PASSWORD` を実際のパスワード（8–128 文字）に変更してください |
| マイグレーションのチェックサム不一致エラー | 適用済みの SQL マイグレーションが編集されています。新しいマイグレーションを追加するか、新規やり直しなら `*.db` を削除してください |
| ログイン時に `503` が返る | 古い SQLite データベースに補助テーブル（例：`user_allowed_groups`）が不足しています。ログで不足テーブル名を確認してください |

---

## 技術スタック

| コンポーネント | 技術 |
|----------------|------|
| バックエンド | Go 1.27.4、Gin、Ent、`modernc.org/sqlite`（純 Go、CGO 不要） |
| フロントエンド | Vue 3、Vite、TailwindCSS（pnpm） |
| データベース | **SQLite 専用**（`backend/migrations/*.sql` は SQLite 方言） |
| キャッシュ | Redis は任意（`redis.enabled: false` でプロセス内組み込み Redis を使用） |
| 配布形式 | フロントエンド組み込み単一バイナリ（`-tags embed`）または単一 Docker コンテナ |

---

## FAQ

**これは Sub2API の SQLite 版ですか？**  
はい。[githubliangliang/sub2api](https://github.com/githubliangliang/sub2api) はローカルの SQLite ファイルに全データを保存する **SQLite 専用版** Sub2API です。

**PostgreSQL は使えますか？**  
いいえ。マイグレーションはすべて SQLite 方言であり、PostgreSQL には対応していません。

**Redis のインストールは必要ですか？**  
不要です。`redis.enabled: false`（または `REDIS_ENABLED=false`）を指定すれば組み込み Redis が動作します。外部 Redis は任意です。

---

## リポジトリとライセンス

- リポジトリ：https://github.com/githubliangliang/sub2api
- リリース：https://github.com/githubliangliang/sub2api/releases
- ライセンス：[LICENSE](./LICENSE)（LGPLv3）
- AI 向けインデックス：[llms.txt](./llms.txt)
