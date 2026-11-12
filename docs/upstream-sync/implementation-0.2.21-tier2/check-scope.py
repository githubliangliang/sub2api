"""Record this batch's exact worktree and verify frozen inputs without git writes."""
import difflib
import hashlib
import json
import re
import subprocess
from pathlib import Path

out = Path(__file__).resolve().parent
root = out.parents[2]
change = root / 'openspec/changes/port-upstream-0.2.21-tier2'
start = json.loads((out / 'start.json').read_text())
base = '63c764704761073e69ff25ba41d64eeab5793cd0'
failures = []

def git(*args):
    return subprocess.check_output(['git', *args], cwd=root)

def digest(data):
    return hashlib.sha256(data).hexdigest()

head = git('rev-parse', 'HEAD').decode().strip()
if head != base:
    failures.append('HEAD changed after the observed independent tier1 commit')

changed = git('diff', '--name-only', base).decode().splitlines()
untracked = git('ls-files', '--others', '--exclude-standard').decode().splitlines()
product = sorted(p for p in set(changed + untracked) if p.startswith(('backend/', 'frontend/')))
allowed_other = {
    'AGENTS.md', 'docs/upstream-sync/README.md',
    'openspec/changes/port-upstream-0.2.21-tier2/README.md',
    'openspec/changes/port-upstream-0.2.21-tier2/tasks.md',
    'openspec/changes/port-upstream-0.2.21-tier2/verification.md',
}
for name in changed + untracked:
    if name in product or name in allowed_other or name.startswith('docs/upstream-sync/implementation-0.2.21-tier2/'):
        continue
    failures.append('outside selected batch: ' + name)

protected_changes = []
for name, expected in start['protected_hashes'].items():
    path = root / name
    actual = digest(path.read_bytes()) if path.is_file() else None
    if actual == expected:
        continue
    protected_changes.append(name)
    # The tier1 implementation and its own final records were completed by an
    # independent session and are now frozen in base. Our three shared files
    # and the current navigation/status entries are the selected exceptions.
    independent = name.startswith(('openspec/changes/port-upstream-0.2.21-tier1/', 'docs/upstream-sync/implementation-0.2.21-tier1/')) or name in {
        'backend/internal/service/openai_responses_sealed_call_test.go',
        'backend/internal/service/tier121_price_test.go',
    }
    if independent:
        if actual != digest(git('show', base + ':' + name)):
            failures.append('independent tier1 completion was modified: ' + name)
    elif name not in product and name not in allowed_other:
        failures.append('frozen input changed: ' + name)

for name in ('source-baseline.md', 'source-feature-map.md', 'proposal.md', 'design.md', 'specs/upstream-0-2-21-tier2-compatibility/spec.md'):
    rel = str((change / name).relative_to(root))
    if digest((change / name).read_bytes()) != start['protected_hashes'][rel]:
        failures.append('frozen specification changed: ' + rel)

for name in product:
    if name.startswith(('backend/ent/', 'backend/cmd/server/', 'backend/resources/')) or name in {
        'backend/go.mod', 'backend/go.sum', 'frontend/package.json', 'frontend/pnpm-lock.yaml', 'frontend/pnpm-workspace.yaml',
    }:
        failures.append('version/dependency/generated/nonselected product: ' + name)
    if name.startswith('backend/migrations/') and name != 'backend/migrations/229_add_usage_logs_api_key_cost_index.sql':
        failures.append('old or unexpected migration changed: ' + name)

tasks = (change / 'tasks.md').read_text()
if tasks.count('- [x]') != 42 or '- [ ]' in tasks:
    failures.append('expected all 42 tasks to be completed')

labels = [
    'backend-unit-exact', 'backend-build-owned-final', 'backend-lint-last',
    'backend-race-complete', 'backend-race-last-guard', 'backend-race-exact', 'sqlite-http-protection',
    'frontend-typecheck-final', 'frontend-lint-final', 'frontend-full-final',
    'frontend-build-final-owned', 'frontend-critical-acceptance', 'browser-owned-final',
    'index-1c1g', 'audit-sqlite-acceptance', 'page-http', 'batch-account-reuse', 'openspec-tier2-final',
]
for label in labels:
    path = out / (label + '.json')
    if not path.exists() or json.loads(path.read_text()).get('exit_code') != 0:
        failures.append('missing/failed required evidence: ' + label)

patch = bytearray(git('diff', '--binary', base, '--', *product))
for name in product:
    if name not in untracked:
        continue
    result = subprocess.run(['git', 'diff', '--no-index', '--binary', '--', '/dev/null', name], cwd=root, capture_output=True)
    if result.returncode not in (0, 1):
        failures.append('cannot record new file: ' + name)
    patch.extend(result.stdout)
(out / 'product.diff').write_bytes(patch)
hashes = {name: digest((root / name).read_bytes()) for name in product}
(out / 'product-hashes.json').write_text(json.dumps(hashes, indent=2) + '\n')

link_files = [change / 'README.md', change / 'verification.md', out / 'README.md']
for path in link_files:
    for target in re.findall(r'\]\(([^)]+)\)', path.read_text()):
        if target.startswith(('https:', 'http:', '#')):
            continue
        destination = (path.parent / target.split('#', 1)[0]).resolve()
        if destination == (out / 'scope-check.json').resolve():
            continue
        if not destination.exists():
            failures.append(f'broken link in {path.name}: {target}')

record = {
    'starting_head': start['head'], 'final_diff_base': base, 'head': head,
    'product_files': len(product), 'product_manifest_sha256': digest((out / 'product-hashes.json').read_bytes()),
    'patch_sha256': digest(patch), 'protected_files_at_start': len(start['protected_hashes']),
    'preexisting_changes_observed': protected_changes, 'tasks_completed': tasks.count('- [x]'),
    'required_evidence': labels, 'failures': failures,
}
(out / 'scope-check.json').write_text(json.dumps(record, indent=2) + '\n')
print(json.dumps(record, indent=2))
raise SystemExit(bool(failures))
