<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API (SQLite fork for a single node)

[![Go](https://img.shields.io/badge/Go-1.27.0-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-only-003B57.svg)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**SQLite-only Sub2API** — a personal / single-node SQLite build of Sub2API at [githubliangliang/sub2api](https://github.com/githubliangliang/sub2api). Database is **SQLite only** (`modernc.org/sqlite`); PostgreSQL is not supported.

This repository is a **SQLite edition of Sub2API for a single node**. At runtime, it opens only a local SQLite file; setting `database.driver` / `DATABASE_DRIVER` to `postgres` is ignored. It is intended for personal use on a single node and does not support PostgreSQL or multiple instances.

Repository: **https://github.com/githubliangliang/sub2api** · Releases: **https://github.com/githubliangliang/sub2api/releases** · [llms.txt](./llms.txt)

[中文](README.md) | [English](README_EN.md) | [日本語](README_JA.md)

</div>

## Changes in this fork

Main differences from upstream:

| Capability | Details |
|------------|---------|
| **SQLite only** | Opens only SQLite (`modernc.org/sqlite`) at runtime. Setting `database.driver` / `DATABASE_DRIVER` to `postgres` is ignored |
| **Setup wizard** | Web / CLI / `AUTO_SETUP` collect only the SQLite file path; host/port/user/ssl are no longer required |
| **Backups** | Admin backups use `VACUUM INTO` to create `*.db.gz` files; images no longer include `pg_dump` / `psql` |
| **Optional Redis** | `REDIS_ENABLED=false` uses Redis embedded in the process (single node only); step 2 of the web wizard has a "Use external Redis" switch, disabled by default |
| **Hide menus** | The `hidden_menu_keys` setting and the admin "Menu Management" page (`/admin/menu`) let you hide individual sidebar entries |
| **Usage persistence** | Adds idempotency indexes for `usage_logs` and the `usage_billing_dedup` table so successful requests still record usage |
| **Payment page cleanup** | Removes frontend requests to the retired `/admin/payment/providers` endpoint |

Further documentation:

| Topic | Document |
|-------|----------|
| Refactoring details | [REFACTOR.md](./REFACTOR.md) |
| Native deployment | [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) |
| Development conventions and pitfalls | [DEV_GUIDE.md](./DEV_GUIDE.md) |
| Removed pages | [REMOVED_PAGES.md](./REMOVED_PAGES.md) |
| Porting upstream changes (**do not** merge the entire repository) | [docs/upstream-sync/README.md](./docs/upstream-sync/README.md) |

> **Note**: This repository **cannot** connect to PostgreSQL. `backend/migrations/*.sql` uses the SQLite dialect. Do not delete already applied `SELECT 1` no-op migrations; their checksums are immutable. This fork is not suitable for high concurrency or multiple instances.

---

## Choosing a deployment method

| Scenario | Method |
|----------|--------|
| Personal use on one VPS | **Method 1**: a single binary + SQLite + embedded Redis |
| Docker already available on a local machine / NAS | **Method 2**: a single SQLite container |
| Multiple instances, high concurrency, or team use | This repository is not suitable (SQLite only, single node) |

Choose `run_mode` before using either method. **`standard` is recommended**; see the next section.

---

## run_mode: standard is recommended

The only valid values are `standard` and `simple`. Other values are normalized to `standard` by `config.NormalizeRunMode`. The application default, `deploy/config.example.yaml`, and `deploy/.env.sqlite.example` all use `standard`.

### How the two modes differ

| Behavior | `standard` (recommended) | `simple` |
|----------|--------------------------|------------|
| Balance / subscription checks | Enabled: balance ≤ 0 or below `billing.minimum_balance_reserve` (default `0.000001`) → `403 INSUFFICIENT_BALANCE` | Skipped entirely |
| Usage and billing | Writes `usage_logs` and deducts the actual token cost from the balance | Writes only `usage_logs` (costs are still calculated and logged, but the balance is unchanged; the log contains `[SIMPLE MODE] Usage recorded (not billed)`) |
| Group scheduling | Isolated by group: a key bound to a group uses only that group's accounts; an unbound key uses only accounts **outside all groups** | Ignores groups and selects all schedulable accounts for the platform |
| Startup side effects | None | Automatically creates missing `anthropic-default` / `openai-default` / `gemini-default` / `grok-default` / `antigravity-default-1,2` groups; raises the administrator's concurrency from 5 to 30 once if it is still 5 |
| Sidebar | All entries | Hides Users, Groups, Channel Management, Subscriptions, and Audit Logs, plus the user pages for Usage Statistics, Batch Image Generation, and Available Channels |
| Direct URL access | All pages accessible | `/admin/groups`, `/admin/subscriptions`, `/admin/redeem`, `/subscriptions`, and `/redeem` redirect to the home page (`/admin/users` is only hidden from the sidebar and remains directly accessible) |
| `/v1/sub2api/billing` | Available | 404 |

### Why standard is recommended even for personal use

- **Groups are the system's only way to isolate accounts and route requests to specific accounts** (group model mappings, group limits, and keys bound to groups). In `simple` mode, these settings can be entered and saved, and appear to work in the UI, but **do not participate in scheduling**. With more accounts, this becomes particularly difficult to troubleshoot.
- Pages that help you understand what happened, such as usage statistics, audit logs, and user management, disappear from the sidebar in `simple` mode.
- Too many menus is a UI issue. This fork provides `hidden_menu_keys` + `/admin/menu` to hide entries individually, giving you more control than switching everything to `simple`.
- The only task `simple` saves you is managing the balance, which the next step addresses.

### Required step for standard: set your own balance

`standard` checks the balance. When the administrator is created, the initial balance comes from `default.user_balance` (upstream default: `0`). Without a balance, the first API request returns `403 INSUFFICIENT_BALANCE`.

- **Before the first startup**: set `default.user_balance` in `config.yaml` to a sufficient value (`deploy/config.personal.sqlite.yaml` already sets `100000`). The administrator is created with this balance.
  - This is also the initial balance for **newly registered users**. If you plan to allow registration, change it back to `0`, then set your own balance separately as described below.
- **If the application has already started, or was installed with Docker `AUTO_SETUP`** (which fixes the administrator's initial balance at `0`): log in, open the admin **User Management** page (`/admin/users`), and increase your own balance (endpoint: `POST /admin/users/:id/balance`).

Actual token costs are deducted from the balance. When it reaches 0, requests return 403 again; increase the balance to continue. View spending in `/admin/usage`.

### When to choose simple

Choose it when you only need request forwarding, do not track costs, and do not use groups to isolate accounts. Before switching back to `standard`, keep two things in mind: the balance is unchanged during `simple` mode (usage logs contain calculated costs, but no balance deductions are recorded), and deductions start only after switching. The default groups created by `simple` remain in the database; these groups are empty, and unbound keys still use accounts outside all groups, so they do not affect `standard`.

---

## Deployment

### Method 1: native deployment (recommended for personal servers)

Goal: personal use on a single node. Build the binary locally, then upload it.

Setup: a single binary + SQLite + no external Redis (no PostgreSQL / Redis processes).

Local requirements: Go `1.27.0` (see `backend/go.mod`), Node 20, and **pnpm** (use pnpm instead of npm).

#### 1. Build locally

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api

# Frontend
cd frontend
pnpm install
pnpm run build
cd ..

# Backend (embed the frontend and strip the binary to reduce size)
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags embed -ldflags="-s -w" \
  -o ../sub2api ./cmd/server
```

This produces a single `sub2api` file in the repository root.

> For ARM servers, change `GOARCH=amd64` to `arm64`.

#### 2. Upload to the server

```bash
# Local machine
scp sub2api root@your-server:/tmp/sub2api

# Server
sudo useradd -r -s /usr/sbin/nologin sub2api 2>/dev/null || true
sudo mkdir -p /opt/sub2api/data
sudo mv /tmp/sub2api /opt/sub2api/sub2api
sudo chmod +x /opt/sub2api/sub2api
sudo chown -R sub2api:sub2api /opt/sub2api
```

#### 3. Configure

```bash
# Copy the template from the repository
sudo cp deploy/config.personal.sqlite.yaml /opt/sub2api/data/config.yaml
sudo nano /opt/sub2api/data/config.yaml
```

**Required changes:**

```yaml
jwt:
  secret: "Generate with openssl rand -hex 32"

totp:
  encryption_key: "Generate with openssl rand -hex 32 (must be 64 hex characters)"

default:
  admin_email: "your-email"
  admin_password: "strong-password"
  user_balance: 100000    # Already set in the template; standard requires > 0, or the first API request returns 403
```

> If `totp.encryption_key` is empty, the application **generates a new random key on every startup**. After a restart, all existing 2FA bindings become invalid immediately, and payment resume tokens are also unusable. Set it before deployment and keep it unchanged afterward; replacing it invalidates all existing bindings. It must be 32 bytes encoded as hex; an invalid format prevents startup.
> On a first deployment with `deploy/deploy-remote.sh`, this key and `jwt.secret` are generated randomly together, so you do not need to fill them in manually.

Key defaults in the template:

| Setting | Value | Purpose |
|---------|-------|---------|
| `run_mode` | `standard` | Full group / billing behavior; hide unwanted menus individually with `/admin/menu` |
| `default.user_balance` | `100000` | Initial balance for the administrator and newly registered users; must be > 0 in standard |
| `default.user_concurrency` | `10` | Concurrency for the administrator / new users; adjust it in User Management if needed |
| `database.driver` | `sqlite` | Compatibility field; runtime always uses sqlite |
| `database.path` | `/opt/sub2api/data/sub2api.db` | Database file |
| `redis.enabled` | `false` | Embedded Redis, single node |
| `ops.enabled` | `true` | Enables error / system logs and daily cleanup at 03:00; aggregation (`aggregation`) and collection caching remain disabled to save CPU |
| Connection pool | 2 | Saves memory |

Configuration template: `deploy/config.personal.sqlite.yaml`.

#### 4. Start automatically with systemd

```bash
sudo cp deploy/sub2api-sqlite.service /etc/systemd/system/sub2api.service
sudo systemctl daemon-reload
sudo systemctl enable --now sub2api
sudo systemctl status sub2api
sudo journalctl -u sub2api -f
```

Unit file: `deploy/sub2api-sqlite.service`.

- Working directory: `/opt/sub2api`
- Data: `DATA_DIR=/opt/sub2api/data`
- Listen address: `0.0.0.0:8080`

#### 5. Access

Open `http://your-server-ip:8080` in a browser.
Log in with the administrator email and password from the configuration.

Allow port 8080 through the firewall, or place an Nginx reverse proxy in front on ports 80/443.

#### 6. Backups / upgrades

- **Copy while stopped**: copy the entire `/opt/sub2api/data` directory (`sub2api.db`, `sub2api.db-wal`, `sub2api.db-shm`, and `config.yaml`).
  - `config.yaml` contains `jwt.secret` and `totp.encryption_key`. Losing it invalidates all sessions and all existing 2FA bindings; include it in the backup.
- **Admin backups** (after configuring S3): `VACUUM INTO` creates a consistent snapshot named `*.db.gz`. Restoring replaces the live `.db` and removes `-wal` / `-shm`.
- **Upgrade**: run `go build` locally again → upload and replace `/opt/sub2api/sub2api` → `sudo systemctl restart sub2api`.

#### One-command deployment (optional)

```bash
cp deploy/deploy.env.example deploy/deploy.env   # Fill in the VPS details
deploy/deploy-remote.sh --setup-key              # Install the SSH public key once; enter the password only once
deploy/deploy-remote.sh                          # Each subsequent run: build → upload → restart → health check; automatic rollback on failure
```

This replaces the binary and restarts the service. The existing remote `config.yaml` and systemd unit are not overwritten.

#### Verification checklist

1. Login succeeds.
2. API Key / Accounts / Groups pages open.
3. An API request succeeds without `403 INSUFFICIENT_BALANCE` (increase the balance if it occurs).
4. Usage is recorded (`/admin/usage` shows requests and costs).
5. Menu Management (`/admin/menu`) opens.
6. `journalctl -u sub2api` does not show a flood of SQL errors.

---

### Method 2: Docker Compose (single SQLite container)

No PostgreSQL / external Redis is required.

#### Prerequisites

- Docker 20.10+
- Docker Compose v2+

#### Steps

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api/deploy

# First run
cp .env.sqlite.example .env.sqlite
# Recommended: change ADMIN_PASSWORD and JWT_SECRET (openssl rand -hex 32)

# Build and start
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# Logs
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite logs -f
```

#### Access

- URL: `http://localhost:8080` (see `SERVER_PORT` in `.env.sqlite` for the port).
- Default administrator: see `ADMIN_EMAIL` / `ADMIN_PASSWORD` in `.env.sqlite`.

> `AUTO_SETUP` fixes the administrator's initial balance at `0` (concurrency: 5). With `RUN_MODE=standard` (the default), the first API request returns `403 INSUFFICIENT_BALANCE`. After logging in, open `/admin/users` to increase your balance and adjust concurrency as needed.

#### Data locations

| Path | Contents |
|------|----------|
| `deploy/data/sub2api.db` | SQLite database |
| `deploy/data/config.yaml` | Runtime configuration |
| `deploy/data/logs/` | Logs |

#### Common commands

```bash
# Stop
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite down

# Rebuild after code changes
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# Status
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite ps
```

#### Environment variables (excerpt)

| Variable | Description |
|----------|-------------|
| `RUN_MODE` | `standard` (default, recommended) / `simple`; see the [run_mode](#run_mode-standard-is-recommended) section for the differences |
| `DATABASE_PATH` | Defaults to `/app/data/sub2api.db` (`DATABASE_DRIVER` is ignored) |
| `REDIS_ENABLED` | `false` = embedded Redis |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | Administrator created by the first `AUTO_SETUP` |
| `JWT_SECRET` | Session key; keep it fixed when recreating the container |
| `TOTP_ENCRYPTION_KEY` | Keep it fixed when using 2FA, or a restart invalidates existing bindings |
| `SERVER_PORT` / `BIND_HOST` | Host mapping; default `8080`, or use `8081` if the local port is occupied |

Full example: `deploy/.env.sqlite.example`.

`deploy/docker-compose.yml` / `docker-compose.local.yml` and similar files are upstream leftovers. This fork still opens only SQLite; do not use these files for personal deployment.

---

## Troubleshooting

| Symptom | Cause / solution |
|---------|------------------|
| API requests return `403 INSUFFICIENT_BALANCE` | Insufficient balance in standard. Increase your balance in `/admin/users`, or set `default.user_balance` before the first startup |
| Scheduling never finds an account / "no available account" | standard isolates accounts by group: keys bound to a group use only that group's accounts; unbound keys use only **accounts outside all groups**. Match key and account group assignments |
| Changing `admin_password` in `config.yaml` leaves the login password unchanged | `default.admin_*` is used only to create the administrator when the database has 0 users; no synchronization happens afterward. Change the password through the profile page in the UI or directly in the `users` table |
| All 2FA bindings become invalid after a restart | `totp.encryption_key` was not fixed, so a new random key is generated on each startup. Set a fixed 32-byte key encoded as hex and back it up with `config.yaml` |
| Startup exits with `default.admin_password must be changed from the example value` | The template's `CHANGE_ME_STRONG_PASSWORD` was not changed. Set a real password (8–128 characters), then restart |
| Startup reports a migration checksum mismatch | An already applied migration file was changed. Add a new migration to resolve it; for a personal database, you can delete `*.db` and start over |
| Login returns 503 | An older database is missing auxiliary tables, such as `user_allowed_groups`. Check the logs for the missing table; the DDL in `EnsureSQLiteAuxTables` must match the migration column for column |

---

## Technology stack (this fork)

| Component | Technology |
|-----------|------------|
| Backend | Go 1.27.0, Gin, Ent, `modernc.org/sqlite` |
| Frontend | Vue 3, Vite, TailwindCSS (pnpm) |
| Database | **SQLite only** (`backend/migrations/*.sql` uses the SQLite dialect) |
| Cache | Optional Redis; disabling external Redis uses an implementation embedded in the process (single node only) |
| Distribution | A single binary (`-tags embed` embeds the frontend) / Docker Compose |

---

## FAQ (search reference)

**Is this the SQLite version of Sub2API?**

Yes. This repository ([githubliangliang/sub2api](https://github.com/githubliangliang/sub2api)) is a **SQLite-only** Sub2API. At runtime, it opens only SQLite and does not connect to PostgreSQL.

**Does it support PostgreSQL?**

No. `backend/migrations/*.sql` uses the SQLite dialect. This repository supports only SQLite on a single node.

**How does it differ from a typical PostgreSQL deployment?**

SQLite only, optional Redis (which can be embedded in the process), configurable menu visibility, and deployment for personal use on a single node. See "Changes in this fork" above.

---

## Repository and license

- Repository (SQLite edition for a single node): https://github.com/githubliangliang/sub2api
- Releases: https://github.com/githubliangliang/sub2api/releases
- License: [LICENSE](./LICENSE) (LGPLv3)
- AI indexing entry: [llms.txt](./llms.txt)
