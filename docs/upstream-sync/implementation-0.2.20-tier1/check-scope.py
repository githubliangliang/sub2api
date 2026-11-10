"""Verify the completed implementation and preservation of frozen intake/history."""
from pathlib import Path
import hashlib,json,re,subprocess
out=Path(__file__).resolve().parent
root=out.parents[2]
start=json.loads((out/'baseline.json').read_text())
before=json.loads((out/'before-hashes.json').read_text())
products=json.loads((out/'product-hashes.json').read_text())
allowed=set(products)|{'AGENTS.md','docs/upstream-sync/README.md'}|{'openspec/changes/port-upstream-0.2.20-tier1/'+n for n in ['README.md','tasks.md','verification.md']}
errors=[]
changed=[]
for name,old in before.items():
 path=root/name
 if not path.is_file() or hashlib.sha256(path.read_bytes()).hexdigest()!=old:
  changed.append(name)
  if name not in allowed:errors.append({'unexpected_change':name})
for name,digest in products.items():
 if hashlib.sha256((root/name).read_bytes()).hexdigest()!=digest:errors.append({'product_hash':name})
checks={}
for label in ['after-targeted-final','additional-boundaries','backend-unit','backend-build','backend-lint-final','backend-race','sqlite-protection','openspec-tier1']:
 data=json.loads((out/(label+'.json')).read_text());checks[label]=data['exit_code']
 if data['exit_code']:errors.append({'failed_check':label})
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
if head!=start['head']:errors.append({'head_changed':head})
change=root/'openspec/changes/port-upstream-0.2.20-tier1'
tasks=(change/'tasks.md').read_text()
if tasks.count('[x]')!=44 or '[ ]' in tasks:errors.append({'tasks':tasks.count('[x]')})
for name in ['source-baseline.md','source-feature-map.md','proposal.md','design.md']:
 path='openspec/changes/port-upstream-0.2.20-tier1/'+name
 if hashlib.sha256((root/path).read_bytes()).hexdigest()!=before[path]:errors.append({'frozen_source':path})
markdown=[out/'README.md',change/'README.md',change/'verification.md',change/'tasks.md']
for path in markdown:
 for target in re.findall(r'\]\(([^)]+)\)',path.read_text()):
  target=target.split('#')[0]
  if not target or re.match(r'\w+://',target):continue
  resolved=(path.parent/target).resolve()
  if resolved==out/'scope-check.json':continue
  if not resolved.exists():errors.append({'broken_link':str(path),'target':target})
formatting=subprocess.check_output(['gofmt','-l',*[str(root/p) for p in products if p.endswith('.go')]],text=True)
if formatting:errors.append({'gofmt':formatting})
diffcheck=subprocess.run(['git','diff','--check'],cwd=root,capture_output=True,text=True)
if diffcheck.returncode:errors.append({'diff_check':diffcheck.stdout})
unit=(out/'backend-unit.log').read_text();race=(out/'backend-race.log').read_text()
report={'result':'PASS' if not errors else 'FAIL','errors':errors,'baseline_head':head,'implementation':'working tree; uncommitted','checked_tasks':44,'frozen_files_checked':len(before),'unchanged_existing_files':len(before)-len(changed),'changed_existing_files':changed,'product_files':len(products),'checks':checks,'unit_packages':len(re.findall(r'^ok\s+',unit,re.M)),'race_packages':len(re.findall(r'^ok\s+',race,re.M)),'tier2_and_intake_unchanged':not any('evidence-0.2.20/' in p or 'tier2/' in p or p.endswith('PORTING-0.2.20.md') for p in changed),'version':'1.1.23','go':'1.27.3','sqlite_max':228,'frontend_dependencies_codegen_unchanged':not any(p.startswith(('frontend/','backend/ent/','backend/migrations/')) or p in ['backend/go.mod','backend/go.sum','backend/cmd/server/wire_gen.go'] for p in changed)}
(out/'scope-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(report,ensure_ascii=False,indent=2))
raise SystemExit(bool(errors))
