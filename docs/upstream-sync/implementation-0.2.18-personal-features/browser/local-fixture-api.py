from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse
import datetime
import json
import threading

OUT = Path(__file__).resolve().parent
OUT.mkdir(parents=True, exist_ok=True)
requests = []
lock = threading.Lock()
now = datetime.datetime.now(datetime.timezone.utc).isoformat()
user = {'id': 1, 'email': 'qa@example.invalid', 'username': 'Local QA', 'role': 'admin', 'status': 'active', 'balance': 100, 'concurrency': 2, 'run_mode': 'standard'}
account = {'id': 101, 'name': 'Codex Local QA', 'platform': 'openai', 'type': 'apikey', 'status': 'error', 'schedulable': True, 'concurrency': 2, 'priority': 1,
           'credentials': {'api_key': 'local-fixture', 'temp_unschedulable_enabled': True, 'temp_unschedulable_rules': [{'error_code': 503, 'keywords': ['capacity'], 'duration_minutes': 5, 'description': '', 'scope': 'model'}]},
           'extra': {'web_search_emulation': 'enabled'}, 'group_ids': [1], 'groups': [], 'account_groups': [], 'created_at': now, 'updated_at': now}
usage = {'id': 1, 'user_id': 1, 'api_key_id': 1, 'account_id': 101, 'request_id': 'local-browser-fixture', 'model': 'gpt-5.2', 'requested_model': 'gpt-5.2',
         'group_id': 1, 'input_tokens': 40, 'output_tokens': 100, 'cache_creation_tokens': 0, 'cache_read_tokens': 0, 'cache_creation_5m_tokens': 0, 'cache_creation_1h_tokens': 0,
         'input_cost': 0, 'output_cost': 0, 'cache_creation_cost': 0, 'cache_read_cost': 0, 'total_cost': 0, 'actual_cost': 0, 'rate_multiplier': 1,
         'long_context_billing_applied': False, 'billing_type': 0, 'billing_mode': 'token', 'request_type': 'stream', 'stream': True, 'duration_ms': 2000, 'first_token_ms': 1500,
         'image_count': 0, 'image_output_tokens': 0, 'image_input_tokens': 0, 'created_at': now, 'user_agent': 'local-browser', 'user': user, 'api_key': {'id': 1, 'name': 'Local key'}, 'account': account}
stats = {'total_requests': 1, 'total_input_tokens': 40, 'total_output_tokens': 100, 'total_cache_tokens': 0, 'total_cache_creation_tokens': 0, 'total_cache_read_tokens': 0,
         'total_tokens': 140, 'total_cost': 0, 'total_actual_cost': 0, 'total_account_cost': 0, 'average_duration_ms': 2000}

