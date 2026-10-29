"""Serve isolated UI fixtures and the current Vite app for Terra acceptance."""
import copy
import datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import threading
from urllib.parse import urlparse

out = Path(__file__).resolve().parent
root = out.parents[3]
now = datetime.datetime.now(datetime.timezone.utc).isoformat()
user = {"id": 1, "email": "terra@example.invalid", "username": "Terra QA", "role": "admin", "status": "active", "balance": 100, "concurrency": 2, "run_mode": "simple"}
group = {"id": 2, "name": "OpenAI Terra QA", "platform": "openai", "status": "active", "subscription_type": "standard", "rate_multiplier": 1}
account = {"id": 101, "name": "Terra QA account", "platform": "openai", "type": "apikey", "status": "active", "schedulable": True, "concurrency": 2, "priority": 1, "credentials": {"api_key": "synthetic-fixture", "base_url": "https://fixture.invalid/v1", "model_mapping": {}}, "extra": {}, "group_ids": [2], "groups": [group], "account_groups": [], "created_at": now, "updated_at": now, "auto_pause_on_expired": False}
key = {"id": 201, "name": "Terra QA key", "key": "sk-terra-synthetic-fixture", "user_id": 1, "group_id": 2, "group": group, "status": "active", "quota": 0, "quota_used": 0, "created_at": now, "updated_at": now, "ip_whitelist": [], "ip_blacklist": []}
requests = []
lock = threading.Lock()


def page(items):
    return {"items": items, "total": len(items), "page": 1, "page_size": 20, "pages": 1}


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def reply(self, data):
        body = json.dumps({"code": 0, "message": "ok", "data": data}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = urlparse(self.path).path
        with lock:
            requests.append({"method": "GET", "path": path})
            (out / "requests.json").write_text(json.dumps(requests, indent=2))
        if path == "/api/v1/settings/public":
            data = {"site_name": "Terra local QA", "run_mode": "simple", "registration_enabled": False, "email_verify_enabled": False, "hidden_menu_keys": [], "custom_menu_items": [], "payment_enabled": False, "purchase_subscription_enabled": False, "api_base_url": "https://gateway.example/v1"}
        elif path == "/api/v1/auth/me":
            data = user
        elif path in ["/setup/status", "/api/v1/setup/status"]:
            data = {"needs_setup": False, "installed": True}
        elif path == "/api/v1/admin/accounts":
            data = page([account])
        elif path == "/api/v1/admin/accounts/101":
            data = account
        elif path == "/api/v1/keys":
            data = page([key])
        elif path.endswith("/groups/all") or path.endswith("/groups/available"):
            data = [group]
        elif path.endswith("/groups/rates"):
            data = {"2": 1}
        elif path.endswith("/models"):
            data = [{"id": name, "display_name": name, "type": "model"} for name in ["gpt-6.1-terra", "gpt-6.1-sol", "gpt-5.4"]]
        elif path.endswith(("/proxies/all", "/tls-fingerprint-profiles")):
            data = []
        elif path.endswith(("/proxies", "/announcements")):
            data = page([])
        elif path == "/api/v1/qa/state":
            data = {"account": account, "key": key, "requests": requests}
        else:
            data = {}
        self.reply(data)

    def mutation(self):
        path = urlparse(self.path).path
        body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0))) or b"{}")
        with lock:
            requests.append({"method": self.command, "path": path, "body": copy.deepcopy(body)})
            (out / "requests.json").write_text(json.dumps(requests, indent=2))
        if path == "/api/v1/admin/accounts/101":
            account.update(body)
            self.reply(account)
        elif path.endswith("/api-keys-usage"):
            self.reply({"stats": {}})
        else:
            self.reply({})

    do_POST = mutation
    do_PUT = mutation
    do_PATCH = mutation


server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
with socket.socket() as probe:
    probe.bind(("127.0.0.1", 0))
    vite_port = probe.getsockname()[1]
env = os.environ.copy()
env.update(VITE_DEV_PROXY_TARGET="http://127.0.0.1:" + str(server.server_port), VITE_DEV_PORT=str(vite_port))
log = (out / "vite.log").open("w")
vite = subprocess.Popen(["pnpm", "run", "dev", "--host", "127.0.0.1", "--strictPort"], cwd=root / "frontend", env=env, stdout=log, stderr=log, start_new_session=True)
state = {"api_port": server.server_port, "vite_port": vite_port, "vite_pid": vite.pid, "fixture_pid": os.getpid()}
(out / "server.json").write_text(json.dumps(state, indent=2))


def stop(*args):
    raise KeyboardInterrupt


signal.signal(signal.SIGTERM, stop)
print(json.dumps(state), flush=True)
try:
    server.serve_forever()
finally:
    server.server_close()
    os.killpg(vite.pid, signal.SIGTERM)
    try:
        vite.wait(timeout=5)
    except subprocess.TimeoutExpired:
        os.killpg(vite.pid, signal.SIGKILL)
        vite.wait()
    log.close()
    (out / "server-stopped.json").write_text(json.dumps({"stopped": True, "vite_exit_code": vite.returncode}, indent=2))
