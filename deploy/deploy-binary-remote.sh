#!/usr/bin/env bash
# 复制仓库根目录的 sub2api 二进制，原子替换后重启服务并检查健康状态。
# 默认目标：admin@18.226.163.236:/home/admin/sub2api/sub2api
# 用法：deploy/deploy-binary-remote.sh [-i SSH_KEY] [--dry-run]
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/../sub2api"
REMOTE="admin@18.226.163.236"
REMOTE_DIR="/home/admin/sub2api"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/test.pem}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-60}"
DRY_RUN=0

die() { printf '[X] %s\n' "$*" >&2; exit 1; }
usage() {
  cat <<'EOF'
用法: deploy/deploy-binary-remote.sh [选项]

复制本地 sub2api 到 admin@18.226.163.236:/home/admin/sub2api/sub2api。
原子替换后重启 systemd 服务 sub2api，并检查 127.0.0.1:8080/health。
发布失败恢复 sub2api.prev；HEALTH_TIMEOUT 可设置健康检查秒数（默认 60）。

  -i, --key FILE    SSH 私钥（默认 ~/.ssh/test.pem，可用 SSH_KEY 覆盖）
      --dry-run     打印复制计划
  -h, --help        显示帮助
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -i|--key)
      [ "$#" -ge 2 ] && [ -n "$2" ] || die "选项 $1 需要一个路径"
      SSH_KEY="$2"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "未知参数：$1（用 --help 查看用法）" ;;
  esac
done
SSH_KEY="${SSH_KEY/#\~/$HOME}"

if [ "$DRY_RUN" = 1 ]; then
  printf '本地二进制：%s\nSSH 密钥：%s\n目标：%s:%s/sub2api\n' \
    "$BINARY" "$SSH_KEY" "$REMOTE" "$REMOTE_DIR"
  printf '步骤：上传二进制 → 备份旧二进制 → 原子替换 → 重启 sub2api → /health 检查\n'
  exit 0
fi

[ -f "$BINARY" ] && [ -s "$BINARY" ] || die "本地二进制不存在或为空：$BINARY"
[ -r "$SSH_KEY" ] || die "SSH 私钥不可读：$SSH_KEY"
[[ "$HEALTH_TIMEOUT" =~ ^[1-9][0-9]*$ ]] || die "HEALTH_TIMEOUT 必须为正整数"
SSH_OPTS=(-o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=30
          -o StrictHostKeyChecking=accept-new -i "$SSH_KEY" -o IdentitiesOnly=yes)
REMOTE_TMP=""
cleanup() {
  if [ -n "$REMOTE_TMP" ]; then
    ssh "${SSH_OPTS[@]}" "$REMOTE" "rm -f -- '$REMOTE_TMP'" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

UPLOAD_PATH="$(ssh "${SSH_OPTS[@]}" "$REMOTE" \
  "test -f '$REMOTE_DIR/sub2api' && sudo -n true && command -v curl >/dev/null && mktemp '$REMOTE_DIR/.sub2api-upload.XXXXXX'")"
[[ "$UPLOAD_PATH" =~ ^/home/admin/sub2api/\.sub2api-upload\.[[:alnum:]]{6}$ ]] \
  || die "远端未返回有效的临时文件路径"
REMOTE_TMP="$UPLOAD_PATH"

printf '复制 %s → %s:%s/sub2api\n' "$BINARY" "$REMOTE" "$REMOTE_DIR"
scp -C "${SSH_OPTS[@]}" "$BINARY" "$REMOTE:$REMOTE_TMP"
ssh "${SSH_OPTS[@]}" "$REMOTE" \
  "bash -s -- '$REMOTE_TMP' '$REMOTE_DIR' '$HEALTH_TIMEOUT'" <<'REMOTE_SCRIPT'
set -Eeuo pipefail
upload_path="$1"
app_dir="$2"
health_timeout="$3"
app_bin="$app_dir/sub2api"
prev_bin="$app_dir/sub2api.prev"

chmod 0755 "$upload_path"
"$upload_path" --version
cp -p -- "$app_bin" "$prev_bin"
mv -fT -- "$upload_path" "$app_bin"

rollback() {
  printf '[X] 发布失败，恢复旧二进制并重启服务\n' >&2
  cp -p -- "$prev_bin" "$app_bin.new"
  mv -fT -- "$app_bin.new" "$app_bin"
  sudo -n systemctl restart sub2api || true
  exit 1
}
if ! sudo -n systemctl restart sub2api; then rollback; fi
printf '已重启 sub2api，等待 /health 就绪（最多 %ss）\n' "$health_timeout"
health_deadline=$((SECONDS + health_timeout))
while [ "$SECONDS" -lt "$health_deadline" ]; do
  if curl --noproxy '*' -fsS --max-time 3 http://127.0.0.1:8080/health >/dev/null 2>&1 \
     && systemctl is-active --quiet sub2api; then
    exit 0
  fi
  sleep 2
done
rollback
REMOTE_SCRIPT
REMOTE_TMP=""
printf '[OK] 发布完成：%s:%s/sub2api（服务已重启，健康检查通过）\n' "$REMOTE" "$REMOTE_DIR"
