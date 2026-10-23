import hashlib
import json
import pathlib
import shutil
import subprocess
import tempfile

root = pathlib.Path.cwd()
watched = [
    root / 'frontend/package.json',
    root / 'frontend/pnpm-lock.yaml',
    root / 'frontend/pnpm-workspace.yaml',
    root / 'frontend/node_modules/.modules.yaml',
]
digest = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
before = {str(path.relative_to(root)): digest(path) for path in watched}
work = pathlib.Path(tempfile.mkdtemp(prefix='sub2api-release-119-node20-'))
shutil.copytree(root / 'frontend', work / 'frontend', ignore=shutil.ignore_patterns('node_modules', 'dist', '*.tsbuildinfo'))
shutil.copytree(root / 'docs/legal', work / 'docs/legal')
copied_lock = digest(work / 'frontend/pnpm-lock.yaml')
script = '''set -eu
node --version
corepack enable
corepack prepare pnpm@9.15.9 --activate
pnpm --version
pnpm install --frozen-lockfile
pnpm run build
'''
command = [
    'rtk', 'proxy', 'docker', 'run', '--rm', '--pull=never',
    '--name', 'sub2api-release119-s07-node20',
    '-e', 'CI=true', '-v', str(work) + ':/work', '-w', '/work/frontend',
    'node:20-alpine', 'sh', '-lc', script,
]
print('isolated_work=' + str(work), flush=True)
print('shared_source_sha256=' + json.dumps(before, sort_keys=True), flush=True)
result = subprocess.run(command)
after = {str(path.relative_to(root)): digest(path) for path in watched}
assert before == after, 'shared frontend source or node_modules metadata changed'
assert copied_lock == digest(work / 'frontend/pnpm-lock.yaml'), 'frozen install rewrote isolated lockfile'
print('shared_frontend_unchanged=true', flush=True)
print('isolated_lockfile_unchanged=true', flush=True)
print('container_exit_code=' + str(result.returncode), flush=True)
if result.returncode == 0:
    output = work / 'backend/internal/web/dist'
    assert (output / 'index.html').is_file()
    print('built_frontend_index_sha256=' + digest(output / 'index.html'), flush=True)
    print('built_frontend_file_count=' + str(sum(1 for p in output.rglob('*') if p.is_file())), flush=True)
raise SystemExit(result.returncode)
