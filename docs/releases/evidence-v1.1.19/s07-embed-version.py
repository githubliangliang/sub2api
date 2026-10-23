import datetime
import hashlib
import json
import os
import pathlib
import subprocess

backend = pathlib.Path.cwd()
root = backend.parent
version = (backend / 'cmd/server/VERSION').read_text().strip()
assert version == '1.1.19', version
head = subprocess.check_output(['rtk', 'proxy', 'git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
resolved = subprocess.check_output(['rtk', 'proxy', 'sh', 'backend/scripts/resolve-version.sh'], cwd=root, text=True).strip()
assert resolved == version, (resolved, version)
print('HEAD=' + head, flush=True)
print('worktree_VERSION=' + version, flush=True)
print('resolved_VERSION=' + resolved, flush=True)
index = backend / 'internal/web/dist/index.html'
assert index.is_file(), 'real frontend dist is absent'
print('embedded_frontend_index_sha256=' + hashlib.sha256(index.read_bytes()).hexdigest(), flush=True)
print('embedded_frontend_file_count=' + str(sum(1 for p in index.parent.rglob('*') if p.is_file())), flush=True)
binary = pathlib.Path('/tmp/sub2api-release-1.1.19/sub2api-v1.1.19-version-check')
stamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
env = os.environ.copy()
env['CGO_ENABLED'] = '0'
env['GOOS'] = 'linux'
env['GOARCH'] = 'amd64'
flags = f'-s -w -X main.Version={version} -X main.Commit={head[:9]} -X main.Date={stamp} -X main.BuildType=release'
command = ['rtk', 'proxy', 'go', 'build', '-tags=embed', '-ldflags=' + flags, '-trimpath', '-o', str(binary), './cmd/server']
print('build_env=' + json.dumps({key: env[key] for key in ['CGO_ENABLED', 'GOOS', 'GOARCH']}), flush=True)
print('build_command=' + json.dumps(command), flush=True)
subprocess.run(command, env=env, check=True)
result = subprocess.run(['rtk', 'proxy', str(binary), '--version'], text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=True)
print(result.stdout, end='', flush=True)
assert 'Sub2API 1.1.19 ' in result.stdout
subprocess.run(['rtk', 'proxy', 'go', 'version', '-m', str(binary)], check=True)
print('binary_sha256=' + hashlib.sha256(binary.read_bytes()).hexdigest(), flush=True)
print('binary_size=' + str(binary.stat().st_size), flush=True)
print('binary_path=' + str(binary), flush=True)
