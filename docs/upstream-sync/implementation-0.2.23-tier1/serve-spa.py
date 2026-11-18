"""Serve only the local build for browser acceptance with synthetic API fixtures."""
import os
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3] / "backend/internal/web/dist"

class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(ROOT), **kwargs)

    def do_GET(self):
        if not (ROOT / self.path.split("?", 1)[0].lstrip("/")).is_file():
            self.path = "/index.html"
        super().do_GET()

    def log_message(self, *args):
        pass

server = ThreadingHTTPServer(("127.0.0.1", 18783), Handler)
print(f"pid={os.getpid()} url=http://127.0.0.1:18783", flush=True)
server.serve_forever()
