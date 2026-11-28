<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API (SQLite Edition)

[![Go](https://img.shields.io/badge/Go-1.27.4-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D.svg)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-Only-003B57.svg)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**Sub2API SQLite Edition** — a lightweight, **SQLite-only** build of Sub2API ([githubliangliang/sub2api](https://github.com/githubliangliang/sub2api)).  
Store everything in a single local SQLite database file (`modernc.org/sqlite`), with optional in-process embedded Redis — no PostgreSQL or external Redis required.

Repository: **https://github.com/githubliangliang/sub2api** · Releases: **https://github.com/githubliangliang/sub2api/releases** · [llms.txt](./llms.txt)

[English](README.md) | [中文](README_CH.md) | [日本語](README_JA.md)

</div>

## About Sub2API SQLite Edition

This repository maintains a **SQLite-only** edition of Sub2API designed for zero-dependency single-node operation:

| Feature | Details |
|---------|---------|
| **SQLite only** | Uses pure-Go SQLite (`modernc.org/sqlite`, no CGO). At runtime it opens only a local SQLite file; `database.driver` / `DATABASE_DRIVER` set to `postgres` is ignored |
| **SQLite-first setup wizard** | Web, CLI (`--setup`), and `AUTO_SETUP` only ask for the SQLite database path instead of PostgreSQL host/port/user/SSL fields |
| **Native SQLite backups** | Admin backups use `VACUUM INTO` to produce consistent `*.db.gz` snapshots; container images no longer bundle `pg_dump` / `psql` |
| **Optional embedded Redis** | Setting `redis.enabled: false` / `REDIS_ENABLED=false` runs an in-process embedded Redis (single instance only). Step 2 of the web wizard leaves "Use external Redis" off by default |
| **Configurable menus** | Hide individual sidebar entries via `hidden_menu_keys` or Admin → Menu Management (`/admin/menu`) |
| **Usage persistence** | Adds SQLite idempotency indexes on `usage_logs` and the `usage_billing_dedup` table so successful requests reliably record usage |
| **Payment page cleanup** | Removes frontend calls to the retired `/admin/payment/providers` endpoint |

> **Important**: This edition **does not** support PostgreSQL or multi-instance horizontal scaling. All migrations in `backend/migrations/*.sql` use the SQLite dialect; do not delete already-applied `SELECT 1` no-op migrations (checksums are immutable).

Further reading:

| Topic | Document |
|-------|----------|
| SQLite refactoring notes | [REFACTOR.md](./REFACTOR.md) |
| Development conventions & pitfalls | [DEV_GUIDE.md](./DEV_GUIDE.md) |
| Removed pages | [REMOVED_PAGES.md](./REMOVED_PAGES.md) |
| Native systemd & remote deploy notes | [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) |
| Upstream sync notes (**do not** merge the full repo) | [docs/upstream-sync/README.md](./docs/upstream-sync/README.md) |

---

## Installation

Choose whichever way fits your environment. None of them require PostgreSQL or an external Redis instance.

### Option 1: Prebuilt Binary (Fastest)

Download the archive for your OS and architecture (Linux, macOS, Windows) from [GitHub Releases](https://github.com/githubliangliang/sub2api/releases), extract it, and run:

```bash
# Start directly — launches the web setup wizard on first run
./sub2api
```

- Open `http://localhost:8080` and complete the setup wizard (choose your SQLite file path and keep external Redis disabled to use the built-in one).
- Prefer terminal setup? Run `./sub2api --setup`.
- Want to store `config.yaml` and the SQLite file in a specific folder? Set `DATA_DIR`:

```bash
DATA_DIR=./data ./sub2api
```

### Option 2: Docker (Single Container)

Run the published image directly with a single persistent volume for SQLite data and config:

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

Or use **Docker Compose** (builds from the local repository):

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api/deploy

cp .env.sqlite.example .env.sqlite
# Recommended: set ADMIN_EMAIL, ADMIN_PASSWORD, and JWT_SECRET (openssl rand -hex 32)

docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite logs -f
```

Common Docker Compose commands:

```bash
# Stop
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite down

# Rebuild after updating code
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# Status
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite ps
```

- Access: `http://localhost:8080` (port controlled by `SERVER_PORT` in `.env.sqlite`).
- Data directory: `deploy/data/` stores `sub2api.db`, `config.yaml`, and `logs/`.
- Key environment variables (full template in `deploy/.env.sqlite.example`):

| Variable | Description |
|----------|-------------|
| `RUN_MODE` | `standard` (default, recommended) or `simple`; see [`run_mode`](#run_mode-standard-recommended) |
| `DATABASE_PATH` | SQLite database path, default `/app/data/sub2api.db` (`DATABASE_DRIVER` is ignored) |
| `REDIS_ENABLED` | `false` enables the in-process embedded Redis |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | Admin credentials created by `AUTO_SETUP` (leave empty to auto-generate and check startup logs) |
| `JWT_SECRET` | Session signing key; keep it fixed so container recreations do not log everyone out |
| `TOTP_ENCRYPTION_KEY` | Fixed 64-char hex key required if using 2FA; otherwise restarts invalidate existing 2FA bindings |
| `SERVER_PORT` / `BIND_HOST` | Host port mapping (default `8080`) |

> Note: `deploy/docker-compose.yml` and `deploy/docker-compose.local.yml` are upstream leftovers. This edition always opens SQLite only; use `deploy/docker-compose.sqlite.yml` instead.

### Option 3: Build from Source

Requires Go `1.27.4` (see `backend/go.mod`), Node.js `20+` (CI uses Node 24), and **pnpm** (do not use npm).

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api

# 1. Build frontend
cd frontend
pnpm install
pnpm run build
cd ..

# 2. Build single binary with embedded frontend (pure Go, no CGO)
cd backend
CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o ../sub2api ./cmd/server
cd ..

# 3. Run
DATA_DIR=./data ./sub2api
```

> Cross-compiling for another platform? Prefix `go build` with `GOOS=linux GOARCH=amd64` (or `arm64`, `darwin`, `windows`).

### Configuration File & Secrets

If you prefer configuring via `config.yaml` before first startup instead of the setup wizard, start from `deploy/config.personal.sqlite.yaml` and copy it to `$DATA_DIR/config.yaml`.

Required fields to customize:

```yaml
database:
  driver: "sqlite"
  path: "./data/sub2api.db"

redis:
  enabled: false              # Use embedded Redis (single instance)

jwt:
  secret: "generate-with-openssl-rand-hex-32"

totp:
  encryption_key: "generate-with-openssl-rand-hex-32"  # Must be 64 hex chars (32 bytes)

default:
  admin_email: "you@example.com"
  admin_password: "a-strong-password"
  user_balance: 100000        # Must be > 0 in standard mode, or API calls return 403
```

> **Keep `totp.encryption_key` fixed**: If left empty, a new random key is generated on every restart, immediately invalidating all existing 2FA bindings. It must be 32 bytes encoded as 64 hex characters (an invalid format prevents startup).

Running continuously as a Linux background service? Sample systemd unit (`deploy/sub2api-sqlite.service`) and helper scripts (`deploy/deploy-remote.sh`) are documented in [deploy/START_NATIVE.md](./deploy/START_NATIVE.md).

---

## `run_mode`: `standard` (Recommended)

Valid values are `standard` and `simple` (any other value is normalized to `standard` by `config.NormalizeRunMode`). The default is `standard`.

| Behavior | `standard` (recommended) | `simple` |
|----------|--------------------------|----------|
| Balance / subscription check | Enabled: balance ≤ 0 or below `billing.minimum_balance_reserve` (default `0.000001`) → `403 INSUFFICIENT_BALANCE` | Skipped |
| Usage & billing | Writes `usage_logs` and deducts actual token cost from balance | Writes `usage_logs` only (cost is logged, balance is unchanged) |
| Group scheduling | Isolated by group: bound keys use only that group's accounts; unbound keys use only accounts **not in any group** | Ignores groups; schedules across all accounts for the platform |
| Startup side effects | None | Auto-creates default groups (`anthropic-default`, `openai-default`, etc.) and bumps admin concurrency from 5 to 30 once |
| Sidebar & routes | Full UI; hide entries individually via `/admin/menu` | Hides Users, Groups, Channels, Subscriptions, Audit Logs, and redirects several admin/user pages |
| `/v1/sub2api/billing` | Available | 404 |

### Why `standard` is recommended

- **Groups are the only way to isolate accounts and route keys** (model mappings, group limits, key-to-group bindings). In `simple` mode, group settings can still be edited in the UI but are ignored at scheduling time.
- Usage statistics, audit logs, and user management stay visible. If the sidebar feels crowded, hide individual items via `/admin/menu` (`hidden_menu_keys`) instead of downgrading to `simple`.

### Set your initial balance in `standard` mode

In `standard` mode, API calls require a positive balance:

- **Via `config.yaml` before first startup**: set `default.user_balance` (e.g., `100000`). If you open public registration later, set it back to `0` and top up your own account manually.
- **Via Web Setup or Docker `AUTO_SETUP`** (admin balance defaults to `0`): log in, open **Admin → Users** (`/admin/users`), and add balance to your account (`POST /admin/users/:id/balance`).

---

## Backup & Upgrade

- **Offline file backup**: stop the process and copy your data directory (`sub2api.db`, `sub2api.db-wal`, `sub2api.db-shm`, and `config.yaml`). Always include `config.yaml` because losing `jwt.secret` and `totp.encryption_key` invalidates all sessions and 2FA bindings.
- **Online S3 backup**: once S3 is configured in Admin, backups run SQLite `VACUUM INTO` to create consistent `*.db.gz` snapshots. Restoring replaces the live `.db` and removes stale `-wal`/`-shm` files.
- **Upgrade**: replace the `sub2api` binary (or pull/rebuild the Docker image) and restart. SQLite migrations apply automatically on startup.

---

## Troubleshooting

| Symptom | Cause & Fix |
|---------|-------------|
| API calls return `403 INSUFFICIENT_BALANCE` | Balance is `0` in `standard` mode. Add balance in `/admin/users` or set `default.user_balance` before first startup |
| No available account found during scheduling | In `standard` mode, keys bound to a group only use that group's accounts, while unbound keys only use accounts **outside all groups**. Align key and account group assignments |
| Changing `admin_password` in `config.yaml` has no effect | `default.admin_*` is only read when the database has 0 users. Update the password from the UI profile page or directly in the `users` table |
| All 2FA bindings stop working after restart | `totp.encryption_key` was not set, causing a random key per startup. Set a fixed 64-char hex key and back it up with `config.yaml` |
| Startup exits with `default.admin_password must be changed from the example value` | Replace the template placeholder `CHANGE_ME_STRONG_PASSWORD` with a real password (8–128 chars) |
| Startup reports a migration checksum mismatch | An already-applied SQL migration was modified. Add a new migration instead, or delete `*.db` if starting fresh |
| Login returns `503` | An older SQLite database is missing auxiliary tables (e.g. `user_allowed_groups`). Check logs for the missing table name |

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.27.4, Gin, Ent, `modernc.org/sqlite` (pure Go, no CGO) |
| Frontend | Vue 3, Vite, TailwindCSS (pnpm) |
| Database | **SQLite only** (`backend/migrations/*.sql` in SQLite dialect) |
| Cache | Optional Redis; embedded in-process Redis when `redis.enabled: false` |
| Distribution | Single embedded binary (`-tags embed`) or single Docker container |

---

## FAQ

**Is this the SQLite version of Sub2API?**  
Yes. [githubliangliang/sub2api](https://github.com/githubliangliang/sub2api) is a **SQLite-only** edition of Sub2API that stores all data in a local SQLite database file.

**Does it support PostgreSQL?**  
No. All migrations use the SQLite dialect and PostgreSQL is not supported.

**Do I need to install Redis?**  
No. Set `redis.enabled: false` (or `REDIS_ENABLED=false`) to use the built-in embedded Redis. An external Redis is optional.

---

## License & Links

- Repository: https://github.com/githubliangliang/sub2api
- Releases: https://github.com/githubliangliang/sub2api/releases
- License: [LICENSE](./LICENSE) (LGPLv3)
- AI index: [llms.txt](./llms.txt)
