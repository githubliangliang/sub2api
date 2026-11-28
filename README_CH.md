<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API（SQLite 版）

[![Go](https://img.shields.io/badge/Go-1.27.4-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D.svg)](https://vuejs.org/)
[![SQLite](https://img.shields.io/badge/SQLite-Only-003B57.svg)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**Sub2API SQLite 版** —— 轻量、**纯 SQLite** 的 Sub2API 构建版本（[githubliangliang/sub2api](https://github.com/githubliangliang/sub2api)）。  
全部数据存储在单个本地 SQLite 文件中（`modernc.org/sqlite`），可选进程内嵌入式 Redis，无需安装 PostgreSQL 或外部 Redis。

仓库：**https://github.com/githubliangliang/sub2api** · 发布页：**https://github.com/githubliangliang/sub2api/releases** · [llms.txt](./llms.txt)

[English](README.md) | [中文](README_CH.md) | [日本語](README_JA.md)

</div>

## 关于 Sub2API SQLite 版

本仓库维护 Sub2API 的 **SQLite 专用版本**，面向零外部数据库依赖的单节点使用场景：

| 特性 | 说明 |
|------|------|
| **纯 SQLite** | 使用纯 Go SQLite 驱动（`modernc.org/sqlite`，无需 CGO）。运行时只打开本地 SQLite 文件，`database.driver` / `DATABASE_DRIVER` 即使写成 `postgres` 也会被忽略 |
| **SQLite 安装向导** | Web、CLI（`--setup`）与 `AUTO_SETUP` 只收集 SQLite 文件路径，不再需要 PostgreSQL 的 host / port / user / ssl 配置 |
| **原生 SQLite 备份** | 管理页备份使用 `VACUUM INTO` 生成一致性 `*.db.gz` 快照；镜像不再携带 `pg_dump` / `psql` |
| **可选嵌入式 Redis** | 设置 `redis.enabled: false` / `REDIS_ENABLED=false` 即可使用进程内嵌入式 Redis（单实例）；Web 安装向导第 2 步默认关闭「使用外部 Redis」 |
| **菜单可隐藏** | 通过 `hidden_menu_keys` 或管理页「菜单管理」（`/admin/menu`）逐项隐藏侧边栏入口 |
| **用量落库加固** | 补齐 SQLite 下 `usage_logs` 幂等索引与 `usage_billing_dedup` 表，避免请求成功但未写入用量 |
| **支付页清理** | 已移除前端对已下线接口 `/admin/payment/providers` 的请求 |

> **注意**：本版本**不支持** PostgreSQL，也不适合多实例横向扩展。`backend/migrations/*.sql` 全部为 SQLite 方言；请勿删除已应用的 `SELECT 1` no-op 迁移（校验和不可变）。

延伸文档：

| 主题 | 文档 |
|------|------|
| SQLite 重构说明 | [REFACTOR.md](./REFACTOR.md) |
| 开发约定与踩坑 | [DEV_GUIDE.md](./DEV_GUIDE.md) |
| 已下线页面 | [REMOVED_PAGES.md](./REMOVED_PAGES.md) |
| 原生 systemd 与远程部署说明 | [deploy/START_NATIVE.md](./deploy/START_NATIVE.md) |
| 上游同步说明（**不要**整仓 merge） | [docs/upstream-sync/README.md](./docs/upstream-sync/README.md) |

---

## 安装方式

以下任选一种即可运行，均不依赖 PostgreSQL 或外部 Redis。

### 方式一：直接下载二进制（最快）

前往 [GitHub Releases](https://github.com/githubliangliang/sub2api/releases) 下载对应系统与架构（Linux / macOS / Windows）的压缩包，解压后直接运行：

```bash
# 直接启动，首次运行自动进入 Web 安装向导
./sub2api
```

- 浏览器打开 `http://localhost:8080` 按向导完成初始化（填写 SQLite 文件路径，保持外部 Redis 关闭即使用嵌入式 Redis）。
- 偏好终端交互安装？运行 `./sub2api --setup`。
- 需要将 `config.yaml` 和 SQLite 数据库放在指定目录？通过 `DATA_DIR` 指定：

```bash
DATA_DIR=./data ./sub2api
```

### 方式二：Docker 单容器

直接使用已发布镜像启动，只需挂载一个目录持久化 SQLite 数据与配置：

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

或使用 **Docker Compose**（基于本地仓库构建）：

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api/deploy

cp .env.sqlite.example .env.sqlite
# 建议设置 ADMIN_EMAIL、ADMIN_PASSWORD 与 JWT_SECRET（openssl rand -hex 32）

docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite logs -f
```

常用 Docker Compose 命令：

```bash
# 停止
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite down

# 更新代码后重建
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite up -d --build

# 查看状态
docker compose -f docker-compose.sqlite.yml --env-file .env.sqlite ps
```

- 访问地址：`http://localhost:8080`（端口由 `.env.sqlite` 中的 `SERVER_PORT` 控制）。
- 数据目录：`deploy/data/` 保存 `sub2api.db`、`config.yaml` 与 `logs/`。
- 常用环境变量（完整模板见 `deploy/.env.sqlite.example`）：

| 变量 | 说明 |
|------|------|
| `RUN_MODE` | `standard`（默认，推荐）或 `simple`；区别见 [`run_mode`](#run_mode推荐-standard) |
| `DATABASE_PATH` | SQLite 数据库路径，默认 `/app/data/sub2api.db`（`DATABASE_DRIVER` 会被忽略） |
| `REDIS_ENABLED` | `false` 表示启用进程内嵌入式 Redis |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | 首次 `AUTO_SETUP` 创建管理员（留空则自动生成并打印到启动日志） |
| `JWT_SECRET` | 会话签名密钥，建议固定以免重建容器后全部退出登录 |
| `TOTP_ENCRYPTION_KEY` | 启用 2FA 时必须固定的 64 位 hex 密钥，否则重启会作废已有绑定 |
| `SERVER_PORT` / `BIND_HOST` | 宿主机端口映射（默认 `8080`） |

> 提示：`deploy/docker-compose.yml` 与 `deploy/docker-compose.local.yml` 为上游遗留文件，本版本运行时仍只打开 SQLite，请使用 `deploy/docker-compose.sqlite.yml`。

### 方式三：从源码编译

需要 Go `1.27.4`（见 `backend/go.mod`）、Node.js `20+`（CI 使用 Node 24）以及 **pnpm**（不要用 npm）。

```bash
git clone https://github.com/githubliangliang/sub2api.git
cd sub2api

# 1. 编译前端
cd frontend
pnpm install
pnpm run build
cd ..

# 2. 编译内嵌前端的单二进制（纯 Go，无需 CGO）
cd backend
CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o ../sub2api ./cmd/server
cd ..

# 3. 启动
DATA_DIR=./data ./sub2api
```

> 需要交叉编译到其他平台时，在 `go build` 前加上 `GOOS=linux GOARCH=amd64`（或 `arm64`、`darwin`、`windows`）即可。

### 配置文件与密钥

如果希望在首次启动前通过 `config.yaml` 预先配置，可参考 `deploy/config.personal.sqlite.yaml`，将其复制到 `$DATA_DIR/config.yaml`。

必改关键项：

```yaml
database:
  driver: "sqlite"
  path: "./data/sub2api.db"

redis:
  enabled: false              # 使用进程内嵌入式 Redis（单实例）

jwt:
  secret: "使用 openssl rand -hex 32 生成"

totp:
  encryption_key: "使用 openssl rand -hex 32 生成"  # 必须为 64 位 hex（32 字节）

default:
  admin_email: "you@example.com"
  admin_password: "强密码"
  user_balance: 100000        # standard 模式必须 > 0，否则首次调 API 报 403
```

> **请固定 `totp.encryption_key`**：留空时程序每次重启都会随机生成新密钥，导致所有已绑定的 2FA 立即失效。格式必须为 32 字节的 64 位十六进制字符串（格式错误会导致启动失败）。

需要作为 Linux 后台服务长期运行时，systemd 单元文件（`deploy/sub2api-sqlite.service`）与发布脚本（`deploy/deploy-remote.sh`）的使用说明见 [deploy/START_NATIVE.md](./deploy/START_NATIVE.md)。

---

## `run_mode`：推荐 `standard`

合法值为 `standard` 与 `simple`（其他值会由 `config.NormalizeRunMode` 归一为 `standard`），默认为 `standard`。

| 行为 | `standard`（推荐） | `simple` |
|------|-------------------|----------|
| 余额 / 订阅校验 | 生效：余额 ≤ 0 或低于 `billing.minimum_balance_reserve`（默认 `0.000001`）→ `403 INSUFFICIENT_BALANCE` | 跳过 |
| 用量与扣费 | 写入 `usage_logs` 并按 token 成本实扣余额 | 仅写入 `usage_logs`（成本照算写入日志，不扣余额） |
| 分组调度 | 按分组隔离：绑了分组的 Key 只用该分组账号；未绑分组的 Key 只用**不属于任何分组**的账号 | 忽略分组，按平台调度全部可用账号 |
| 启动副作用 | 无 | 自动创建默认分组（`anthropic-default`、`openai-default` 等），并将并发仍为 5 的管理员一次性提至 30 |
| 侧边栏与页面 | 全量显示；不需要的入口可在 `/admin/menu` 逐项隐藏 | 隐藏用户、分组、渠道、订阅、审计日志，并重定向部分管理/用户页面 |
| `/v1/sub2api/billing` | 可用 | 404 |

### 为什么推荐使用 `standard`

- **分组是系统内唯一的账号隔离与定向路由方式**（分组模型映射、分组限额、Key 绑分组）。`simple` 模式下这些配置虽可在界面保存，但调度时会被完全忽略。
- 用量统计、审计日志与用户管理完整保留。若觉得菜单偏多，直接通过 `/admin/menu`（`hidden_menu_keys`）按需隐藏，无需降级到 `simple`。

### `standard` 模式设置初始余额

`standard` 模式调用 API 会校验余额：

- **首次启动前通过 `config.yaml`**：将 `default.user_balance` 设为大于 0 的数值（如 `100000`）。后续若开放公开注册，再改回 `0` 并单独给管理员充值。
- **通过 Web 向导或 Docker `AUTO_SETUP` 安装**（管理员默认余额为 `0`）：登录后前往 **管理页 → 用户管理**（`/admin/users`）给自己增加余额（接口 `POST /admin/users/:id/balance`）。

---

## 备份与升级

- **停机文件备份**：停止进程后拷贝整个数据目录（`sub2api.db`、`sub2api.db-wal`、`sub2api.db-shm` 及 `config.yaml`）。务必同时备份 `config.yaml`，丢失 `jwt.secret` 和 `totp.encryption_key` 会导致所有会话和 2FA 绑定失效。
- **在线 S3 备份**：在管理页配置 S3 后，备份会使用 SQLite `VACUUM INTO` 生成一致性 `*.db.gz` 快照；恢复时会自动替换线上 `.db` 并清理旧 `-wal`/`-shm`。
- **升级**：替换 `sub2api` 二进制文件（或拉取/重建 Docker 镜像）并重启，SQLite 数据库迁移会在启动时自动执行。

---

## 常见问题

| 现象 | 原因与处理 |
|------|------------|
| 调 API 报 `403 INSUFFICIENT_BALANCE` | `standard` 模式下余额为 `0`。在 `/admin/users` 增加余额，或在首次启动前设置 `default.user_balance` |
| 调度不到账号 / 提示无可用账号 | `standard` 模式按分组隔离：绑了分组的 Key 只用该分组账号，未绑分组的 Key 只用**不属于任何分组**的账号。请对齐 Key 与账号的分组归属 |
| 修改 `config.yaml` 里的 `admin_password` 后登录仍是旧密码 | `default.admin_*` 仅在数据库为 0 用户时用于建号，之后不再同步。请在 UI 个人资料页修改密码，或直接更新 `users` 表 |
| 重启后所有 2FA 失效 | 未配置 `totp.encryption_key`，导致每次启动随机生成密钥。请固定 64 位 hex 密钥并随 `config.yaml` 备份 |
| 启动报错 `default.admin_password must be changed from the example value` | 未修改模板中的 `CHANGE_ME_STRONG_PASSWORD`，请改成实际密码（8–128 字符） |
| 启动报迁移 checksum 不一致 | 修改了已应用的 SQL 迁移文件。请通过新增迁移文件处理；若是全新测试库可直接删掉 `*.db` 重建 |
| 登录返回 `503` | 旧库缺少辅助表（如 `user_allowed_groups`），请查看日志确认缺失的表名 |

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.27.4、Gin、Ent、`modernc.org/sqlite`（纯 Go，无需 CGO） |
| 前端 | Vue 3、Vite、TailwindCSS（pnpm） |
| 数据库 | **纯 SQLite**（`backend/migrations/*.sql` 为 SQLite 方言） |
| 缓存 | Redis 可选；`redis.enabled: false` 时使用进程内嵌入式 Redis |
| 分发 | 内嵌前端单二进制（`-tags embed`）或单 Docker 容器 |

---

## FAQ

**这是 SQLite 版 Sub2API 吗？**  
是。[githubliangliang/sub2api](https://github.com/githubliangliang/sub2api) 是 **纯 SQLite** 版本的 Sub2API，所有数据均保存在本地 SQLite 数据库文件中。

**支持 PostgreSQL 吗？**  
不支持。全部迁移脚本均为 SQLite 方言，运行时只连接 SQLite。

**必须安装 Redis 吗？**  
不需要。配置 `redis.enabled: false`（或 `REDIS_ENABLED=false`）即可启用进程内嵌入式 Redis，外部 Redis 为可选项。

---

## 仓库与许可证

- 仓库：https://github.com/githubliangliang/sub2api
- 发布页：https://github.com/githubliangliang/sub2api/releases
- 许可证：[LICENSE](./LICENSE)（LGPLv3）
- AI 索引入口：[llms.txt](./llms.txt)
