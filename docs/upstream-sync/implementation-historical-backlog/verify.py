#!/usr/bin/env python3
"""Check implementation scope, protected inputs and recorded acceptance gates."""
from pathlib import Path
import gzip
import hashlib
import json
import re
import subprocess

out = Path(__file__).resolve().parent
root = out.parents[2]
initial = json.loads(gzip.decompress((out / 'initial-state.json.gz').read_bytes()))
digest = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
implementation = {}
for line in (out / 'implementation-files.sha256').read_text().splitlines():
    value, name = line.split('  ', 1)
    assert digest(root / name) == value, name
    implementation[name] = value
assert len(implementation) == 18
known = set(initial['tracked']) | set(initial['untracked'])
current = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'], cwd=root).decode().split('\0')
assert not [name for name in current if name.startswith(('backend/', 'frontend/')) and name not in known and name not in implementation]
documentation = {'AGENTS.md', 'docs/upstream-sync/README.md'}
for backup, name in (('initial-AGENTS.md.gz', 'AGENTS.md'), ('initial-upstream-README.md.gz', 'docs/upstream-sync/README.md')):
    original = gzip.decompress((out / backup).read_bytes())
    assert hashlib.sha256(original).hexdigest() == initial['tracked'][name]
    assert (root / name).read_bytes().endswith(original.split(b'\n\n', 1)[1]), name
protected_tracked = 0
for name, value in initial['tracked'].items():
    if name not in implementation and name not in documentation:
        assert digest(root / name) == value, name
        protected_tracked += 1
changes = ['port-upstream-historical-backlog-tier1', 'port-upstream-historical-backlog-tier2']
selected_documents = {f'openspec/changes/{change}/{name}.md' for change in changes for name in ('README', 'design', 'tasks', 'verification')}
protected_untracked = 0
for name, value in initial['untracked'].items():
    if name not in selected_documents:
        assert digest(root / name) == value, name
        protected_untracked += 1
old_tasks = [name for name in initial['tracked'] if name.startswith('openspec/changes/') and '/archive/' not in name and name.endswith('/tasks.md')]
assert len(old_tasks) == 15
assert sum(len(re.findall(r'^- \[x\]', (root / name).read_text(), re.M)) for name in old_tasks) == 456
assert not any(re.search(r'^- \[ \]', (root / name).read_text(), re.M) for name in old_tasks)
selected_tasks = []
for change in changes:
    text = (root / 'openspec/changes' / change / 'tasks.md').read_text()
    selected_tasks.append(len(re.findall(r'^- \[x\]', text, re.M)))
    assert not re.search(r'^- \[ \]', text, re.M)
    for name in ('README.md', 'verification.md'):
        path = root / 'openspec/changes' / change / name
        for target in re.findall(r'\]\(([^)]+)\)', path.read_text()):
            target = target.split('#', 1)[0]
            if target and not re.match(r'[a-z]+:', target):
                resolved = (path.parent / target).resolve()
                assert resolved.exists() or resolved == out / 'validation.json', (str(path), target)
assert selected_tasks == [5, 6]
assert (root / 'backend/cmd/server/VERSION').read_text().strip() == '1.1.28'
assert subprocess.check_output(['git', 'rev-parse', 'v1.1.28^{}'], cwd=root).decode().strip() == initial['head']
required = ['backend-unit-final', 'backend-handler-unit-final', 'backend-race-complete', 'ops-worker-shutdown-fixed',
            'backend-lint-publish', 'backend-build-publish', 'frontend-typecheck', 'frontend-lint', 'frontend-build',
            'frontend-combination', 'openspec-tier1-complete', 'openspec-tier2-complete',
            'query-cost-baseline-serial', 'query-cost-current-serial', 'baseline-replay-fixed']
for name in required:
    data = json.loads((out / (name + '.json')).read_text())
    assert data['exit_code'] == 0, name
for record in sorted(out.glob('*.json')):
    data = json.loads(record.read_text())
    if not isinstance(data, dict) or 'command' not in data:
        continue
    for stream in ('stdout', 'stderr'):
        if stream + '_log' in data:
            path = out / data[stream + '_log']
            raw = gzip.decompress(path.read_bytes()) if path.suffix == '.gz' else path.read_bytes()
            assert hashlib.sha256(raw).hexdigest() == data[stream + '_sha256'], path.name
report = {'head_at_start': initial['head'], 'implementation_and_test_files': len(implementation),
          'protected_tracked_files': protected_tracked, 'initial_untracked_files': len(initial['untracked']),
          'protected_initial_untracked_files': protected_untracked, 'previous_changes': 15, 'previous_completed_tasks': 456,
          'selected_completed_tasks': selected_tasks, 'active_completed_tasks_total': 467, 'required_gates': required,
          'version': '1.1.28', 'source_schema_dependencies_generated_and_initial_diagnostics_preserved': True}
(out / 'validation.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(report, ensure_ascii=False))
