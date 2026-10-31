"""Local UI acceptance API; all accounts and credentials are synthetic."""

import copy
import datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import threading
from urllib.parse import urlparse


OUT = Path(__file__).resolve().parent / "browser"
OUT.mkdir(exist_ok=True)
lock = threading.Lock()
requests = []
fail_next = False
user = {"id": 1, "email": "qa@example.invalid", "username": "Local QA", "role": "admin", "status": "active", "balance": 10, "concurrency": 2, "run_mode": "simple"}


def stamp():
    return datetime.datetime.now(datetime.timezone.utc).isoformat()


def catalog(model, state="fresh"):
    now = datetime.datetime.now(datetime.timezone.utc)
    return {"state": state, "endpoint": "https://cli-chat-proxy.grok.com/v1/models",
            "fetched_at": now.isoformat(), "expires_at": (now + datetime.timedelta(hours=6)).isoformat(),
            "models": [model], "last_attempt": {"status": "success", "started_at": now.isoformat(), "finished_at": now.isoformat()},
            "mappings": [{"requested": "friendly", "upstream": model, "status": "listed"},
                         {"requested": "missing", "upstream": "grok-4.5", "status": "not_listed"}]}


def account(id, name):
    return {"id": id, "name": name, "notes": "", "platform": "grok", "type": "oauth", "status": "active",
            "schedulable": True, "concurrency": 1, "priority": 1, "rate_multiplier": 1,
            "credentials": {"model_mapping": {"friendly": "grok-4.6", "missing": "grok-4.5"}},
            "credentials_status": {"has_access_token": True, "has_refresh_token": True},
            "extra": {}, "proxy_id": None, "group_ids": [], "groups": [], "account_groups": [],
            "created_at": stamp(), "updated_at": stamp(), "auto_pause_on_expired": False,
            "expires_at": None, "grok_model_diagnostics": catalog("grok-4.6")}


accounts = {101: account(101, "Grok diagnostics A"), 102: account(102, "Grok diagnostics B")}


def page(items):
    return {"items": items, "total": len(items), "page": 1, "page_size": 20, "pages": 1}


def detail(item):
    result = copy.deepcopy(item)
    snapshot = result["grok_model_diagnostics"]
    listed = set(snapshot["models"])
    snapshot["mappings"] = [{"requested": name, "upstream": target,
                             "status": "listed" if target in listed else "not_listed"}
                            for name, target in result["credentials"].get("model_mapping", {}).items()]
    return result


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def respond(self, data, status=200):
        body = json.dumps({"code": 0 if status == 200 else status, "message": "fixture" if status == 200 else "Synthetic sync failure", "data": data}).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def payload(self):
        size = int(self.headers.get("Content-Length", "0"))
        return json.loads(self.rfile.read(size)) if size else {}

    def record(self, method, path, body=None):
        with lock:
            requests.append({"method": method, "path": path, "body": body})
            (OUT / "requests.json").write_text(json.dumps(requests, ensure_ascii=False, indent=2))

    def do_GET(self):
        path = urlparse(self.path).path
        self.record("GET", path)
        if path == "/api/v1/settings/public":
            data = {"site_name": "Local compatibility acceptance", "registration_enabled": False, "email_verify_enabled": False,
                    "ops_monitoring_enabled": True, "payment_enabled": False, "purchase_subscription_enabled": False,
                    "hidden_menu_keys": [], "custom_menu_items": [], "run_mode": "simple"}
        elif path == "/api/v1/auth/me":
            data = user
        elif path in ["/setup/status", "/api/v1/setup/status"]:
            data = {"needs_setup": False, "installed": True}
        elif path == "/api/v1/admin/accounts":
            data = page([{key: value for key, value in item.items() if key != "grok_model_diagnostics"} for item in accounts.values()])
        elif path.startswith("/api/v1/admin/accounts/") and path.split("/")[-1].isdigit():
            data = detail(accounts[int(path.split("/")[-1])])
        elif path.endswith("/groups/all") or path.endswith("/proxies/all") or path.endswith("/tls-fingerprint-profiles"):
            data = []
        elif path.endswith("/groups") or path.endswith("/proxies") or path.endswith("/announcements"):
            data = page([])
        elif path.endswith("/admin/settings"):
            data = {"account_quota_notify_enabled": False}
        elif "web-search" in path or "websearch" in path:
            data = {"enabled": False, "providers": []}
        elif path.endswith("/usage"):
            data = {"five_hour": None, "seven_day": None}
        elif path.endswith("/models"):
            data = [{"id": "grok-4.6", "type": "model", "display_name": "Grok"}]
        else:
            data = {}
        self.respond(data)

    def do_POST(self):
        global fail_next
        path = urlparse(self.path).path
        body = self.payload()
        self.record("POST", path, body)
        if path == "/__fixture/fail-next-sync":
            fail_next = True
            return self.respond({})
        if path.endswith("/models/sync-upstream"):
            id = int(path.split("/")[-3])
            if fail_next:
                fail_next = False
                accounts[id]["grok_model_diagnostics"]["last_attempt"] = {"status": "failed", "error_code": "transport", "started_at": stamp(), "finished_at": stamp()}
                return self.respond({}, 502)
            accounts[id]["grok_model_diagnostics"] = catalog("grok-4.6")
            return self.respond({"models": ["grok-4.6"], "discovery_source": "upstream", "grok_model_diagnostics": accounts[id]["grok_model_diagnostics"]})
        if path.endswith("/usage/batch"):
            return self.respond({"usage": {}, "errors": {}})
        self.respond({})

    def do_PUT(self):
        path = urlparse(self.path).path
        body = self.payload()
        self.record("PUT", path, body)
        if path.startswith("/api/v1/admin/accounts/"):
            id = int(path.split("/")[-1])
            accounts[id].update({key: value for key, value in body.items() if key not in ["credentials", "extra"]})
            accounts[id]["credentials"].update(body.get("credentials", {}))
            accounts[id]["extra"].update(body.get("extra", {}))
            accounts[id]["updated_at"] = stamp()
            self.respond(detail(accounts[id]))
        else:
            self.respond({})


print("Synthetic UI API on 127.0.0.1:39971", flush=True)
ThreadingHTTPServer(("127.0.0.1", 39971), Handler).serve_forever()
