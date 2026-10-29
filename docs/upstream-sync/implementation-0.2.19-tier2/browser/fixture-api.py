"""Local UI fixture; it never contacts a business instance or upstream provider."""
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse, parse_qs
import datetime
import json
import os
import signal
import socket
import subprocess
import threading

out = Path(__file__).resolve().parent
root = out.parents[3]
now = datetime.datetime.now(datetime.timezone.utc).isoformat()
user = {"id": 1, "email": "qa@example.invalid", "username": "Tier2 QA", "role": "admin", "status": "active", "balance": 100, "concurrency": 2, "run_mode": "standard"}
groups = [{"id": i, "name": name, "platform": platform, "status": "active"} for i, name, platform in [(1,"Claude QA","anthropic"),(2,"OpenAI QA","openai"),(3,"Gemini QA","gemini")]]
channel = {"id": 11, "name": "Tier2 search QA", "description": "Local fixture", "status": "active", "group_ids": [1,2], "model_pricing": [], "model_mapping": {}, "features_config": {"web_search_emulation": {"anthropic": True, "openai": True, "future": True}, "unrelated": {"keep": True}}}
accounts = [{"id": i, "name": name, "platform": "gemini", "type": kind, "status": "active", "schedulable": True, "concurrency": 2, "priority": 1, "credentials": credentials, "extra": {}, "group_ids": [3], "groups": [groups[2]], "account_groups": [], "created_at": now, "updated_at": now, "auto_pause_on_expired": False} for i, name, kind, credentials in [(101,"Gemini ordinary QA","apikey",{"api_key":"synthetic-fixture"}),(102,"Gemini legacy QA","oauth",{"oauth_type":"google_one"})]]
requests = []
lock = threading.Lock()
search_enabled = True
def page(items): return {"items": items, "total": len(items), "page": 1, "page_size": 20, "pages": 1}

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args): pass
    def reply(self, data):
        body = json.dumps({"code": 0, "message": "ok", "data": data}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def do_GET(self):
        parsed = urlparse(self.path); path = parsed.path
        with lock:
            requests.append({"method":"GET", "path":path, "query":parse_qs(parsed.query)})
            (out / "requests.json").write_text(json.dumps(requests, indent=2))
        if path == "/api/v1/settings/public":
            data = {"site_name":"Tier2 local QA", "registration_enabled":False, "email_verify_enabled":False, "hidden_menu_keys":[], "custom_menu_items":[], "payment_enabled":False, "purchase_subscription_enabled":False}
        elif path == "/api/v1/auth/me": data = user
        elif path in ["/setup/status","/api/v1/setup/status"]: data = {"needs_setup":False,"installed":True}
        elif path == "/api/v1/admin/channels": data = page([channel])
        elif path == "/api/v1/admin/channels/11": data = channel
        elif path == "/api/v1/admin/groups/all": data = groups
        elif path == "/api/v1/admin/accounts": data = page(accounts)
        elif path.endswith("/102/models"):
            data = [{"id":i,"display_name":i,"type":"model"} for i in ["gemini-2.5-flash","gemini-2.5-pro","gemini-2.0-flash"]]
        elif path.endswith("/101/models"):
            data = [{"id":i,"display_name":i,"type":"model"} for i in ["gemini-3.1-flash-image","gemini-3.5-flash","gemini-3-flash-preview","gemini-3.1-pro-preview","gemini-3.1-pro-preview-customtools"]]
        elif path == "/api/v1/admin/accounts/101": data = accounts[0]
        elif path == "/api/v1/admin/accounts/102": data = accounts[1]
        elif path.endswith("/web-search-emulation"): data = {"enabled":search_enabled,"providers":[{"type":"brave","api_key_configured":True}]}
        elif path.endswith(("/proxies/all","/tls-fingerprint-profiles","/available")): data = []
        elif path.endswith(("/keys","/proxies","/announcements")): data = page([])
        elif path == "/api/v1/qa/state": data = {"channel":channel,"requests":requests,"search_enabled":search_enabled}
        else: data = {}
        self.reply(data)
    def mutation(self):
        global search_enabled
        path = urlparse(self.path).path
        body = json.loads(self.rfile.read(int(self.headers.get("Content-Length",0))) or b"{}")
        with lock:
            requests.append({"method":self.command,"path":path,"features_config":body.get("features_config"),"model_id":body.get("model_id")})
            (out / "requests.json").write_text(json.dumps(requests, indent=2))
        if path == "/api/v1/admin/channels/11": channel.update(body); self.reply(channel)
        elif path == "/api/v1/qa/search": search_enabled = body["enabled"]; self.reply({"enabled":search_enabled})
        elif path.endswith("/test"):
            payload = "data: "+json.dumps({"type":"test_complete","success":True,"model":body.get("model_id")})+"\n\n"
            self.send_response(200); self.send_header("Content-Type","text/event-stream"); self.end_headers(); self.wfile.write(payload.encode())
        else: self.reply({})
    do_POST = mutation
    do_PUT = mutation
    do_PATCH = mutation

server = ThreadingHTTPServer(("127.0.0.1",0), Handler)
with socket.socket() as probe:
    probe.bind(("127.0.0.1",0)); vite_port = probe.getsockname()[1]
env = os.environ.copy()
env.update(VITE_DEV_PROXY_TARGET="http://127.0.0.1:"+str(server.server_port),VITE_DEV_PORT=str(vite_port))
log = (out / "vite.log").open("w")
vite = subprocess.Popen(["pnpm","run","dev","--host","127.0.0.1","--strictPort"],cwd=root/"frontend",env=env,stdout=log,stderr=log,start_new_session=True)
state = {"api_port":server.server_port,"vite_port":vite_port,"vite_pid":vite.pid,"fixture_pid":os.getpid()}
(out / "server.json").write_text(json.dumps(state,indent=2))
def stop(*args): raise KeyboardInterrupt
signal.signal(signal.SIGTERM,stop)
print(json.dumps(state),flush=True)
try: server.serve_forever()
finally:
    server.server_close(); os.killpg(vite.pid, signal.SIGTERM)
    try: vite.wait(timeout=5)
    except subprocess.TimeoutExpired: vite.kill(); vite.wait()
    log.close(); (out/"server-stopped.json").write_text(json.dumps({"stopped":True,"vite_exit_code":vite.returncode},indent=2))
