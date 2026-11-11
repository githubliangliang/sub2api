"""Verify final product snapshot, gates, frozen sources and selected task scope."""
from pathlib import Path
import hashlib,json,re,subprocess
out=Path(__file__).resolve().parent
root=out.parents[2]
before=json.loads((out/'before-hashes.json').read_text());products=json.loads((out/'product-hashes.json').read_text());base=json.loads((out/'baseline.json').read_text())
allowed=set(products)|{'AGENTS.md','docs/upstream-sync/README.md'}|{'openspec/changes/port-upstream-0.2.20-tier2/'+n for n in ['README.md','tasks.md','verification.md']}
errors=[];changed=[]
for name,digest in before.items():
 file=root/name
 if not file.is_file() or hashlib.sha256(file.read_bytes()).hexdigest()!=digest:
  changed.append(name)
  if name not in allowed:errors.append({'unexpected_change':name})
for name,digest in products.items():
 if hashlib.sha256((root/name).read_bytes()).hexdigest()!=digest:errors.append({'stale_product_hash':name})
labels=['acceptance-unit','acceptance-build','acceptance-lint','backend-race','final-additions-race','cache-final-race','last-terminal-final','sqlite-protection','settlement-current','final-boundaries','prompt-http','patch-budget-check','quota-worker','server-cleanup-final','frontend-typecheck','frontend-lint-final','frontend-full-final','frontend-critical','frontend-build','browser-acceptance','memory-1g-stress','wire-generate-final','thinking-benchmark','delegation-benchmark','thinking-fuzz','export-cancel-check','openspec-tier2']
checks={}
for label in labels:
 file=out/(label+'.json')
 if not file.exists():errors.append({'missing_result':label});continue
 data=json.loads(file.read_text());checks[label]=data['exit_code']
 if data['exit_code']!=0:errors.append({'check_failed':label})
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
if head!=base['head']:errors.append({'head_changed':head})
change=root/'openspec/changes/port-upstream-0.2.20-tier2';tasks=(change/'tasks.md').read_text()
if tasks.count('[x]')!=38 or '[ ]' in tasks:errors.append({'completed_tasks':tasks.count('[x]')})
for file in [change/'README.md',change/'tasks.md',change/'verification.md',out/'README.md']:
 for target in re.findall(r'\]\(([^)]+)\)',file.read_text()):
  target=target.split('#')[0]
  if not target or re.match(r'\w+://',target):continue
  path=(file.parent/target).resolve()
  if path==out/'scope-check.json':continue
  if not path.exists():errors.append({'broken_link':str(file),'target':target})
fmt=subprocess.check_output(['gofmt','-l',*[str(root/p) for p in products if p.endswith('.go')]],text=True)
if fmt:errors.append({'gofmt':fmt})
result=subprocess.run(['git','diff','--check'],cwd=root,capture_output=True,text=True)
if result.returncode:errors.append({'diff_check':result.stdout})
status=subprocess.check_output(['git','status','--porcelain=v1','--untracked-files=all'],cwd=root,text=True)
for line in status.splitlines():
 name=line[3:]
 if name not in allowed and not name.startswith('docs/upstream-sync/implementation-0.2.20-tier2/'):
  errors.append({'unexpected_status':name})
unit=(out/'acceptance-unit.log').read_text()
report={'result':'PASS' if not errors else 'FAIL','errors':errors,'baseline_head':head,'implementation':'working tree; uncommitted','tasks_completed':38,'protected_existing_files':len(before),'changed_existing_files':changed,'product_files':len(products),'checks':checks,'unit_packages':len(re.findall(r'^ok\s+',unit,re.M)),'frontend':'2186 passed / 2 existing skipped; critical179 / 2','sources_and_history_unchanged':not any(p.startswith(('docs/upstream-sync/evidence-0.2.20/','openspec/changes/port-upstream-0.2.20-tier1/')) or p.endswith('PORTING-0.2.20.md') for p in changed),'version':'1.1.23','go':'1.27.3','sqlite_max':228,'auth_snapshot':21,'wire':'generated, shared digest cache cleanup','dependency_schema_ent_unchanged':not any(p.startswith(('backend/ent/','backend/migrations/')) or p in ['backend/go.mod','backend/go.sum','frontend/package.json','frontend/pnpm-lock.yaml'] for p in changed),'caveat':'1GiB stress covers body-budget layer; browser API is synthetic; no business deployment'}
(out/'scope-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
(out/'checks-summary.json').write_text(json.dumps({label:json.loads((out/(label+'.json')).read_text()) for label in labels if (out/(label+'.json')).exists()},ensure_ascii=False,indent=2)+'\n')
print(json.dumps(report,ensure_ascii=False,indent=2));raise SystemExit(bool(errors))
