<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API（SQLite 単一ノード版フォーク）

[![Go](https://img.shields.io/badge/Go-1.27.0-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-only-003B57.svg)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**SQLite 専用 Sub2API** — [githubliangliang/sub2api](https://github.com/githubliangliang/sub2api) で公開している、個人利用 / 単一ノード向けの Sub2API の SQLite 版です。データベースは **SQLite のみ**（`modernc.org/sqlite`）で、PostgreSQL には対応していません。

本リポジトリは Sub2API の **SQLite 単一ノード版**です。実行時にはローカルの SQLite ファイルだけを開き、`database.driver` / `DATABASE_DRIVER` に `postgres` を指定しても無視されます。個人の単一ノード利用を想定しており、PostgreSQL / 複数インスタンスには対応していません。

リポジトリ：**https://github.com/githubliangliang/sub2api** · リリース：**https://github.com/githubliangliang/sub2api/releases** · [llms.txt](./llms.txt)

[中文](README.md) | [English](README_EN.md) | [日本語](README_JA.md)

</div>

## 本フォークの変更点

上流版との主な違い：

| 機能 | 内容 |
|------|------|
| **SQLite のみ** | 実行時には SQLite（`modernc.org/sqlite`）だけを開きます。`database.driver` / `DATABASE_DRIVER` に `postgres` を指定しても無視されます |
| **セットアップウィザード** | Web / CLI / `AUTO_SETUP` では SQLite ファイルのパスだけを設定します。host/port/user/ssl は不要です |
| **バックアップ** | 管理画面のバックアップでは `VACUUM INTO` を使って `*.db.gz` を作成します。イメージには `pg_dump` / `psql` を含めません |
| **Redis は任意** | `REDIS_ENABLED=false` の場合はプロセス内の組み込み Redis を使います（単一ノードのみ）。Web ウィザードのステップ 2 に「外部 Redis を使用」スイッチがあり、初期状態ではオフです |
| **メニューの非表示** | 設定項目 `hidden_menu_keys` と管理画面の「メニュー管理」（`/admin/menu`）で、サイドバーの項目を個別に非表示にできます |
| **利用量の記録** | `usage_logs` の冪等性インデックスと `usage_billing_dedup` テーブルを補い、リクエストが成功しても利用量が記録されない問題を防ぎます |
| **決済ページの整理** | 廃止された `/admin/payment/providers` エンドポイントへのフロントエンドのリクエストを削除しています |

関連ドキュメント：

| 内容 | ドキュメント |
|------|--------------|
| リファクタリングの詳細 | [REFACTOR.md](./REFACTOR.md) |
| ネイティブデプロイ | [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) |
| 開発上の規約と注意点 | [DEV_GUIDE.md](./DEV_GUIDE.md) |
| 廃止したページ | [REMOVED_PAGES.md](./REMOVED_PAGES.md) |
| 上流変更の取り込み（リポジトリ全体をマージ**しない**） | [docs/upstream-sync/README.md](./docs/upstream-sync/README.md) |

> **注意**：本リポジトリは PostgreSQL に接続**できません**。`backend/migrations/*.sql` は SQLite 方言です。適用済みの `SELECT 1` no-op マイグレーションを削除しないでください。チェックサムは変更できません。高い同時実行数 / 複数インスタンスには適していません。

---

## デプロイ方法の選び方

| 利用環境 | 方法 |
|----------|------|
| 1 台の VPS で個人利用 | **方法 1**：単一バイナリ + SQLite + 組み込み Redis |
| ローカル環境 / NAS に Docker がある | **方法 2**：SQLite の単一コンテナ |
| 複数インスタンス、高い同時実行数、チーム利用 | 本リポジトリは適していません（SQLite のみ、単一ノード） |

どちらの方法でも、先に `run_mode` を決めてください。**`standard` を推奨**します。次の節を参照してください。

---

## run_mode：standard を推奨

有効な値は `standard` と `simple` だけです。それ以外の値は `config.NormalizeRunMode` によって `standard` に正規化されます。プログラムのデフォルト値、`deploy/config.example.yaml`、`deploy/.env.sqlite.example` はすべて `standard` です。

### 2 つのモードの違い

| 動作 | `standard`（推奨） | `simple` |
|------|---------------------|------------|
| 残高 / サブスクリプションの確認 | 有効：残高 ≤ 0、または `billing.minimum_balance_reserve`（デフォルト `0.000001`）未満の場合 → `403 INSUFFICIENT_BALANCE` | すべてスキップ |
| 利用量と課金 | `usage_logs` に記録し、トークンの実コストを残高から差し引きます | `usage_logs` への記録のみです。コストは計算してログに残しますが、残高は変更しません。ログには `[SIMPLE MODE] Usage recorded (not billed)` と出力します |
| グループ別のスケジューリング | グループで分離します。キーにグループが設定されていればそのグループのアカウントだけを使い、未設定なら**どのグループにも属していないアカウント**だけを使います | グループを無視し、プラットフォームごとのスケジュール可能な全アカウントから選びます |
| 起動時の副作用 | なし | 不足している `anthropic-default` / `openai-default` / `gemini-default` / `grok-default` / `antigravity-default-1,2` グループを自動作成します。管理者の同時実行数がまだ 5 なら、一度だけ 30 に引き上げます |
| サイドバー | 全項目を表示 | ユーザー、グループ、チャネル管理、サブスクリプション、監査ログに加え、ユーザー側の利用量統計、一括画像生成、利用可能なチャネルを非表示にします |
| URL を直接入力 | すべてアクセス可能 | `/admin/groups`、`/admin/subscriptions`、`/admin/redeem`、`/subscriptions`、`/redeem` はホームにリダイレクトされます。`/admin/users` はサイドバーに表示されないだけで、直接アクセスできます |
| `/v1/sub2api/billing` | 利用可能 | 404 |

### 個人利用でも standard を推奨する理由

- **グループは、このシステムでアカウントを分離し、特定のアカウントにリクエストを振り分ける唯一の手段です**（グループのモデルマッピング、グループの上限、キーへのグループ設定）。`simple` でも設定を入力・保存でき、UI 上は有効に見えますが、**スケジューリングには使われません**。アカウントが増えると、この点の調査に時間がかかります。
- 利用量統計、監査ログ、ユーザー管理など、何が起きたかを確認するためのページが、`simple` ではサイドバーから消えます。
- メニューが多すぎる場合は UI 側で調整できます。本フォークでは `hidden_menu_keys` + `/admin/menu` で項目を個別に非表示にでき、全体を `simple` に切り替えるより細かく制御できます。
- `simple` で省ける作業は残高の管理だけです。これは次の設定で対応できます。

### standard で必要な設定：自分の残高を設定する

`standard` は残高を確認します。管理者作成時の初期残高には `default.user_balance`（上流版のデフォルトは `0`）を使います。残高を設定しないと、最初の API リクエストで `403 INSUFFICIENT_BALANCE` が返ります。

- **初回起動前**：`config.yaml` の `default.user_balance` を十分な値に設定してください。`deploy/config.personal.sqlite.yaml` では `100000` を設定済みです。管理者はこの残高で作成されます。
  - この値は**新規登録ユーザー**の初期残高にもなります。登録を開放する場合は `0` に戻し、以下の方法で自分の残高だけを別途設定してください。
- **すでに起動した場合、または Docker の `AUTO_SETUP` でセットアップした場合**（この方法では管理者の初期残高は `0` 固定）：ログイン後、管理画面 → **ユーザー管理**（`/admin/users`）→ 自分の残高を増やしてください。エンドポイントは `POST /admin/users/:id/balance` です。

残高からトークンの実コストが差し引かれます。0 になると再び 403 が返るので、残高を増やしてください。`/admin/usage` で使用した金額を確認できます。

### simple を選ぶ場合

リクエストを転送できればよく、コストを確認せず、グループでアカウントを分離しない場合に選びます。`standard` に戻す前には 2 点に注意してください。`simple` の間は残高が変更されません。利用量ログには計算したコストが記録されますが、残高の差し引きは記録されず、切り替えた後に差し引きが始まります。また、`simple` が起動時に作成したデフォルトグループはデータベースに残ります。これらのグループは空で、グループ未設定のキーは引き続き「どのグループにも属していないアカウント」を使うため、`standard` には影響しません。

---

## デプロイ

### 方法 1：ネイティブデプロイ（個人用サーバーに推奨）

目的：個人の単一ノード利用。ローカルでバイナリをビルドし、サーバーにアップロードします。

構成：単一バイナリ + SQLite + 外部 Redis なし（PostgreSQL / Redis のプロセスは起動しません）。

ローカル環境に必要なもの：Go `1.27.0`（`backend/go.mod` を参照）、Node 20、**pnpm**（npm ではなく pnpm を使ってください）。

#### 1. ローカルでビルド

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api

# フロントエンド
cd frontend
pnpm install
pnpm run build
cd ..

# バックエンド（フロントエンドを組み込み、strip でサイズを削減）
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags embed -ldflags="-s -w" \
  -o ../sub2api ./cmd/server
```

リポジトリのルートに単一ファイル `sub2api` が生成されます。

> ARM サーバーでは `GOARCH=amd64` を `arm64` に変更してください。

#### 2. サーバーにアップロード

```bash
# ローカル環境
scp sub2api root@あなたのサーバー:/tmp/sub2api

# サーバー
sudo useradd -r -s /usr/sbin/nologin sub2api 2>/dev/null || true
sudo mkdir -p /opt/sub2api/data
sudo mv /tmp/sub2api /opt/sub2api/sub2api
sudo chmod +x /opt/sub2api/sub2api
sudo chown -R sub2api:sub2api /opt/sub2api
```

#### 3. 設定

```bash
# リポジトリ内のテンプレートをコピー
sudo cp deploy/config.personal.sqlite.yaml /opt/sub2api/data/config.yaml
sudo nano /opt/sub2api/data/config.yaml
```

**変更が必要な項目：**

```yaml
jwt:
  secret: "openssl rand -hex 32 で生成"

totp:
  encryption_key: "openssl rand -hex 32 で生成（16 進数で 64 文字が必須）"

default:
  admin_email: "あなたのメールアドレス"
  admin_password: "強力なパスワード"
  user_balance: 100000    # テンプレートで設定済み。standard では > 0 が必須。未設定なら最初の API リクエストで 403
```

> `totp.encryption_key` が空の場合、プログラムは**起動するたびに新しいランダムなキーを生成**します。再起動すると既存の 2FA 設定はすべて直ちに無効になり、決済の resume token も使えなくなります。デプロイ前に設定し、その後は変更しないでください。変更すると既存の設定はすべて無効になります。形式は 32 バイトの 16 進数表現で、形式が正しくないとサービスは起動できません。
> `deploy/deploy-remote.sh` で初回デプロイする場合、このキーは `jwt.secret` とともにランダムに自動生成されるため、手動入力は不要です。

テンプレートの主な初期設定：

| 項目 | 値 | 用途 |
|------|----|------|
| `run_mode` | `standard` | グループ / 課金の動作をすべて有効にします。不要なメニューは `/admin/menu` で個別に非表示にしてください |
| `default.user_balance` | `100000` | 管理者と新規登録ユーザーの初期残高。standard では > 0 が必須です |
| `default.user_concurrency` | `10` | 管理者 / 新規ユーザーの同時実行数。足りない場合はユーザー管理で変更できます |
| `database.driver` | `sqlite` | 互換性用の項目。実行時は常に sqlite を使います |
| `database.path` | `/opt/sub2api/data/sub2api.db` | データベースファイル |
| `redis.enabled` | `false` | 組み込み Redis、単一ノード |
| `ops.enabled` | `true` | エラー / システムログと毎日 03:00 の定期クリーンアップを有効にします。集計（`aggregation`）と収集キャッシュは引き続き無効で、CPU 使用量を抑えます |
| コネクションプール | 2 | メモリ使用量を抑えます |

設定テンプレート：`deploy/config.personal.sqlite.yaml`。

#### 4. systemd で自動起動

```bash
sudo cp deploy/sub2api-sqlite.service /etc/systemd/system/sub2api.service
sudo systemctl daemon-reload
sudo systemctl enable --now sub2api
sudo systemctl status sub2api
sudo journalctl -u sub2api -f
```

ユニットファイル：`deploy/sub2api-sqlite.service`。

- 作業ディレクトリ：`/opt/sub2api`
- データ：`DATA_DIR=/opt/sub2api/data`
- 待ち受けアドレス：`0.0.0.0:8080`

#### 5. アクセス

ブラウザーで `http://サーバーIP:8080` を開きます。
設定した管理者のメールアドレスとパスワードでログインしてください。

ファイアウォールで 8080 を許可してください。または、前段に Nginx のリバースプロキシを置いて 80/443 を使います。

#### 6. バックアップ / アップグレード

- **停止中にコピー**：`/opt/sub2api/data` ディレクトリ全体（`sub2api.db`、`sub2api.db-wal`、`sub2api.db-shm`、`config.yaml`）をコピーしてください。
  - `config.yaml` には `jwt.secret` と `totp.encryption_key` が含まれます。失うとすべてのセッションと既存の 2FA 設定が無効になるため、必ず一緒にバックアップしてください。
- **管理画面のバックアップ**（S3 設定後）：`VACUUM INTO` で整合性のあるスナップショットを作成し、オブジェクト名は `*.db.gz` になります。復元時には稼働中の `.db` を置き換え、`-wal` / `-shm` を削除します。
- **アップグレード**：ローカルで再度 `go build` → アップロードして `/opt/sub2api/sub2api` を上書き → `sudo systemctl restart sub2api`。

#### 1 コマンドでデプロイ（任意）

```bash
cp deploy/deploy.env.example deploy/deploy.env   # VPS の情報を入力
deploy/deploy-remote.sh --setup-key              # SSH 公開鍵を初回だけ設定。パスワード入力も 1 回だけ
deploy/deploy-remote.sh                          # 以降は毎回：ビルド → アップロード → 再起動 → ヘルスチェック。失敗時は自動ロールバック
```

バイナリの置き換えと再起動だけを行います。リモートにある既存の `config.yaml` と systemd ユニットは上書きしません。

#### 確認項目

1. ログインできる。
2. API キー / アカウント / グループのページを開ける。
3. API リクエストが成功し、`403 INSUFFICIENT_BALANCE` が返らない。返る場合は残高を増やす。
4. 利用量が記録される。`/admin/usage` にリクエストとコストが表示される。
5. メニュー管理（`/admin/menu`）を開ける。
6. `journalctl -u sub2api` に SQL エラーが大量に出力されていない。

---

### 方法 2：Docker Compose（SQLite の単一コンテナ）

PostgreSQL / 外部 Redis は不要です。

#### 前提条件

- Docker 20.10+
- Docker Compose v2+

#### 手順

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api/deploy

# 初回
cp .env.sqlite.example .env.sqlite
# ADMIN_PASSWORD、JWT_SECRET の変更を推奨（openssl rand -hex 32）

# ビルドして起動
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# ログ
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite logs -f
```

#### アクセス

- URL：`http://localhost:8080`（ポートは `.env.sqlite` の `SERVER_PORT` を参照）。
- 初期管理者：`.env.sqlite` の `ADMIN_EMAIL` / `ADMIN_PASSWORD` を参照。

> `AUTO_SETUP` で作成した管理者の初期残高は `0` 固定です（同時実行数は 5）。`RUN_MODE=standard`（デフォルト）では、最初の API リクエストで `403 INSUFFICIENT_BALANCE` が返ります。ログイン後、`/admin/users` で自分の残高を増やし、必要に応じて同時実行数を調整してください。

#### データの保存場所

| パス | 内容 |
|------|------|
| `deploy/data/sub2api.db` | SQLite データベース |
| `deploy/data/config.yaml` | 実行時の設定 |
| `deploy/data/logs/` | ログ |

#### よく使うコマンド

```bash
# 停止
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite down

# コード変更後に再ビルド
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# 状態
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite ps
```

#### 環境変数（抜粋）

| 変数 | 説明 |
|------|------|
| `RUN_MODE` | `standard`（デフォルト、推奨）/ `simple`。違いは上の [run_mode](#run_modestandard-を推奨) の節を参照してください |
| `DATABASE_PATH` | デフォルトは `/app/data/sub2api.db`。`DATABASE_DRIVER` は無視されます |
| `REDIS_ENABLED` | `false` = 組み込み Redis |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | 初回の `AUTO_SETUP` で作成する管理者 |
| `JWT_SECRET` | セッション用のキー。コンテナを再作成する場合も固定してください |
| `TOTP_ENCRYPTION_KEY` | 2FA を使う場合は固定が必須です。固定しないと再起動で既存の設定が無効になります |
| `SERVER_PORT` / `BIND_HOST` | ホスト側のマッピング。デフォルトは `8080`。ローカルで使用中の場合は `8081` などに変更してください |

完全な例：`deploy/.env.sqlite.example`。

`deploy/docker-compose.yml` / `docker-compose.local.yml` などは上流版から残っているファイルです。本フォークのプロセスは引き続き SQLite だけを開くため、個人用デプロイではこれらを使わないでください。

---

## トラブルシューティング

| 症状 | 原因 / 対処 |
|------|-------------|
| API リクエストで `403 INSUFFICIENT_BALANCE` が返る | standard で残高が不足しています。`/admin/users` で自分の残高を増やすか、初回起動前に `default.user_balance` を設定してください |
| アカウントがスケジュールされない / 利用可能なアカウントがない | standard はグループでアカウントを分離します。グループ設定済みのキーはそのグループのアカウントだけを使い、未設定のキーは**どのグループにも属していないアカウント**だけを使います。キーとアカウントのグループ設定を合わせてください |
| `config.yaml` の `admin_password` を変更してもログイン時のパスワードが変わらない | `default.admin_*` はデータベースのユーザー数が 0 のときに管理者を作成するためだけに使い、その後は同期しません。UI のプロフィール画面、または `users` テーブルを直接変更してパスワードを更新してください |
| 再起動後にすべての 2FA 設定が無効になる | `totp.encryption_key` が固定されておらず、起動のたびにランダムなキーを生成しています。32 バイトのキーを 16 進数表現で固定し、`config.yaml` と一緒にバックアップしてください |
| 起動直後に終了し、`default.admin_password must be changed from the example value` が出る | テンプレートの `CHANGE_ME_STRONG_PASSWORD` を変更していません。実際に使うパスワード（8–128 文字）に変更してから起動してください |
| 起動時にマイグレーションのチェックサム不一致が出る | 適用済みのマイグレーションファイルが変更されています。新しいマイグレーションを追加して解決してください。個人用データベースなら `*.db` を削除して作り直す方法もあります |
| ログインで 503 が返る | 古いデータベースに補助テーブル（例：`user_allowed_groups`）がありません。ログで不足しているテーブルを確認してください。`EnsureSQLiteAuxTables` の DDL はマイグレーションと列ごとに一致している必要があります |

---

## 技術スタック（本フォーク）

| コンポーネント | 技術 |
|----------------|------|
| バックエンド | Go 1.27.0、Gin、Ent、`modernc.org/sqlite` |
| フロントエンド | Vue 3、Vite、TailwindCSS（pnpm） |
| データベース | **SQLite のみ**（`backend/migrations/*.sql` は SQLite 方言） |
| キャッシュ | Redis は任意。外部 Redis を無効にするとプロセス内の組み込み実装を使います（単一ノードのみ） |
| 配布 | 単一バイナリ（`-tags embed` でフロントエンドを組み込み）/ Docker Compose |

---

## FAQ（検索用）

**これは SQLite 版の Sub2API ですか？**

はい。本リポジトリ（[githubliangliang/sub2api](https://github.com/githubliangliang/sub2api)）は **SQLite 専用**の Sub2API です。実行時には SQLite だけを開き、PostgreSQL には接続しません。

**PostgreSQL に対応していますか？**

いいえ。`backend/migrations/*.sql` は SQLite 方言です。本リポジトリは単一ノードの SQLite のみを対象にしています。

**一般的な PostgreSQL デプロイとの違いは何ですか？**

SQLite のみ、Redis は任意（プロセス内への組み込みが可能）、メニューの非表示に対応し、個人の単一ノードデプロイを想定しています。上の「本フォークの変更点」を参照してください。

---

## リポジトリとライセンス

- リポジトリ（SQLite 単一ノード版）：https://github.com/githubliangliang/sub2api
- リリース：https://github.com/githubliangliang/sub2api/releases
- ライセンス：[LICENSE](./LICENSE)（LGPLv3）
- AI 向けインデックス：[llms.txt](./llms.txt)
