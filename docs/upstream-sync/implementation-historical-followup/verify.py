#!/usr/bin/env python3
"""Verify selected implementation, preserved inputs and actual acceptance records."""
from pathlib import Path
import gzip
import hashlib
import json
import re
import subprocess

out = Path(__file__).resolve().parent
root = out.parents[2]
initial_path = out / 'initial-state.json'
initial = json.loads(initial_path.read_bytes() if initial_path.exists() else gzip.decompress((out / 'initial-state.json.gz').read_bytes()))
def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()
implementation = {}
for line in (out / 'implementation-files.sha256').read_text().splitlines():
    expected, name = line.split('  ', 1)
    assert digest(root / name) == expected, name
    implementation[name] = expected
assert len(implementation) == 21
selected = ['port-upstream-historical-followup-tier1', 'port-upstream-historical-followup-tier2']
mutable_docs = {'AGENTS.md', 'docs/upstream-sync/README.md'} | {
    f'openspec/changes/{change}/{name}.md'
    for change in selected for name in ('README', 'design', 'tasks', 'verification')
}
protected = []
for name, expected in initial['hashes'].items():
    if name not in implementation and name not in mutable_docs:
        assert digest(root / name) == expected, name
        protected.append(name)
for name, backup in [('AGENTS.md', 'initial-AGENTS.md.gz'), ('docs/upstream-sync/README.md', 'initial-upstream-README.md.gz')]:
    original = gzip.decompress((out / backup).read_bytes())
    assert hashlib.sha256(original).hexdigest() == initial['hashes'][name]
    assert (root / name).read_bytes().endswith(original.split(b'\n\n', 1)[1]), name
old_changes = {name: tasks for name, tasks in initial['active_tasks'].items() if name.split('/')[-1] not in selected}
assert len(old_changes) == 17
assert sum(t['complete'] for t in old_changes.values()) == 467
for name, counts in old_changes.items():
    text = (root / name / 'tasks.md').read_text()
    assert text.count('- [x]') == counts['complete']
    assert text.count('- [ ]') == 0
for change, count in zip(selected, [13, 14]):
    text = (root / 'openspec/changes' / change / 'tasks.md').read_text()
    assert text.count('- [x]') == count
    assert '- [ ]' not in text
current = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'], cwd=root).decode().split('\0')
known = set(initial['tracked']) | set(initial['untracked'])
assert not [name for name in current if name.startswith(('backend/', 'frontend/')) and name not in known and name not in implementation]
assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root).decode().strip() == initial['head']
assert subprocess.check_output(['git', 'write-tree'], cwd=root).decode().strip() == initial['index']
assert subprocess.check_output(['git', 'show-ref'], cwd=root).decode().strip() == initial['refs']
assert (root / 'backend/cmd/server/VERSION').read_text().strip() == '1.1.28'
gates = ['baseline-replay-after', 'codex-route-paths-after', 'http-sqlite', 'backend-unit-final',
         'backend-admin-final', 'backend-race-final', 'backend-protection-final', 'backend-lint-complete',
         'backend-build-final', 'frontend-unit-final', 'frontend-typecheck-complete',
         'frontend-lint-complete', 'frontend-build-final', 'openspec-tier1-complete', 'openspec-tier2-complete']
for name in gates:
    data = json.loads((out / (name + '.json')).read_text())
    assert data['exit_code'] == 0, name
frontend = json.loads((out / 'frontend-unit-final-summary.json').read_text())
assert frontend['success'] and frontend['numFailedTests'] == 0 and frontend['numPassedTests'] == 2329
assert frontend['numPendingTests'] == 2
unit = json.loads((out / 'backend-unit-final.json').read_text())
assert unit['top_level'] == {'pass': 10829, 'skip': 16}
for name, errors in [('user-editors/browser/results.json', 'errors'), ('image-template/browser/result.json', 'pageErrors'), ('user-editors/browser/stepup/results.json', 'errors')]:
    data = json.loads((out / name).read_text())
    assert not data[errors], name
for record in sorted(out.glob('*.json')):
    data = json.loads(record.read_text())
    if not isinstance(data, dict) or 'stdout_log' not in data:
        continue
    for stream in ('stdout', 'stderr'):
        path = out / data[stream + '_log']
        raw = gzip.decompress(path.read_bytes()) if path.suffix == '.gz' else path.read_bytes()
        assert hashlib.sha256(raw).hexdigest() == data[stream + '_sha256'], path.name
link_docs = [out / 'README.md', root / 'AGENTS.md', root / 'docs/upstream-sync/README.md'] + [
    root / 'openspec/changes' / change / name for change in selected for name in ('README.md', 'verification.md')
]
for path in link_docs:
    for target in re.findall(r'\]\(([^)]+)\)', path.read_text()):
        target = target.split('#', 1)[0]
        if target and not re.match(r'[a-z]+:', target):
            resolved = (path.parent / target).resolve()
            assert resolved.exists() or resolved == out / 'validation.json', (str(path), target)
protected_source = json.loads((out / 'protected-source-check.json').read_text())
assert protected_source['previous_implementation_hashes'] == 18 and protected_source['matching']
assert protected_source['whole_PR_sources'] == 24 and protected_source['all_source_patch_hashes_match']
result = {
    'head_at_start': initial['head'], 'source_files': 9, 'new_regression_test_files': 12,
    'tracked_at_start': len(initial['tracked']), 'untracked_at_start': len(initial['untracked']),
    'protected_initial_files': len(protected), 'previous_active_changes': 17,
    'previous_completed_tasks': 467, 'selected_completed_tasks': [13, 14],
    'active_changes_after': 19, 'completed_tasks_after': 494, 'required_gates': gates,
    'version': '1.1.28', 'go': '1.27.4', 'sqlite_max': 230, 'auth_snapshot': 21,
    'frozen_history_sources_baselines_dependencies_schema_generated_resource_defaults_and_fast_diagnostic_preserved': True,
    'publication_archive_business_instance_operations': False,
}
(out / 'validation.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(result, ensure_ascii=False))