def page(items):
    return {'items': items, 'total': len(items), 'page': 1, 'page_size': 20, 'pages': 1}

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        parsed = urlparse(self.path)
        path, query = parsed.path, parse_qs(parsed.query)
        with lock:
            requests.append({'method': 'GET', 'path': path, 'query': query})
            (OUT / 'requests-completion.json').write_text(json.dumps(requests, indent=2))
        if path == '/api/v1/settings/public':
            data = {'site_name': 'Local feature QA', 'registration_enabled': False, 'email_verify_enabled': False, 'ops_monitoring_enabled': True, 'ops_realtime_monitoring_enabled': False,
                    'payment_enabled': False, 'risk_control_enabled': False, 'hidden_menu_keys': [], 'custom_menu_items': [], 'purchase_subscription_enabled': False}
        elif path == '/api/v1/auth/me':
            data = user
        elif path in ['/setup/status', '/api/v1/setup/status']:
            data = {'needs_setup': False, 'installed': True}
        elif path in ['/api/v1/admin/usage', '/api/v1/usage']:
            data = page([usage])
        elif path.endswith('/usage/stats'):
            data = stats
        elif path == '/api/v1/admin/accounts':
            data = page([account] if not query.get('platform') or query['platform'][0] == 'openai' else [])
        elif path == '/api/v1/admin/accounts/101':
            data = account
        elif path == '/api/v1/admin/groups/all' or path == '/api/v1/groups/available':
            data = [{'id': 1, 'name': 'Local group', 'platform': 'openai', 'status': 'active'}]
        elif path == '/api/v1/admin/groups':
            data = page([{'id': 1, 'name': 'Local group', 'platform': 'openai', 'status': 'active'}])
        elif path == '/api/v1/admin/users':
            data = page([user])
        elif path.endswith('/web-search-emulation') or 'websearch' in path or 'web-search' in path:
            data = {'enabled': True, 'providers': [{'type': 'brave', 'api_key_configured': True}]}
        elif path.endswith('/keys') or path.endswith('/proxies') or path.endswith('/announcements'):
            data = page([])
        elif path.endswith('/proxies/all') or path.endswith('/tls-fingerprint-profiles'):
            data = []
        elif path.endswith('/admin/settings'):
            data = {'ops_monitoring_enabled': True, 'ops_realtime_monitoring_enabled': False, 'account_quota_notify_enabled': False}
        elif path.endswith('/realtime-traffic'):
            data = {'enabled': False, 'summary': None}
        elif path.endswith('/requests'):
            data = page([
                {'kind': 'success', 'account_id': 101, 'platform': 'openai', 'model': 'gpt-5.2', 'request_id': 'local-success', 'created_at': now, 'duration_ms': 2000, 'first_token_ms': 1500, 'status_code': 200},
                {'kind': 'error', 'account_id': 101, 'platform': 'openai', 'model': 'gpt-5.2', 'request_id': 'local-error', 'created_at': now, 'duration_ms': 3000, 'first_token_ms': None, 'status_code': 503, 'error_id': 12},
                {'kind': 'error', 'account_id': 999, 'platform': 'openai', 'model': 'gpt-5.2', 'request_id': 'local-deleted', 'created_at': now, 'duration_ms': 3000, 'status_code': 503},
                {'kind': 'success', 'account_id': None, 'platform': 'openai', 'model': 'gpt-5.2', 'request_id': 'local-unattributed', 'created_at': now, 'duration_ms': 1000, 'status_code': 200}
            ])
        elif path.endswith('/snapshot-v2'):
            data = {'generated_at': now, 'overview': {'health_score': 100, 'success_count': 1, 'request_count_total': 4, 'request_count_sla': 4, 'error_count_total': 3, 'business_limited_count': 0, 'error_count_sla': 3, 'token_consumed': 140, 'sla': 25, 'error_rate': 75, 'upstream_error_rate': 75, 'upstream_error_count_excl_429_529': 3, 'upstream_429_count': 0, 'upstream_529_count': 0, 'qps': {'current': 1, 'peak': 1, 'avg': 1}, 'tps': {'current': 70, 'peak': 70, 'avg': 70}, 'duration': {'p99_ms': 2000, 'avg_ms': 2000}, 'ttft': {'p99_ms': 1500, 'avg_ms': 1500}}, 'throughput_trend': {'points': [], 'by_platform': [], 'top_groups': []}, 'error_trend': {'points': []}}
        elif path.endswith('/error-trend') or path.endswith('/switch-trend') or path.endswith('/throughput-trend'):
            data = {'points': [], 'by_platform': [], 'top_groups': []}
        elif path.endswith('/concurrency'):
            data = {'accounts': [], 'groups': [], 'users': [], 'totals': {}}
        elif path.endswith('/metrics-thresholds'):
            data = {}

        else:
            data = {}
        body = json.dumps({'code': 0, 'message': 'ok', 'data': data}).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_POST(self):
        size = int(self.headers.get('Content-Length', '0'))
        if size:
            self.rfile.read(size)
        body = json.dumps({'code': 0, 'data': {}}).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

print('Local fixture API on 127.0.0.1:33887', flush=True)
ThreadingHTTPServer(('127.0.0.1', 33887), Handler).serve_forever()
