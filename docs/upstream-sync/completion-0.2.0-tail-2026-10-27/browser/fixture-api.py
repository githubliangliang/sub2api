"""Loopback fixtures for task 9.14; no business service or credentials are used."""

import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import threading
from urllib.parse import urlparse

parser = argparse.ArgumentParser()
parser.add_argument('--port', type=int, default=34917)
parser.add_argument('--frontend-port', type=int, default=34916)
parser.add_argument('--baseline-module', type=Path)
args = parser.parse_args()
out = Path(__file__).resolve().parent
requests = []
lock = threading.Lock()
created_at = '2026-10-27T00:00:00Z'
user = {'id': 7, 'email': 'browser@example.invalid', 'username': '个人用户',
        'role': 'user', 'status': 'active', 'balance': 0, 'concurrency': 1, 'run_mode': 'simple'}
svg = ('<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">'
       '<circle cx="12" cy="12" r="10" fill="#0d9488" onclick="window.__qaXss=true"/>'
       '<path d="M7 12l3 3 7-7" fill="none" stroke="white" stroke-width="2"/>'
       '</svg>')
markdown = '''## 使用须知

欢迎使用个人服务。**重要内容**与普通段落应清晰显示。

- 保留列表条目
- 保留[帮助链接](/legal/qa-policy)

| 项目 | 说明 |
| --- | --- |
| 文本 | 正常显示 |

`inline code`

<img src="/api/v1/qa/check.svg" alt="确认图标" width="32" height="32" onload="window.__qaXss=true">
<a href="javascript:window.__qaXss=true">说明链接</a>
'''
custom_markdown = '# 使用指南\n\n' + markdown + '''

## 嵌入内容

<iframe src="/api/v1/qa/embed" title="本地嵌入内容" style="width:100%;max-width:360px;height:120px;border:0" frameborder="0" allowfullscreen></iframe>
'''
announcements = [
    {'id': 21, 'title': '服务使用说明', 'content': markdown, 'notify_mode': 'popup',
     'created_at': created_at, 'updated_at': created_at, 'read_at': None},
    {'id': 22, 'title': '帮助与文档', 'content': markdown, 'notify_mode': 'silent',
     'created_at': created_at, 'updated_at': created_at, 'read_at': None},
]
settings = {
    'site_name': '个人网关', 'site_logo': '', 'registration_enabled': False,
    'email_verify_enabled': False, 'payment_enabled': False, 'purchase_subscription_enabled': False,
    'risk_control_enabled': False, 'backend_mode_enabled': False,
    'ops_monitoring_enabled': False, 'hidden_menu_keys': [],
    'login_agreement_documents': [{'id': 'qa-policy', 'title': '服务使用说明', 'content_md': markdown}],
    'login_agreement_updated_at': '2026-10-27',
    'custom_menu_items': [
        {'id': 'qa-guide', 'label': '使用指南', 'icon_svg': svg, 'url': 'md:qa-guide',
         'page_slug': 'qa-guide', 'visibility': 'user', 'sort_order': 0},
        {'id': 'qa-embed', 'label': '嵌入页面', 'icon_svg': svg,
         'url': f'http://127.0.0.1:{args.frontend_port}/api/v1/qa/embed',
         'visibility': 'user', 'sort_order': 1},
    ],
}
embed = ('<!doctype html><html lang="zh"><meta charset="utf-8">'
         '<style>body{font:16px system-ui;margin:0;padding:20px;background:#f0fdfa;color:#115e59}</style>'
         '<h2 style="margin:0 0 10px">嵌入内容正常显示</h2><p style="margin:0">本地文档与 iframe 内容可读。</p></html>')


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args):
        pass

    def record(self):
        with lock:
            requests.append({'method': self.command, 'path': self.path,
                             'authorization_present': bool(self.headers.get('Authorization'))})
            (out / 'requests.json').write_text(json.dumps(requests, ensure_ascii=False, indent=2) + '\n')

    def send(self, body, content_type='application/json', status=200):
        if isinstance(body, str):
            body = body.encode()
        self.send_response(status)
        self.send_header('Content-Type', content_type)
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def data(self, value):
        self.send(json.dumps({'code': 0, 'message': 'ok', 'data': value}, ensure_ascii=False))

    def do_GET(self):
        self.record()
        path = urlparse(self.path).path
        if path == '/api/v1/settings/public':
            self.data(settings)
        elif path in ['/api/v1/user', '/api/v1/auth/me']:
            self.data(user)
        elif path in ['/setup/status', '/api/v1/setup/status']:
            self.data({'needs_setup': False, 'installed': True})
        elif path == '/api/v1/subscriptions/active':
            self.data([])
        elif path == '/api/v1/announcements':
            self.data(announcements)
        elif path == '/api/v1/pages/qa-guide':
            self.send(custom_markdown, 'text/markdown; charset=utf-8')
        elif path == '/api/v1/qa/embed':
            self.send(embed, 'text/html; charset=utf-8')
        elif path == '/api/v1/qa/check.svg':
            self.send(svg.replace(' onclick="window.__qaXss=true"', ''), 'image/svg+xml')
        elif path == '/api/v1/qa/inputs':
            self.data({'markdown': markdown, 'custom_markdown': custom_markdown, 'svg': svg})
        elif path == '/api/v1/qa/baseline-purify.mjs' and args.baseline_module:
            self.send(args.baseline_module.read_bytes(), 'text/javascript')
        else:
            self.send(json.dumps({'code': 404, 'message': 'Unsupported fixture endpoint'}), status=404)

    def do_POST(self):
        self.record()
        size = int(self.headers.get('Content-Length', '0'))
        if size:
            self.rfile.read(size)
        path = urlparse(self.path).path
        match = next((item for item in announcements if path == f'/api/v1/announcements/{item["id"]}/read'), None)
        if match:
            match['read_at'] = created_at
            self.data({'message': 'ok'})
        else:
            self.send(json.dumps({'code': 404, 'message': 'Unsupported fixture endpoint'}), status=404)


print(f'Fixture API ready on 127.0.0.1:{args.port}', flush=True)
ThreadingHTTPServer(('127.0.0.1', args.port), Handler).serve_forever()
