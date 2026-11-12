from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

root = Path(__file__).resolve().parents[3] / 'backend/internal/web/dist'

class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(root), **kwargs)

    def do_GET(self):
        if not (root / self.path.split('?', 1)[0].lstrip('/')).is_file():
            self.path = '/index.html'
        super().do_GET()

    def log_message(self, *args):
        pass

ThreadingHTTPServer(('127.0.0.1', 18763), Handler).serve_forever()
